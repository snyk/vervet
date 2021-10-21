package cmd_test

import (
	"io/ioutil"
	"os"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/snyk/vervet/cmd"
	"github.com/snyk/vervet/testdata"
)

func TestScaffold(t *testing.T) {
	c := qt.New(t)
	dstDir := c.Mkdir()
	cd(c, dstDir)
	// Create an API project from a scaffold
	err := cmd.App.Run([]string{"vervet", "scaffold", "init", testdata.Path("test-scaffold")})
	c.Assert(err, qt.IsNil)
	// Generate a new resource version in the project
	err = cmd.App.Run([]string{"vervet", "version", "new", "--version", "2021-10-01", "v3", "foo"})
	c.Assert(err, qt.IsNil)
	for _, item := range []string{".vervet/templates/README.tmpl", ".vervet.yaml", ".vervet/extras/foo", ".vervet/extras/bar/bar"} {
		_, err = os.Stat(item)
		c.Assert(err, qt.IsNil)
	}
	readme, err := ioutil.ReadFile("v3/resources/foo/2021-10-01/README")
	c.Assert(err, qt.IsNil)
	c.Assert(string(readme), qt.Equals, `
This is a generated scaffold for version 2021-10-01~wip of the
foo resource in API v3.

`[1:])
}
