package generator

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
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
	for _, prefix := range []string{"", "{{.Cwd}}/"} {
		c.Run(fmt.Sprintf("template prefix %q", prefix), func(c *qt.C) {
			setup(c)

			configBuf, err := ioutil.ReadFile(".vervet.yaml")
			c.Assert(err, qt.IsNil)
			proj, err := config.Load(bytes.NewBuffer(configBuf))
			c.Assert(err, qt.IsNil)

			out := c.TempDir()

			versionReadme := `
version-readme:
  scope: version
  filename: "{{.Here}}/{{.API}}/{{.Resource}}/{{.Version.DateString}}/README"
  template: "` + prefix + `.vervet/resource/version/README.tmpl"
`
			os.Setenv("VERVET_TEMPLATE_test-value", "bad wolf")
			t.Cleanup(func() {
				os.Unsetenv("VERVET_TEMPLATE_test-value")
			})
			generatorsConf, err := config.LoadGenerators(bytes.NewBufferString(versionReadme))
			c.Assert(err, qt.IsNil)

			genReadme, err := New(generatorsConf["version-readme"], Debug(true), Here(out))
			c.Assert(err, qt.IsNil)

			resources, err := MapResources(proj)
			c.Assert(err, qt.IsNil)
			files, err := genReadme.Execute(resources)
			c.Assert(err, qt.IsNil)
			c.Assert(files, qt.ContentEquals, []string{
				out + "/testdata/hello-world/2021-06-01/README",
				out + "/testdata/hello-world/2021-06-07/README",
				out + "/testdata/hello-world/2021-06-13/README",
				out + "/testdata/projects/2021-06-04/README",
				out + "/testdata/projects/2021-08-20/README",
			})

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
					test.resource+" resource in API testdata.\n\n"+
					"An environment test value of bad wolf has been provided\n"+
					"for this scaffold.")
			}
		})
	}
}

func TestResourceScope(t *testing.T) {
	c := qt.New(t)
	for _, prefix := range []string{"", "{{.Cwd}}/"} {
		c.Run(fmt.Sprintf("template prefix %q", prefix), func(c *qt.C) {
			setup(c)

			configBuf, err := ioutil.ReadFile(".vervet.yaml")
			c.Assert(err, qt.IsNil)
			proj, err := config.Load(bytes.NewBuffer(configBuf))
			c.Assert(err, qt.IsNil)

			out := c.TempDir()

			versionReadme := `
resource-routes:
  scope: resource
  filename: "{{.Here}}/{{ .API }}/{{ .Resource }}/routes.ts"
  template: "` + prefix + `.vervet/resource/routes.ts.tmpl"
`
			os.Setenv("VERVET_TEMPLATE_test-value", "bad wolf")
			t.Cleanup(func() {
				os.Unsetenv("VERVET_TEMPLATE_test-value")
			})
			generatorsConf, err := config.LoadGenerators(bytes.NewBufferString(versionReadme))
			c.Assert(err, qt.IsNil)

			genReadme, err := New(generatorsConf["resource-routes"], Debug(true), Here(out))
			c.Assert(err, qt.IsNil)

			resources, err := MapResources(proj)
			c.Assert(err, qt.IsNil)
			files, err := genReadme.Execute(resources)
			c.Assert(err, qt.IsNil)
			c.Assert(files, qt.ContentEquals, []string{
				out + "/testdata/hello-world/routes.ts",
				out + "/testdata/projects/routes.ts",
			})

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
// An environment test value of bad wolf has been provided
// for this scaffold.`[1:])

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
// An environment test value of bad wolf has been provided
// for this scaffold.`[1:])
		})
	}
}

