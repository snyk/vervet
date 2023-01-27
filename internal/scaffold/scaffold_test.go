package scaffold_test

import (
	"os"
	"path/filepath"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/snyk/vervet/v5/internal/scaffold"
	"github.com/snyk/vervet/v5/testdata"
)

func TestScaffold(t *testing.T) {
	c := qt.New(t)
	dstDir := c.TempDir()
	s, err := scaffold.New(dstDir, testdata.Path("test-scaffold"))
	c.Assert(err, qt.IsNil)
	err = s.Organize()
	c.Assert(err, qt.IsNil)
	readmeTmpl, err := os.ReadFile(filepath.Join(dstDir, ".vervet", "templates", "README.tmpl"))
	c.Assert(err, qt.IsNil)
	c.Assert(string(readmeTmpl), qt.Equals, `
This is a generated scaffold for version {{ .Version }}~{{ .Stability }} of the
{{ .Resource }} resource in API {{ .API }}.

`[1:])
}
