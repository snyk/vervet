package vervet_test

import (
	"context"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/snyk/vervet/v8"
	"github.com/snyk/vervet/v8/testdata"
)

func TestNewDocumentFile(t *testing.T) {
	c := qt.New(t)
	doc, err := vervet.NewDocumentFile(testdata.Path("resources/_examples/hello-world/2021-06-01/spec.yaml"))
	c.Assert(err, qt.IsNil)
	c.Assert(doc.Paths.Len(), qt.Equals, 1)
	c.Assert(doc.Paths.Value("/examples/hello-world/{id}"), qt.Not(qt.IsNil))
	c.Assert(doc.Components.Schemas["HelloWorld"], qt.Not(qt.IsNil))
	c.Assert(doc.Validate(context.TODO()), qt.IsNil)
}

func TestDocumentVersion(t *testing.T) {
	c := qt.New(t)
	doc, err := vervet.NewDocumentFile(testdata.Path("resources/_examples/hello-world/2021-06-01/spec.yaml"))
	c.Assert(err, qt.IsNil)

	version, err := doc.Version()
	c.Assert(err, qt.IsNil)
	c.Assert(version, qt.Equals, vervet.MustParseVersion("2021-06-01"))
}
