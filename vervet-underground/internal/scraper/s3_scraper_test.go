package scraper_test

import (
	"context"
	"net/http/httptest"
	"testing"
	"time"

	qt "github.com/frankban/quicktest"
	"github.com/getkin/kin-openapi/openapi3"

	"vervet-underground/config"
	"vervet-underground/internal/scraper"
	"vervet-underground/internal/storage/s3"
	s3testing "vervet-underground/internal/storage/s3/testing"
)

func TestS3Scraper(t *testing.T) {
	c := qt.New(t)
	s3Cfg := s3testing.Setup(c)

	ctx := context.Background()
	petfoodService, animalsService := setupHttpServers(c)
	tests := []struct {
		name, version, digest string
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
	st, err := s3.New(ctx, s3Cfg)
	c.Assert(err, qt.IsNil)
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
		ok, err := st.HasVersion(ctx, test.name, test.version, test.digest)
		c.Assert(err, qt.IsNil)
		c.Assert(ok, qt.IsTrue)
	}

	c.Assert(len(st.Versions()), qt.Equals, 4)
	for _, version := range st.Versions() {
		specData, err := st.Version(ctx, version)
		c.Assert(err, qt.IsNil)
		l := openapi3.NewLoader()
		spec, err := l.LoadFromData(specData)
		c.Assert(err, qt.IsNil)
		c.Assert(spec, qt.IsNotNil)
		c.Assert(len(spec.Paths), qt.Equals, collatedPaths[version])
	}
}

func TestS3ScraperCollation(t *testing.T) {
	c := qt.New(t)
	s3Cfg := s3testing.Setup(c)

	ctx := context.Background()
	petfoodService := httptest.NewServer(petfood.Handler())
	c.Cleanup(petfoodService.Close)

	animalsService := httptest.NewServer(animals.Handler())
	c.Cleanup(animalsService.Close)

	tests := []struct {
		name, version, digest string
	}{{
		"petfood", "2021-09-01", "sha256:I20cAQ3VEjDrY7O0B678yq+0pYN2h3sxQy7vmdlo4+w=",
	}, {
		"animals", "2021-10-16", "sha256:P1FEFvnhtxJSqXr/p6fMNKE+HYwN6iwKccBGHIVZbyg=",
	}}

	cfg := &config.ServerConfig{
		Services: []config.ServiceConfig{{
			Name: "petfood", URL: petfoodService.URL,
		}, {
			Name: "animals", URL: animalsService.URL,
		}},
	}
	st, err := s3.New(ctx, s3Cfg)
	c.Assert(err, qt.IsNil)
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
		ok, err := st.HasVersion(ctx, test.name, test.version, test.digest)
		c.Assert(err, qt.IsNil)
		c.Assert(ok, qt.IsTrue)
	}

	c.Assert(len(st.Versions()), qt.Equals, 4)
	for _, version := range st.Versions() {
		specData, err := st.Version(ctx, version)
		c.Assert(err, qt.IsNil)
		l := openapi3.NewLoader()
		spec, err := l.LoadFromData(specData)
		c.Assert(err, qt.IsNil)
		c.Assert(spec, qt.IsNotNil)
		c.Assert(len(spec.Paths), qt.Equals, collatedPaths[version])
	}
}
