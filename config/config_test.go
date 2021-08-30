package config_test

import (
	"bytes"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/snyk/vervet/config"
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
apis:
  test:
    resources:
      - linter: apitest-resource
        path: testdata/resources
        excludes:
          - testdata/resources/schemas/**
    overlays:
      - template: |-
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
		Version: "1",
		Linters: map[string]*config.Linter{
			"apitest-resource": &config.Linter{
				Name:        "apitest-resource",
				Description: "Test resource rules",
				Spectral: &config.SpectralLinter{
					Rules: []string{
						"resource-rules.yaml",
					},
				},
			},
			"apitest-compiled": &config.Linter{
				Name:        "apitest-compiled",
				Description: "Test compiled rules",
				Spectral: &config.SpectralLinter{
					Rules: []string{
						"compiled-rules.yaml",
					},
				},
			},
		},
		APIs: map[string]*config.API{
			"test": &config.API{
				Name: "test",
				Resources: []*config.ResourceSet{{
					Linter:   "apitest-resource",
					Path:     "testdata/resources",
					Excludes: []string{"testdata/resources/schemas/**"},
				}},
				Overlays: []*config.Overlay{{
					Template: `
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
      - template: |-
          servers:
            - url: ${API_BASE_URL}
              description: Test API
    output:
      path: testdata/output
`)
	proj, err := config.Load(conf)
	c.Assert(err, qt.IsNil)
	c.Assert(proj, qt.DeepEquals, &config.Project{
		Version: "1",
		Linters: map[string]*config.Linter{},
		APIs: map[string]*config.API{
			"test": &config.API{
				Name: "test",
				Resources: []*config.ResourceSet{{
					Path:     "testdata/resources",
					Excludes: []string{"testdata/resources/schemas/**"},
				}},
				Overlays: []*config.Overlay{{
					Template: `
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
	}}
	for i := range tests {
		c.Logf("test#%d: %s", i, tests[i].conf)
		_, err := config.Load(bytes.NewBufferString(tests[i].conf))
		c.Assert(err, qt.ErrorMatches, tests[i].err)
	}
}
