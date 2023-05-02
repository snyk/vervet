package vervet

import (
	"reflect"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/mitchellh/reflectwalk"
)

// RefIndex indexes the distinct references used in an OpenAPI document.
type RefIndex struct {
	refs map[string]struct{}
}

// NewRefIndex returns a new reference index on an OpenAPI document.
func NewRefIndex(doc *openapi3.T) (*RefIndex, error) {
	ix := &RefIndex{refs: map[string]struct{}{}}
	if err := ix.index(doc); err != nil {
		return nil, err
	}
	return ix, nil
}

func (ix *RefIndex) index(doc *openapi3.T) error {
	return reflectwalk.Walk(doc, ix)
}

// HasRef returns whether the indexed document contains the given ref.
func (ix *RefIndex) HasRef(ref string) bool {
	_, ok := ix.refs[ref]
	return ok
}

// Struct implements reflectwalk.StructWalker.
func (ix *RefIndex) Struct(v reflect.Value) error {
	if !v.CanInterface() {
		return nil
	}

	switch val := v.Addr().Interface().(type) {
	case *openapi3.SchemaRef:
		ix.refs[val.Ref] = struct{}{}
	case *openapi3.ParameterRef:
		ix.refs[val.Ref] = struct{}{}
	case *openapi3.HeaderRef:
		ix.refs[val.Ref] = struct{}{}
	case *openapi3.RequestBodyRef:
		ix.refs[val.Ref] = struct{}{}
	case *openapi3.ResponseRef:
		ix.refs[val.Ref] = struct{}{}
	case *openapi3.SecuritySchemeRef:
		ix.refs[val.Ref] = struct{}{}
	case *openapi3.ExampleRef:
		ix.refs[val.Ref] = struct{}{}
	case *openapi3.LinkRef:
		ix.refs[val.Ref] = struct{}{}
	case *openapi3.CallbackRef:
		ix.refs[val.Ref] = struct{}{}
	}
	return nil
}

// StructField implements reflectwalk.StructWalker.
func (*RefIndex) StructField(field reflect.StructField, v reflect.Value) error {
	return nil
}
