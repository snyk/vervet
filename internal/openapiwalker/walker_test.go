package openapiwalker

import (
	"encoding/json"
	"testing"

	qt "github.com/frankban/quicktest"
	"github.com/getkin/kin-openapi/openapi3"
)

func TestProcessRefs(t *testing.T) {
	c := qt.New(t)

	c.Run("processes example refs in order", func(c *qt.C) {
		doc := &openapi3.T{}
		doc.Components = &openapi3.Components{
			Examples: openapi3.Examples{
				"key-3": {
					Ref: "value-3",
				},
				"key-1": {
					Ref: "value-1",
				},
				"key-2": {
					Ref: "value-2",
				},
			},
		}

		recorder := &recordRefProcessor{}
		err := ProcessRefs(doc, recorder)
		c.Assert(err, qt.IsNil)
		c.Assert(toJson(c, recorder.ExampleRefs), qt.JSONEquals, []*openapi3.ExampleRef{
			{
				Ref: "value-1",
			},
			{
				Ref: "value-2",
			},
			{
				Ref: "value-3",
			},
		})
	})
	c.Run("processes callback refs in order", func(c *qt.C) {
		doc := &openapi3.T{}
		doc.Components = &openapi3.Components{
			Callbacks: openapi3.Callbacks{
				"key-3": {
					Ref: "value-3",
				},
				"key-1": {
					Ref: "value-1",
				},
				"key-2": {
					Ref: "value-2",
				},
			},
		}

		recorder := &recordRefProcessor{}
		err := ProcessRefs(doc, recorder)
		c.Assert(err, qt.IsNil)
		c.Assert(toJson(c, recorder.CallbackRefs), qt.JSONEquals, []*openapi3.CallbackRef{
			{
				Ref: "value-1",
			},
			{
				Ref: "value-2",
			},
			{
				Ref: "value-3",
			},
		})
	})
	c.Run("processes header refs in order", func(c *qt.C) {
		doc := &openapi3.T{}
		doc.Components = &openapi3.Components{
			Headers: openapi3.Headers{
				"key-3": {
					Ref: "value-3",
				},
				"key-1": {
					Ref: "value-1",
				},
				"key-2": {
					Ref: "value-2",
				},
			},
		}

		recorder := &recordRefProcessor{}
		err := ProcessRefs(doc, recorder)
		c.Assert(err, qt.IsNil)
		c.Assert(toJson(c, recorder.HeaderRefs), qt.JSONEquals, []*openapi3.HeaderRef{
			{
				Ref: "value-1",
			},
			{
				Ref: "value-2",
			},
			{
				Ref: "value-3",
			},
		})
	})
	c.Run("processes parameter refs in order", func(c *qt.C) {
		doc := &openapi3.T{}
		doc.Components = &openapi3.Components{
			Parameters: openapi3.ParametersMap{
				"key-3": {
					Ref: "value-3",
				},
				"key-1": {
					Ref: "value-1",
				},
				"key-2": {
					Ref: "value-2",
				},
			},
		}

		recorder := &recordRefProcessor{}
		err := ProcessRefs(doc, recorder)
		c.Assert(err, qt.IsNil)
		c.Assert(toJson(c, recorder.ParameterRefs), qt.JSONEquals, []*openapi3.ParameterRef{
			{
				Ref: "value-1",
			},
			{
				Ref: "value-2",
			},
			{
				Ref: "value-3",
			},
		})
	})
	c.Run("processes request bodies in order", func(c *qt.C) {
		doc := &openapi3.T{}
		doc.Components = &openapi3.Components{
			RequestBodies: openapi3.RequestBodies{
				"key-3": {
					Ref: "value-3",
				},
				"key-1": {
					Ref: "value-1",
				},
				"key-2": {
					Ref: "value-2",
				},
			},
		}

		recorder := &recordRefProcessor{}
		err := ProcessRefs(doc, recorder)
		c.Assert(err, qt.IsNil)
		c.Assert(toJson(c, recorder.RequestBodyRefs), qt.JSONEquals, []*openapi3.RequestBodyRef{
			{
				Ref: "value-1",
			},
			{
				Ref: "value-2",
			},
			{
				Ref: "value-3",
			},
		})
	})
	c.Run("processes response bodies in order", func(c *qt.C) {
		doc := &openapi3.T{}
		doc.Components = &openapi3.Components{
			Responses: openapi3.ResponseBodies{
				"key-3": {
					Ref: "value-3",
				},
				"key-1": {
					Ref: "value-1",
				},
				"key-2": {
					Ref: "value-2",
				},
			},
		}

		recorder := &recordRefProcessor{}
		err := ProcessRefs(doc, recorder)
		c.Assert(err, qt.IsNil)
		c.Assert(toJson(c, recorder.ResponseRefs), qt.JSONEquals, []*openapi3.ResponseRef{
			{
				Ref: "value-1",
			},
			{
				Ref: "value-2",
			},
			{
				Ref: "value-3",
			},
		})
	})
	c.Run("processes schema refs in order", func(c *qt.C) {
		doc := &openapi3.T{}
		doc.Components = &openapi3.Components{
			Schemas: openapi3.Schemas{
				"key-3": {
					Ref: "value-3",
				},
				"key-1": {
					Ref: "value-1",
				},
				"key-2": {
					Ref: "value-2",
				},
			},
		}

		recorder := &recordRefProcessor{}
		err := ProcessRefs(doc, recorder)
		c.Assert(err, qt.IsNil)
		c.Assert(toJson(c, recorder.SchemaRefs), qt.JSONEquals, []*openapi3.SchemaRef{
			{
				Ref: "value-1",
			},
			{
				Ref: "value-2",
			},
			{
				Ref: "value-3",
			},
		})
	})
	c.Run("processes security schema refs in order", func(c *qt.C) {
		doc := &openapi3.T{}
		doc.Components = &openapi3.Components{
			SecuritySchemes: openapi3.SecuritySchemes{
				"key-3": {
					Ref: "value-3",
				},
				"key-1": {
					Ref: "value-1",
				},
				"key-2": {
					Ref: "value-2",
				},
			},
		}

		recorder := &recordRefProcessor{}
		err := ProcessRefs(doc, recorder)
		c.Assert(err, qt.IsNil)
		c.Assert(toJson(c, recorder.SecuritySchemeRefs), qt.JSONEquals, []*openapi3.SecuritySchemeRef{
			{
				Ref: "value-1",
			},
			{
				Ref: "value-2",
			},
			{
				Ref: "value-3",
			},
		})
	})
	c.Run("processes link refs in order", func(c *qt.C) {
		doc := &openapi3.T{}
		doc.Components = &openapi3.Components{
			Links: openapi3.Links{
				"key-3": {
					Ref: "value-3",
				},
				"key-1": {
					Ref: "value-1",
				},
				"key-2": {
					Ref: "value-2",
				},
			},
		}

		recorder := &recordRefProcessor{}
		err := ProcessRefs(doc, recorder)
		c.Assert(err, qt.IsNil)
		c.Assert(toJson(c, recorder.LinkRefs), qt.JSONEquals, []*openapi3.LinkRef{
			{
				Ref: "value-1",
			},
			{
				Ref: "value-2",
			},
			{
				Ref: "value-3",
			},
		})
	})
}

