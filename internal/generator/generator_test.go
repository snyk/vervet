package generator

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/snyk/vervet/config"
	"github.com/snyk/vervet/testdata"
)

func setup(c *qt.C) {
	cwd, err := os.Getwd()
	c.Assert(err, qt.IsNil)

	err = os.Chdir(testdata.Path("."))
	c.Assert(err, qt.IsNil)
	c.Cleanup(func() {
		err := os.Chdir(cwd)
		c.Assert(err, qt.IsNil)
	})
}

func TestGenerators(t *testing.T) {
	c := qt.New(t)
	setup(c)

	generated := c.Mkdir()
	configBuf, err := ioutil.ReadFile(testdata.Path(".vervet.yaml"))
	c.Assert(err, qt.IsNil)
	configBuf = bytes.ReplaceAll(configBuf, []byte("generated/{{"), []byte(generated+"/{{"))
	proj, err := config.Load(bytes.NewBuffer(configBuf))
	c.Assert(err, qt.IsNil)

	genMap, err := NewMap(proj, Debug(true))
	c.Assert(err, qt.IsNil)

	scope := &VersionScope{
		API:       "testdata",
		Resource:  "foo",
		Version:   "2021-09-01",
		Stability: "beta",
	}

	// Smoke test generated files. Fixture test will catch the rest.

	// Test a single-file generator with no data
	readme := genMap["version-readme"]
	err = readme.Run(scope)
	c.Assert(err, qt.IsNil)
	contents, err := ioutil.ReadFile(filepath.Join(generated, "foo/2021-09-01/README"))
	c.Assert(err, qt.IsNil)
	c.Log(string(contents))
	c.Assert(string(contents), qt.Equals, `
This is a generated scaffold for version 2021-09-01~beta of the
foo resource in API testdata.

`[1:])

	// Generate an OpenAPI spec, used by the following generator
	spec := genMap["version-spec"]
	err = spec.Run(scope)
	c.Assert(err, qt.IsNil)

	// Test a multi-file generator with data
	controller := genMap["version-controller"]
	err = controller.Run(scope)
	c.Assert(err, qt.IsNil)
	contents, err = ioutil.ReadFile(filepath.Join(generated, "foo/2021-09-01/createFoo.ts"))
	c.Assert(err, qt.IsNil)
	c.Assert(string(contents), qt.Contains, `export const createFoo = async (`)
}

func TestVersionScope(t *testing.T) {
	c := qt.New(t)
	s := &VersionScope{
		API:      "someapi",
		Resource: "somerc",
		Version:  "abc",
	}
	c.Assert(s.validate(), qt.ErrorMatches, `invalid version "abc"`)
	s = &VersionScope{
		API:       "someapi",
		Resource:  "somerc",
		Version:   "2021-07-01",
		Stability: "shaky",
	}
	c.Assert(s.validate(), qt.ErrorMatches, `invalid stability "shaky"`)
}
