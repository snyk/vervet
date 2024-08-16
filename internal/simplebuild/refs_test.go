package simplebuild_test

import (
	"fmt"
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

		rr := simplebuild.NewRefResolver()
		err := rr.ResolveRefs(&doc)
		c.Assert(err, qt.IsNil)

		c.Assert(doc.Components.Parameters["foo"].Value, qt.Equals, param)
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

		rr := simplebuild.NewRefResolver()
		err := rr.ResolveRefs(&doc)
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

		rr := simplebuild.NewRefResolver()
		err := rr.ResolveRefs(&doc)
		c.Assert(err, qt.IsNil)

		c.Assert(doc.Components.Parameters["foo"], qt.IsNil)
	})

	c.Run("conflicting components get renamed", func(c *qt.C) {
		paramA := &openapi3.Parameter{
			Name: "fooname",
		}
		pathA := openapi3.PathItem{
			Parameters: []*openapi3.ParameterRef{{
				Ref:   "#/components/parameters/fooo",
				Value: paramA,
			}},
		}
		paramB := &openapi3.Parameter{
			Name: "barname",
		}
		pathB := openapi3.PathItem{
			Parameters: []*openapi3.ParameterRef{{
				Ref:   "#/components/parameters/fooo",
				Value: paramB,
			}},
		}
		doc := openapi3.T{
			Paths: openapi3.NewPaths(openapi3.WithPath("/foo", &pathA), openapi3.WithPath("/bar", &pathB)),
		}

		rr := simplebuild.NewRefResolver()
		err := rr.ResolveRefs(&doc)
		c.Assert(err, qt.IsNil)

		c.Assert(doc.Paths.Value("/foo").Parameters[0].Ref, qt.Not(qt.Equals), doc.Paths.Value("/bar").Parameters[0].Ref)
		c.Assert(doc.Components.Parameters, qt.HasLen, 2)
	})

	c.Run("comparable components get merged", func(c *qt.C) {
		paramA := &openapi3.Parameter{
			Name: "fooname",
		}
		pathA := openapi3.PathItem{
			Parameters: []*openapi3.ParameterRef{{
				Ref:   "#/components/parameters/fooo",
				Value: paramA,
			}},
		}
		paramB := &openapi3.Parameter{
			Name: "fooname",
		}
		pathB := openapi3.PathItem{
			Parameters: []*openapi3.ParameterRef{{
				Ref:   "#/components/parameters/fooo",
				Value: paramB,
			}},
		}
		doc := openapi3.T{
			Paths: openapi3.NewPaths(openapi3.WithPath("/foo", &pathA), openapi3.WithPath("/bar", &pathB)),
		}

		rr := simplebuild.NewRefResolver()
		err := rr.ResolveRefs(&doc)
		c.Assert(err, qt.IsNil)

		out, _ := doc.MarshalJSON()
		fmt.Println()
		fmt.Println(string(out))
		fmt.Println()

		c.Assert(doc.Paths.Value("/foo").Parameters[0].Ref, qt.Equals, doc.Paths.Value("/bar").Parameters[0].Ref)
		c.Assert(doc.Components.Parameters, qt.HasLen, 1)
	})
}
