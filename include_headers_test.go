package vervet_test

import (
	"context"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/snyk/vervet/v5"
	"github.com/snyk/vervet/v5/testdata"
)

func TestCommonResponseHeaders(t *testing.T) {
	c := qt.New(t)
	specFile := testdata.Path("resources/_examples/hello-world/2021-06-13/spec.yaml")
	doc, err := vervet.NewDocumentFile(specFile)
	c.Assert(err, qt.IsNil)
	err = doc.Validate(context.TODO())
	c.Assert(err, qt.IsNil)

	// Headers are not included
	pathItem := doc.Paths.Find("/examples/hello-world")
	c.Assert(pathItem, qt.Not(qt.IsNil))
	resp := pathItem.Post.Responses.Status(201).Value
	c.Assert(resp, qt.Not(qt.IsNil))
	// There's a Location header defined outside of the common includes
	c.Assert(resp.Headers, qt.HasLen, 1)

	err = vervet.IncludeHeaders(doc)
	c.Assert(err, qt.IsNil)

	// Included header refs are resolved
	pathItem = doc.Paths.Find("/examples/hello-world")
	c.Assert(pathItem, qt.Not(qt.IsNil))
	resp = pathItem.Post.Responses.Status(201).Value
	c.Assert(resp, qt.Not(qt.IsNil))
	// Now add 3 more common included headers
	c.Assert(resp.Headers, qt.HasLen, 4)
	for _, name := range []string{"snyk-version-requested", "snyk-version-served", "snyk-request-id"} {
		// All of these headers are string type
		c.Assert(resp.Headers[name].Value.Schema.Value.Type, qt.Equals, "string")
	}
}
