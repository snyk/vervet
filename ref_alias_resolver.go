package vervet

import (
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

// refAliasResolver rewrites references in an OpenAPI document object to local
// references, so that the spec is self-contained.
type refAliasResolver struct {
	doc        *openapi3.T
	refAliases map[string]string
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
func (l *refAliasResolver) resolve() {
	for _, path := range l.doc.Paths.InMatchingOrder() {
		for _, operation := range l.doc.Paths.Value(path).Operations() {
			for _, parameter := range operation.Parameters {
				parameter.Ref = l.resolveRefAlias(parameter.Ref)
			}
			if operation.RequestBody != nil {
				operation.RequestBody.Ref = l.resolveRefAlias(operation.RequestBody.Ref)
			}
			for _, response := range operation.Responses.Map() {
				response.Ref = l.resolveRefAlias(response.Ref)
				if response.Value != nil {
					for _, mediaType := range response.Value.Content {
						mediaType.Schema.Ref = l.resolveRefAlias(mediaType.Schema.Ref)
						for _, properties := range mediaType.Schema.Value.Properties {
							properties.Ref = l.resolveRefAlias(properties.Ref)
						}
					}
				}
			}
		}
	}
}
