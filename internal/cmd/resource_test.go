package cmd_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/snyk/vervet/v5/internal/cmd"
	"github.com/snyk/vervet/v5/testdata"
)

func cd(c *qt.C, path string) {
	cwd, err := os.Getwd()
	c.Assert(err, qt.IsNil)
	err = os.Chdir(path)
	c.Assert(err, qt.IsNil)
	c.Cleanup(func() {
		err := os.Chdir(cwd)
		c.Assert(err, qt.IsNil)
	})
}

func TestResourceFiles(t *testing.T) {
	c := qt.New(t)
	tmp := c.TempDir()
	tmpFile := filepath.Join(tmp, "out")
	c.Run("cmd", func(c *qt.C) {
		output, err := os.Create(tmpFile)
		c.Assert(err, qt.IsNil)
		defer output.Close()
		c.Patch(&os.Stdout, output)
		cd(c, testdata.Path("."))
		err = cmd.Vervet.Run([]string{"vervet", "resource", "files"})
		c.Assert(err, qt.IsNil)
	})
	out, err := ioutil.ReadFile(tmpFile)
	c.Assert(err, qt.IsNil)
	c.Assert(string(out), qt.Equals, `
resources/_examples/hello-world/2021-06-01/spec.yaml
resources/_examples/hello-world/2021-06-07/spec.yaml
resources/_examples/hello-world/2021-06-13/spec.yaml
resources/projects/2021-06-04/spec.yaml
resources/projects/2021-08-20/spec.yaml
`[1:])
}

func TestResourceInfo(t *testing.T) {
	c := qt.New(t)
	tmp := c.TempDir()
	tmpFile := filepath.Join(tmp, "out")
	c.Run("cmd", func(c *qt.C) {
		output, err := os.Create(tmpFile)
		c.Assert(err, qt.IsNil)
		defer output.Close()
		c.Patch(&os.Stdout, output)
		cd(c, testdata.Path("."))
		err = cmd.Vervet.Run([]string{"vervet", "resource", "info"})
		c.Assert(err, qt.IsNil)
	})
	out, err := ioutil.ReadFile(tmpFile)
	c.Assert(err, qt.IsNil)
	c.Assert(string(out), qt.Equals, `
+----------+-------------+-------------------------+--------------------------------------+--------+-------------------+
|   API    |  RESOURCE   |         VERSION         |                 PATH                 | METHOD |     OPERATION     |
+----------+-------------+-------------------------+--------------------------------------+--------+-------------------+
| testdata | hello-world | 2021-06-01~experimental | /examples/hello-world/{id}           | GET    | helloWorldGetOne  |
| testdata | projects    | 2021-06-04~experimental | /orgs/{orgId}/projects               | GET    | getOrgsProjects   |
| testdata | hello-world | 2021-06-07~experimental | /examples/hello-world/{id}           | GET    | helloWorldGetOne  |
| testdata | hello-world | 2021-06-13~beta         | /examples/hello-world                | POST   | helloWorldCreate  |
| testdata | hello-world | 2021-06-13~beta         | /examples/hello-world/{id}           | GET    | helloWorldGetOne  |
| testdata | projects    | 2021-08-20~experimental | /orgs/{org_id}/projects/{project_id} | DELETE | deleteOrgsProject |
+----------+-------------+-------------------------+--------------------------------------+--------+-------------------+
`[1:])
}

func TestResourceInfoResource(t *testing.T) {
	c := qt.New(t)
	tmp := c.TempDir()
	tmpFile := filepath.Join(tmp, "out")
	c.Run("cmd", func(c *qt.C) {
		output, err := os.Create(tmpFile)
		c.Assert(err, qt.IsNil)
		defer output.Close()
		c.Patch(&os.Stdout, output)
		cd(c, testdata.Path("."))
		err = cmd.Vervet.Run([]string{"vervet", "resource", "info", "testdata", "projects"})
		c.Assert(err, qt.IsNil)
	})
	out, err := ioutil.ReadFile(tmpFile)
	c.Assert(err, qt.IsNil)
	c.Assert(string(out), qt.Equals, `
+----------+----------+-------------------------+--------------------------------------+--------+-------------------+
|   API    | RESOURCE |         VERSION         |                 PATH                 | METHOD |     OPERATION     |
+----------+----------+-------------------------+--------------------------------------+--------+-------------------+
| testdata | projects | 2021-06-04~experimental | /orgs/{orgId}/projects               | GET    | getOrgsProjects   |
| testdata | projects | 2021-08-20~experimental | /orgs/{org_id}/projects/{project_id} | DELETE | deleteOrgsProject |
+----------+----------+-------------------------+--------------------------------------+--------+-------------------+
`[1:])
}