func TestFunctions(t *testing.T) {
	c := qt.New(t)
	setup(c)

	configBuf, err := ioutil.ReadFile(".vervet.yaml")
	c.Assert(err, qt.IsNil)
	proj, err := config.Load(bytes.NewBuffer(configBuf))
	c.Assert(err, qt.IsNil)

	out := c.TempDir()
	c.Assert(ioutil.WriteFile(out+"/tsfuncs.js", []byte(`
function tsType(oasType) {
	switch (String(oasType)) {
	case "string":
		return "string";
	case "integer":
		return "number";
	case "boolean":
		return "boolean";
	}
	console.log("warning: failed to resolve type of", oasType);
	return "any";
}
`[1:]), 0666), qt.IsNil)
	c.Assert(ioutil.WriteFile(out+"/models.ts.tmpl", []byte(`
{{- /*

	Template interfaceProperties produces the contents of an interface.

*/ -}}
{{- define "interfaceProperties" -}}
{{- range $propName, $prop := .Properties }}
{{ $propName }}: {{ template "schemaTypeDef" $prop.Value }},
{{- end -}}
{{- end -}}

{{- /*

	Template resolveSchemaRef either resolves a local component ref to a
	generated type name, or emits an inline type declaration.

*/ -}}
{{- define "resolveSchemaRef" -}}
{{- if .Ref }}{{ .Ref | basename }}
{{- else }}{{ template "schemaTypeDef" .Value }}
{{- end -}}
{{- end -}}

{{- /*

	Template schemaTypeDef produces the definition of a type.

*/ -}}
{{- define "schemaTypeDef" -}}

{{- if isOneOf . }}
{{- range $i, $oneOf := .OneOf }} {{ if ne $i 0 }}| {{ end }}{{ template "resolveSchemaRef" $oneOf }}{{ end }}

{{- else if isAnyOf . }}
{{- range $i, $anyOf := .AnyOf }} {{ if ne $i 0 }}| {{ end }}{{ template "resolveSchemaRef" $anyOf }}{{ end }}

{{- else if isAllOf . }}
{{- range $i, $allOf := .AllOf }} {{ if ne $i 0 }}| {{ end }}{{ template "resolveSchemaRef" $allOf }}{{ end }}

{{- else if isAssociativeArray . }}{
  [key: string]: object;
}

{{- else if eq .Type "object" }}{
{{- include "interfaceProperties" . | indent 2 }}
}

{{- else if eq .Type "array" }}Array<{{ template "resolveSchemaRef" .Items }}>

{{- else }}{{ .Type | tsType }}

{{- end -}}
{{- end -}}

{{- /*

	Template schemaTypeDecl produces a complete Typescript type declaration.
	It might be an interface, union, intersection or alias.

*/ -}}
{{- define "schemaTypeDecl" -}}

{{- if isAssociativeArray .Schema.Value }}
export interface {{ .Name }} {{ template "schemaTypeDef" .Schema.Value }};

{{- else if eq .Schema.Value.Type "object" }}
export interface {{ .Name }} {{ template "schemaTypeDef" .Schema.Value }};

{{- else }}
export type {{ .Name }} = {{ template "schemaTypeDef" .Schema.Value }};

{{- end -}}
{{- end -}}

{{- /*

	Top-level template.

*/ -}}
{{ range $schemaName, $schema := .ResourceVersion.Document.Components.Schemas -}}
{{- if $schema.Value -}}
{{ with $ctx := map "Name" $schemaName "Schema" $schema }}{{ template "schemaTypeDecl" $ctx }}{{ end }}
{{ end -}}
{{ end -}}
`[1:]), 0666), qt.IsNil)

	versionReadme := `
version-models:
  scope: version
  filename: "{{.Here}}/{{.API}}/{{.Resource}}/{{.Version.DateString}}/models.ts"
  functions: "{{.Here}}/tsfuncs.js"
  template: "{{.Here}}/models.ts.tmpl"
`
	generatorsConf, err := config.LoadGenerators(bytes.NewBufferString(versionReadme))
	c.Assert(err, qt.IsNil)

	genReadme, err := New(generatorsConf["version-models"], Debug(true), Here(out))
	c.Assert(err, qt.IsNil)

	resources, err := MapResources(proj)
	c.Assert(err, qt.IsNil)
	files, err := genReadme.Execute(resources)
	c.Assert(err, qt.IsNil)
	c.Assert(files, qt.ContentEquals, []string{
		out + "/testdata/hello-world/2021-06-01/models.ts",
		out + "/testdata/hello-world/2021-06-07/models.ts",
		out + "/testdata/hello-world/2021-06-13/models.ts",
		out + "/testdata/projects/2021-06-04/models.ts",
		out + "/testdata/projects/2021-08-20/models.ts",
	})

	jsFile, err := ioutil.ReadFile(out + "/testdata/projects/2021-06-04/models.ts")
	c.Assert(err, qt.IsNil)
	c.Assert(string(jsFile), qt.Equals, `
export type ActualVersion = string;

export interface Error {
  detail: string,
  id: string,
  meta: {
    [key: string]: object;
  },
  source: {
    parameter: string,
    pointer: string,
  },
  status: string,
};

export interface ErrorDocument {
  errors: Array<Error>,
  jsonapi: {
    version: string,
  },
};

export interface JsonApi {
  version: string,
};

export type LinkProperty =  string | {
  href: string,
  meta: {
    [key: string]: object;
  },
};

export interface Links {
  first:  string | {
    href: string,
    meta: {
      [key: string]: object;
    },
  },
  last:  string | {
    href: string,
    meta: {
      [key: string]: object;
    },
  },
  next:  string | {
    href: string,
    meta: {
      [key: string]: object;
    },
  },
  prev:  string | {
    href: string,
    meta: {
      [key: string]: object;
    },
  },
  related:  string | {
    href: string,
    meta: {
      [key: string]: object;
    },
  },
  self:  string | {
    href: string,
    meta: {
      [key: string]: object;
    },
  },
};

export interface Meta {
  [key: string]: object;
};

export interface Project {
  attributes: {
    created: string,
    hostname: string,
    name: string,
    origin: string,
    status: string,
    type: string,
  },
  id: string,
  type: string,
};

export type QueryVersion = string;
`)
}

