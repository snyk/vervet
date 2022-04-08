package config_test

import (
	"bytes"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/snyk/vervet/v4/config"
)

func TestLoad(t *testing.T) {
	c := qt.New(t)
	conf := bytes.NewBufferString(`
version: "1"
linters:
  apitest-resource:
    description: Test resource rules
    spectral:
      rules:
        - resource-rules.yaml
  apitest-compiled:
    description: Test compiled rules
    spectral:
      rules:
        - compiled-rules.yaml
      extraArgs: ['--format', 'json', '-v']
  ci-rules:
    optic-ci:
      original: target-branch
      ciContext: ci-context.json
      uploadResults: true
apis:
  test:
    resources:
      - linter: apitest-resource
        path: testdata/resources
        excludes:
          - testdata/resources/schemas/**
      - linter: ci-rules
        path: testdata/resources
        excludes:
          - testdata/resources/schemas/**
    overlays:
      - inline: |-
          servers:
            - url: ${API_BASE_URL}
              description: Test API
    output:
      path: testdata/output
      linter: apitest-compiled
`)
	proj, err := config.Load(conf)
	c.Assert(err, qt.IsNil)
	c.Assert(proj, qt.DeepEquals, &config.Project{
		Version:    "1",
		Generators: config.Generators{},
		Linters: config.Linters{
			"apitest-resource": {
				Name:        "apitest-resource",
				Description: "Test resource rules",
				Spectral: &config.SpectralLinter{
					Rules: []string{
						"resource-rules.yaml",
					},
					ExtraArgs: []string{"--format", "text"},
				},
			},
			"apitest-compiled": {
				Name:        "apitest-compiled",
				Description: "Test compiled rules",
				Spectral: &config.SpectralLinter{
					Rules: []string{
						"compiled-rules.yaml",
					},
					ExtraArgs: []string{"--format", "json", "-v"},
				},
			},
			"ci-rules": {
				Name: "ci-rules",
				OpticCI: &config.OpticCILinter{
					Image:         "snyk/sweater-comb:latest",
					Original:      "target-branch",
					CIContext:     "ci-context.json",
					UploadResults: true,
				},
			},
		},
		APIs: config.APIs{
			"test": {
				Name: "test",
				Resources: []*config.ResourceSet{{
					Linter:   "apitest-resource",
					Path:     "testdata/resources",
					Excludes: []string{"testdata/resources/schemas/**"},
				}, {
					Linter:   "ci-rules",
					Path:     "testdata/resources",
					Excludes: []string{"testdata/resources/schemas/**"},
				}},
				Overlays: []*config.Overlay{{
					Inline: `
servers:
  - url: ${API_BASE_URL}
    description: Test API`[1:],
				}},
				Output: &config.Output{
					Path:   "testdata/output",
					Linter: "apitest-compiled",
				},
			},
		},
	})
}

func TestLoadNoLinters(t *testing.T) {
	c := qt.New(t)
	conf := bytes.NewBufferString(`
version: "1"
apis:
  test:
    resources:
      - path: testdata/resources
        excludes:
          - testdata/resources/schemas/**
    overlays:
      - inline: |-
          servers:
            - url: ${API_BASE_URL}
              description: Test API
    output:
      path: testdata/output
`)
	proj, err := config.Load(conf)
	c.Assert(err, qt.IsNil)
	c.Assert(proj, qt.DeepEquals, &config.Project{
		Version:    "1",
		Generators: config.Generators{},
		Linters:    config.Linters{},
		APIs: config.APIs{
			"test": {
				Name: "test",
				Resources: []*config.ResourceSet{{
					Path:     "testdata/resources",
					Excludes: []string{"testdata/resources/schemas/**"},
				}},
				Overlays: []*config.Overlay{{
					Inline: `
servers:
  - url: ${API_BASE_URL}
    description: Test API`[1:],
				}},
				Output: &config.Output{
					Path: "testdata/output",
				},
			},
		},
	})
}

func TestLoadErrors(t *testing.T) {
	c := qt.New(t)
	tests := []struct {
		conf string
		err  string
	}{{
		conf: `version: "2"`,
		err:  `unsupported version "2"`,
	}, {
		conf: `version: "1"`,
		err:  `no apis defined`,
	}, {
		conf: `
version: "1"
apis:
  testapi:
    resources: []`[1:],
		err: `no resources defined \(apis\.testapi\.resources\)`,
	}, {
		conf: `
version: "1"
apis:
  testapi:
    resources:
      - path: resources
        linter: foo`[1:],
		err: `linter "foo" not found \(apis\.testapi\.resources\[0\]\.linter\)`,
	}, {
		conf: `
version: "1"
linters:
  ci:
    optic-ci: {}
apis:
  testapi:
    resources:
      - path: resources
        linter: ci
        linter-overrides:
          foo:
            2021-09-01:
              optic-ci: {}
`[1:],
		err: `optic linter does not support overrides \(apis\.testapi\.resources\[0\]\.linter-overrides\.foo\.2021-09-01\)`,
	}, {
		conf: `
version: "1"
linters:
  ci:
    optic-ci: {}
apis:
  testapi:
    resources:
      - path: resources
        linter: ci
    output:
      path: /somewhere/else
      linter: ci
`[1:],
		err: `optic linter does not yet support compiled specs \(apis\.testapi\.output\.linter\)`,
	}, {
		conf: `
version: "1"
linters:
  ci:
    optic-ci: {}
apis:
  testapi:
    resources:
      - path: resources
        linter: ci
    output:
      path: /somewhere/else
      paths:
        - /another/place
        - /and/another
      linter: ci
`[1:],
		err: `output should specify one of 'path' or 'paths', not both \(apis\.testapi\.output\)`,
	}, {
		conf: `
linters:
  ci:
`[1:],
		err: `missing linter definition \(linters\.ci\)`,
	}, {
		err: `no apis defined`,
	}}
	for i := range tests {
		c.Logf("test#%d: %s", i, tests[i].conf)
		_, err := config.Load(bytes.NewBufferString(tests[i].conf))
		c.Assert(err, qt.ErrorMatches, tests[i].err)
	}
}
