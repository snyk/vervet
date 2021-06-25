package cmd_test

import (
	"context"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/snyk/vervet"
	"github.com/snyk/vervet/cmd"
)

func TestCompile(t *testing.T) {
	c := qt.New(t)
	dstDir := c.Mkdir()
	err := cmd.App.Run([]string{"vervet", "compile", "../testdata", dstDir})
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
