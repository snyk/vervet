package vervet

import (
	"github.com/getkin/kin-openapi/openapi3"

	"github.com/snyk/vervet/v8/internal/openapiwalker"
)

// Inliner inlines the component.
type Inliner struct {
	refs map[string]struct{}
}

func (in *Inliner) ProcessCallbackRef(ref *openapi3.CallbackRef) error {
	if in.matched(ref.Ref) {
		return RemoveRefs(ref)
	}
	return nil
}

func (in *Inliner) ProcessExampleRef(ref *openapi3.ExampleRef) error {
	if in.matched(ref.Ref) {
		return RemoveRefs(ref)
	}
	return nil
}

func (in *Inliner) ProcessHeaderRef(ref *openapi3.HeaderRef) error {
	if in.matched(ref.Ref) {
		return RemoveRefs(ref)
	}
	return nil
}

func (in *Inliner) ProcessLinkRef(ref *openapi3.LinkRef) error {
	if in.matched(ref.Ref) {
		return RemoveRefs(ref)
	}
	return nil
}

func (in *Inliner) ProcessParameterRef(ref *openapi3.ParameterRef) error {
	if in.matched(ref.Ref) {
		return RemoveRefs(ref)
	}
	return nil
}

func (in *Inliner) ProcessRequestBodyRef(ref *openapi3.RequestBodyRef) error {
	if in.matched(ref.Ref) {
		return RemoveRefs(ref)
	}
	return nil
}

func (in *Inliner) ProcessResponseRef(ref *openapi3.ResponseRef) error {
	if in.matched(ref.Ref) {
		return RemoveRefs(ref)
	}
	return nil
}

func (in *Inliner) ProcessSchemaRef(ref *openapi3.SchemaRef) error {
	if in.matched(ref.Ref) {
		return RemoveRefs(ref)
	}
	return nil
}

func (in *Inliner) ProcessSecuritySchemeRef(ref *openapi3.SecuritySchemeRef) error {
	if in.matched(ref.Ref) {
		return RemoveRefs(ref)
	}
	return nil
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
	if len(in.refs) == 0 {
		return nil
	}

	return openapiwalker.ProcessRefs(doc, in)
}

func (in *Inliner) matched(ref string) bool {
	_, match := in.refs[ref]
	return match
}

// RemoveRefs removes all $ref locations from an OpenAPI document object
// fragment. If the reference has already been resolved, this has the effect of
// "inlining" the formerly referenced object when serializing the OpenAPI
// document.
func RemoveRefs(target interface{}) error {
	return openapiwalker.ProcessRefs(target, clearRefs{})
}

type clearRefs struct {
}

func (c clearRefs) ProcessCallbackRef(ref *openapi3.CallbackRef) error {
	ref.Ref = ""
	return nil
}

func (c clearRefs) ProcessExampleRef(ref *openapi3.ExampleRef) error {
	ref.Ref = ""
	return nil
}

func (c clearRefs) ProcessHeaderRef(ref *openapi3.HeaderRef) error {
	ref.Ref = ""
	return nil
}

func (c clearRefs) ProcessLinkRef(ref *openapi3.LinkRef) error {
	ref.Ref = ""
	return nil
}

func (c clearRefs) ProcessParameterRef(ref *openapi3.ParameterRef) error {
	ref.Ref = ""
	return nil
}

func (c clearRefs) ProcessRequestBodyRef(ref *openapi3.RequestBodyRef) error {
	ref.Ref = ""
	return nil
}

func (c clearRefs) ProcessResponseRef(ref *openapi3.ResponseRef) error {
	ref.Ref = ""
	return nil
}

func (c clearRefs) ProcessSchemaRef(ref *openapi3.SchemaRef) error {
	ref.Ref = ""
	return nil
}

func (c clearRefs) ProcessSecuritySchemeRef(ref *openapi3.SecuritySchemeRef) error {
	ref.Ref = ""
	return nil
}
