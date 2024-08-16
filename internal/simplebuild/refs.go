package simplebuild

import (
	"fmt"
	"reflect"
	"slices"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"

	"github.com/snyk/vervet/v8"
	"github.com/snyk/vervet/v8/internal/openapiwalker"
)

// Refs are an OpenAPI concept where you can define part of a spec then use a
// JSON reference
// [https://datatracker.ietf.org/doc/html/draft-pbryan-zyp-json-ref-03] to
// include that block in another part of the document.
//
// For example a component might live at the top level then that is consumed
// elsewhere:
//
// components:
//
//	parameters:
//	  foo:
//	    name: fooparam
//	    in: query
//
// paths:
//
//	/foo:
//	  parameters:
//	    - $ref: "#/components/parameters/foo"
//
// openapi3 has several *Ref types which have Ref and Value fields, the Ref
// field is the string from the original document and Value is the block it
// points to if the ref is resolved, hen loading our documents we do this
// resolution to populate all Value fields.
//
// When serialising an openapi3.*Ref, if the Ref field is set then the Value
// field is ignored. Therefore we have two options, either add the components
// back into the document at the appropriate position or inline them. As some
// components are likely to be reused several times, we elect to do the former
// where possible.
//
// This class walks a given object and recursively copy any refs it finds back
// into the document at the path they are referenced from.
type refResolver struct {
	doc     *openapi3.T
	renames map[string]string
}

func NewRefResolver() refResolver {
	return refResolver{renames: make(map[string]string)}
}

func (rr *refResolver) deRef(orignalRef string, component any) (string, error) {
	newRef, err := rr.deref(orignalRef, reflect.ValueOf(component))
	if err != nil {
		return "", err
	}
	if newRef != orignalRef {
		rr.renames[newRef] = orignalRef
	}
	return newRef, nil
}

func (rr *refResolver) ProcessCallbackRef(ref *openapi3.CallbackRef) error {
	if ref.Ref == "" {
		return nil
	}
	component := &openapi3.CallbackRef{
		Value: ref.Value,
	}
	var err error
	ref.Ref, err = rr.deRef(ref.Ref, component)
	return err
}

func (rr *refResolver) ProcessExampleRef(ref *openapi3.ExampleRef) error {
	if ref.Ref == "" {
		return nil
	}
	component := &openapi3.ExampleRef{
		Value: ref.Value,
	}
	var err error
	ref.Ref, err = rr.deRef(ref.Ref, component)
	return err
}

func (rr *refResolver) ProcessHeaderRef(ref *openapi3.HeaderRef) error {
	if ref.Ref == "" {
		return nil
	}
	component := &openapi3.HeaderRef{
		Value: ref.Value,
	}
	var err error
	ref.Ref, err = rr.deRef(ref.Ref, component)
	return err
}

func (rr *refResolver) ProcessLinkRef(ref *openapi3.LinkRef) error {
	if ref.Ref == "" {
		return nil
	}
	component := &openapi3.LinkRef{
		Value: ref.Value,
	}
	var err error
	ref.Ref, err = rr.deRef(ref.Ref, component)
	return err
}

func (rr *refResolver) ProcessParameterRef(ref *openapi3.ParameterRef) error {
	if ref.Ref == "" {
		return nil
	}
	component := &openapi3.ParameterRef{
		Value: ref.Value,
	}
	var err error
	ref.Ref, err = rr.deRef(ref.Ref, component)
	return err
}

func (rr *refResolver) ProcessRequestBodyRef(ref *openapi3.RequestBodyRef) error {
	if ref.Ref == "" {
		return nil
	}
	component := &openapi3.RequestBodyRef{
		Value: ref.Value,
	}
	var err error
	ref.Ref, err = rr.deRef(ref.Ref, component)
	return err
}

func (rr *refResolver) ProcessResponseRef(ref *openapi3.ResponseRef) error {
	if ref.Ref == "" {
		return nil
	}
	component := &openapi3.ResponseRef{
		Value: ref.Value,
	}
	var err error
	ref.Ref, err = rr.deRef(ref.Ref, component)
	return err
}

func (rr *refResolver) ProcessSchemaRef(ref *openapi3.SchemaRef) error {
	if ref.Ref == "" {
		return nil
	}
	component := &openapi3.SchemaRef{
		Value: ref.Value,
	}
	var err error
	ref.Ref, err = rr.deRef(ref.Ref, component)
	return err
}

