package vervet

import (
	"reflect"
	"strings"

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
func newRefAliasResolver(doc *openapi3.T) *refAliasResolver {
	res := &refAliasResolver{doc: doc}
	if doc.Components == nil {
		return res
	}

	res.refAliases = make(map[string]string, len(doc.Components.Extensions))

	for refAlias, extValue := range doc.Components.Extensions {
		contents, ok := extValue.(map[string]interface{})
		if !ok {
			continue
		}
		ref, ok := contents["$ref"].(string)
		if ok && ref != "" {
			res.refAliases["#/components/"+refAlias] = ref
		}
	}

	return res
}

func (l *refAliasResolver) resolveRefAlias(ref string) string {
	if ref != "" && ref[0] == '#' {
		for refAlias, refTarget := range l.refAliases {
			if strings.HasPrefix(ref, refAlias) {
				return strings.Replace(ref, refAlias, refTarget+"#", 1)
			}
		}
	}
	return ref
}

// resolve rewrites all references in the OpenAPI document to local references.
func (l *refAliasResolver) resolve() error {
	return reflectwalk.Walk(l.doc, l)
}

// Struct implements reflectwalk.StructWalker.
func (l *refAliasResolver) Struct(v reflect.Value) error {
	l.curRefType, l.curRefField = v, v.FieldByName("Ref")
	return nil
}

// StructField implements reflectwalk.StructWalker.
func (l *refAliasResolver) StructField(sf reflect.StructField, v reflect.Value) error {
	if !l.curRefField.IsValid() {
		return nil
	}
	ref := l.curRefField.String()
	if ref == "" {
		return nil
	}
	ref = l.resolveRefAlias(ref)
	l.curRefField.Set(reflect.ValueOf(ref))
	return nil
}
