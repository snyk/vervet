package scaffold_test

import (
	"io/ioutil"
	"path/filepath"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/snyk/vervet/internal/scaffold"
	"github.com/snyk/vervet/testdata"
)

func TestScaffold(t *testing.T) {
	c := qt.New(t)
	dstDir := c.Mkdir()
	s, err := scaffold.New(dstDir, testdata.Path("test-scaffold"))
	c.Assert(err, qt.IsNil)
	err = s.Organize()
	c.Assert(err, qt.IsNil)
	readmeTmpl, err := ioutil.ReadFile(filepath.Join(dstDir, ".vervet", "templates", "README.tmpl"))
	c.Assert(err, qt.IsNil)
	c.Assert(string(readmeTmpl), qt.Equals, `
This is a generated scaffold for version {{ .Version }}~{{ .Stability }} of the
{{ .Resource }} resource in API {{ .API }}.

`[1:])
}
