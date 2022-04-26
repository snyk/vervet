package gcs_test

import (
	"bytes"
	"os"
	"strconv"
	"testing"

	qt "github.com/frankban/quicktest"
	"github.com/rs/zerolog/log"

	"vervet-underground/internal/storage"
	"vervet-underground/internal/storage/gcs"
)

const (
	gcsEndpoint = "http://localhost:4443/storage/v1/"
	gcsRegion   = "US-CENTRAL1" // https://cloud.google.com/storage/docs/locations#location-r
	projectId   = "test"
	bucketName  = "vervet-underground-specs"
)

var cfg = &gcs.Config{
	GcsRegion:   gcsRegion,
	GcsEndpoint: gcsEndpoint,
	BucketName:  bucketName,
	Credentials: gcs.StaticKeyCredentials{
		ProjectId: projectId,
	},
}

func cleanup() {
	// cleanup
	client, err := gcs.New(cfg)
	if err != nil {
		log.Error().Err(err).Msg("failed to initialize GCS storage")
		return
	}
	st, ok := client.(*gcs.Storage)
	if !ok {
		log.Error().Msg("failed to cast to GCS storage")
		return
	}
	revs, err := st.ListObjects("", "")

	if err != nil {
		log.Error().Err(err).Msg("failed to List Objects")
		return
	}
	for _, rev := range revs {
		if rev.Name != "" {
			err := st.DeleteObject(rev.Name)
			if err != nil {
				log.Error().Err(err).Msgf("failed to delete Object %s", rev.Prefix+"/"+rev.Name)
			}
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

	st, err := gcs.New(cfg)
	c.Assert(err, qt.IsNil)
	client, ok := st.(*gcs.Storage)
	c.Assert(ok, qt.IsTrue)

	data := []byte("this is some data stored as a byte slice in Go Lang!")

	// convert byte slice to io.Reader
	reader := bytes.NewReader(data)
	obj, err := client.PutObject("dummy.txt", reader)
	c.Assert(err, qt.IsNil)

	// Assert
	c.Assert(obj, qt.IsNotNil)
	c.Assert(obj.ObjectName(), qt.Not(qt.Equals), "")
}

func TestGetObject(t *testing.T) {
	// Arrange
	c := setup(t)

	st, err := gcs.New(cfg)
	c.Assert(err, qt.IsNil)
	client, ok := st.(*gcs.Storage)
	c.Assert(ok, qt.IsTrue)

	data := "this is some data stored as a byte slice in Go Lang!"

	// convert byte slice to io.Reader
	reader := bytes.NewReader([]byte(data))
	obj, err := client.PutObject(storage.CollatedVersionsFolder+"spec.txt", reader)
	c.Assert(err, qt.IsNil)
	c.Assert(obj.ObjectName(), qt.Equals, storage.CollatedVersionsFolder+"spec.txt")

	// Assert
	res, err := client.GetObject(storage.CollatedVersionsFolder + "spec.txt")
	c.Assert(err, qt.IsNil)
	c.Assert(string(res), qt.Equals, data)

	// Fail silently
	res, err = client.GetObject(storage.CollatedVersionsFolder + "dummy.txt")
	c.Assert(err, qt.IsNil)
	c.Assert(string(res), qt.Equals, "")
}

func TestListObjectsAndPrefixes(t *testing.T) {
	// Arrange
	c := setup(t)

	st, err := gcs.New(cfg)
	c.Assert(err, qt.IsNil)
	client, ok := st.(*gcs.Storage)
	c.Assert(ok, qt.IsTrue)

	objects, err := client.ListObjects(storage.CollatedVersionsFolder, "")
	c.Assert(err, qt.IsNil)
	c.Assert(len(objects), qt.Equals, 0)

	data := "this is some data stored as a byte slice in Go Lang!"

	// convert byte slice to io.Reader
	reader := bytes.NewReader([]byte(data))
	_, err = client.PutObject(storage.CollatedVersionsFolder+"2022-02-02/spec.txt", reader)
	c.Assert(err, qt.IsNil)

	// Assert
	res, err := client.ListObjects(storage.CollatedVersionsFolder, "")
	c.Assert(err, qt.IsNil)
	c.Assert(len(res), qt.Equals, 1)

	versions, err := client.ListCollatedVersions()
	c.Assert(err, qt.IsNil)
	c.Assert(versions, qt.Contains, "2022-02-02")
}