func (rr *refResolver) ProcessSecuritySchemeRef(ref *openapi3.SecuritySchemeRef) error {
	if ref.Ref == "" {
		return nil
	}
	component := &openapi3.SecuritySchemeRef{
		Value: ref.Value,
	}
	var err error
	ref.Ref, err = rr.deRef(ref.Ref, component)
	return err
}

// ResolveRefs recursively finds all ref objects in the current documents paths
// and makes sure they are valid by copying the referenced component to the
// documents components section.
//
// WARNING: this will mutate references so if references are shared between
// documents make sure that any other documents are serialised before resolving
// refs. This method only ensures the current document is correct.
func (rr *refResolver) ResolveRefs(doc *openapi3.T) error {
	// Refs use a full path eg #/components/schemas/..., to avoid having a
	// special case at the top level we pass the entire document and trust the
	// refs to not reference parts of the document they shouldn't.
	rr.doc = doc
	return openapiwalker.ProcessRefs(doc, rr)
}

func (rr *refResolver) deref(ref string, value reflect.Value) (string, error) {
	path := strings.Split(ref, "/")
	if path[0] != "#" {
		// All refs should have been resolved to the local document when
		// loading so if we hit this case then we have not loaded the document
		// correctly.
		return "", fmt.Errorf("external ref %s", ref)
	}

	field := reflect.ValueOf(rr.doc)
	newRef, err := deref(path[1:], field, value, rr.renames)
	if err != nil {
		return "", err
	}
	slices.Reverse(newRef)
	newRefStr := fmt.Sprintf("#/%s", strings.Join(newRef, "/"))
	return newRefStr, nil
}

func deref(path []string, field, value reflect.Value, renames map[string]string) ([]string, error) {
	if len(path) == 0 {
		field.Set(value.Elem())
		return []string{}, nil
	}

	newName := path[0]
	nextField, err := getField(newName, field)
	if err != nil {
		return nil, fmt.Errorf("invalid ref: %w", err)
	}

	// Lookup if we already have a component in the same document with the same
	// name, if they conflict then we need to rename the current component
	if len(path) == 1 {
		// Name might have changed on previous documents but previous
		// collisions are no longer present. Always start from the original
		// name to make sure we aren't leaving unessisary gaps.
		originalName, ok := renames[newName]
		if ok {
			newName = originalName
			nextField, err = getField(newName, field)
			if err != nil {
				return nil, fmt.Errorf("invalid ref: %w", err)
			}
		}
		suffix := 0
		prevName := newName
		// If the component is the same as the one we have already then it
		// isn't a problem, we can merge them.
		for !isZero(nextField) && !vervet.ComponentsEqual(nextField.Interface(), value.Interface()) {
			newName = fmt.Sprintf("%s~%d", prevName, suffix)
			nextField, err = getField(newName, field)
			if err != nil {
				return nil, fmt.Errorf("renaming ref: %w", err)
			}
			suffix += 1
		}
	}

	// If the container for the next layer doesn't exist then we have to create
	// it.
	if isZero(nextField) {
		if field.Kind() == reflect.Map {
			nextField = reflect.New(field.Type().Elem().Elem())
			field.SetMapIndex(reflect.ValueOf(newName), nextField)
		} else {
			var newValue reflect.Value
			if nextField.Kind() == reflect.Map {
				newValue = reflect.MakeMap(nextField.Type())
			} else {
				newValue = reflect.New(nextField.Type().Elem())
			}
			nextField.Set(newValue)
		}
	}
	if field.Kind() == reflect.Map {
		nextField = nextField.Elem()
	}

	newRef, err := deref(path[1:], nextField, value, renames)
	return append(newRef, newName), err
}

func isZero(field reflect.Value) bool {
	if !field.IsValid() {
		return true
	}
	if field.Kind() == reflect.Pointer {
		return field.IsNil()
	}
	return field.IsZero()
}

func getField(tag string, object reflect.Value) (reflect.Value, error) {
	if object.Kind() == reflect.Map {
		fieldName := reflect.ValueOf(tag)
		return object.MapIndex(fieldName), nil
	}

	reflectedObject := object.Type().Elem()
	if reflectedObject.Kind() != reflect.Struct {
		return reflect.Value{}, fmt.Errorf("object is not a struct")
	}
	for idx := 0; idx < reflectedObject.NumField(); idx++ {
		structField := reflectedObject.Field(idx)
		yamlTag := structField.Tag.Get("yaml")
		// Remove tag options (eg "omitempty")
		yamlField := strings.SplitN(yamlTag, ",", 2)[0]
		if yamlField == tag {
			return object.Elem().FieldByName(structField.Name), nil
		}
	}
	return reflect.Value{}, fmt.Errorf("field %s not found on object", tag)
}
