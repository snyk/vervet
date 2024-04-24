package disk

import (
	"context"
	"fmt"
	"testing"
	"time"

	qt "github.com/frankban/quicktest"

	"vervet-underground/internal/storage"
)

var t0 = time.Date(2021, time.December, 3, 20, 49, 51, 0, time.UTC)

func TestNotifyVersions(t *testing.T) {
	c := qt.New(t)
	s := New("/tmp/specs")
	c.Cleanup(func() {
		ds, ok := s.(*Storage)
		if !ok {
			return
		}
		err := ds.Cleanup()
		c.Assert(err, qt.IsNil)
	})
	ctx := context.Background()
	err := s.NotifyVersions(ctx, "petfood", []string{"2021-09-01", "2021-09-16"}, t0)
	c.Assert(err, qt.IsNil)
	// TODO: verify side-effects when there are some...
}

func TestHasVersion(t *testing.T) {
	c := qt.New(t)
	s := New("/tmp/specs")
	c.Cleanup(func() {
		ds, ok := s.(*Storage)
		if !ok {
			return
		}
		err := ds.Cleanup()
		c.Assert(err, qt.IsNil)
	})
	ctx := context.Background()
	const cricketsDigest = "sha256:mWpHX0/hIZS9mVd8eobfHWm6OkUsKZLiqd6ShRnNzA4="
	const geckosDigest = "sha256:c5JD7m0g4DVhoaX4z8HFcTP8S/yUOEsjgP8ECkuEHqM="
	for _, digest := range []string{cricketsDigest, geckosDigest} {
		ok, err := s.HasVersion(ctx, "petfood", "2021-09-16", digest)
		c.Assert(err, qt.IsNil)
		c.Assert(ok, qt.IsFalse)
	}
	err := s.NotifyVersion(ctx, "petfood", "2021-09-16", []byte("crickets"), t0)
	c.Assert(err, qt.IsNil)
	err = s.NotifyVersion(ctx, "animals", "2021-09-16", []byte("geckos"), t0)
	c.Assert(err, qt.IsNil)

	tests := []struct {
		service, version, digest string
		shouldHave               bool
	}{
		{"petfood", "2021-09-16", cricketsDigest, true},
		{"animals", "2021-09-16", geckosDigest, true},
		{"petfood", "2021-09-16", geckosDigest, false},
		{"animals", "2021-09-16", cricketsDigest, false},
		{"petfood", "2021-10-16", cricketsDigest, false},
		{"animals", "2021-09-17", geckosDigest, false},
	}
	for i, t := range tests {
		c.Logf("test#%d: %v", i, t)
		ok, err := s.HasVersion(ctx, t.service, t.version, t.digest)
		c.Assert(err, qt.IsNil)
		c.Assert(ok, qt.Equals, t.shouldHave)
	}
}

const spec = `{"components":{},"info":{"title":"ServiceA API","version":"0.0.0"},` +
	`"openapi":"3.0.0","paths":{"/test":{"get":{"operation":"getTest",` +
	`"responses":{"204":{"description":"An empty response"}},"summary":"Test endpoint"}}}}`

const emptySpec = `{"components":{},"info":{"title":"","version":""},"openapi":"","paths":null}`

func TestCollateVersions(t *testing.T) {
	c := qt.New(t)
	s := New("/tmp/specs")
	c.Cleanup(func() {
		ds, ok := s.(*Storage)
		if !ok {
			return
		}
		err := ds.Cleanup()
		c.Assert(err, qt.IsNil)
	})

	ctx := context.Background()
	err := s.NotifyVersion(ctx, "petfood", "2021-09-16", []byte(emptySpec), t0)
	c.Assert(err, qt.IsNil)

	serviceFilter := map[string]bool{"petfood": true}
	err = s.CollateVersions(ctx, serviceFilter)
	c.Assert(err, qt.IsNil)
	before, err := s.Version(ctx, "2021-09-16")
	c.Assert(err, qt.IsNil)
	c.Assert(string(before), qt.Equals, emptySpec)

	content, err := s.Version(ctx, "2021-01-01")
	c.Assert(err.Error(), qt.Equals, fmt.Errorf("no matching version").Error())
	c.Assert(content, qt.IsNil)

	err = s.NotifyVersion(ctx, "petfood", "2021-09-16", []byte(spec), t0.Add(time.Second))
	c.Assert(err, qt.IsNil)
	err = s.CollateVersions(ctx, serviceFilter)
	c.Assert(err, qt.IsNil)

	after, err := s.Version(ctx, "2021-09-16")
	c.Assert(err, qt.IsNil)
	c.Assert(string(after), qt.Equals, spec)
}

func TestDiskStorageCollateVersion(t *testing.T) {
	c := qt.New(t)
	s := New("/tmp/specs")
	c.Cleanup(func() {
		ds, ok := s.(*Storage)
		if !ok {
			return
		}
		err := ds.Cleanup()
		c.Assert(err, qt.IsNil)
	})

	storage.AssertCollateVersion(c, s)
}