func toJson(c *qt.C, obj any) []byte {
	jsonOutput, err := json.Marshal(obj)
	c.Assert(err, qt.IsNil)
	return jsonOutput
}

type recordRefProcessor struct {
	CallbackRefs       []*openapi3.CallbackRef
	ExampleRefs        []*openapi3.ExampleRef
	HeaderRefs         []*openapi3.HeaderRef
	LinkRefs           []*openapi3.LinkRef
	ParameterRefs      []*openapi3.ParameterRef
	RequestBodyRefs    []*openapi3.RequestBodyRef
	ResponseRefs       []*openapi3.ResponseRef
	SchemaRefs         []*openapi3.SchemaRef
	SecuritySchemeRefs []*openapi3.SecuritySchemeRef
}

func (r *recordRefProcessor) ProcessCallbackRef(ref *openapi3.CallbackRef) error {
	r.CallbackRefs = append(r.CallbackRefs, ref)
	return nil
}

func (r *recordRefProcessor) ProcessExampleRef(ref *openapi3.ExampleRef) error {
	r.ExampleRefs = append(r.ExampleRefs, ref)
	return nil
}

func (r *recordRefProcessor) ProcessHeaderRef(ref *openapi3.HeaderRef) error {
	r.HeaderRefs = append(r.HeaderRefs, ref)
	return nil
}

func (r *recordRefProcessor) ProcessLinkRef(ref *openapi3.LinkRef) error {
	r.LinkRefs = append(r.LinkRefs, ref)
	return nil
}

func (r *recordRefProcessor) ProcessParameterRef(ref *openapi3.ParameterRef) error {
	r.ParameterRefs = append(r.ParameterRefs, ref)
	return nil
}

func (r *recordRefProcessor) ProcessRequestBodyRef(ref *openapi3.RequestBodyRef) error {
	r.RequestBodyRefs = append(r.RequestBodyRefs, ref)
	return nil
}

func (r *recordRefProcessor) ProcessResponseRef(ref *openapi3.ResponseRef) error {
	r.ResponseRefs = append(r.ResponseRefs, ref)
	return nil
}

func (r *recordRefProcessor) ProcessSchemaRef(ref *openapi3.SchemaRef) error {
	r.SchemaRefs = append(r.SchemaRefs, ref)
	return nil
}

func (r *recordRefProcessor) ProcessSecuritySchemeRef(ref *openapi3.SecuritySchemeRef) error {
	r.SecuritySchemeRefs = append(r.SecuritySchemeRefs, ref)
	return nil
}
