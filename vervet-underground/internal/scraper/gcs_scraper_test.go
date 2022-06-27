package scraper_test

import (
	"context"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
	"time"

	qt "github.com/frankban/quicktest"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/rs/zerolog/log"

	"vervet-underground/config"
	"vervet-underground/internal/scraper"
	"vervet-underground/internal/storage/gcs"
)

const (
	gcsEndpoint   = "http://localhost:4443/storage/v1/"
	gcsRegion     = "US-CENTRAL1" // https://cloud.google.com/storage/docs/locations#location-r
	projectId     = "test"
	gcsBucketName = "vervet-underground-specs"
)

var gcsCfg = &gcs.Config{
	GcsRegion:   gcsRegion,
	GcsEndpoint: gcsEndpoint,
	BucketName:  gcsBucketName,
	Credentials: gcs.StaticKeyCredentials{
		ProjectId: projectId,
	},
}

func gcsCleanup() {
	// cleanup
	ctx := context.Background()
	client, err := gcs.New(ctx, gcsCfg)
	if err != nil {
		log.Error().Err(err).Msg("failed to initialize GCS storage")
		return
	}
	st, ok := client.(*gcs.Storage)
	if !ok {
		log.Error().Msg("failed to cast to GCS storage")
		return
	}
	revs, err := st.ListObjects(ctx, "", "")

	if err != nil {
		log.Error().Err(err).Msg("failed to List Objects")
		return
	}
	for _, rev := range revs {
		if rev.Name != "" {
			err := st.DeleteObject(ctx, rev.Name)
			if err != nil {
				log.Error().Err(err).Msgf("failed to delete Object %s", rev.Prefix+"/"+rev.Name)
			}
		}
	}
}

func isGcsCIEnabled(t *testing.T) bool {
	t.Helper()

	ci, err := strconv.ParseBool(os.Getenv("CI"))
	return err == nil || ci
}

func gcsSetup(t *testing.T) *qt.C {
	t.Helper()
	c := qt.New(t)
	if isGcsCIEnabled(t) {
		c.Skip("CI not enabled")
	}
	c.Cleanup(gcsCleanup)
	return c
}

func TestGCSScraper(t *testing.T) {
	// Arrange
	c := gcsSetup(t)
	ctx := context.Background()
	petfoodService, animalsService := setupHttpServers(c)
	tests := []struct {
		service, version, digest string
	}{
		{petfoodService.URL, "2021-09-01", "sha256:I20cAQ3VEjDrY7O0B678yq+0pYN2h3sxQy7vmdlo4+w="},
		{animalsService.URL, "2021-10-16", "sha256:P1FEFvnhtxJSqXr/p6fMNKE+HYwN6iwKccBGHIVZbyg="},
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

func TestGCSScraperCollation(t *testing.T) {
	// Arrange
	c := gcsSetup(t)

	ctx := context.Background()
	petfoodService := httptest.NewServer(petfood.Handler())
	c.Cleanup(petfoodService.Close)

	animalsService := httptest.NewServer(animals.Handler())
	c.Cleanup(animalsService.Close)

	tests := []struct {
		service, version, digest string
	}{
		{petfoodService.URL, "2021-09-01", "sha256:I20cAQ3VEjDrY7O0B678yq+0pYN2h3sxQy7vmdlo4+w="},
		{animalsService.URL, "2021-10-16", "sha256:P1FEFvnhtxJSqXr/p6fMNKE+HYwN6iwKccBGHIVZbyg="},
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
