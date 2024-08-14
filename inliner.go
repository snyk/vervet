package vervet

import (
	"reflect"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/mitchellh/reflectwalk"
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
					removeRefsResponseAndChildren(response)
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
		removeHeaderRefsAndChildren(headerRef)
	}
	if headerRef.Value != nil {
		in.checkSchemaRef(headerRef.Value.Schema)
	}
}

func removeHeaderRefsAndChildren(headerRef *openapi3.HeaderRef) {
	headerRef.Ref = ""
	if headerRef.Value != nil {
		removeRefsForSchemaAndChildren(headerRef.Value.Schema)
		removeRefsForExamples(headerRef.Value.Examples)
	}
}

func (in *Inliner) checkParameters(parameters openapi3.Parameters) {
	for _, parameterRef := range parameters {
		in.checkParameterRef(parameterRef)
	}
}

func (in *Inliner) checkParameterRef(parameterRef *openapi3.ParameterRef) {
	if in.matched(parameterRef.Ref) {
		removeParameterRefsAndChildren(parameterRef)
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
			removeRefsForSchemaAndChildren(schema)
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

func (in *Inliner) checkResponseBody(body *openapi3.RequestBodyRef) {
	if body != nil {
		if in.matched(body.Ref) {
			body.Ref = ""
			removeRefsContentAndChildren(body.Value.Content)
		}
		in.checkContent(body.Value.Content)
	}
}

func removeRefsForSchemaAndChildren(schemas ...*openapi3.SchemaRef) {
	if schemas == nil {
		return
	}
	for _, schema := range schemas {
		if schema == nil {
			return
		}
		schema.Ref = ""
		if schema.Value != nil {
			removeRefsForSchemaAndChildren(schemaRefsFromSchemas(schema.Value.Properties)...)
			removeRefsForSchemaAndChildren(schema.Value.Items)
			removeRefsForSchemaAndChildren(schema.Value.AllOf...)
			removeRefsForSchemaAndChildren(schema.Value.AnyOf...)
			removeRefsForSchemaAndChildren(schema.Value.OneOf...)
			removeRefsForSchemaAndChildren(schema.Value.Not)
		}
	}
}

func schemaRefsFromSchemas(properties openapi3.Schemas) []*openapi3.SchemaRef {
	refs := make([]*openapi3.SchemaRef, 0, len(properties))
	for _, ref := range properties {
		refs = append(refs, ref)
	}
	return refs
}

func removeParameterRefsAndChildren(parameter *openapi3.ParameterRef) {
	parameter.Ref = ""
	removeRefsForSchemaAndChildren(parameter.Value.Schema)
	removeRefsForExamples(parameter.Value.Examples)
}

func removeRefsForExamples(examples openapi3.Examples) {
	for _, example := range examples {
		example.Ref = ""
	}
}

func removeRefsResponseAndChildren(response *openapi3.ResponseRef) {
	response.Ref = ""
	removeRefsContentAndChildren(response.Value.Content)
	for _, ref := range response.Value.Headers {
		removeHeaderRefsAndChildren(ref)
	}
}

func removeRefsContentAndChildren(content openapi3.Content) {
	for _, mediaType := range content {
		removeRefsForSchemaAndChildren(mediaType.Schema)
		removeRefsForExamples(mediaType.Examples)
	}
}

// RefRemover removes the ref from the component.
type RefRemover struct {
	target interface{}
}

func NewRefRemover(target interface{}) *RefRemover {
	return &RefRemover{target: target}
}

// RemoveRef removes all $ref locations from an OpenAPI document object
// fragment. If the reference has already been resolved, this has the effect of
// "inlining" the formerly referenced object when serializing the OpenAPI
// document.
func (rr *RefRemover) RemoveRef() error {
	return reflectwalk.Walk(rr.target, rr)
}

// Struct implements reflectwalk.StructWalker.
func (rr *RefRemover) Struct(v reflect.Value) error {
	if !v.CanInterface() {
		return nil
	}
	switch v.Interface().(type) {
	case openapi3.SchemaRef:
		valPointer := v.Addr().Interface().(*openapi3.SchemaRef)
		valPointer.Ref = ""
	case openapi3.ParameterRef:
		valPointer := v.Addr().Interface().(*openapi3.ParameterRef)
		valPointer.Ref = ""
	case openapi3.HeaderRef:
		valPointer := v.Addr().Interface().(*openapi3.HeaderRef)
		valPointer.Ref = ""
	case openapi3.RequestBodyRef:
		valPointer := v.Addr().Interface().(*openapi3.RequestBodyRef)
		valPointer.Ref = ""
	case openapi3.ResponseRef:
		valPointer := v.Addr().Interface().(*openapi3.ResponseRef)
		valPointer.Ref = ""
	case openapi3.SecuritySchemeRef:
		valPointer := v.Addr().Interface().(*openapi3.SecuritySchemeRef)
		valPointer.Ref = ""
	case openapi3.ExampleRef:
		valPointer := v.Addr().Interface().(*openapi3.ExampleRef)
		valPointer.Ref = ""
	case openapi3.LinkRef:
		valPointer := v.Addr().Interface().(*openapi3.LinkRef)
		valPointer.Ref = ""
	case openapi3.CallbackRef:
		valPointer := v.Addr().Interface().(*openapi3.CallbackRef)
		valPointer.Ref = ""
	}

	return nil
}

// StructField implements reflectwalk.StructWalker.
func (rr *RefRemover) StructField(field reflect.StructField, v reflect.Value) error {
	return nil
}
