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
	for _, path := range doc.Paths.InMatchingOrder() {
		for _, operation := range doc.Paths.Value(path).Operations() {
			for _, parameter := range operation.Parameters {
				parameter.Ref = in.removeIfMatched(parameter.Ref)
				in.checkSchemaRef(parameter.Value.Schema)
				for _, exampleRef := range parameter.Value.Examples {
					exampleRef.Ref = in.removeIfMatched(exampleRef.Ref)
				}
			}
			if operation.RequestBody != nil {
				operation.RequestBody.Ref = in.removeIfMatched(operation.RequestBody.Ref)
			}
			for _, response := range operation.Responses.Map() {
				response.Ref = in.removeIfMatched(response.Ref)
				if response.Value != nil {
					for _, mediaType := range response.Value.Content {
						in.checkSchemaRef(mediaType.Schema)
						for _, example := range mediaType.Examples {
							example.Ref = in.removeIfMatched(example.Ref)
						}
					}
				}
			}
		}

	}
	return nil
}

func (in *Inliner) checkSchemaRef(schema *openapi3.SchemaRef) {
	schema.Ref = in.removeIfMatched(schema.Ref)
	for _, properties := range schema.Value.Properties {
		properties.Ref = in.removeIfMatched(properties.Ref)
	}
}
func (in *Inliner) removeIfMatched(ref string) string {
	if _, match := in.refs[ref]; match {
		return ""
	}
	return ref
}

// Struct implements reflectwalk.StructWalker.
func (in *Inliner) Struct(v reflect.Value) error {
	if !v.CanInterface() {
		return nil
	}
	switch val := v.Interface().(type) {
	case openapi3.SchemaRef:
		if _, ok := in.refs[val.Ref]; ok {
			valPointer := v.Addr().Interface().(*openapi3.SchemaRef)
			refRemover := NewRefRemover(valPointer)
			err := refRemover.RemoveRef()
			if err != nil {
				return err
			}
		}
	case openapi3.ParameterRef:
		if _, ok := in.refs[val.Ref]; ok {
			valPointer := v.Addr().Interface().(*openapi3.ParameterRef)
			refRemover := NewRefRemover(valPointer)
			err := refRemover.RemoveRef()
			if err != nil {
				return err
			}
		}
	case openapi3.HeaderRef:
		if _, ok := in.refs[val.Ref]; ok {
			valPointer := v.Addr().Interface().(*openapi3.HeaderRef)
			refRemover := NewRefRemover(valPointer)
			err := refRemover.RemoveRef()
			if err != nil {
				return err
			}
		}
	case openapi3.RequestBodyRef:
		if _, ok := in.refs[val.Ref]; ok {
			valPointer := v.Addr().Interface().(*openapi3.RequestBodyRef)
			refRemover := NewRefRemover(valPointer)
			err := refRemover.RemoveRef()
			if err != nil {
				return err
			}
		}
	case openapi3.ResponseRef:
		if _, ok := in.refs[val.Ref]; ok {
			valPointer := v.Addr().Interface().(*openapi3.ResponseRef)
			refRemover := NewRefRemover(valPointer)
			err := refRemover.RemoveRef()
			if err != nil {
				return err
			}
		}
	case openapi3.SecuritySchemeRef:
		if _, ok := in.refs[val.Ref]; ok {
			valPointer := v.Addr().Interface().(*openapi3.SecuritySchemeRef)
			refRemover := NewRefRemover(valPointer)
			err := refRemover.RemoveRef()
			if err != nil {
				return err
			}
		}
	case openapi3.ExampleRef:
		if _, ok := in.refs[val.Ref]; ok {
			valPointer := v.Addr().Interface().(*openapi3.ExampleRef)
			refRemover := NewRefRemover(valPointer)
			err := refRemover.RemoveRef()
			if err != nil {
				return err
			}
		}
	case openapi3.LinkRef:
		if _, ok := in.refs[val.Ref]; ok {
			valPointer := v.Addr().Interface().(*openapi3.LinkRef)
			refRemover := NewRefRemover(valPointer)
			err := refRemover.RemoveRef()
			if err != nil {
				return err
			}
		}
	case openapi3.CallbackRef:
		if _, ok := in.refs[val.Ref]; ok {
			valPointer := v.Addr().Interface().(*openapi3.CallbackRef)
			refRemover := NewRefRemover(valPointer)
			err := refRemover.RemoveRef()
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// StructField implements reflectwalk.StructWalker.
func (in *Inliner) StructField(field reflect.StructField, v reflect.Value) error {
	return nil
}

// RefRemover removes the ref from the component.
type RefRemover struct {
	target interface{}
}

// NewRefRemover returns a new RefRemover instance.
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
