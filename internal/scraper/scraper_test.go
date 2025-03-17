package scraper_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	qt "github.com/frankban/quicktest"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"

	"github.com/snyk/vervet/v8/config"
	"github.com/snyk/vervet/v8/internal/scraper"
	"github.com/snyk/vervet/v8/internal/storage/disk"
)

var (
	t0            = time.Date(2021, time.December, 3, 20, 49, 51, 0, time.UTC)
	collatedPaths = map[string]int{
		"2021-09-01~experimental": 1, // publicly documented version
		"2021-09-16":              2,
		"2021-10-01":              3,
		"2021-10-16":              4,
		"2024-09-09~experimental": 5, // publicly undocumented version
	}

	petfood = &testService{
		versions: []string{"2021-09-01~experimental", "2021-09-16", "2024-09-09~experimental"},
		contents: map[string]string{
			"2021-09-01~experimental": `{"paths":{"/crickets": {"get": {}}}}`,
			"2021-09-16":              `{"paths":{"/crickets": {"get": {}}, "/kibble": {"get": {}}}}`,
			"2024-09-09~experimental": `{"paths":{"/newexperiment": {"get": {}}}}`,
		},
	}
	animals = &testService{
		versions: []string{"2021-01-01", "2021-10-01", "2021-10-16"},
		contents: map[string]string{
			"2021-01-01": `{"paths":{"/legacy": {"get": {}}}}`,
			"2021-10-01": `{"paths":{"/geckos": {"get": {}}}}`,
			"2021-10-16": `{"paths":{"/geckos": {"get": {}}, "/puppies": {"get": {}}}}`,
		},
	}
)

type testService struct {
	versions []string
	contents map[string]string
}

func setupHttpServers(c *qt.C) (*httptest.Server, *httptest.Server) {
	petfoodService := httptest.NewServer(petfood.Handler())
	c.Cleanup(petfoodService.Close)

	animalsService := httptest.NewServer(animals.Handler())
	c.Cleanup(animalsService.Close)
	return petfoodService, animalsService
}

func (t *testService) Handler() http.Handler {
	r := mux.NewRouter()
	r.HandleFunc("/openapi", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode(&t.versions)
		if err != nil {
			log.Fatal().Err(err).Msg("test openapi handler failed to reply")
		}
	})
	r.HandleFunc("/openapi/{version}", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, err := w.Write([]byte(t.contents[mux.Vars(r)["version"]]))
		if err != nil {
			log.Fatal().Err(err).Msg("test openapi/version handler failed to reply")
		}
	})
	return r
}

func TestScraper(t *testing.T) {
	c := qt.New(t)

	petfoodService, animalsService := setupHttpServers(c)
	tests := []struct {
		name, version, digest string
	}{
		{"petfood", "2021-09-01~experimental", "sha256:zCgJaPeR8R21wsAlYn46xO6NE3XJiyFtLnYrP4DpM3U="},
		{"petfood", "2024-09-09~experimental", "sha256:zCgJaPeR8R21wsAlYn46xO6NE3XJiyFtLnYrP4DpM3U="},
		{"animals", "2021-10-16", "sha256:hcv2i7awT6CcSCecw9WrYBokFyzYNVaQArGgqHqdj7s="},
	}

	cfg := &config.ServerConfig{
		Services: []config.ServiceConfig{{
			Name: "petfood", URL: petfoodService.URL,
		}, {
			Name: "animals", URL: animalsService.URL,
		}},
	}
	st := disk.New("/tmp/specs")
	c.Cleanup(func() {
		ds, ok := st.(*disk.Storage)
		if !ok {
			return
		}
		err := ds.Cleanup()
		c.Assert(err, qt.IsNil)
	})

	sc, err := scraper.New(cfg, st, scraper.Clock(func() time.Time { return t0 }))
	c.Assert(err, qt.IsNil)

	// Cancel the scrape context after a timeout so we don't hang the test
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
	c.Cleanup(cancel)

	// No version digests should be known
	for _, test := range tests {
		ok, err := st.HasVersion(ctx, test.name, test.version, test.digest)
		c.Assert(err, qt.IsNil)
		c.Assert(ok, qt.IsFalse)
	}

	// Run the scrape
	err = sc.Run(ctx)
	c.Assert(err, qt.IsNil)

	// Version digests now known to storage
	for _, test := range tests {
		if !scraper.IsPubliclyDocumented(test.version) {
			ok, err := st.HasVersion(ctx, test.name, test.version, test.digest)
			c.Assert(err, qt.IsNil)
			c.Assert(ok, qt.IsFalse, qt.Commentf("publicly undocumented version %s should not be included", test.version))
		} else {
			ok, err := st.HasVersion(ctx, test.name, test.version, test.digest)
			c.Assert(err, qt.IsNil)
			c.Assert(ok, qt.IsTrue)
		}
	}

	vi, err := st.VersionIndex(ctx)
	c.Assert(err, qt.IsNil)
	c.Assert(len(vi.Versions()), qt.Equals, 4)
	for _, version := range vi.Versions() {
		specData, err := st.Version(ctx, version.String())
		c.Assert(err, qt.IsNil)
		l := openapi3.NewLoader()
		spec, err := l.LoadFromData(specData)
		c.Assert(err, qt.IsNil)
		c.Assert(spec, qt.IsNotNil)
		c.Assert(spec.Paths.Len(), qt.Equals, collatedPaths[version.String()])
	}
}

