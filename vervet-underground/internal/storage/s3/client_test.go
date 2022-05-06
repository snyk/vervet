package s3_test

import (
	"bytes"
	"context"
	"os"
	"strconv"
	"testing"

	qt "github.com/frankban/quicktest"
	"github.com/rs/zerolog/log"

	"vervet-underground/internal/storage"
	"vervet-underground/internal/storage/s3"
)

const (
	bucketName           = "vervet-underground-specs"
	localstackAccessKey  = "test"
	localstackSecretKey  = "test"
	localstackSessionKey = "test"
	awsEndpoint          = "http://localhost:4566"
	awsRegion            = "us-east-1"
)

var cfg = &s3.Config{
	AwsRegion:   awsRegion,
	AwsEndpoint: awsEndpoint,
	BucketName:  bucketName,
	Credentials: s3.StaticKeyCredentials{
		AccessKey:  localstackAccessKey,
		SecretKey:  localstackSecretKey,
		SessionKey: localstackSessionKey,
	},
}

func cleanup() {
	// cleanup
	ctx := context.Background()
	client, err := s3.New(ctx, cfg)
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

func isCIEnabled(t *testing.T) bool {
	t.Helper()

	ci, err := strconv.ParseBool(os.Getenv("CI"))
	return err == nil || ci
}

func setup(t *testing.T) *qt.C {
	t.Helper()
	c := qt.New(t)
	if isCIEnabled(t) {
		c.Skip("CI not enabled")
	}
	c.Cleanup(cleanup)
	return c
}

func TestPutObject(t *testing.T) {
	// Arrange
	c := setup(t)

	ctx := context.Background()
	st, err := s3.New(ctx, cfg)
	c.Assert(err, qt.IsNil)
	client, ok := st.(*s3.Storage)
	c.Assert(ok, qt.IsTrue)

	data := []byte("this is some data stored as a byte slice in Go Lang!")

	// convert byte slice to io.Reader
	reader := bytes.NewReader(data)
	obj, err := client.PutObject(ctx, "dummy", reader)
	c.Assert(err, qt.IsNil)

	// Assert
	c.Assert(obj, qt.IsNotNil)
	c.Assert(obj.ETag, qt.IsNotNil)
}

func TestGetObject(t *testing.T) {
	// Arrange
	c := setup(t)

	ctx := context.Background()
	st, err := s3.New(ctx, cfg)
	c.Assert(err, qt.IsNil)
	client, ok := st.(*s3.Storage)
	c.Assert(ok, qt.IsTrue)

	data := "this is some data stored as a byte slice in Go Lang!"

	// convert byte slice to io.Reader
	reader := bytes.NewReader([]byte(data))
	_, err = client.PutObject(ctx, storage.CollatedVersionsFolder+"spec.txt", reader)
	c.Assert(err, qt.IsNil)

	// Assert
	res, err := client.GetObject(ctx, storage.CollatedVersionsFolder+"spec.txt")
	c.Assert(err, qt.IsNil)
	c.Assert(string(res), qt.Equals, data)

	// Fail silently
	res, err = client.GetObject(ctx, storage.CollatedVersionsFolder+"dummy.txt")
	c.Assert(err, qt.IsNil)
	c.Assert(string(res), qt.Equals, "")
}

func TestListObjectsAndPrefixes(t *testing.T) {
	// Arrange
	c := setup(t)

	ctx := context.Background()
	st, err := s3.New(ctx, cfg)
	c.Assert(err, qt.IsNil)
	client, ok := st.(*s3.Storage)
	c.Assert(ok, qt.IsTrue)

	objects, err := client.ListObjects(ctx, storage.CollatedVersionsFolder, "/")
	c.Assert(err, qt.IsNil)
	c.Assert(objects.Contents, qt.IsNil)

	data := "this is some data stored as a byte slice in Go Lang!"

	// convert byte slice to io.Reader
	reader := bytes.NewReader([]byte(data))
	_, err = client.PutObject(ctx, storage.CollatedVersionsFolder+"2022-02-02/spec.txt", reader)
	c.Assert(err, qt.IsNil)

	// Assert
	res, err := client.ListObjects(ctx, storage.CollatedVersionsFolder, "/")
	c.Assert(err, qt.IsNil)
	c.Assert(res.Contents, qt.IsNil)

	versions, err := client.ListCollatedVersions(ctx)
	c.Assert(err, qt.IsNil)
	c.Assert(versions, qt.Contains, "2022-02-02")
}
