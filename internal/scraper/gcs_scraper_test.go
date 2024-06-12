package scraper_test

import (
	"context"
	"net/http/httptest"
	"testing"
	"time"

	qt "github.com/frankban/quicktest"
	"github.com/getkin/kin-openapi/openapi3"

	"github.com/snyk/vervet/v6/config"
	"github.com/snyk/vervet/v6/internal/scraper"
	"github.com/snyk/vervet/v6/internal/storage/gcs"
	gcstesting "github.com/snyk/vervet/v6/internal/storage/gcs/testing"
)

func TestGCSScraper(t *testing.T) {
	// Arrange
	c := qt.New(t)
	gcsCfg := gcstesting.Setup(c)
	ctx := context.Background()
	petfoodService, animalsService := setupHttpServers(c)
	tests := []struct {
		service, version, digest string
	}{
		{"petfood", "2021-09-01", "sha256:I20cAQ3VEjDrY7O0B678yq+0pYN2h3sxQy7vmdlo4+w="},
		{"animals", "2021-10-16", "sha256:P1FEFvnhtxJSqXr/p6fMNKE+HYwN6iwKccBGHIVZbyg="},
	}

	cfg := &config.ServerConfig{
		Services: []config.ServiceConfig{{
			Name: "petfood", URL: petfoodService.URL,
		}, {
			Name: "animals", URL: animalsService.URL,
		}},
	}

	client, err := gcs.New(ctx, gcsCfg)
	c.Assert(err, qt.IsNil)
	st, ok := client.(*gcs.Storage)
	c.Assert(ok, qt.IsTrue)

	sc, err := scraper.New(cfg, st, scraper.Clock(func() time.Time { return t0 }))
	c.Assert(err, qt.IsNil)

	// Cancel the scrape context after a timeout so we don't hang the test
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
	c.Cleanup(cancel)

	// No version digests should be known
	for _, test := range tests {
		ok, err := st.HasVersion(ctx, test.service, test.version, test.digest)
		c.Assert(err, qt.IsNil)
		c.Assert(ok, qt.IsFalse)
	}

	// Run the scrape
	err = sc.Run(ctx)
	c.Assert(err, qt.IsNil)

	// Version digests now known to storage
	for _, test := range tests {
		ok, err := st.HasVersion(ctx, test.service, test.version, test.digest)
		c.Assert(err, qt.IsNil)
		c.Assert(ok, qt.IsTrue)
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
		c.Assert(len(spec.Paths), qt.Equals, collatedPaths[version.String()])
	}
}

func TestGCSScraperCollation(t *testing.T) {
	// Arrange
	c := qt.New(t)
	gcsCfg := gcstesting.Setup(c)

	ctx := context.Background()
	petfoodService := httptest.NewServer(petfood.Handler())
	c.Cleanup(petfoodService.Close)

	animalsService := httptest.NewServer(animals.Handler())
	c.Cleanup(animalsService.Close)

	tests := []struct {
		service, version, digest string
	}{
		{"petfood", "2021-09-01", "sha256:I20cAQ3VEjDrY7O0B678yq+0pYN2h3sxQy7vmdlo4+w="},
		{"animals", "2021-10-16", "sha256:P1FEFvnhtxJSqXr/p6fMNKE+HYwN6iwKccBGHIVZbyg="},
	}

	cfg := &config.ServerConfig{
		Services: []config.ServiceConfig{{
			Name: "petfood", URL: petfoodService.URL,
		}, {
			Name: "animals", URL: animalsService.URL,
		}},
	}

	client, err := gcs.New(ctx, gcsCfg)
	c.Assert(err, qt.IsNil)
	st, ok := client.(*gcs.Storage)
	c.Assert(ok, qt.IsTrue)

	sc, err := scraper.New(cfg, st, scraper.Clock(func() time.Time { return t0 }))
	c.Assert(err, qt.IsNil)

	// Cancel the scrape context after a timeout so we don't hang the test
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
	c.Cleanup(cancel)

	// Run the scrape
	err = sc.Run(ctx)
	c.Assert(err, qt.IsNil)

	// Version digests now known to storage
	for _, test := range tests {
		ok, err := st.HasVersion(ctx, test.service, test.version, test.digest)
		c.Assert(err, qt.IsNil)
		c.Assert(ok, qt.IsTrue)
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
		c.Assert(len(spec.Paths), qt.Equals, collatedPaths[version.String()])
	}
}
