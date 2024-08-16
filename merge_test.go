package vervet_test

import (
	"testing"

	qt "github.com/frankban/quicktest"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/snyk/vervet/v8"
	"github.com/snyk/vervet/v8/testdata"
)

var openapiCmp = qt.CmpEquals(cmpopts.IgnoreUnexported(
	openapi3.ExampleRef{},
	openapi3.HeaderRef{},
	openapi3.ParameterRef{},
	openapi3.RequestBodyRef{},
	openapi3.ResponseRef{},
	openapi3.Schema{},
	openapi3.SchemaRef{},
	openapi3.SecuritySchemeRef{},
))

func TestMergeComponents(t *testing.T) {
	c := qt.New(t)
	c.Run("component without replace and conflict", func(c *qt.C) {
		src := mustLoadFile(c, "merge_test_conflict.yaml")
		dst := mustLoadFile(c, "merge_test_dst.yaml")
		err := vervet.Merge(dst, src, false)
		c.Assert(err, qt.IsNotNil)
	})
	c.Run("component without replace", func(c *qt.C) {
		src := mustLoadFile(c, "merge_test_src.yaml")
		dstOrig := mustLoadFile(c, "merge_test_dst.yaml")
		dst := mustLoadFile(c, "merge_test_dst.yaml")
		err := vervet.Merge(dst, src, false)
		c.Assert(err, qt.IsNil)

		c.Assert(dst.Components.Schemas["Foo"], openapiCmp, dstOrig.Components.Schemas["Foo"])
		c.Assert(dst.Components.Schemas["Bar"], openapiCmp, src.Components.Schemas["Bar"])
		c.Assert(dst.Components.Schemas["Baz"], openapiCmp, dstOrig.Components.Schemas["Baz"])

		c.Assert(dst.Components.Parameters["Foo"], openapiCmp, dstOrig.Components.Parameters["Foo"])
		c.Assert(dst.Components.Parameters["Bar"], openapiCmp, src.Components.Parameters["Bar"])
		c.Assert(dst.Components.Parameters["Baz"], openapiCmp, dstOrig.Components.Parameters["Baz"])

		c.Assert(dst.Components.Headers["Foo"], openapiCmp, dstOrig.Components.Headers["Foo"])
		c.Assert(dst.Components.Headers["Bar"], openapiCmp, src.Components.Headers["Bar"])
		c.Assert(dst.Components.Headers["Baz"], openapiCmp, dstOrig.Components.Headers["Baz"])

		c.Assert(dst.Components.RequestBodies["Foo"], openapiCmp, dstOrig.Components.RequestBodies["Foo"])
		c.Assert(dst.Components.RequestBodies["Bar"], openapiCmp, src.Components.RequestBodies["Bar"])
		c.Assert(dst.Components.RequestBodies["Baz"], openapiCmp, dstOrig.Components.RequestBodies["Baz"])

		c.Assert(dst.Components.Responses["200"], openapiCmp, dstOrig.Components.Responses["200"])
		c.Assert(dst.Components.Responses["201"], openapiCmp, src.Components.Responses["201"])
		c.Assert(dst.Components.Responses["202"], openapiCmp, dstOrig.Components.Responses["202"])

		c.Assert(dst.Components.SecuritySchemes["Foo"], openapiCmp, dstOrig.Components.SecuritySchemes["Foo"])
		c.Assert(dst.Components.SecuritySchemes["Bar"], openapiCmp, src.Components.SecuritySchemes["Bar"])
		c.Assert(dst.Components.SecuritySchemes["Baz"], openapiCmp, dstOrig.Components.SecuritySchemes["Baz"])

		c.Assert(dst.Components.Examples["Foo"], openapiCmp, dstOrig.Components.Examples["Foo"])
		c.Assert(dst.Components.Examples["Bar"], openapiCmp, src.Components.Examples["Bar"])
		c.Assert(dst.Components.Examples["Baz"], openapiCmp, dstOrig.Components.Examples["Baz"])
	})
	c.Run("component with replace", func(c *qt.C) {
		src := mustLoadFile(c, "merge_test_src.yaml")
		dstOrig := mustLoadFile(c, "merge_test_dst.yaml")
		dst := mustLoadFile(c, "merge_test_dst.yaml")
		err := vervet.Merge(dst, src, true)
		c.Assert(err, qt.IsNil)

		c.Assert(dst.Components.Schemas["Foo"], openapiCmp, src.Components.Schemas["Foo"])
		c.Assert(dst.Components.Schemas["Bar"], openapiCmp, src.Components.Schemas["Bar"])
		c.Assert(dst.Components.Schemas["Baz"], openapiCmp, dstOrig.Components.Schemas["Baz"])

		c.Assert(dst.Components.Parameters["Foo"], openapiCmp, src.Components.Parameters["Foo"])
		c.Assert(dst.Components.Parameters["Bar"], openapiCmp, src.Components.Parameters["Bar"])
		c.Assert(dst.Components.Parameters["Baz"], openapiCmp, dstOrig.Components.Parameters["Baz"])

		c.Assert(dst.Components.Headers["Foo"], openapiCmp, src.Components.Headers["Foo"])
		c.Assert(dst.Components.Headers["Bar"], openapiCmp, src.Components.Headers["Bar"])
		c.Assert(dst.Components.Headers["Baz"], openapiCmp, dstOrig.Components.Headers["Baz"])

		c.Assert(dst.Components.RequestBodies["Foo"], openapiCmp, src.Components.RequestBodies["Foo"])
		c.Assert(dst.Components.RequestBodies["Bar"], openapiCmp, src.Components.RequestBodies["Bar"])
		c.Assert(dst.Components.RequestBodies["Baz"], openapiCmp, dstOrig.Components.RequestBodies["Baz"])

		c.Assert(dst.Components.RequestBodies["200"], openapiCmp, src.Components.RequestBodies["200"])
		c.Assert(dst.Components.RequestBodies["201"], openapiCmp, src.Components.RequestBodies["201"])
		c.Assert(dst.Components.RequestBodies["202"], openapiCmp, dstOrig.Components.RequestBodies["202"])

		c.Assert(dst.Components.SecuritySchemes["Foo"], openapiCmp, src.Components.SecuritySchemes["Foo"])
		c.Assert(dst.Components.SecuritySchemes["Bar"], openapiCmp, src.Components.SecuritySchemes["Bar"])
		c.Assert(dst.Components.SecuritySchemes["Baz"], openapiCmp, dstOrig.Components.SecuritySchemes["Baz"])

		c.Assert(dst.Components.Examples["Foo"], openapiCmp, src.Components.Examples["Foo"])
		c.Assert(dst.Components.Examples["Bar"], openapiCmp, src.Components.Examples["Bar"])
		c.Assert(dst.Components.Examples["Baz"], openapiCmp, dstOrig.Components.Examples["Baz"])
	})
	c.Run("component with missing sections", func(c *qt.C) {
		src := mustLoadFile(c, "merge_test_src.yaml")
		dstOrig := mustLoadFile(c, "merge_test_dst_missing_components.yaml")
		dst := mustLoadFile(c, "merge_test_dst_missing_components.yaml")
		err := vervet.Merge(dst, src, true)
		c.Assert(err, qt.IsNil)

		c.Assert(dst.Components.Schemas["Foo"], openapiCmp, src.Components.Schemas["Foo"])
		c.Assert(dst.Components.Schemas["Bar"], openapiCmp, src.Components.Schemas["Bar"])
		c.Assert(dst.Components.Schemas["Baz"], openapiCmp, dstOrig.Components.Schemas["Baz"])

		c.Assert(dst.Components.Parameters["Foo"], openapiCmp, src.Components.Parameters["Foo"])
		c.Assert(dst.Components.Parameters["Bar"], openapiCmp, src.Components.Parameters["Bar"])
		c.Assert(dst.Components.Parameters["Baz"], openapiCmp, dstOrig.Components.Parameters["Baz"])

		c.Assert(dst.Components.Headers["Foo"], openapiCmp, src.Components.Headers["Foo"])
		c.Assert(dst.Components.Headers["Bar"], openapiCmp, src.Components.Headers["Bar"])
		c.Assert(dst.Components.Headers["Baz"], openapiCmp, dstOrig.Components.Headers["Baz"])

		c.Assert(dst.Components.RequestBodies["Foo"], openapiCmp, src.Components.RequestBodies["Foo"])
		c.Assert(dst.Components.RequestBodies["Bar"], openapiCmp, src.Components.RequestBodies["Bar"])
		c.Assert(dst.Components.RequestBodies["Baz"], openapiCmp, dstOrig.Components.RequestBodies["Baz"])

		c.Assert(dst.Components.RequestBodies["200"], openapiCmp, src.Components.RequestBodies["200"])
		c.Assert(dst.Components.RequestBodies["201"], openapiCmp, src.Components.RequestBodies["201"])
		c.Assert(dst.Components.RequestBodies["202"], openapiCmp, dstOrig.Components.RequestBodies["202"])

		c.Assert(dst.Components.SecuritySchemes["Foo"], openapiCmp, src.Components.SecuritySchemes["Foo"])
		c.Assert(dst.Components.SecuritySchemes["Bar"], openapiCmp, src.Components.SecuritySchemes["Bar"])
		c.Assert(dst.Components.SecuritySchemes["Baz"], openapiCmp, dstOrig.Components.SecuritySchemes["Baz"])

		c.Assert(dst.Components.Examples["Foo"], openapiCmp, src.Components.Examples["Foo"])
		c.Assert(dst.Components.Examples["Bar"], openapiCmp, src.Components.Examples["Bar"])
		c.Assert(dst.Components.Examples["Baz"], openapiCmp, dstOrig.Components.Examples["Baz"])
	})
}

