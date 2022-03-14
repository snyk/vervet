package vervet

import (
	"encoding/json"
	"reflect"
	"strings"

	"github.com/getkin/kin-openapi/jsoninfo"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/mitchellh/reflectwalk"
)

// refAliasResolver rewrites references in an OpenAPI document object to local
// references, so that the spec is self-contained.
type refAliasResolver struct {
	doc         *openapi3.T
	refAliases  map[string]string
	curRefType  reflect.Value
	curRefField reflect.Value
}

// newRefAliasResolver returns a new refAliasResolver.
func newRefAliasResolver(doc *openapi3.T) (*refAliasResolver, error) {
	refAliases := map[string]string{}
	for refAlias, extValue := range doc.Components.ExtensionProps.Extensions {
		contents, ok := extValue.(json.RawMessage)
		if !ok {
			continue
		}
		dec, err := jsoninfo.NewObjectDecoder(contents)
		if err != nil {
			return nil, err
		}
		var ref openapi3.Ref
		if err := doc.Components.ExtensionProps.DecodeWith(dec, &ref); err == nil && ref.Ref != "" {
			refAliases["#/components/"+refAlias] = ref.Ref
		}
	}

	return &refAliasResolver{doc: doc, refAliases: refAliases}, nil
}

func (r *refAliasResolver) resolveRefAlias(ref string) string {
	if ref != "" && ref[0] == '#' {
		for refAlias, refTarget := range r.refAliases {
			if strings.HasPrefix(ref, refAlias) {
				return strings.Replace(ref, refAlias, refTarget+"#", 1)
			}
		}
	}
	return ref
}

// resolve rewrites all references in the OpenAPI document to local references.
func (r *refAliasResolver) resolve() error {
	return reflectwalk.Walk(r.doc, r)
}

// Struct implements reflectwalk.StructWalker
func (r *refAliasResolver) Struct(v reflect.Value) error {
	r.curRefType, r.curRefField = v, v.FieldByName("Ref")
	return nil
}

// StructField implements reflectwalk.StructWalker
func (r *refAliasResolver) StructField(sf reflect.StructField, v reflect.Value) error {
	if !r.curRefField.IsValid() {
		return nil
	}
	ref := r.curRefField.String()
	if ref == "" {
		return nil
	}
	ref = r.resolveRefAlias(ref)
	r.curRefField.Set(reflect.ValueOf(ref))
	return nil
}
