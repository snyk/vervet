package config_test

import (
	"bytes"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/snyk/vervet/v5/config"
)

func TestLoad(t *testing.T) {
	c := qt.New(t)
	conf := bytes.NewBufferString(`
version: "1"
apis:
  test:
    resources:
      - path: testdata/resources
        excludes:
          - testdata/resources/schemas/**
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
		APIs: config.APIs{
			"test": {
				Name: "test",
				Resources: []*config.ResourceSet{{
					Path:     "testdata/resources",
					Excludes: []string{"testdata/resources/schemas/**"},
				}, {
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
    output:
      path: /somewhere/else
      paths:
        - /another/place
        - /and/another
`[1:],
		err: `output should specify one of 'path' or 'paths', not both \(apis\.testapi\.output\)`,
	}, {
		err: `no apis defined`,
	}}
	for i := range tests {
		c.Logf("test#%d: %s", i, tests[i].conf)
		_, err := config.Load(bytes.NewBufferString(tests[i].conf))
		c.Assert(err, qt.ErrorMatches, tests[i].err)
	}
}
