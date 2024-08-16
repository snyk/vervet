package vervet

import (
	"strings"

	"github.com/getkin/kin-openapi/openapi3"

	"github.com/snyk/vervet/v8/internal/openapiwalker"
)

// refAliasResolver rewrites references in an OpenAPI document object to local
// references, so that the spec is self-contained.
type refAliasResolver struct {
	doc        *openapi3.T
	refAliases map[string]string
}

func (l *refAliasResolver) ProcessCallbackRef(ref *openapi3.CallbackRef) error {
	ref.Ref = l.resolveRefAlias(ref.Ref)
	return nil
}

func (l *refAliasResolver) ProcessExampleRef(ref *openapi3.ExampleRef) error {
	ref.Ref = l.resolveRefAlias(ref.Ref)
	return nil
}

func (l *refAliasResolver) ProcessHeaderRef(ref *openapi3.HeaderRef) error {
	ref.Ref = l.resolveRefAlias(ref.Ref)
	return nil
}

func (l *refAliasResolver) ProcessLinkRef(ref *openapi3.LinkRef) error {
	ref.Ref = l.resolveRefAlias(ref.Ref)
	return nil
}

func (l *refAliasResolver) ProcessParameterRef(ref *openapi3.ParameterRef) error {
	ref.Ref = l.resolveRefAlias(ref.Ref)
	return nil
}

func (l *refAliasResolver) ProcessRequestBodyRef(ref *openapi3.RequestBodyRef) error {
	ref.Ref = l.resolveRefAlias(ref.Ref)
	return nil
}

func (l *refAliasResolver) ProcessResponseRef(ref *openapi3.ResponseRef) error {
	ref.Ref = l.resolveRefAlias(ref.Ref)
	return nil
}

func (l *refAliasResolver) ProcessSchemaRef(ref *openapi3.SchemaRef) error {
	ref.Ref = l.resolveRefAlias(ref.Ref)
	return nil
}

func (l *refAliasResolver) ProcessSecuritySchemeRef(ref *openapi3.SecuritySchemeRef) error {
	ref.Ref = l.resolveRefAlias(ref.Ref)
	return nil
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
	return openapiwalker.ProcessRefs(l.doc, l)
}
