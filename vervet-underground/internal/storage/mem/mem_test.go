package mem

import (
	"fmt"
	"testing"
	"time"

	qt "github.com/frankban/quicktest"
)

var t0 = time.Date(2021, time.December, 3, 20, 49, 51, 0, time.UTC)

func TestNotifyVersions(t *testing.T) {
	c := qt.New(t)
	s := New()
	err := s.NotifyVersions("petfood", []string{"2021-09-01", "2021-09-16"}, t0)
	c.Assert(err, qt.IsNil)
	// TODO: verify side-effects when there are some...
}

func TestHasVersion(t *testing.T) {
	c := qt.New(t)
	s := New()

	const cricketsDigest = "sha256:mWpHX0/hIZS9mVd8eobfHWm6OkUsKZLiqd6ShRnNzA4="
	const geckosDigest = "sha256:c5JD7m0g4DVhoaX4z8HFcTP8S/yUOEsjgP8ECkuEHqM="
	for _, digest := range []string{cricketsDigest, geckosDigest} {
		ok, err := s.HasVersion("petfood", "2021-09-16", digest)
		c.Assert(err, qt.IsNil)
		c.Assert(ok, qt.IsFalse)
	}
	err := s.NotifyVersion("petfood", "2021-09-16", []byte("crickets"), t0)
	c.Assert(err, qt.IsNil)
	err = s.NotifyVersion("animals", "2021-09-16", []byte("geckos"), t0)
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
		ok, err := s.HasVersion(t.service, t.version, t.digest)
		c.Assert(err, qt.IsNil)
		c.Assert(ok, qt.Equals, t.shouldHave)
	}
}

const spec = `
openapi: 3.0.0
info:
  title: ServiceA API
  version: 0.0.0
paths:
  /test:
    get:
      operation: getTest
      summary: Test endpoint
      responses:
        '204':
          description: An empty response
`
const emptySpec = `{"components":{},"info":null,"openapi":"","paths":null}`

func TestCollateVersions(t *testing.T) {
	c := qt.New(t)
	s := New()

	err := s.NotifyVersion("petfood", "2021-09-16", []byte(spec), t0)
	c.Assert(err, qt.IsNil)

	err = s.CollateVersions()
	c.Assert(err, qt.IsNil)

	tests := []struct {
		version        string
		empty          bool
		errorCondition qt.Checker
	}{
		{"2021-09-16", false, qt.Equals},
		{"2021-01-01", true, qt.Not(qt.Equals)},
		{"2021-09-16", false, qt.Equals},
	}
	for i, t := range tests {
		c.Run(fmt.Sprintf("test#%d: %v", i, t), func(c *qt.C) {
			content, err := s.Version(t.version)
			c.Assert(err, t.errorCondition, nil)
			if t.empty {
				c.Assert(string(content), t.errorCondition, emptySpec)
				return
			}
			c.Assert(string(content), qt.Not(qt.Equals), emptySpec)
		})
	}
}