func TestDryRun(t *testing.T) {
	c := qt.New(t)
	for _, prefix := range []string{"", "{{.Cwd}}/"} {
		c.Run(fmt.Sprintf("template prefix %q", prefix), func(c *qt.C) {
			setup(c)

			configBuf, err := ioutil.ReadFile(".vervet.yaml")
			c.Assert(err, qt.IsNil)
			proj, err := config.Load(bytes.NewBuffer(configBuf))
			c.Assert(err, qt.IsNil)

			out := c.TempDir()

			versionReadme := `
version-readme:
  scope: version
  filename: "{{.Here}}/{{.API}}/{{.Resource}}/{{.Version.DateString}}/README"
  template: "` + prefix + `.vervet/resource/version/README.tmpl"
`
			generatorsConf, err := config.LoadGenerators(bytes.NewBufferString(versionReadme))
			c.Assert(err, qt.IsNil)

			genReadme, err := New(generatorsConf["version-readme"], Debug(true), Here(out), DryRun(true))
			c.Assert(err, qt.IsNil)

			resources, err := MapResources(proj)
			c.Assert(err, qt.IsNil)
			files, err := genReadme.Execute(resources)
			c.Assert(err, qt.IsNil)
			c.Assert(files, qt.ContentEquals, []string{
				out + "/testdata/hello-world/2021-06-01/README",
				out + "/testdata/hello-world/2021-06-07/README",
				out + "/testdata/hello-world/2021-06-13/README",
				out + "/testdata/projects/2021-06-04/README",
				out + "/testdata/projects/2021-08-20/README",
			})

			actualFiles, err := filepath.Glob(out + "/*/*/*/README")
			c.Assert(err, qt.IsNil)
			c.Assert(actualFiles, qt.HasLen, 0)
		})
	}
}