func TestScraperWithLegacy(t *testing.T) {
	c := qt.New(t)

	_, animalsService := setupHttpServers(c)
	tests := []struct {
		name, version, digest string
	}{
		{"animals", "2021-01-01", "sha256:XX2f9c3iySLCw54rJ/CZs+ZK6IQy7GXNY4nSOyu2QG4="},
	}

	cfg := &config.ServerConfig{
		Services: []config.ServiceConfig{
			{
				Name: "animals", URL: animalsService.URL,
			},
		},
	}
	st := disk.New("/tmp/specs")
	c.Cleanup(func() {
		ds, ok := st.(*disk.Storage)
		if !ok {
			return
		}
		err := ds.Cleanup()
		c.Assert(err, qt.IsNil)
	})
	sc, err := scraper.New(cfg, st, scraper.Clock(func() time.Time { return t0 }))
	c.Assert(err, qt.IsNil)

	// Cancel the scrape context after a timeout so we don't hang the test
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
	c.Cleanup(cancel)

	// Run the scrape
	err = sc.Run(ctx)
	c.Assert(err, qt.IsNil)

	// Legacy (default) version should not be stored
	for _, test := range tests {
		ok, err := st.HasVersion(ctx, test.name, test.version, test.digest)
		c.Assert(err, qt.IsNil)
		c.Assert(ok, qt.IsFalse)
	}
}

func TestEmptyScrape(t *testing.T) {
	c := qt.New(t)
	cfg := &config.ServerConfig{
		Services: nil,
	}
	st := disk.New("/tmp/specs")
	c.Cleanup(func() {
		ds, ok := st.(*disk.Storage)
		if !ok {
			return
		}
		err := ds.Cleanup()
		c.Assert(err, qt.IsNil)
	})
	sc, err := scraper.New(cfg, st, scraper.Clock(func() time.Time { return t0 }))
	c.Assert(err, qt.IsNil)

	// Cancel the scrape context after a timeout so we don't hang the test
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
	c.Cleanup(cancel)

	// Run the scrape
	err = sc.Run(ctx)
	c.Assert(err, qt.IsNil)
}

func TestScrapeClientError(t *testing.T) {
	c := qt.New(t)
	cfg := &config.ServerConfig{
		Services: []config.ServiceConfig{{Name: "nope", URL: "http://example.com/nope"}},
	}
	st := disk.New("/tmp/specs")
	c.Cleanup(func() {
		ds, ok := st.(*disk.Storage)
		if !ok {
			return
		}
		err := ds.Cleanup()
		c.Assert(err, qt.IsNil)
	})
	sc, err := scraper.New(cfg, st,
		scraper.Clock(func() time.Time { return t0 }),
		scraper.HTTPClient(&http.Client{
			Transport: &errorTransport{},
		}),
	)
	c.Assert(err, qt.IsNil)

	// Cancel after a short timeout so we don't hang the test
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*1)
	c.Cleanup(cancel)

	// Run the scrape
	err = sc.Run(ctx)
	c.Assert(err, qt.ErrorMatches, `.*: bad wolf`)
}

type errorTransport struct{}

func (*errorTransport) RoundTrip(*http.Request) (*http.Response, error) {
	return nil, fmt.Errorf("bad wolf")
}
