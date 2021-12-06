// Package scraper provides support for scraping OpenAPI versions from
// services.
package scraper

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/multierr"

	"vervet-underground/config"
	"vervet-underground/internal/storage"
)

// Storage defines the storage functionality needed in order to store service
// API version spec snapshots.
type Storage interface {
	// NotifyVersions tells the storage which versions are currently available.
	// This is the primary mechanism by which the storage layer discovers and
	// processes versions which are removed post-sunset.
	NotifyVersions(ctx context.Context, name string, versions []string, scrapeTime time.Time) error

	// HasVersion returns whether the storage has already stored the service
	// API spec version at the given content digest.
	HasVersion(ctx context.Context, name string, version string, digest string) (bool, error)

	// NotifyVersion tells the storage to store the given version contents at
	// the scrapeTime. The storage implementation must detect and ignore
	// duplicate version contents, as some services may not provide content
	// digest headers in their responses.
	NotifyVersion(ctx context.Context, name string, version string, contents []byte, scrapeTime time.Time) error
}

// Scraper gets OpenAPI specs from a collection of services and updates storage
// accordingly.
type Scraper struct {
	storage  Storage
	services []service
	http     *http.Client
	timeNow  func() time.Time
}

type service struct {
	base string
	url  *url.URL
}

// Option defines an option that may be specified when creating a new Scraper.
type Option func(*Scraper) error

// New returns a new Scraper instance.
func New(cfg *config.ServerConfig, storage Storage, options ...Option) (*Scraper, error) {
	s := &Scraper{
		storage: storage,
		http:    &http.Client{},
		timeNow: time.Now,
	}
	s.services = make([]service, len(cfg.Services))
	for i := range cfg.Services {
		u, err := url.Parse(cfg.Services[i] + "/openapi")
		if err != nil {
			return nil, errors.Wrapf(err, "invalid service %q", cfg.Services[i])
		}
		s.services[i] = service{base: cfg.Services[i], url: u}
	}
	for i := range options {
		err := options[i](s)
		if err != nil {
			return nil, err
		}
	}
	return s, nil
}

// HTTPClient is a Scraper constructor Option that allows providing an
// *http.Client instance. This may be used to configure the transport and
// timeouts on the HTTP client.
func HTTPClient(cl *http.Client) Option {
	return func(s *Scraper) error {
		s.http = cl
		return nil
	}
}

// Clock is a Scraper constructor Option that allows providing an alternative
// clock used to determine the scrape timestamps used to record changes in
// service spec versions.
func Clock(c func() time.Time) Option {
	return func(s *Scraper) error {
		s.timeNow = c
		return nil
	}
}

// Run executes the OpenAPI version scraping on all configured services.
func (s *Scraper) Run(ctx context.Context) error {
	scrapeTime := s.timeNow().UTC()
	errCh := make(chan error, len(s.services))
	for i := range s.services {
		svc := s.services[i]
		go func() {
			errCh <- s.scrape(ctx, scrapeTime, svc)
		}()
	}
	var errs error
	for _ = range s.services {
		err := <-errCh
		errs = multierr.Append(errs, err)
	}
	close(errCh)
	return errs
}

func (s *Scraper) scrape(ctx context.Context, scrapeTime time.Time, svc service) error {
	versions, err := s.getVersions(ctx, svc)
	if err != nil {
		return errors.WithStack(err)
	}
	err = s.storage.NotifyVersions(ctx, svc.base, versions, scrapeTime)
	if err != nil {
		return errors.WithStack(err)
	}
	for i := range versions {
		// TODO: we might run this concurrently per live service pod if/when
		// we're more k8s aware, but we won't do that yet.
		contents, isNew, err := s.getNewVersion(ctx, svc, versions[i])
		if err != nil {
			return errors.WithStack(err)
		}
		if !isNew {
			continue
		}
		err = s.storage.NotifyVersion(ctx, svc.base, versions[i], contents, scrapeTime)
		if err != nil {
			return errors.WithStack(err)
		}
	}
	return nil
}

func (s *Scraper) getVersions(ctx context.Context, svc service) ([]string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", svc.url.String(), nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create request")
	}
	resp, err := s.http.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "request failed")
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, httpError(resp)
	}
	var versions []string
	err = json.NewDecoder(resp.Body).Decode(&versions)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return versions, nil
}

func httpError(r *http.Response) error {
	if contents, err := ioutil.ReadAll(r.Body); err == nil {
		return errors.Errorf("request failed: HTTP %d: %s", r.StatusCode, string(contents))
	}
	return errors.Errorf("request failed: HTTP %d", r.StatusCode)
}

func (s *Scraper) getNewVersion(ctx context.Context, svc service, version string) ([]byte, bool, error) {
	isNew, err := s.hasNewVersion(ctx, svc, version)
	if err != nil {
		return nil, false, errors.WithStack(err)
	}
	if !isNew {
		return nil, false, nil
	}

	req, err := http.NewRequestWithContext(ctx, "GET", svc.url.String()+"/"+version, nil)
	if err != nil {
		return nil, false, errors.Wrap(err, "failed to create request")
	}
	resp, err := s.http.Do(req)
	if err != nil {
		return nil, false, errors.Wrap(err, "request failed")
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, false, httpError(resp)
	}
	if ct := resp.Header.Get("Content-Type"); ct != "application/json" {
		return nil, false, errors.Errorf("unexpected content type: %s", ct)
	}
	respContents, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, false, errors.WithStack(err)
	}
	// For now, let's just see if the response can be unmarshaled
	// TODO: Load w/kin-openapi and validate it?
	var doc map[string]interface{}
	err = json.Unmarshal(respContents, &doc)
	if err != nil {
		return nil, false, errors.WithStack(err)
	}
	return respContents, true, nil
}

func (s *Scraper) hasNewVersion(ctx context.Context, svc service, version string) (bool, error) {
	// Check Digest to see if there's a new version
	req, err := http.NewRequestWithContext(ctx, "HEAD", svc.url.String()+"/"+version, nil)
	if err != nil {
		return false, errors.Wrap(err, "failed to create request")
	}
	resp, err := s.http.Do(req)
	if err != nil {
		return false, errors.Wrap(err, "request failed")
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusMethodNotAllowed {
		// Not supporting HEAD is fine, we'll just come back with a GET
		return true, nil
	}
	if resp.StatusCode != http.StatusOK {
		return false, httpError(resp)
	}
	if ct := resp.Header.Get("Content-Type"); ct != "application/json" {
		return false, errors.Errorf("unexpected content type: %s", ct)
	}
	digest := storage.DigestHeader(resp.Header.Get("Digest"))
	if digest == "" {
		// Not providing a digest is fine, we'll just come back with a GET
		return true, nil
	}
	return s.storage.HasVersion(ctx, svc.base, version, digest)
}
