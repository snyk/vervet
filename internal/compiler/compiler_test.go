package compiler

import (
	"bytes"
	"context"
	"io/ioutil"
	"os"
	"testing"
	"text/template"

	qt "github.com/frankban/quicktest"

	"github.com/snyk/vervet/v4/config"
	"github.com/snyk/vervet/v4/internal/files"
	"github.com/snyk/vervet/v4/internal/linter"
	"github.com/snyk/vervet/v4/testdata"
)

func setup(c *qt.C) {
	c.Setenv("API_BASE_URL", "https://example.com/api/v3")
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
linters:
  resource-rules:
    spectral:
      rules:
        - 'node_modules/@snyk/sweater-comb/resource.yaml'
  compiled-rules:
    spectral:
      rules:
        - 'node_modules/@snyk/sweater-comb/compiled.yaml'
apis:
  v3-api:
    resources:
      - linter: resource-rules
        path: 'testdata/resources'
        excludes:
          - 'testdata/resources/schemas/**'
    overlays:
      - include: 'testdata/resources/include.yaml'
      - inline: |-
          servers:
            - url: ${API_BASE_URL}
              description: Snyk API
    output:
      path: {{ . }}
      linter: compiled-rules
`[1:]))

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
	err = ioutil.WriteFile(outputPath+"/goof", []byte("goof"), 0777)
	c.Assert(err, qt.IsNil)

	proj, err := config.Load(bytes.NewBuffer(configBuf.Bytes()))
	c.Assert(err, qt.IsNil)
	compiler, err := New(ctx, proj, LinterFactory(func(context.Context, *config.Linter) (linter.Linter, error) {
		return &mockLinter{}, nil
	}))
	c.Assert(err, qt.IsNil)

	// Assert constructor set things up as expected
	c.Assert(compiler.apis, qt.HasLen, 1)
	c.Assert(compiler.linters, qt.HasLen, 2)
	v3Api := compiler.apis["v3-api"]
	c.Assert(v3Api, qt.Not(qt.IsNil))
	c.Assert(v3Api.resources, qt.HasLen, 1)
	c.Assert(v3Api.resources[0].sourceFiles, qt.Contains, "testdata/resources/projects/2021-06-04/spec.yaml")
	c.Assert(v3Api.overlayIncludes, qt.HasLen, 1)
	c.Assert(v3Api.overlayIncludes[0].Paths, qt.HasLen, 2)
	c.Assert(v3Api.overlayInlines[0].Servers[0].URL, qt.Contains, "https://example.com/api/v3", qt.Commentf("environment variable interpolation"))
	c.Assert(v3Api.output, qt.Not(qt.IsNil))

	// LintResources stage
	err = compiler.LintResourcesAll(ctx)
	c.Assert(err, qt.IsNil)
	c.Assert(compiler.linters["resource-rules"].(*mockLinter).runs, qt.HasLen, 1)
	c.Assert(compiler.linters["compiled-rules"].(*mockLinter).runs, qt.HasLen, 0)
	c.Assert(compiler.linters["resource-rules"].(*mockLinter).runs[0], qt.Contains, "testdata/resources/projects/2021-06-04/spec.yaml")

	// Build stage
	err = compiler.BuildAll(ctx)
	c.Assert(err, qt.IsNil)

	// Verify created files/folders are as expected
	// Look for existence of /2021-06-01~experimental
	_, err = os.Stat(outputPath + "/2021-06-01~experimental")
	c.Assert(err, qt.IsNil)

	// Look for absence of /2021-06-01 folder (ga)
	_, err = os.Stat(outputPath + "/2021-06-01")
	c.Assert(os.IsNotExist(err), qt.IsTrue)

	// Build output was cleaned up
	_, err = ioutil.ReadFile(outputPath + "/goof")
	c.Assert(err, qt.ErrorMatches, ".*/goof: no such file or directory")

	// LintOutput stage
	err = compiler.LintOutputAll(ctx)
	c.Assert(err, qt.IsNil)
	c.Assert(compiler.linters["resource-rules"].(*mockLinter).runs, qt.HasLen, 1)
	c.Assert(compiler.linters["compiled-rules"].(*mockLinter).runs, qt.HasLen, 1)
	c.Assert(compiler.linters["compiled-rules"].(*mockLinter).runs[0], qt.Contains, outputPath+"/2021-06-04~experimental/spec.yaml")
	c.Assert(compiler.linters["compiled-rules"].(*mockLinter).runs[0], qt.Contains, outputPath+"/2021-06-04~experimental/spec.json")
}

type mockLinter struct {
	runs     [][]string
	override *config.Linter
	err      error
}

func (l *mockLinter) Match(rcConfig *config.ResourceSet) ([]string, error) {
	return files.LocalFSSource{}.Match(rcConfig)
}

func (l *mockLinter) Run(ctx context.Context, root string, paths ...string) error {
	l.runs = append(l.runs, paths)
	return l.err
}

func (l *mockLinter) WithOverride(ctx context.Context, cfg *config.Linter) (linter.Linter, error) {
	nl := &mockLinter{
		override: cfg,
	}
	return nl, nil
}
