package simplebuild

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/mitchellh/reflectwalk"
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
	derefed := reflect.New(v.Type())
	reflect.Indirect(derefed).FieldByName("Value").Set(value)

	return rr.deref(refLoc, derefed)
}

// Implements reflectwalk.StructWalker. We work on whole structs so there is
// nothing to do here.
func (rr *refResolver) StructField(sf reflect.StructField, v reflect.Value) error {
	return nil
}

func (rr *refResolver) deref(ref string, value reflect.Value) error {
	path := strings.Split(ref, "/")
	if path[0] != "#" {
		// All refs should have been resolved to the local document when
		// loading so if we hit this case then we have not loaded the document
		// correctly.
		return fmt.Errorf("external ref %s", ref)
	}

	field := reflect.ValueOf(rr.doc)
	return deref(path[1:], field, value)
}

func deref(path []string, field, value reflect.Value) error {
	if len(path) == 0 {
		field.Set(value.Elem())
		return nil
	}
	if len(path) == 1 {
		fmt.Println("setting last value", path[0])
		//return trySetField(path[0], field, value)
	}

	// Maps are a special case since the key also needs to be created.
	if field.Kind() == reflect.Map {
		newValue := reflect.New(field.Type().Elem().Elem())
		fieldName := reflect.ValueOf(path[0])
		oldVal := field.MapIndex(fieldName)
		fmt.Println("setting map key", oldVal, oldVal.IsValid())
		if oldVal.IsValid() {
			// Value already exists
			return deref(path[1:], oldVal.Elem(), value)
		}
		field.SetMapIndex(fieldName, newValue)
		return deref(path[1:], newValue.Elem(), value)
	}
	// else we assume we are working on a struct
	field, err := getField(path[0], field)
	if err != nil {
		return fmt.Errorf("invalid ref: %w", err)
	}

	// A lot of the openapi3.T fields are pointers so if this is the first
	// time we have encountered an object of this type we need to create
	// the container.
	if field.Kind() == reflect.Map && field.IsZero() {
		newValue := reflect.MakeMap(field.Type())
		field.Set(newValue)
	} else if field.IsNil() {
		newValue := reflect.New(field.Type().Elem())
		field.Set(newValue)
	}

	return deref(path[1:], field, value)
}

/*
 *func trySetField(name string, field, value reflect.Value) error {
 *    return nil
 *}
 */

func getField(tag string, object reflect.Value) (reflect.Value, error) {
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
