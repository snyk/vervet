package vervet_test

import (
	"context"
	"testing"
	"time"

	qt "github.com/frankban/quicktest"

	. "github.com/snyk/vervet/v7"
	"github.com/snyk/vervet/v7/testdata"
)

func TestResource(t *testing.T) {
	c := qt.New(t)
	eps, err := LoadResourceVersions(testdata.Path("resources/_examples/hello-world"))
	c.Assert(err, qt.IsNil)
	c.Assert(eps.Versions(), qt.DeepEquals, VersionSlice{{
		Date:      time.Date(2021, time.June, 1, 0, 0, 0, 0, time.UTC),
		Stability: StabilityExperimental,
	}, {
		Date:      time.Date(2021, time.June, 7, 0, 0, 0, 0, time.UTC),
		Stability: StabilityExperimental,
	}, {
		Date:      time.Date(2021, time.June, 13, 0, 0, 0, 0, time.UTC),
		Stability: StabilityBeta,
	}})
	for _, v := range eps.Versions() {
		e, err := eps.At(v.String())
		c.Assert(err, qt.IsNil)
		c.Assert(e.Validate(context.Background()), qt.IsNil)
		c.Assert(e.Version, qt.DeepEquals, v)
	}
}

func TestVersionRangesHelloWorld(t *testing.T) {
	c := qt.New(t)
	eps, err := LoadResourceVersions(testdata.Path("resources/_examples/hello-world"))
	c.Assert(err, qt.IsNil)
	tests := []struct {
		query, match string
	}{{
		query: "2021-07-01~experimental",
		match: "2021-06-13~beta",
	}, {
		query: "2021-07-01~beta",
		match: "2021-06-13~beta",
	}, {
		query: "2021-06-08~experimental",
		match: "2021-06-07~experimental",
	}}
	for _, t := range tests {
		e, err := eps.At(t.query)
		c.Assert(err, qt.IsNil)
		c.Assert(e.Version.String(), qt.Equals, t.match)
	}
}

func TestVersionRangesProjects(t *testing.T) {
	c := qt.New(t)
	eps, err := LoadResourceVersions(testdata.Path("resources/projects"))
	c.Assert(err, qt.IsNil)
	c.Assert(eps.Versions(), qt.HasLen, 3)
	tests := []struct {
		query, match, err string
	}{{
		query: "2021-07-01~experimental",
		match: "2021-06-04~experimental",
	}, {
		query: "2021-07-01~beta",
		err:   `no matching version`,
	}, {
		query: "2021-07-01",
		err:   `no matching version`,
	}}
	for i, t := range tests {
		c.Logf("test#%d: %#v", i, t)
		e, err := eps.At(t.query)
		if t.err != "" {
			c.Assert(err, qt.ErrorMatches, t.err)
		} else {
			c.Assert(err, qt.IsNil)
			c.Assert(e.Version.String(), qt.Equals, t.match)
		}
	}
}

func TestIsExtensionNotFound(t *testing.T) {
	c := qt.New(t)
	eps, err := LoadResourceVersions(testdata.Path("resources/_examples/hello-world"))
	c.Assert(err, qt.IsNil)
	resource, err := eps.At("2021-06-04~experimental")
	c.Assert(err, qt.IsNil)

	_, err = ExtensionString(resource.Extensions, ExtSnykApiVersion)
	c.Assert(IsExtensionNotFound(err), qt.IsTrue)

	_, err = ExtensionString(resource.Extensions, "some-bogus-value")
	c.Assert(IsExtensionNotFound(err), qt.IsTrue)

	_, err = ExtensionString(resource.Extensions, ExtSnykApiStability)
	c.Assert(IsExtensionNotFound(err), qt.IsFalse)
}

func TestLoadResourceVersionsWithDuplicateSpecs(t *testing.T) {
	c := qt.New(t)
	resourceVersions, err := LoadResourceVersions(testdata.Path("duplicate-specs"))
	dirPath := testdata.Path("duplicate-specs")
	c.Assert(resourceVersions, qt.IsNil)
	c.Assert(err, qt.IsNotNil)
	c.Assert(err, qt.ErrorMatches, "duplicate spec found in "+dirPath+"/2022-08-31")
}

func TestResourceVersionsAtSunset(t *testing.T) {
	c := qt.New(t)
	eps, err := LoadResourceVersions(testdata.Path("sunset-specs"))
	c.Assert(err, qt.IsNil)
	c.Assert(eps.Versions(), qt.HasLen, 2)
	tests := []struct {
		query, match, err string
	}{{
		query: "2023-01-01~experimental",
		match: "2023-01-01~experimental",
	}, {
		query: "2023-01-10~experimental",
		match: "2023-01-01~experimental",
	}, {
		query: "2023-02-01~experimental",
		match: "2023-02-01~experimental",
	}, {
		query: "2023-03-01~experimental",
		err:   "no matching version",
	}}
	for i, t := range tests {
		c.Logf("test#%d: %#v", i, t)
		e, err := eps.At(t.query)
		if t.err != "" {
			c.Assert(err, qt.ErrorMatches, t.err)
		} else {
			c.Assert(err, qt.IsNil)
			c.Assert(e.Version.String(), qt.Equals, t.match)
		}
	}
}