func TestMergeTags(t *testing.T) {
	srcYaml := `
tags:
  - name: foo
    description: foo resource (src)
  - name: bar
    description: bar resource (src)
`
	dstYaml := `
tags:
  - name: foo
    description: foo resource (dst)
  - name: baz
    description: baz resource (dst)
`
	c := qt.New(t)
	c.Run("tags without replace", func(c *qt.C) {
		src := mustLoad(c, srcYaml)
		dst := mustLoad(c, dstYaml)
		err := vervet.Merge(dst, src, false)
		c.Assert(err, qt.IsNil)
		c.Assert(dst.Tags, qt.DeepEquals, openapi3.Tags{{
			Name:        "bar",
			Description: "bar resource (src)",
		}, {
			Name:        "baz",
			Description: "baz resource (dst)",
		}, {
			Name:        "foo",
			Description: "foo resource (dst)",
		}})
	})
	c.Run("tags with replace", func(c *qt.C) {
		src := mustLoad(c, srcYaml)
		dst := mustLoad(c, dstYaml)
		err := vervet.Merge(dst, src, true)
		c.Assert(err, qt.IsNil)
		c.Assert(dst.Tags, qt.DeepEquals, openapi3.Tags{{
			Name:        "bar",
			Description: "bar resource (src)",
		}, {
			Name:        "baz",
			Description: "baz resource (dst)",
		}, {
			Name:        "foo",
			Description: "foo resource (src)",
		}})
	})
}

