package generator

import (
	"bytes"
	"io/ioutil"
	"os"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/snyk/vervet/v3/config"
	"github.com/snyk/vervet/v3/testdata"
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

func TestVersionScope(t *testing.T) {
	c := qt.New(t)
	setup(c)

	configBuf, err := ioutil.ReadFile(".vervet.yaml")
	c.Assert(err, qt.IsNil)
	proj, err := config.Load(bytes.NewBuffer(configBuf))
	c.Assert(err, qt.IsNil)

	out := c.TempDir()

	versionReadme := `
version-readme:
  scope: version
  filename: "` + out + `/{{.API}}/{{.Resource.Name}}/{{.Resource.Version.DateString}}/README"
  template: ".vervet/resource/version/README.tmpl"
`
	generatorsConf, err := config.LoadGenerators(bytes.NewBufferString(versionReadme))
	c.Assert(err, qt.IsNil)

	genReadme, err := New(generatorsConf["version-readme"], Debug(true))
	c.Assert(err, qt.IsNil)

	resources, err := MapResources(proj)
	c.Assert(err, qt.IsNil)
	err = genReadme.Execute(resources)
	c.Assert(err, qt.IsNil)

	for _, test := range []struct {
		resource, version string
	}{{
		"projects", "2021-06-04",
	}, {
		"projects", "2021-08-20",
	}} {
		readme, err := ioutil.ReadFile(out + "/testdata/" + test.resource + "/" + test.version + "/README")
		c.Assert(err, qt.IsNil)
		c.Assert(string(readme), qt.Equals, ``+
			`This is a generated scaffold for version `+test.version+"~experimental of the\n"+
			test.resource+" resource in API testdata.\n\n")
	}
}

func TestResourceScope(t *testing.T) {
	c := qt.New(t)
	setup(c)

	configBuf, err := ioutil.ReadFile(".vervet.yaml")
	c.Assert(err, qt.IsNil)
	proj, err := config.Load(bytes.NewBuffer(configBuf))
	c.Assert(err, qt.IsNil)

	out := c.TempDir()

	versionReadme := `
resource-routes:
  scope: resource
  filename: "` + out + `/{{ .API }}/{{ .ResourceVersions.Name }}/routes.ts"
  template: ".vervet/resource/routes.ts.tmpl"
`
	generatorsConf, err := config.LoadGenerators(bytes.NewBufferString(versionReadme))
	c.Assert(err, qt.IsNil)

	genReadme, err := New(generatorsConf["resource-routes"], Debug(true))
	c.Assert(err, qt.IsNil)

	resources, err := MapResources(proj)
	c.Assert(err, qt.IsNil)
	err = genReadme.Execute(resources)
	c.Assert(err, qt.IsNil)

	routes, err := ioutil.ReadFile(out + "/testdata/hello-world/routes.ts")
	c.Assert(err, qt.IsNil)
	c.Assert(string(routes), qt.Equals, ""+
		"import type * as express from 'express';\n"+
		"import type { V3Request, V3Response } from '../../../../framework';\n"+
		"// TODO: route hello-world 2021-06-01~experimental\n"+
		"\n"+
		"// TODO: route hello-world 2021-06-07~experimental\n"+
		"\n"+
		"// TODO: route hello-world 2021-06-13~beta\n")

	routes, err = ioutil.ReadFile(out + "/testdata/projects/routes.ts")
	c.Assert(err, qt.IsNil)
	c.Assert(string(routes), qt.Equals, ""+
		"import type * as express from 'express';\n"+
		"import type { V3Request, V3Response } from '../../../../framework';\n"+
		"// TODO: route projects 2021-06-04~experimental\n"+
		"\n"+
		"// TODO: route projects 2021-08-20~experimental\n")
}
