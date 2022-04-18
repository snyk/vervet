package generate_test

import (
	"os"
	"testing"
	"testing/fstest"

	qt "github.com/frankban/quicktest"

	"github.com/snyk/vervet/v4/generate"
	"github.com/snyk/vervet/v4/testdata"
)

func TestGenerateFS(t *testing.T) {
	c := qt.New(t)
	out := c.TempDir()

	fs := fstest.MapFS{
		"generator.yaml": &fstest.MapFile{
			Data: []byte(`
version-readme:
  scope: version
  filename: "` + out + `/{{.API}}/{{.Resource}}/{{.Version.DateString}}/README"
  template: "/README.tmpl"
`),
			Mode: 0666,
		},
		"README.tmpl": &fstest.MapFile{
			Data: []byte(`
This is a generated scaffold for version {{ .Version.String }} of the
{{ .Resource }} resource in API {{ .API }}.
`),
			Mode: 0666,
		},
	}

	params := generate.GeneratorParams{
		ProjectDir:     testdata.Path("."),
		ConfigFile:     testdata.Path(".vervet.yaml"),
		GeneratorsFile: "/generator.yaml",
		Generators:     []string{"version-readme"},
		FS:             fs,
	}
	err := generate.Generate(params)
	c.Assert(err, qt.IsNil)

	contents, err := os.ReadFile(out + "/testdata/hello-world/2021-06-01/README")
	c.Assert(err, qt.IsNil)
	c.Assert(string(contents), qt.Equals, `
This is a generated scaffold for version 2021-06-01~experimental of the
hello-world resource in API testdata.
`)
}
