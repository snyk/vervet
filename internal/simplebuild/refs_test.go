package simplebuild_test

import (
	"testing"

	qt "github.com/frankban/quicktest"
	"github.com/getkin/kin-openapi/openapi3"

	"github.com/snyk/vervet/v8/internal/simplebuild"
)

func TestResolveRefs(t *testing.T) {
	c := qt.New(t)

	c.Run("copies ref value into referenced location", func(c *qt.C) {
		param := &openapi3.Parameter{}
		path := openapi3.PathItem{
			Parameters: []*openapi3.ParameterRef{{
				Ref:   "#/components/parameters/foo",
				Value: param,
			}},
		}
		doc := openapi3.T{
			Paths: openapi3.NewPaths(openapi3.WithPath("/foo", &path)),
		}

		rr := simplebuild.NewRefResolver(&doc)
		err := rr.Resolve(path)
		c.Assert(err, qt.IsNil)

		c.Assert(doc.Components.Parameters["foo"].Value, qt.Equals, param)
	})

	c.Run("ignores refs on other parts of the doc", func(c *qt.C) {
		param := &openapi3.Parameter{}
		pathA := openapi3.PathItem{
			Parameters: []*openapi3.ParameterRef{{
				Ref:   "#/components/parameters/foo",
				Value: param,
			}},
		}
		pathB := openapi3.PathItem{
			Parameters: []*openapi3.ParameterRef{{
				Ref:   "#/components/parameters/bar",
				Value: param,
			}},
		}
		doc := openapi3.T{
			Paths: openapi3.NewPaths(openapi3.WithPath("/foo", &pathA), openapi3.WithPath("/bar", &pathB)),
		}

		rr := simplebuild.NewRefResolver(&doc)
		err := rr.Resolve(pathA)
		c.Assert(err, qt.IsNil)

		c.Assert(doc.Components.Parameters["bar"], qt.IsNil)
	})

	c.Run("merges refs from successive calls", func(c *qt.C) {
		paramA := &openapi3.Parameter{}
		pathA := openapi3.PathItem{
			Parameters: []*openapi3.ParameterRef{{
				Ref:   "#/components/parameters/foo",
				Value: paramA,
			}},
		}
		paramB := &openapi3.Parameter{}
		pathB := openapi3.PathItem{
			Parameters: []*openapi3.ParameterRef{{
				Ref:   "#/components/parameters/bar",
				Value: paramB,
			}},
		}
		doc := openapi3.T{

			Paths: openapi3.NewPaths(openapi3.WithPath("/foo", &pathA), openapi3.WithPath("/bar", &pathB)),
		}

		rr := simplebuild.NewRefResolver(&doc)
		err := rr.Resolve(pathA)
		c.Assert(err, qt.IsNil)
		err = rr.Resolve(pathB)
		c.Assert(err, qt.IsNil)

		c.Assert(doc.Components.Parameters["foo"].Value, qt.Equals, paramA)
		c.Assert(doc.Components.Parameters["bar"].Value, qt.Equals, paramB)
	})

	c.Run("recursively resolves components", func(c *qt.C) {
		schema := &openapi3.Schema{}
		param := &openapi3.Parameter{
			Schema: &openapi3.SchemaRef{
				Ref:   "#/components/schemas/foo",
				Value: schema,
			},
		}
		path := openapi3.PathItem{
			Parameters: []*openapi3.ParameterRef{{
				Ref:   "#/components/parameters/foo",
				Value: param,
			}},
		}
		doc := openapi3.T{
			Paths: openapi3.NewPaths(openapi3.WithPath("/foo", &path)),
		}

		rr := simplebuild.NewRefResolver(&doc)
		err := rr.Resolve(path)
		c.Assert(err, qt.IsNil)

		c.Assert(doc.Components.Parameters["foo"].Value, qt.Equals, param)
		c.Assert(doc.Components.Schemas["foo"].Value, qt.Equals, schema)
	})

	c.Run("ignores ref objects with no ref value", func(c *qt.C) {
		param := &openapi3.Parameter{}
		path := openapi3.PathItem{
			Parameters: []*openapi3.ParameterRef{{
				Value: param,
			}},
		}
		doc := openapi3.T{
			Components: &openapi3.Components{},
			Paths:      openapi3.NewPaths(openapi3.WithPath("/foo", &path)),
		}

		rr := simplebuild.NewRefResolver(&doc)
		err := rr.Resolve(path)
		c.Assert(err, qt.IsNil)

		c.Assert(doc.Components.Parameters["foo"], qt.IsNil)
	})
}
