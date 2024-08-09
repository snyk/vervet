// Package scraper provides support for scraping OpenAPI versions from
// services.
package scraper

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog/log"
	"go.uber.org/multierr"

	"github.com/snyk/vervet/v7/config"
	"github.com/snyk/vervet/v7/internal/storage"
)

// Scraper gets OpenAPI specs from a collection of services and updates storage
// accordingly.
type Scraper struct {
	storage       storage.Storage
	services      []service
	http          *http.Client
	timeNow       func() time.Time
	serviceFilter map[string]bool
}

type service struct {
	base string
	url  *url.URL
	name string
}

// Option defines an option that may be specified when creating a new Scraper.
type Option func(*Scraper) error

// New returns a new Scraper instance.
func New(cfg *config.ServerConfig, store storage.Storage, options ...Option) (*Scraper, error) {
	client := &http.Client{
		Timeout:   time.Second * 15,
		Transport: DurationTransport(http.DefaultTransport),
	}

	s := &Scraper{
		storage: store,
		http:    client,
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
	s.serviceFilter = make(map[string]bool)
	for i := range cfg.Services {
		s.serviceFilter[cfg.Services[i].Name] = true
		u, err := url.Parse(cfg.Services[i].URL + "/openapi")
		if err != nil {
			return errors.Wrapf(err, "invalid service %q", cfg.Services[i].Name)
		}
		// Handle for local/smaller deployments and tests
		s.services[i] = service{base: cfg.Services[i].URL, url: u, name: cfg.Services[i].Name}
		if u.Hostname() == "localhost" || u.Hostname() == "127.0.0.1" {
			s.services[i] = service{base: u.Host, url: u, name: cfg.Services[i].Name}
		}
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
	var errs error
	scrapeTime := s.timeNow().UTC()
	defer func() {
		metrics.runDuration.Observe(time.Since(scrapeTime).Seconds())
		if errs != nil {
			metrics.runError.Inc()
		}
	}()

	errCh := make(chan error, len(s.services))
	for i := range s.services {
		svc := s.services[i]
		go func() {
			timer := prometheus.NewTimer(metrics.scrapeDuration.WithLabelValues(svc.base))
			defer timer.ObserveDuration()

			log.Debug().Str("service", svc.name).Msg("started scrape")
			err := s.scrape(ctx, scrapeTime, svc)
			if err != nil {
				metrics.scrapeError.WithLabelValues(svc.base).Inc()
				log.Error().Str("service", svc.name).Err(err).Msg("error scraping service")
			}
			log.Debug().Str("service", svc.name).Msg("finished scrape")
			errCh <- err
		}()
	}

	for range s.services {
		err := <-errCh
		errs = multierr.Append(errs, err)
	}
	close(errCh)
	if errs != nil {
		return errs
	} else {
		err := s.collateVersions(ctx)
		errs = multierr.Append(errs, err)
	}
	return errs
}

func (s *Scraper) scrape(ctx context.Context, scrapeTime time.Time, svc service) error {
	versions, err := s.getVersions(ctx, svc)
	if err != nil {
		return errors.WithStack(err)
	}

	for i := range versions {
		// Skip if the version is not publicly documented
		if !IsPubliclyDocumented(versions[i]) {
			continue
		}

		// Skip if it's a legacy api using the default legacy version.
		if isLegacyVersion(versions[i]) {
			continue
		}

		contents, isNew, err := s.getNewVersion(ctx, svc, versions[i])
		if err != nil {
			return errors.WithStack(err)
		}
		if !isNew {
			continue
		}

		err = s.storage.NotifyVersion(ctx, svc.name, versions[i], contents, scrapeTime)
		if err != nil {
			return errors.WithStack(err)
		}
	}
	return nil
}

func (s *Scraper) collateVersions(ctx context.Context) error {
	return s.storage.CollateVersions(ctx, s.serviceFilter)
}

func (s *Scraper) getVersions(ctx context.Context, svc service) ([]string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", svc.url.String(), http.NoBody)
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
	if contents, err := io.ReadAll(r.Body); err == nil {
		return errors.Errorf("request failed: HTTP %d: %s", r.StatusCode, string(contents))
	}
	return errors.Errorf("request failed: HTTP %d", r.StatusCode)
}

func (s *Scraper) getNewVersion(ctx context.Context,
	svc service,
	version string) (respContents []byte,
	isNew bool,
	err error,
) {
	// TODO: Services don't emit HEAD currently with compiled vervet
	//       will need to enforce down the line
	isNew, err = s.hasNewVersion(ctx, svc, version)
	if err != nil {
		return nil, false, errors.WithStack(err)
	}
	if !isNew {
		return nil, isNew, nil
	}

	req, err := http.NewRequestWithContext(ctx, "GET", svc.url.String()+"/"+version, http.NoBody)
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
	respContents, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, false, errors.WithStack(err)
	}
	// For now, let's just see if the response can be unmarshaled
	// TODO: Load w/kin-openapi and remove excluded elements
	var doc map[string]interface{}
	err = json.Unmarshal(respContents, &doc)
	if err != nil {
		return nil, false, errors.WithStack(err)
	}
	return respContents, true, nil
}

func (s *Scraper) hasNewVersion(ctx context.Context, svc service, version string) (bool, error) {
	// Check Digest to see if there's a new version
	req, err := http.NewRequestWithContext(ctx, "HEAD", svc.url.String()+"/"+version, http.NoBody)
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
	return s.storage.HasVersion(ctx, svc.name, version, digest)
}

// isLegacyVersion is used to identify legacy APIs which should be excluded.
func isLegacyVersion(version string) bool {
	// This default version predates vervet's creation date.
	return version == "2021-01-01"
}

func IsPubliclyDocumented(version string) bool {
	const publiclyDocumentedDate = "2024-09-08"
	if !IsExperimentalVersion(version) {
		return true
	}
	parts := strings.Split(version, "~")
	if len(parts) > 0 {
		versionDate := parts[0]
		if versionDate <= publiclyDocumentedDate {
			return true
		}
	}
	return false
}

func IsExperimentalVersion(version string) bool {
	return strings.HasSuffix(version, "~experimental")
}
