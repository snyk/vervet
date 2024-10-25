package compiler

import (
	"bytes"
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"text/template"
	"time"

	qt "github.com/frankban/quicktest"

	"github.com/snyk/vervet/v8"
	"github.com/snyk/vervet/v8/config"
	"github.com/snyk/vervet/v8/testdata"
)

func setup(c *qt.C) {
	c.Setenv("API_BASE_URL", "https://example.com/api/rest")
	cwd, err := os.Getwd()
	c.Assert(err, qt.IsNil)
	err = os.Chdir(testdata.Path(".."))
	c.Assert(err, qt.IsNil)
	c.Cleanup(func() {
		err := os.Chdir(cwd)
		c.Assert(err, qt.IsNil)
	})
}

var configTemplate = template.Must(template.New("vervet.yaml").Parse(`
apis:
  rest-api:
    resources:
      - path: 'testdata/resources'
        excludes:
          - 'testdata/resources/schemas/**'
    overlays:
      - include: 'testdata/resources/include.yaml'
      - inline: |-
          servers:
            - url: ${API_BASE_URL}
              description: Test REST API
    output:
      path: {{ . }}
`[1:]))

var configTemplateWithPaths = template.Must(template.New("vervet.yaml").Parse(`
apis:
  rest-api:
    resources:
      - path: 'testdata/resources'
        excludes:
          - 'testdata/resources/schemas/**'
    overlays:
      - include: 'testdata/resources/include.yaml'
      - inline: |-
          servers:
            - url: ${API_BASE_URL}
              description: Test REST API
    output:
      paths:
{{- range . }}
        - {{ . }}
{{- end }}
`[1:]))

var path = "/goof"

// Sanity-check the compiler at lifecycle stages in a simple scenario. This
// isn't meant to be a comprehensive end-to-end test of the compiler; those are
// done with fixtures. These are easier to break out, debug, and add specific
// asserts when the e2e fixtures fail.
func TestCompilerSmoke(t *testing.T) {
	c := qt.New(t)
	setup(c)
	ctx := context.Background()
	outputPath := c.TempDir()
	var configBuf bytes.Buffer
	err := configTemplate.Execute(&configBuf, outputPath)
	c.Assert(err, qt.IsNil)

	// Create a file that should be removed prior to build
	err = os.WriteFile(outputPath+path, []byte("goof"), 0777)
	c.Assert(err, qt.IsNil)

	proj, err := config.Load(bytes.NewBuffer(configBuf.Bytes()))
	c.Assert(err, qt.IsNil)
	compiler, err := New(ctx, proj)
	c.Assert(err, qt.IsNil)

	// Assert constructor set things up as expected
	c.Assert(compiler.apis, qt.HasLen, 1)
	restApi := compiler.apis["rest-api"]
	c.Assert(restApi, qt.Not(qt.IsNil))
	c.Assert(restApi.resources, qt.HasLen, 1)
	c.Assert(restApi.resources[0].sourceFiles, qt.Contains, "testdata/resources/projects/2021-06-04/spec.yaml")
	c.Assert(restApi.overlayIncludes, qt.HasLen, 1)
	c.Assert(restApi.overlayIncludes[0].Paths.Len(), qt.Equals, 2)
	c.Assert(
		restApi.overlayInlines[0].Servers[0].URL,
		qt.Contains,
		"https://example.com/api/rest",
		qt.Commentf("environment variable interpolation"),
	)
	c.Assert(restApi.output, qt.Not(qt.IsNil))

	// Build stage
	err = compiler.BuildAll(ctx, vervet.MustParseVersion("2024-06-01"))
	c.Assert(err, qt.IsNil)

	// Verify created files/folders are as expected
	// Look for existence of /2021-06-01~experimental
	refOutputPath := testdata.Path("output")
	assertOutputsEqual(c, refOutputPath, outputPath)

	// Look for absence of /2021-06-01 folder (ga)
	_, err = os.Stat(outputPath + "/2021-06-01")
	c.Assert(os.IsNotExist(err), qt.IsTrue)

	// Build output was cleaned up
	_, err = os.ReadFile(outputPath + path)
	c.Assert(err, qt.ErrorMatches, ".*/goof: no such file or directory")
}

