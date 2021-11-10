package vervet_test

import (
	"testing"

	qt "github.com/frankban/quicktest"

	. "github.com/snyk/vervet"
	"github.com/snyk/vervet/testdata"
)

func TestSpecs(t *testing.T) {
	c := qt.New(t)
	specs, err := LoadSpecVersions(testdata.Path("resources"))
	c.Assert(err, qt.IsNil)
	versions := specs.Versions()
	c.Assert(versions, qt.HasLen, 4)
	c.Assert(versions, qt.ContentEquals, []Version{
		MustParseVersion("2021-06-01~experimental"),
		MustParseVersion("2021-06-04~experimental"),
		MustParseVersion("2021-06-07~experimental"),
		MustParseVersion("2021-06-13~beta"),
	})

	type expectResourceVersion struct {
		version string
		path    string
	}
	tests := []struct {
		query       string
		hasVersions []expectResourceVersion
	}{{
		query: "2021-07-01~experimental",
		hasVersions: []expectResourceVersion{{
			version: "2021-06-13~beta",
			path:    "/examples/hello-world",
		}, {
			version: "2021-06-13~beta",
			path:    "/examples/hello-world/{id}",
		}, {
			version: "2021-06-04~experimental",
			path:    "/orgs/{orgId}/projects",
		}},
	}, {
		query: "2021-07-01~wip",
		hasVersions: []expectResourceVersion{{
			version: "2021-06-13~beta",
			path:    "/examples/hello-world",
		}, {
			version: "2021-06-13~beta",
			path:    "/examples/hello-world/{id}",
		}, {
			version: "2021-06-04~experimental",
			path:    "/orgs/{orgId}/projects",
		}},
	}, {
		query: "2021-07-01~beta",
		hasVersions: []expectResourceVersion{{
			version: "2021-06-13~beta",
			path:    "/examples/hello-world",
		}, {
			version: "2021-06-13~beta",
			path:    "/examples/hello-world/{id}",
		}},
	}}
	for i, t := range tests {
		c.Logf("test#%d: %#v", i, t)
		spec, err := specs.At(t.query)
		c.Assert(err, qt.IsNil)
		_, err = ExtensionString(spec.ExtensionProps, ExtSnykApiStability)
		c.Assert(err, qt.ErrorMatches, `extension "x-snyk-api-stability" not found`)
		c.Assert(IsExtensionNotFound(err), qt.IsTrue)
		m := map[expectResourceVersion]bool{}
		for path, pathItem := range spec.Paths {
			pathVersionStr, err := ExtensionString(pathItem.ExtensionProps, ExtSnykApiVersion)
			c.Assert(err, qt.IsNil)
			c.Assert(IsExtensionNotFound(err), qt.IsFalse)
			m[expectResourceVersion{version: pathVersionStr, path: path}] = true
		}
		c.Assert(m, qt.HasLen, len(t.hasVersions))
		for _, hasVersion := range t.hasVersions {
			c.Assert(m[hasVersion], qt.IsTrue)
		}
	}
}
