package vervet

import (
	"fmt"
	"github.com/getkin/kin-openapi/openapi3"
)

// Inliner inlines the component.
type Inliner struct {
	refs map[string]struct{}
}

// NewInliner returns a new Inliner instance.
func NewInliner() *Inliner {
	return &Inliner{refs: map[string]struct{}{}}
}

// AddRef adds a JSON Reference URI to the set of references to be inlined.
func (in *Inliner) AddRef(ref string) {
	in.refs[ref] = struct{}{}
}

// Inline inlines all the JSON References previously indicated with AddRef in
// the given OpenAPI document.
func (in *Inliner) Inline(doc *openapi3.T) error {
	if len(in.refs) == 0 {
		return nil
	}

	for _, pathItem := range doc.Paths.Map() {
		in.checkParameters(pathItem.Parameters)

		for _, operation := range pathItem.Operations() {
			in.checkParameters(operation.Parameters)
			in.checkResponseBody(operation.RequestBody)
			for _, response := range operation.Responses.Map() {
				if in.matched(response.Ref) {
					RemoveRefs(response)
				}
				if response.Value != nil {
					for _, headerRef := range response.Value.Headers {
						in.checkHeaderRefMatch(headerRef)
					}
					in.checkContent(response.Value.Content)
				}
			}
		}
	}
	return nil
}

func (in *Inliner) matched(ref string) bool {
	_, match := in.refs[ref]
	return match
}

func (in *Inliner) removeIfMatched(ref string) string {
	if in.matched(ref) {
		return ""
	}
	return ref
}

func (in *Inliner) checkContent(content openapi3.Content) {
	for _, mediaType := range content {
		in.checkSchemaRef(mediaType.Schema)
		for _, example := range mediaType.Examples {
			example.Ref = in.removeIfMatched(example.Ref)
		}
	}
}

func (in *Inliner) checkHeaderRefMatch(headerRef *openapi3.HeaderRef) {
	if in.matched(headerRef.Ref) {
		RemoveRefs(headerRef)
	}
	if headerRef.Value != nil {
		in.checkSchemaRef(headerRef.Value.Schema)
	}
}

func (in *Inliner) checkParameters(parameters openapi3.Parameters) {
	for _, parameterRef := range parameters {
		in.checkParameterRef(parameterRef)
	}
}

func (in *Inliner) checkParameterRef(parameterRef *openapi3.ParameterRef) {
	if in.matched(parameterRef.Ref) {
		RemoveRefs(parameterRef)
	}
	in.checkSchemaRef(parameterRef.Value.Schema)
}

func (in *Inliner) checkSchemaRef(schemas ...*openapi3.SchemaRef) {
	if schemas == nil {
		return
	}
	for _, schema := range schemas {
		if schema == nil {
			return
		}
		if in.matched(schema.Ref) {
			RemoveRefs(schema)
		}
		if schema.Value != nil {
			in.checkSchemaRef(schemaRefsFromSchemas(schema.Value.Properties)...)
			in.checkSchemaRef(schema.Value.Items)
			in.checkSchemaRef(schema.Value.AllOf...)
			in.checkSchemaRef(schema.Value.AnyOf...)
			in.checkSchemaRef(schema.Value.OneOf...)
			in.checkSchemaRef(schema.Value.Not)
		}
	}
}

func schemaRefsFromSchemas(properties openapi3.Schemas) []*openapi3.SchemaRef {
	refs := []*openapi3.SchemaRef{}
	for _, ref := range properties {
		refs = append(refs, ref)
	}
	return refs
}

func (in *Inliner) checkResponseBody(body *openapi3.RequestBodyRef) {
	if body != nil {
		if in.matched(body.Ref) {
			body.Ref = ""
			RemoveRefs(body.Value.Content)
		}
		in.checkContent(body.Value.Content)
	}
}

// RemoveRefs removes all $ref locations from an OpenAPI document object
// fragment. If the reference has already been resolved, this has the effect of
// "inlining" the formerly referenced object when serializing the OpenAPI
// document.
func RemoveRefs(target interface{}) {
	switch v := target.(type) {
	case nil:
		return
	case openapi3.Content:
		for _, mediaType := range v {
			RemoveRefs(mediaType.Schema)
			RemoveRefs(mediaType.Examples)
		}

	case openapi3.Schemas:
		for _, schema := range v {
			RemoveRefs(schema)
		}
	case openapi3.SchemaRefs:
		for _, schema := range v {
			RemoveRefs(schema)
		}
	case openapi3.Headers:
		for _, header := range v {
			RemoveRefs(header)
		}
	case openapi3.Parameter:
		RemoveRefs(v.Schema)

	case openapi3.Examples:
		for _, example := range v {
			RemoveRefs(example)
		}

	case *openapi3.ExampleRef:
		v.Ref = ""

	case *openapi3.ParameterRef:
		if v == nil {
			return
		}
		v.Ref = ""
		if v.Value != nil {
			RemoveRefs(v.Value.Schema)
			RemoveRefs(v.Value.Content)
			RemoveRefs(v.Value.Examples)
		}
	case *openapi3.ResponseRef:
		if v == nil {
			return
		}
		v.Ref = ""
		if v.Value != nil {
			RemoveRefs(v.Value.Content)
			RemoveRefs(v.Value.Headers)
		}
	case *openapi3.SchemaRef:
		if v == nil {
			return
		}
		v.Ref = ""
		if v.Value != nil {
			RemoveRefs(v.Value.Properties)
			RemoveRefs(v.Value.Items)
			RemoveRefs(v.Value.AllOf)
			RemoveRefs(v.Value.AnyOf)
			RemoveRefs(v.Value.OneOf)
			RemoveRefs(v.Value.Not)
		}
	case *openapi3.HeaderRef:
		if v == nil {
			return
		}
		v.Ref = ""
		if v.Value != nil {
			RemoveRefs(v.Value.Parameter)
		}
	default:
		//intentional panic, have covered all the types in kin-openapi v0.127.0
		//might fail in the future if new types are added, should be caught in tests
		panic(fmt.Sprintf("unhandled type %v", target))
	}
}
