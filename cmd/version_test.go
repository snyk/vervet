package cmd_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/snyk/vervet/v3"
	"github.com/snyk/vervet/v3/cmd"
	"github.com/snyk/vervet/v3/testdata"
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

func copyToDir(c *qt.C, srcFile, dstDir string) {
	buf, err := ioutil.ReadFile(srcFile)
	c.Assert(err, qt.IsNil)
	err = ioutil.WriteFile(filepath.Join(dstDir, filepath.Base(srcFile)), buf, 0666)
	c.Assert(err, qt.IsNil)
}

func TestVersionFiles(t *testing.T) {
	c := qt.New(t)
	tmp := c.TempDir()
	tmpFile := filepath.Join(tmp, "out")
	c.Run("cmd", func(c *qt.C) {
		output, err := os.Create(tmpFile)
		c.Assert(err, qt.IsNil)
		defer output.Close()
		c.Patch(&os.Stdout, output)
		cd(c, testdata.Path("."))
		err = cmd.Vervet.Run([]string{"vervet", "version", "files"})
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

func TestVersionList(t *testing.T) {
	c := qt.New(t)
	tmp := c.TempDir()
	tmpFile := filepath.Join(tmp, "out")
	c.Run("cmd", func(c *qt.C) {
		output, err := os.Create(tmpFile)
		c.Assert(err, qt.IsNil)
		defer output.Close()
		c.Patch(&os.Stdout, output)
		cd(c, testdata.Path("."))
		err = cmd.Vervet.Run([]string{"vervet", "version", "list"})
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

func TestVersionNew(t *testing.T) {
	c := qt.New(t)

	// Set up vervet project directory
	projectDir := c.TempDir()
	copyToDir(c, testdata.Path(".vervet.yaml"), projectDir)
	copyToDir(c, testdata.Path("compiled-rules.yaml"), projectDir)
	copyToDir(c, testdata.Path("resource-rules.yaml"), projectDir)
	versionTemplateDir := filepath.Join(projectDir, ".vervet", "resource", "version")
	c.Assert(os.MkdirAll(versionTemplateDir, 0777), qt.IsNil)
	copyToDir(c, testdata.Path(".vervet/resource/version/README.tmpl"), versionTemplateDir)
	copyToDir(c, testdata.Path(".vervet/resource/version/controller.ts.tmpl"), versionTemplateDir)
	copyToDir(c, testdata.Path(".vervet/resource/version/index.ts.tmpl"), versionTemplateDir)
	copyToDir(c, testdata.Path(".vervet/resource/version/spec.yaml.tmpl"), versionTemplateDir)
	cd(c, projectDir)

	// Set up vervet app for testing.
	prompt := testPrompt{}
	testApp := cmd.NewApp(cmd.VervetParams{
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
		Prompt: &prompt,
	})

	// Test to ensure spec is created correctly.
	checkVersion := func(c *qt.C, versionName string) {
		versions, err := vervet.LoadResourceVersions(filepath.Join(projectDir, "generated", versionName))
		c.Assert(err, qt.IsNil)
		c.Assert(versions.Name(), qt.Equals, versionName)
		c.Assert(versions.Versions(), qt.HasLen, 1)
		rc, err := versions.At(versions.Versions()[0].String())
		c.Assert(err, qt.IsNil)
		c.Assert(rc.Paths, qt.HasLen, 2)
	}

	// Running it with all args creates the new spec.
	err := testApp.Run([]string{"vervet", "version", "new", "testdata", "foo"})
	c.Assert(err, qt.IsNil)
	checkVersion(c, "foo")

	// Running it without resource prompts for resource name.
	prompt.ReturnEntry = "baz"
	err = testApp.Run([]string{"vervet", "version", "new", "testdata"})
	c.Assert(err, qt.IsNil)
	checkVersion(c, "baz")

	// Running it with no args prompts for api and resource.
	prompt.ReturnSelect = "testdata"
	prompt.ReturnEntry = "bar"
	err = testApp.Run([]string{"vervet", "version", "new"})
	c.Assert(err, qt.IsNil)
	checkVersion(c, "bar")
}
