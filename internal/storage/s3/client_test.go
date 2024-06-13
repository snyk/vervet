package s3_test

import (
	"bytes"
	"context"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/snyk/vervet/v7/internal/storage"
	"github.com/snyk/vervet/v7/internal/storage/s3"
	s3testing "github.com/snyk/vervet/v7/internal/storage/s3/testing"
)

func TestPutObject(t *testing.T) {
	// Arrange
	c := qt.New(t)
	cfg := s3testing.Setup(c)

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
	c := qt.New(t)
	cfg := s3testing.Setup(c)

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
	c := qt.New(t)
	cfg := s3testing.Setup(c)

	ctx := context.Background()
	st, err := s3.New(ctx, cfg)
	c.Assert(err, qt.IsNil)
	client, ok := st.(*s3.Storage)
	c.Assert(ok, qt.IsTrue)

	objects, err := client.ListObjects(ctx, storage.CollatedVersionsFolder, "/")
	c.Assert(err, qt.IsNil)
	c.Assert(objects, qt.Not(qt.IsNil))
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

func TestHandleAwsError(t *testing.T) {
	c := qt.New(t)
	cfg := s3testing.Setup(c)
	ctx := context.Background()

	st, err := s3.New(ctx, cfg)
	c.Assert(err, qt.IsNil)
	client, ok := st.(*s3.Storage)
	c.Assert(ok, qt.IsTrue)

	// Fail silently
	res, err := client.GetObject(ctx, storage.CollatedVersionsFolder+"dummy.txt")
	c.Assert(err, qt.IsNil)
	c.Assert(string(res), qt.Equals, "")
}

func TestS3StorageCollateVersion(t *testing.T) {
	c := qt.New(t)
	cfg := s3testing.Setup(c)
	ctx := context.Background()
	s, err := s3.New(ctx, cfg)
	c.Assert(err, qt.IsNil)
	storage.AssertCollateVersion(c, s)
}
