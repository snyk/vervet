package simplebuild

import (
	"fmt"
	"reflect"
	"slices"
	"strings"
	"unicode"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/mitchellh/reflectwalk"

	"github.com/snyk/vervet/v7"
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
	doc *openapi3.T
}

func NewRefResolver(doc *openapi3.T) refResolver {
	return refResolver{doc: doc}
}

func (rr *refResolver) Resolve(from any) error {
	return reflectwalk.Walk(from, rr)
}

// Implements reflectwalk.StructWalker. This function is called for every
// struct found when walking.
func (rr *refResolver) Struct(v reflect.Value) error {
	ref := v.FieldByName("Ref")
	value := v.FieldByName("Value")
	if !ref.IsValid() || !value.IsValid() {
		// This isn't a openapi3.*Ref so nothing to do
		return nil
	}
	refLoc := ref.String()
	if refLoc == "" {
		// This ref has been inlined
		return nil
	}
	// Create a new *Ref object to avoid mutating the original
	component := reflect.New(v.Type())
	reflect.Indirect(component).FieldByName("Value").Set(value)
	newRef, err := rr.deref(refLoc, component)
	if err != nil {
		return err
	}

	if newRef != refLoc {
		// TODO: other documents have references to this same object which we
		// are mutating. Any previously generated document is now wrong.
		ref.Set(reflect.ValueOf(newRef))
	}

	return nil
}

// Implements reflectwalk.StructWalker. We work on whole structs so there is
// nothing to do here.
func (rr *refResolver) StructField(sf reflect.StructField, v reflect.Value) error {
	return nil
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
	newRef, err := deref(path[1:], field, value)
	if err != nil {
		return "", err
	}
	slices.Reverse(newRef)
	newRefStr := fmt.Sprintf("#/%s", strings.Join(newRef, "/"))
	return newRefStr, nil
}

func deref(path []string, field, value reflect.Value) ([]string, error) {
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
		// collisions are no longer present. Always start from 0 to make sure
		// we aren't leaving unessisary gaps. Some components are already
		// numbers, eg "400" responses, in which case assume they don't
		// conflict. TODO: fix that
		if !unicode.IsDigit(rune(newName[0])) {
			newName = strings.TrimRightFunc(newName, unicode.IsDigit)
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
			newName = fmt.Sprintf("%s%d", prevName, suffix)
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

	newRef, err := deref(path[1:], nextField, value)
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
