// Package scraper provides support for scraping OpenAPI versions from
// services.
package scraper

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
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
	NotifyVersions(name string, versions []string, scrapeTime time.Time) error

	// CollateVersions tells the storage to execute the compilation and
	// update all VU-formatted specs from all services and their
	// respective versions gathered.
	CollateVersions() error

	// HasVersion returns whether the storage has already stored the service
	// API spec version at the given content digest.
	HasVersion(name string, version string, digest string) (bool, error)

	// NotifyVersion tells the storage to store the given version contents at
	// the scrapeTime. The storage implementation must detect and ignore
	// duplicate version contents, as some services may not provide content
	// digest headers in their responses.
	NotifyVersion(name string, version string, contents []byte, scrapeTime time.Time) error

	// Versions fetches the Storage Versions compiled by VU
	Versions() []string

	// Version fetches the Storage Version spec compiled by VU
	Version(version string) ([]byte, error)
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
		http:    &http.Client{Timeout: time.Second * 15},
		timeNow: time.Now,
	}
	err := setupScraper(s, cfg, options)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func setupScraper(s *Scraper, cfg *config.ServerConfig, options []Option) error {
	s.services = make([]service, len(cfg.Services))
	for i := range cfg.Services {
		u, err := url.Parse(cfg.Services[i] + "/openapi")
		if err != nil {
			return errors.Wrapf(err, "invalid service %q", cfg.Services[i])
		}
		s.services[i] = service{base: cfg.Services[i], url: u}
	}
	for i := range options {
		err := options[i](s)
		if err != nil {
			return err
		}
	}
	return nil
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
	for range s.services {
		err := <-errCh
		errs = multierr.Append(errs, err)
	}
	close(errCh)
	if errs != nil {
		return errs
	} else {
		err := s.collateVersions()
		errs = multierr.Append(errs, err)
	}
	return errs
}

func (s *Scraper) scrape(ctx context.Context, scrapeTime time.Time, svc service) error {
	versions, err := s.getVersions(svc)
	if err != nil {
		return errors.WithStack(err)
	}
	err = s.storage.NotifyVersions(svc.base, versions, scrapeTime)
	if err != nil {
		return errors.WithStack(err)
	}
	for i := range versions {
		// TODO: we might run this concurrently per live service pod if/when
		// we're more k8s aware, but we won't do that yet.
		contents, isNew, err := s.getNewVersion(svc, versions[i])
		if err != nil {
			return errors.WithStack(err)
		}
		if !isNew {
			continue
		}
		err = s.storage.NotifyVersion(svc.base, versions[i], contents, scrapeTime)
		if err != nil {
			return errors.WithStack(err)
		}
	}
	return nil
}

func (s *Scraper) collateVersions() error {
	return s.storage.CollateVersions()
}

func (s *Scraper) getVersions(svc service) ([]string, error) {
	req, err := http.NewRequest("GET", svc.url.String(), nil)
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

func (s *Scraper) getNewVersion(svc service, version string) ([]byte, bool, error) {
	// TODO: Services don't emit HEAD currently with compiled vervet
	//       will need to enforce down the line
	isNew, err := s.hasNewVersion(svc, version)
	if err != nil {
		return nil, false, errors.WithStack(err)
	}
	if !isNew {
		return nil, false, nil
	}

	req, err := http.NewRequest("GET", svc.url.String()+"/"+version, nil)
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
	if ct := resp.Header.Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
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

func (s *Scraper) hasNewVersion(svc service, version string) (bool, error) {
	// Check Digest to see if there's a new version
	req, err := http.NewRequest("HEAD", svc.url.String()+"/"+version, nil)
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

	// Can be formed similarly: "application/json; charset: utf-8"
	if ct := resp.Header.Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
		return false, errors.Errorf("unexpected content type: %s", ct)
	}
	digest := storage.DigestHeader(resp.Header.Get("Digest"))
	if digest == "" {
		// Not providing a digest is fine, we'll just come back with a GET
		return true, nil
	}
	return s.storage.HasVersion(svc.base, version, digest)
}

func (s *Scraper) Versions() []string {
	return s.storage.Versions()
}

func (s *Scraper) Version(version string) ([]byte, error) {
	return s.storage.Version(version)
}
