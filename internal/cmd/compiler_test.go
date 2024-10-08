package cmd_test

import (
	"context"
	"os"
	"testing"

	qt "github.com/frankban/quicktest"
	"github.com/getkin/kin-openapi/openapi3"

	"github.com/snyk/vervet/v8"
	"github.com/snyk/vervet/v8/internal/cmd"
	"github.com/snyk/vervet/v8/testdata"
)

var specFile = "/spec.yaml"

func TestBuild(t *testing.T) {
	c := qt.New(t)
	dstDir := c.TempDir()
	err := cmd.Vervet.Run([]string{"vervet", "build", testdata.Path("resources"), dstDir})
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
		c.Run("built version "+test.version, func(c *qt.C) {
			doc, err := vervet.NewDocumentFile(dstDir + "/" + test.version + specFile)
			c.Assert(err, qt.IsNil)
			c.Assert(doc.Validate(context.TODO()), qt.IsNil)
			for _, path := range test.paths {
				c.Assert(doc.Paths.Value(path), qt.Not(qt.IsNil))
			}
		})
	}
}

func TestBuildConflict(t *testing.T) {
	c := qt.New(t)
	dstDir := c.TempDir()
	err := cmd.Vervet.Run([]string{"vervet", "build", testdata.Path("conflict"), dstDir})
	c.Assert(err, qt.ErrorMatches, `failed to load spec versions: conflict: .*`)
}

func TestBuildInclude(t *testing.T) {
	c := qt.New(t)
	dstDir := c.TempDir()
	err := cmd.Vervet.Run(
		[]string{
			"vervet",
			"build",
			"-I",
			testdata.Path("resources/include.yaml"),
			testdata.Path("resources"),
			dstDir,
		},
	)
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
		// Load just-built OpenAPI YAML file
		doc, err := vervet.NewDocumentFile(dstDir + "/" + test.version + specFile)
		c.Assert(err, qt.IsNil)

		expected, err := os.ReadFile(testdata.Path("output/" + test.version + "/spec.json"))
		c.Assert(err, qt.IsNil)

		// Servers will differ between the fixture output and the above, since
		// testdata/.vervet.yaml contains an overlay that modifies the servers:
		// section. This patches the output to match expected.
		doc.Servers = []*openapi3.Server{{
			Description: "Test REST API",
			URL:         "https://example.com/api/rest",
		}}

		c.Assert(expected, qt.JSONEquals, doc)
	}
}