func TestMergeTopLevel(t *testing.T) {
	srcYaml := `
info:
  title: Src
  version: src
security:
  - Foo: []
  - Bar:
      - read
      - write
servers:
  - url: https://example.com/foo
    description: Foo (src)
  - url: https://example.com/bar
    description: Bar (src)
x-extension:
  key0: value0
  key1: value1
`
	dstYaml := `
info:
  title: Dst
  version: dst
security:
  - Foo:
     - up
     - down
  - Baz:
     - strange
     - crunchy
servers:
  - url: https://example.com/foo
    description: Foo (dst)
  - url: https://example.com/baz
    description: Baz (dst)
x-extension:
  key1: value11
  key2: value2
`
	c := qt.New(t)
	c.Run("without replace", func(c *qt.C) {
		src := mustLoad(c, srcYaml)
		dst := mustLoad(c, dstYaml)
		err := vervet.Merge(dst, src, false)
		c.Assert(err, qt.IsNil)
		c.Assert(dst.Info, qt.DeepEquals, &openapi3.Info{
			Title:   "Dst",
			Version: "dst",
		})
		c.Assert(dst.Security, qt.DeepEquals, openapi3.SecurityRequirements{{
			"Foo": []string{"up", "down"},
		}, {
			"Baz": []string{"strange", "crunchy"},
		}})
		c.Assert(dst.Servers, qt.DeepEquals, openapi3.Servers{{
			URL:         "https://example.com/foo",
			Description: "Foo (dst)",
		}, {
			URL:         "https://example.com/baz",
			Description: "Baz (dst)",
		}})
		c.Assert(dst.Extensions, qt.DeepEquals, map[string]interface{}{
			"x-extension": map[string]interface{}{
				"key1": "value11",
				"key2": "value2",
			},
		})
	})
	c.Run("with replace", func(c *qt.C) {
		src := mustLoad(c, srcYaml)
		dst := mustLoad(c, dstYaml)
		err := vervet.Merge(dst, src, true)
		c.Assert(err, qt.IsNil)
		c.Assert(dst.Info, qt.DeepEquals, &openapi3.Info{
			Title:   "Src",
			Version: "src",
		})
		c.Assert(dst.Security, qt.DeepEquals, openapi3.SecurityRequirements{{
			"Foo": []string{},
		}, {
			"Bar": []string{"read", "write"},
		}})
		c.Assert(dst.Servers, qt.DeepEquals, openapi3.Servers{{
			URL:         "https://example.com/foo",
			Description: "Foo (src)",
		}, {
			URL:         "https://example.com/bar",
			Description: "Bar (src)",
		}})
		c.Assert(dst.Extensions, qt.DeepEquals, map[string]interface{}{
			"x-extension": map[string]interface{}{
				"key0": "value0",
				"key1": "value1",
			},
		})
	})
}

func TestMergeIntoEmpty(t *testing.T) {
	c := qt.New(t)
	srcYaml := `
info:
  title: Src
  version: src
paths:
  /foo:
    get:
    description: get a foo
    responses:
      200:
      contents:
        application/json:
        schema:
          type: object
`
	src := mustLoad(c, srcYaml)
	dst := &openapi3.T{}
	err := vervet.Merge(dst, src, false)
	c.Assert(err, qt.IsNil)
	c.Assert(dst.Paths.Len(), qt.Equals, 1)
}

func mustLoadFile(c *qt.C, path string) *openapi3.T {
	doc, err := vervet.NewDocumentFile(testdata.Path(path))
	c.Assert(err, qt.IsNil)
	return doc.T
}

func mustLoad(c *qt.C, s string) *openapi3.T {
	doc, err := openapi3.NewLoader().LoadFromData([]byte(s))
	c.Assert(err, qt.IsNil)
	return doc
}
