package scraper_test

import (
	"context"
	qt "github.com/frankban/quicktest"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/rs/zerolog/log"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
	"time"


	"vervet-underground/internal/scraper"
	"vervet-underground/internal/service"
	"vervet-underground/internal/storage/s3"
)

const (
	localstackAccessKey  = "test"
	localstackSecretKey  = "test"
	localstackSessionKey = "test"
	awsEndpoint          = "http://localhost:4566"
	awsRegion            = "us-east-1"
)

var s3Cfg = &s3.Config{
	AwsRegion:   awsRegion,
	AwsEndpoint: awsEndpoint,
	Credentials: s3.StaticKeyCredentials{
		AccessKey:  localstackAccessKey,
		SecretKey:  localstackSecretKey,
		SessionKey: localstackSessionKey,
	},
}

func isCIEnabled() bool {
	ci, err := strconv.ParseBool(os.Getenv("CI"))
	return err == nil || ci
}

func cleanup() {
	// cleanup
	client, err := s3.New(s3Cfg)
	if err != nil {
		log.Error().Err(err).Msg("failed to initialize S3 storage")
		return
	}
	st, ok := client.(*s3.Storage)
	if !ok {
		log.Error().Err(err).Msg("failed to cast to S3 storage")
		return
	}
	revs, err := st.ListObjects("", "")

	if err != nil {
		log.Error().Err(err).Msg("failed to List Objects")
		return
	}
	for _, rev := range revs.Contents {
		err = st.DeleteObject(*rev.Key)
		if err != nil {
			log.Error().Err(err).Msgf("failed to delete Object %s", *rev.Key)
		}
	}
}

func TestS3Scraper(t *testing.T) {
	c := qt.New(t)
	if isCIEnabled() {
		c.Skip("CI not enabled")
	}
	c.Cleanup(cleanup)

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
	st, err := s3.New(s3Cfg)
	c.Assert(err, qt.IsNil)
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

func TestS3ScraperCollation(t *testing.T) {
	c := qt.New(t)
	if isCIEnabled() {
		c.Skip("CI not enabled")
	}
	c.Cleanup(cleanup)

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

	reg := service.NewRegistry(
		service.StaticServiceLoader([]string{
			petfoodService.URL,
			animalsService.URL,
		}))

	st, err := s3.New(s3Cfg)
	c.Assert(err, qt.IsNil)
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
