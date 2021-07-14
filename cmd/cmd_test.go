package cmd_test

import (
	"context"
	"io/ioutil"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/snyk/vervet"
	"github.com/snyk/vervet/cmd"
)

func TestCompile(t *testing.T) {
	c := qt.New(t)
	dstDir := c.Mkdir()
	err := cmd.App.Run([]string{"vervet", "compile", "../testdata/resources", dstDir})
	c.Assert(err, qt.IsNil)
	tests := []struct {
		version string
		paths   []string
	}{{
		version: "2021-06-01",
		paths:   []string{"/examples/hello-world/{id}"},
	}, {
		version: "2021-06-07",
		paths:   []string{"/examples/hello-world/{id}"},
	}, {
		version: "beta",
		paths:   []string{"/examples/hello-world/{id}"},
	}, {
		version: "experimental",
		paths:   []string{"/examples/hello-world/{id}", "/orgs/{orgId}/projects"},
	}}
	for _, test := range tests {
		c.Run("compiled version "+test.version, func(c *qt.C) {
			doc, err := vervet.LoadSpecFile(dstDir + "/" + test.version + "/spec.yaml")
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
	dstDir := c.Mkdir()
	err := cmd.App.Run([]string{"vervet", "compile", "-I", "../testdata/resources/include.yaml", "../testdata/resources", dstDir})
	c.Assert(err, qt.IsNil)

	tests := []struct {
		version string
	}{{
		version: "2021-06-01",
	}, {
		version: "2021-06-07",
	}, {
		version: "beta",
	}, {
		version: "experimental",
	}}
	for _, test := range tests {
		// Load just-compiled OpenAPI YAML file
		doc, err := vervet.LoadSpecFile(dstDir + "/" + test.version + "/spec.yaml")
		c.Assert(err, qt.IsNil)

		expected, err := ioutil.ReadFile("../testdata/output/" + test.version + "/spec.json")
		c.Assert(err, qt.IsNil)

		c.Assert(expected, qt.JSONEquals, doc)
	}
}

func TestCompileConflict(t *testing.T) {
	c := qt.New(t)
	dstDir := c.Mkdir()
	err := cmd.App.Run([]string{"vervet", "compile", "../testdata/conflict", dstDir})
	c.Assert(err, qt.ErrorMatches, `conflict: .*`)
}
