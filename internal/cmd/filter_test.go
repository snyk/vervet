package cmd_test

import (
	"os"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/snyk/vervet/v5"
	"github.com/snyk/vervet/v5/internal/cmd"
	"github.com/snyk/vervet/v5/testdata"
)

func TestFilterInclude(t *testing.T) {
	c := qt.New(t)
	tmpOut := c.TempDir()

	c.Run("filter include hello-world", func(c *qt.C) {
		stdout, err := os.Create(tmpOut + "/spec.yaml")
		c.Assert(err, qt.IsNil)
		defer stdout.Close()
		c.Patch(&os.Stdout, stdout)
		err = cmd.Vervet.Run(
			[]string{
				"vervet",
				"filter",
				"-I",
				"/examples/hello-world/{id}",
				testdata.Path("output/2021-06-01~experimental/spec.json"),
			},
		)
		c.Assert(err, qt.IsNil)
	})

	out, err := os.ReadFile(tmpOut + "/spec.yaml")
	c.Assert(err, qt.IsNil)
	c.Log(string(out))

	doc, err := vervet.NewDocumentFile(tmpOut + "/spec.yaml")
	c.Assert(err, qt.IsNil)

	// Included paths and their referenced components are present
	c.Assert(doc.Paths["/examples/hello-world/{id}"], qt.Not(qt.IsNil))
	c.Assert(doc.Components.Schemas["HelloWorld"], qt.Not(qt.IsNil))

	// Not-included paths are not present
	c.Assert(doc.Paths["/openapi"], qt.IsNil)
}

func XestFilterExclude(t *testing.T) {
	c := qt.New(t)
	tmpOut := c.TempDir()

	c.Run("filter hello-world", func(c *qt.C) {
		stdout, err := os.Create(tmpOut + "/spec.yaml")
		c.Assert(err, qt.IsNil)
		defer stdout.Close()
		app := cmd.NewApp(&cmd.CLIApp, cmd.VervetParams{
			Stdin:  os.Stdin,
			Stdout: stdout,
			Stderr: os.Stderr,
			Prompt: cmd.Prompt{},
		})
		err = app.Run(
			[]string{
				"vervet",
				"filter",
				"-X",
				"/examples/hello-world/{id}",
				testdata.Path("output/2021-06-01~experimental/spec.json"),
			},
		)
		c.Assert(err, qt.IsNil)
	})

	doc, err := vervet.NewDocumentFile(tmpOut + "/spec.yaml")
	c.Assert(err, qt.IsNil)

	// Excluded paths and components only these reference, are not present
	c.Assert(doc.Paths["/examples/hello-world/{id}"], qt.IsNil)
	c.Assert(doc.Components.Schemas["HelloWorld"], qt.IsNil)

	// Not-excluded paths and referenced components are present
	c.Assert(doc.Paths["/openapi"], qt.Not(qt.IsNil))
	c.Assert(doc.Paths["/openapi/{version}"], qt.Not(qt.IsNil))
	c.Assert(doc.Components.Headers["VersionRequestedResponseHeader"], qt.Not(qt.IsNil))
}
