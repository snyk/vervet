package generator

import (
	"bytes"
	"io/ioutil"
	"os"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/snyk/vervet/v4/config"
	"github.com/snyk/vervet/v4/testdata"
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
  filename: "` + out + `/{{.API}}/{{.Resource}}/{{.Version.DateString}}/README"
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
  filename: "` + out + `/{{ .API }}/{{ .Resource }}/routes.ts"
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
	c.Assert(string(routes), qt.Equals, `
import { versions } from '@snyk/rest-node-libs';
import * as v2021_06_13 './2021-06-13';
import * as v2021_06_01 './2021-06-01';
import * as v2021_06_07 './2021-06-07';
import * as v2021_06_13 './2021-06-13';

export const helloWorldCreate = versions([
  {
    handler: v2021_06_13.helloWorldCreate,
    version: '2021-06-13~beta',
  },
]);
export const helloWorldGetOne = versions([
  {
    handler: v2021_06_01.helloWorldGetOne,
    version: '2021-06-01~experimental',
  },

  {
    handler: v2021_06_07.helloWorldGetOne,
    version: '2021-06-07~experimental',
  },

  {
    handler: v2021_06_13.helloWorldGetOne,
    version: '2021-06-13~beta',
  },
]);
`[1:])

	routes, err = ioutil.ReadFile(out + "/testdata/projects/routes.ts")
	c.Assert(err, qt.IsNil)
	c.Assert(string(routes), qt.Equals, `
import { versions } from '@snyk/rest-node-libs';
import * as v2021_08_20 './2021-08-20';
import * as v2021_06_04 './2021-06-04';

export const deleteOrgsProject = versions([
  {
    handler: v2021_08_20.deleteOrgsProject,
    version: '2021-08-20~experimental',
  },
]);
export const getOrgsProjects = versions([
  {
    handler: v2021_06_04.getOrgsProjects,
    version: '2021-06-04~experimental',
  },
]);
`[1:])
}
