package vervet

import (
	"log"
	"path"
	"reflect"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/mitchellh/reflectwalk"
)

// Localize rewrites all references in an OpenAPI document to local references.
func Localize(doc *openapi3.T) error {
	l := newLocalizer(doc)
	return l.localize()
}

// localizer rewrites references in an OpenAPI document object to local
// references, so that the spec is self-contained.
type localizer struct {
	doc *openapi3.T

	curRefType    reflect.Value
	curRefField   reflect.Value
	curValueField reflect.Value
}

// newLocalizer returns a new localizer.
func newLocalizer(doc *openapi3.T) *localizer {
	return &localizer{doc: doc}
}

// localize rewrites all references in the OpenAPI document to local references.
func (l *localizer) localize() error {
	err := reflectwalk.Walk(l.doc, l)
	if err != nil {
		return err
	}
	// Some of the localized components may have non-localized references,
	// since they were just added to the document object tree in the prior
	// walk. Brute-forcing them into the fold...
	return reflectwalk.Walk(l.doc.Components, l)
}

// Struct implements reflectwalk.StructWalker
func (l *localizer) Struct(v reflect.Value) error {
	l.curRefType, l.curRefField, l.curValueField = v, v.FieldByName("Ref"), v.FieldByName("Value")
	return nil
}

// StructField implements reflectwalk.StructWalker
func (l *localizer) StructField(sf reflect.StructField, v reflect.Value) error {
	if !l.curRefField.IsValid() || !l.curValueField.IsValid() {
		return nil
	}
	refPath := l.curRefField.String()
	if refPath == "" {
		return nil
	}
	// TODO: Resolve unique names from external component refs, URI basename
	// may not be good enough.
	refBase := path.Base(refPath)
	if isLocalRef(refPath) {
		return nil
	}

	switch refObj := l.curRefType.Addr().Interface().(type) {
	case *openapi3.SchemaRef:
		refObj.Ref = "#/components/schemas/" + refBase
		if l.doc.Components.Schemas == nil {
			l.doc.Components.Schemas = map[string]*openapi3.SchemaRef{}
		}
		if l.doc.Components.Schemas[refBase] == nil {
			l.doc.Components.Schemas[refBase] = &openapi3.SchemaRef{Value: refObj.Value}
		}
	case *openapi3.ParameterRef:
		refObj.Ref = "#/components/parameters/" + refBase
		if l.doc.Components.Parameters == nil {
			l.doc.Components.Parameters = map[string]*openapi3.ParameterRef{}
		}
		if l.doc.Components.Parameters[refBase] == nil {
			l.doc.Components.Parameters[refBase] = &openapi3.ParameterRef{Value: refObj.Value}
		}
	case *openapi3.LinkRef:
		refObj.Ref = "#/components/links/" + refBase
		if l.doc.Components.Links == nil {
			l.doc.Components.Links = map[string]*openapi3.LinkRef{}
		}
		if l.doc.Components.Links[refBase] == nil {
			l.doc.Components.Links[refBase] = &openapi3.LinkRef{Value: refObj.Value}
		}
	case *openapi3.RequestBodyRef:
		refObj.Ref = "#/components/requests/" + refBase
		if l.doc.Components.RequestBodies == nil {
			l.doc.Components.RequestBodies = map[string]*openapi3.RequestBodyRef{}
		}
		if l.doc.Components.RequestBodies[refBase] == nil {
			l.doc.Components.RequestBodies[refBase] = &openapi3.RequestBodyRef{Value: refObj.Value}
		}
	case *openapi3.ResponseRef:
		refObj.Ref = "#/components/responses/" + refBase
		if l.doc.Components.Responses == nil {
			l.doc.Components.Responses = map[string]*openapi3.ResponseRef{}
		}
		if l.doc.Components.Responses[refBase] == nil {
			l.doc.Components.Responses[refBase] = &openapi3.ResponseRef{Value: refObj.Value}
		}
	case *openapi3.HeaderRef:
		refObj.Ref = "#/components/headers/" + refBase
		if l.doc.Components.Headers == nil {
			l.doc.Components.Headers = map[string]*openapi3.HeaderRef{}
		}
		if l.doc.Components.Headers[refBase] == nil {
			l.doc.Components.Headers[refBase] = &openapi3.HeaderRef{Value: refObj.Value}
		}
	default:
		log.Printf("warning, unsupported ref type %T", refObj)
	}
	return nil
}

// isLocalRef returns whether the reference is localized.
func isLocalRef(s string) bool {
	return strings.HasPrefix(s, "#/components/")
}
