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
	"vervet-underground/internal/storage/s3"
)

const (
	localstackAccessKey  = "test"
	localstackSecretKey  = "test"
	localstackSessionKey = "test"
	awsEndpoint          = "http://localhost:4566"
	awsRegion            = "us-east-1"
	awsBucketName        = "vervet-underground-specs"
)

var s3Cfg = &s3.Config{
	AwsRegion:   awsRegion,
	AwsEndpoint: awsEndpoint,
	BucketName:  awsBucketName,
	Credentials: s3.StaticKeyCredentials{
		AccessKey:  localstackAccessKey,
		SecretKey:  localstackSecretKey,
		SessionKey: localstackSessionKey,
	},
}

func isCIEnabled(t *testing.T) bool {
	t.Helper()

	ci, err := strconv.ParseBool(os.Getenv("CI"))
	return err == nil || ci
}

func s3cleanup() {
	// cleanup
	ctx := context.Background()
	client, err := s3.New(ctx, s3Cfg)
	if err != nil {
		log.Error().Err(err).Msg("failed to initialize S3 storage")
		return
	}
	st, ok := client.(*s3.Storage)
	if !ok {
		log.Error().Err(err).Msg("failed to cast to S3 storage")
		return
	}
	revs, err := st.ListObjects(ctx, "", "")

	if err != nil {
		log.Error().Err(err).Msg("failed to List Objects")
		return
	}
	for _, rev := range revs.Contents {
		err = st.DeleteObject(ctx, *rev.Key)
		if err != nil {
			log.Error().Err(err).Msgf("failed to delete Object %s", *rev.Key)
		}
	}
}

func s3Setup(t *testing.T) *qt.C {
	t.Helper()
	c := qt.New(t)
	if isCIEnabled(t) {
		c.Skip("CI not enabled")
	}
	c.Cleanup(s3cleanup)
	return c
}

func TestS3Scraper(t *testing.T) {
	c := s3Setup(t)

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
	c := s3Setup(t)

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
