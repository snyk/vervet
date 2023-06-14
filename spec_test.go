package vervet_test

import (
	"testing"

	qt "github.com/frankban/quicktest"
	"github.com/getkin/kin-openapi/openapi3"

	. "github.com/snyk/vervet/v5"
	"github.com/snyk/vervet/v5/testdata"
)

func TestSpecs(t *testing.T) {
	c := qt.New(t)
	specs, err := LoadSpecVersions(testdata.Path("resources"))
	c.Assert(err, qt.IsNil)
	versions := specs.Versions()
	c.Assert(versions, qt.ContentEquals, VersionSlice{
		MustParseVersion("2021-06-01~experimental"),
		MustParseVersion("2021-06-04~experimental"),
		MustParseVersion("2021-06-07~experimental"),
		MustParseVersion("2021-06-13~experimental"),
		MustParseVersion("2021-06-13~beta"),
		MustParseVersion("2021-08-20~experimental"),
		MustParseVersion("2021-08-20~beta"),
		MustParseVersion("2023-06-01~experimental"),
		MustParseVersion("2023-06-01~beta"),
	})

	type expectResourceVersion struct {
		version string
		path    string
		opFunc  func(*openapi3.PathItem) *openapi3.Operation
	}
	tests := []struct {
		query, match string
		hasVersions  []expectResourceVersion
		err          string
	}{{
		query: "2021-07-01~experimental",
		match: "2021-06-13~experimental",
		hasVersions: []expectResourceVersion{{
			version: "2021-06-13~beta",
			path:    "/examples/hello-world",
			opFunc:  func(p *openapi3.PathItem) *openapi3.Operation { return p.Post },
		}, {
			version: "2021-06-13~beta",
			path:    "/examples/hello-world/{id}",
			opFunc:  func(p *openapi3.PathItem) *openapi3.Operation { return p.Get },
		}, {
			version: "2021-06-04~experimental",
			path:    "/orgs/{orgId}/projects",
			opFunc:  func(p *openapi3.PathItem) *openapi3.Operation { return p.Get },
		}},
	}, {
		query: "2021-09-01~experimental",
		match: "2021-08-20~experimental",
		hasVersions: []expectResourceVersion{{
			version: "2021-06-13~beta",
			path:    "/examples/hello-world",
			opFunc:  func(p *openapi3.PathItem) *openapi3.Operation { return p.Post },
		}, {
			version: "2021-06-13~beta",
			path:    "/examples/hello-world/{id}",
			opFunc:  func(p *openapi3.PathItem) *openapi3.Operation { return p.Get },
		}, {
			version: "2021-06-04~experimental",
			path:    "/orgs/{orgId}/projects",
			opFunc:  func(p *openapi3.PathItem) *openapi3.Operation { return p.Get },
		}, {
			version: "2021-08-20~experimental",
			path:    "/orgs/{org_id}/projects/{project_id}",
			opFunc:  func(p *openapi3.PathItem) *openapi3.Operation { return p.Delete },
		}},
	}, {
		query: "2021-07-01~wip",
		match: "2021-06-13~experimental",
	}, {
		query: "2021-06-01",
		err:   "no matching version",
	}, {
		query: "2021-07-01~beta",
		match: "2021-06-13~beta",
		hasVersions: []expectResourceVersion{{
			version: "2021-06-13~beta",
			path:    "/examples/hello-world",
			opFunc:  func(p *openapi3.PathItem) *openapi3.Operation { return p.Post },
		}, {
			version: "2021-06-13~beta",
			path:    "/examples/hello-world/{id}",
			opFunc:  func(p *openapi3.PathItem) *openapi3.Operation { return p.Get },
		}},
	}}
	for i, t := range tests {
		c.Logf("test#%d: %#v", i, t)
		spec, err := specs.At(MustParseVersion(t.query))
		if t.err != "" {
			c.Assert(err, qt.ErrorMatches, t.err)
			continue
		}
		c.Assert(err, qt.IsNil)
		_, err = ExtensionString(spec.Extensions, ExtSnykApiStability)
		c.Assert(err, qt.ErrorMatches, `extension "x-snyk-api-stability" not found`)
		c.Assert(IsExtensionNotFound(err), qt.IsTrue)
		version, err := ExtensionString(spec.Extensions, ExtSnykApiVersion)
		c.Assert(err, qt.IsNil)
		c.Assert(version, qt.Equals, t.match)
		for _, expected := range t.hasVersions {
			pathItem := spec.Paths[expected.path]
			c.Assert(pathItem, qt.Not(qt.IsNil))
			op := expected.opFunc(pathItem)
			c.Assert(op, qt.Not(qt.IsNil))
			versionStr, err := ExtensionString(op.Extensions, ExtSnykApiVersion)
			c.Assert(err, qt.IsNil)
			c.Assert(versionStr, qt.Equals, expected.version)
		}
	}
}