func TestCompilerSmokePaths(t *testing.T) {
	c := qt.New(t)
	setup(c)
	ctx := context.Background()
	outputPaths := []string{c.TempDir(), c.TempDir()}
	var configBuf bytes.Buffer
	err := configTemplateWithPaths.Execute(&configBuf, outputPaths)
	c.Assert(err, qt.IsNil)

	// Create a file that should be removed prior to build
	err = os.WriteFile(outputPaths[0]+path, []byte("goof"), 0777)
	c.Assert(err, qt.IsNil)

	proj, err := config.Load(bytes.NewBuffer(configBuf.Bytes()))
	c.Assert(err, qt.IsNil)
	compiler, err := New(ctx, proj)
	c.Assert(err, qt.IsNil)

	// Build stage
	err = compiler.BuildAll(ctx, vervet.MustParseVersion("2024-06-01"))
	c.Assert(err, qt.IsNil)

	refOutputPath := testdata.Path("output")
	// Verify created files/folders are as expected
	for _, outputPath := range outputPaths {
		assertOutputsEqual(c, refOutputPath, outputPath)

		// Build output was cleaned up
		_, err = os.ReadFile(outputPath + path)
		c.Assert(err, qt.ErrorMatches, ".*/goof: no such file or directory")
	}
}

func TestCompilerStopsAtStopDate(t *testing.T) {
	c := qt.New(t)
	setup(c)
	ctx := context.Background()
	outputPath := c.TempDir()
	var configBuf bytes.Buffer
	err := configTemplate.Execute(&configBuf, outputPath)
	c.Assert(err, qt.IsNil)

	// Create a file that should be removed prior to build
	err = os.WriteFile(outputPath+path, []byte("goof"), 0777)
	c.Assert(err, qt.IsNil)

	proj, err := config.Load(bytes.NewBuffer(configBuf.Bytes()))
	c.Assert(err, qt.IsNil)
	compiler, err := New(ctx, proj)
	c.Assert(err, qt.IsNil)

	// Assert constructor set things up as expected
	c.Assert(compiler.apis, qt.HasLen, 1)
	restApi := compiler.apis["rest-api"]
	c.Assert(restApi, qt.Not(qt.IsNil))
	c.Assert(restApi.resources, qt.HasLen, 1)
	c.Assert(restApi.resources[0].sourceFiles, qt.Contains, "testdata/resources/projects/2021-06-04/spec.yaml")
	c.Assert(restApi.overlayIncludes, qt.HasLen, 1)
	c.Assert(restApi.overlayIncludes[0].Paths.Len(), qt.Equals, 2)
	c.Assert(
		restApi.overlayInlines[0].Servers[0].URL,
		qt.Contains,
		"https://example.com/api/rest",
		qt.Commentf("environment variable interpolation"),
	)
	c.Assert(restApi.output, qt.Not(qt.IsNil))

	// Build stage
	err = compiler.BuildAll(ctx, vervet.MustParseVersion("2021-06-04"))
	c.Assert(err, qt.IsNil)

	// Verify created files/folders are as expected
	// Look for existence of /2021-06-01~experimental
	refOutputPath := restApi.output.paths[0]
	assertOutputsEqual(c, refOutputPath, outputPath)

	// Look for absence of /2021-06-01 folder (ga)
	_, err = os.Stat(outputPath + "/2021-06-01")
	c.Assert(os.IsNotExist(err), qt.IsTrue)

	// Look for presence of /2021-06-01~experimental folder
	_, err = os.Stat(outputPath + "/2021-06-01~experimental")
	c.Assert(err, qt.IsNil)

	// Look for absence of /2021-06-04~experimental folder, as its on the stop date
	_, err = os.Stat(outputPath + "/2021-06-04~experimental")
	c.Assert(os.IsNotExist(err), qt.IsTrue,
		qt.Commentf("2021-06-04~experimental should not be created, its after the stop date"))

	// Build output was cleaned up
	_, err = os.ReadFile(outputPath + path)
	c.Assert(err, qt.ErrorMatches, ".*/goof: no such file or directory")
}

func assertOutputsEqual(c *qt.C, refDir, testDir string) {
	pivotDate, err := time.Parse("2006-01-02", "2024-10-15")
	c.Assert(err, qt.IsNil)

	err = fs.WalkDir(os.DirFS(refDir), ".", func(path string, d fs.DirEntry, err error) error {
		c.Assert(err, qt.IsNil)
		if d.IsDir() {
			return nil
		}
		if d.Name() != "spec.yaml" {
			// only comparing compiled specs here
			return nil
		}
		specDate, err := time.Parse("2006-01-02", path[:10])
		c.Assert(err, qt.IsNil)
		if !specDate.Before(pivotDate) {
			// After pivot date we exclusively use simplebuild so don't emit
			// specs from this pipeline
			return nil
		}

		outputFile, err := os.ReadFile(filepath.Join(testDir, path))
		c.Assert(err, qt.IsNil)
		refFile, err := os.ReadFile(filepath.Join(refDir, path))
		c.Assert(err, qt.IsNil)
		c.Assert(string(outputFile), qt.Equals, string(refFile), qt.Commentf("%s", path))
		return nil
	})
	c.Assert(err, qt.IsNil)
}
