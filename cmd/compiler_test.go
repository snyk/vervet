package cmd_test

import (
	"context"
	"io/ioutil"
	"testing"

	qt "github.com/frankban/quicktest"
	"github.com/getkin/kin-openapi/openapi3"

	"github.com/snyk/vervet"
	"github.com/snyk/vervet/cmd"
	"github.com/snyk/vervet/testdata"
)

func TestCompile(t *testing.T) {
	c := qt.New(t)
	dstDir := c.TempDir()
	err := cmd.Vervet.Run([]string{"vervet", "compile", testdata.Path("resources"), dstDir})
	c.Assert(err, qt.IsNil)
	tests := []struct {
		version string
		paths   []string
	}{{
		version: "2021-06-01~experimental",
		paths:   []string{"/examples/hello-world/{id}"},
	}, {
		version: "2021-06-07~experimental",
		paths:   []string{"/examples/hello-world/{id}"},
	}, {
		version: "2021-06-13~beta",
		paths:   []string{"/examples/hello-world", "/examples/hello-world/{id}"},
	}, {
		version: "2021-06-04~experimental",
		paths:   []string{"/examples/hello-world/{id}", "/orgs/{orgId}/projects"},
	}}
	for _, test := range tests {
		c.Run("compiled version "+test.version, func(c *qt.C) {
			doc, err := vervet.NewDocumentFile(dstDir + "/" + test.version + "/spec.yaml")
			c.Assert(err, qt.IsNil)
			c.Assert(doc.Validate(context.TODO()), qt.IsNil)
			for _, path := range test.paths {
				c.Assert(doc.Paths[path], qt.Not(qt.IsNil))
			}
		})
	}
}

func TestCompileInclude(t *testing.T) {
	c := qt.New(t)
	dstDir := c.TempDir()
	err := cmd.Vervet.Run([]string{"vervet", "compile", "-I", testdata.Path("resources/include.yaml"), testdata.Path("resources"), dstDir})
	c.Assert(err, qt.IsNil)

	tests := []struct {
		version string
	}{{
		version: "2021-06-01~experimental",
	}, {
		version: "2021-06-07~experimental",
	}, {
		version: "2021-06-13~beta",
	}, {
		version: "2021-06-04~experimental",
	}}
	for _, test := range tests {
		c.Assert(err, qt.IsNil)
		// Load just-compiled OpenAPI YAML file
		doc, err := vervet.NewDocumentFile(dstDir + "/" + test.version + "/spec.yaml")
		c.Assert(err, qt.IsNil)

		expected, err := ioutil.ReadFile(testdata.Path("output/" + test.version + "/spec.json"))
		c.Assert(err, qt.IsNil)

		// Servers will differ between the fixture output and the above, since
		// testdata/.vervet.yaml contains an overlay that modifies the servers:
		// section. This patches the output to match expected.
		doc.Servers = []*openapi3.Server{{
			Description: "Test API v3",
			URL:         "https://example.com/api/v3",
		}}

		c.Assert(expected, qt.JSONEquals, doc)
	}
}

func TestCompileConflict(t *testing.T) {
	c := qt.New(t)
	dstDir := c.TempDir()
	err := cmd.Vervet.Run([]string{"vervet", "compile", "../testdata/conflict", dstDir})
	c.Assert(err, qt.ErrorMatches, `failed to load spec versions: conflict: .*`)
}
