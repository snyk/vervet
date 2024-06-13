package vervet_test

import (
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/snyk/vervet/v7"
	"github.com/snyk/vervet/v7/testdata"
)

func TestRefIndexSource(t *testing.T) {
	c := qt.New(t)
	doc, err := vervet.NewDocumentFile(testdata.Path("resources/_examples/hello-world/2021-06-01/spec.yaml"))
	c.Assert(err, qt.IsNil)
	ix, err := vervet.NewRefIndex(doc.T)
	c.Assert(err, qt.IsNil)

	c.Assert(ix.HasRef("#/components/schemas/HelloWorld"), qt.IsTrue)
	c.Assert(ix.HasRef("#/components/headers/DeprecationHeader"), qt.IsFalse)
	c.Assert(ix.HasRef("#/components/schemas/Nope"), qt.IsFalse)
}

func TestRefIndexCompiled(t *testing.T) {
	c := qt.New(t)
	doc, err := vervet.NewDocumentFile(testdata.Path("output/2021-06-01~experimental/spec.json"))
	c.Assert(err, qt.IsNil)
	ix, err := vervet.NewRefIndex(doc.T)
	c.Assert(err, qt.IsNil)

	c.Assert(ix.HasRef("#/components/schemas/HelloWorld"), qt.IsTrue)
	c.Assert(ix.HasRef("#/components/headers/DeprecationHeader"), qt.IsTrue)
	c.Assert(ix.HasRef("#/components/schemas/Nope"), qt.IsFalse)
}
