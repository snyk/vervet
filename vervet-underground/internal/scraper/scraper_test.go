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

	"vervet-underground/internal/scraper"
	"vervet-underground/internal/service"
	"vervet-underground/internal/storage/mem"
)

var (
	t0            = time.Date(2021, time.December, 3, 20, 49, 51, 0, time.UTC)
	collatedPaths = map[string]int{
		"2021-09-01": 1,
		"2021-09-16": 2,
		"2021-10-01": 3,
		"2021-10-16": 4,
	}

	petfood = &testService{
		versions: []string{"2021-09-01", "2021-09-16"},
		contents: map[string]string{
			"2021-09-01": `{"paths":{"/crickets": {}}}`,
			"2021-09-16": `{"paths":{"/crickets": {}, "/kibble": {}}}`,
		},
	}
	animals = &testService{
		versions: []string{"2021-10-01", "2021-10-16"},
		contents: map[string]string{
			"2021-10-01": `{"paths":{"/geckos": {}}}`,
			"2021-10-16": `{"paths":{"/geckos": {}, "/puppies": {}}}`,
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
		service, version, digest string
	}{
		{petfoodService.URL, "2021-09-01", "sha256:I20cAQ3VEjDrY7O0B678yq+0pYN2h3sxQy7vmdlo4+w="},
		{animalsService.URL, "2021-10-16", "sha256:P1FEFvnhtxJSqXr/p6fMNKE+HYwN6iwKccBGHIVZbyg="},
	}

	reg := service.NewRegistry(
		service.StaticServiceLoader([]string{
			petfoodService.URL,
			animalsService.URL,
		}))
	st := mem.New()
	sc, err := scraper.New(reg, st, scraper.Clock(func() time.Time { return t0 }))
	c.Assert(err, qt.IsNil)

	// Cancel the scrape context after a timeout so we don't hang the test
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
	c.Cleanup(cancel)

	// No version digests should be known
	for _, test := range tests {
		ok, err := st.HasVersion(test.service, test.version, test.digest)
		c.Assert(err, qt.IsNil)
		c.Assert(ok, qt.IsFalse)
	}

	// Run the scrape
	err = sc.Run(ctx)
	c.Assert(err, qt.IsNil)

	// Version digests now known to storage
	for _, test := range tests {
		ok, err := st.HasVersion(test.service, test.version, test.digest)
		c.Assert(err, qt.IsNil)
		c.Assert(ok, qt.IsTrue)
	}

	c.Assert(len(st.Versions()), qt.Equals, 4)
	for _, version := range st.Versions() {
		specData, err := st.Version(version)
		c.Assert(err, qt.IsNil)
		l := openapi3.NewLoader()
		spec, err := l.LoadFromData(specData)
		c.Assert(err, qt.IsNil)
		c.Assert(spec, qt.IsNotNil)
		c.Assert(len(spec.Paths), qt.Equals, collatedPaths[version])
	}
}

func TestEmptyScrape(t *testing.T) {
	c := qt.New(t)
	reg := service.NewRegistry(service.StaticServiceLoader(nil))
	st := mem.New()
	sc, err := scraper.New(reg, st, scraper.Clock(func() time.Time { return t0 }))
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
	reg := service.NewRegistry(service.StaticServiceLoader([]string{"http://example.com/nope"}))
	st := mem.New()
	sc, err := scraper.New(reg, st,
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

func TestScraperCollation(t *testing.T) {
	c := qt.New(t)

	petfoodService, animalsService := setupHttpServers(c)
	tests := []struct {
		service, version, digest string
	}{
		{petfoodService.URL, "2021-09-01", "sha256:I20cAQ3VEjDrY7O0B678yq+0pYN2h3sxQy7vmdlo4+w="},
		{animalsService.URL, "2021-10-16", "sha256:P1FEFvnhtxJSqXr/p6fMNKE+HYwN6iwKccBGHIVZbyg="},
	}

	reg := service.NewRegistry(
		service.StaticServiceLoader([]string{
			petfoodService.URL,
			animalsService.URL,
		}))
	st := mem.New()
	sc, err := scraper.New(reg, st, scraper.Clock(func() time.Time { return t0 }))
	c.Assert(err, qt.IsNil)

	// Cancel the scrape context after a timeout so we don't hang the test
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
	c.Cleanup(cancel)

	// Run the scrape
	err = sc.Run(ctx)
	c.Assert(err, qt.IsNil)

	// Version digests now known to storage
	for _, test := range tests {
		ok, err := st.HasVersion(test.service, test.version, test.digest)
		c.Assert(err, qt.IsNil)
		c.Assert(ok, qt.IsTrue)
	}

	collated, err := st.GetCollatedVersionSpecs()
	c.Assert(err, qt.IsNil)
	c.Assert(len(collated), qt.Equals, 4)

	c.Assert(len(st.Versions()), qt.Equals, 4)
	for _, version := range st.Versions() {
		specData, err := st.Version(version)
		c.Assert(err, qt.IsNil)
		l := openapi3.NewLoader()
		spec, err := l.LoadFromData(specData)
		c.Assert(err, qt.IsNil)
		c.Assert(spec, qt.IsNotNil)
		c.Assert(len(spec.Paths), qt.Equals, collatedPaths[version])
	}
}
