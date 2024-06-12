package gcs_test

import (
	"bytes"
	"context"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/snyk/vervet/v6/internal/storage"
	"github.com/snyk/vervet/v6/internal/storage/gcs"
	gcstesting "github.com/snyk/vervet/v6/internal/storage/gcs/testing"
)

func TestPutObject(t *testing.T) {
	c := qt.New(t)
	cfg := gcstesting.Setup(c)
	ctx := context.Background()
	st, err := gcs.New(ctx, cfg)
	c.Assert(err, qt.IsNil)
	client, ok := st.(*gcs.Storage)
	c.Assert(ok, qt.IsTrue)

	data := []byte("this is some data stored as a byte slice in Go Lang!")

	// convert byte slice to io.Reader
	reader := bytes.NewReader(data)
	err = client.PutObject(ctx, "dummy.txt", reader)
	c.Assert(err, qt.IsNil)
}

func TestGetObject(t *testing.T) {
	c := qt.New(t)
	cfg := gcstesting.Setup(c)

	// Arrange
	ctx := context.Background()
	st, err := gcs.New(ctx, cfg)
	c.Assert(err, qt.IsNil)
	client, ok := st.(*gcs.Storage)
	c.Assert(ok, qt.IsTrue)

	data := "this is some data stored as a byte slice in Go Lang!"

	// convert byte slice to io.Reader
	reader := bytes.NewReader([]byte(data))
	err = client.PutObject(ctx, storage.CollatedVersionsFolder+"spec.txt", reader)
	c.Assert(err, qt.IsNil)

	// Assert
	objects, err := client.ListObjects(ctx, storage.CollatedVersionsFolder, "")
	c.Assert(err, qt.IsNil)
	c.Assert(len(objects), qt.Equals, 1)
	c.Assert(objects[0].Name, qt.Equals, storage.CollatedVersionsFolder+"spec.txt")

	res, err := client.GetObject(ctx, storage.CollatedVersionsFolder+"spec.txt")
	c.Assert(err, qt.IsNil)
	c.Assert(res, qt.Not(qt.IsNil))
	c.Assert(string(res), qt.Equals, data)

	// Fail silently
	res, err = client.GetObject(ctx, storage.CollatedVersionsFolder+"dummy.txt")
	c.Assert(err, qt.IsNil)
	c.Assert(string(res), qt.Equals, "")
}

func TestListObjectsAndPrefixes(t *testing.T) {
	c := qt.New(t)
	cfg := gcstesting.Setup(c)

	// Arrange
	ctx := context.Background()
	st, err := gcs.New(ctx, cfg)
	c.Assert(err, qt.IsNil)
	client, ok := st.(*gcs.Storage)
	c.Assert(ok, qt.IsTrue)

	objects, err := client.ListObjects(ctx, storage.CollatedVersionsFolder, "")
	c.Assert(err, qt.IsNil)
	c.Assert(len(objects), qt.Equals, 0)

	data := "this is some data stored as a byte slice in Go Lang!"

	// convert byte slice to io.Reader
	reader := bytes.NewReader([]byte(data))
	err = client.PutObject(ctx, storage.CollatedVersionsFolder+"2022-02-02/spec.txt", reader)
	c.Assert(err, qt.IsNil)

	// Assert
	res, err := client.ListObjects(ctx, storage.CollatedVersionsFolder, "")
	c.Assert(err, qt.IsNil)
	c.Assert(len(res), qt.Equals, 1)

	versions, err := client.ListCollatedVersions(ctx)
	c.Assert(err, qt.IsNil)
	c.Assert(versions, qt.Contains, "2022-02-02")
}

func TestCollateVersion(t *testing.T) {
	c := qt.New(t)
	cfg := gcstesting.Setup(c)

	ctx := context.Background()
	s, err := gcs.New(ctx, cfg)
	c.Assert(err, qt.IsNil)
	storage.AssertCollateVersion(c, s)
}
