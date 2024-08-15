package vervet

import (
	"github.com/getkin/kin-openapi/openapi3"

	"github.com/snyk/vervet/v7/internal/openapiwalker"
)

// Inliner inlines the component.
type Inliner struct {
	refs map[string]struct{}
}

func (in *Inliner) ProcessCallbackRef(ref *openapi3.CallbackRef) {
	if in.matched(ref.Ref) {
		RemoveRefs(ref)
	}
}

func (in *Inliner) ProcessExampleRef(ref *openapi3.ExampleRef) {
	if in.matched(ref.Ref) {
		RemoveRefs(ref)
	}
}

func (in *Inliner) ProcessHeaderRef(ref *openapi3.HeaderRef) {
	if in.matched(ref.Ref) {
		RemoveRefs(ref)
	}
}

func (in *Inliner) ProcessLinkRef(ref *openapi3.LinkRef) {
	if in.matched(ref.Ref) {
		RemoveRefs(ref)
	}
}

func (in *Inliner) ProcessParameterRef(ref *openapi3.ParameterRef) {
	if in.matched(ref.Ref) {
		RemoveRefs(ref)
	}
}

func (in *Inliner) ProcessRequestBodyRef(ref *openapi3.RequestBodyRef) {
	if in.matched(ref.Ref) {
		RemoveRefs(ref)
	}
}

func (in *Inliner) ProcessResponseRef(ref *openapi3.ResponseRef) {
	if in.matched(ref.Ref) {
		RemoveRefs(ref)
	}
}

func (in *Inliner) ProcessSchemaRef(ref *openapi3.SchemaRef) {
	if in.matched(ref.Ref) {
		RemoveRefs(ref)
	}
}

func (in *Inliner) ProcessSecuritySchemeRef(ref *openapi3.SecuritySchemeRef) {
	if in.matched(ref.Ref) {
		RemoveRefs(ref)
	}
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
	openapiwalker.ProcessRefs(doc, in)

	return nil
}

func (in *Inliner) matched(ref string) bool {
	_, match := in.refs[ref]
	return match
}

// RemoveRefs removes all $ref locations from an OpenAPI document object
// fragment. If the reference has already been resolved, this has the effect of
// "inlining" the formerly referenced object when serializing the OpenAPI
// document.
func RemoveRefs(target interface{}) {
	openapiwalker.ProcessRefs(target, clearRefs{})
}

type clearRefs struct {
}

func (c clearRefs) ProcessCallbackRef(ref *openapi3.CallbackRef) {
	ref.Ref = ""
}

func (c clearRefs) ProcessExampleRef(ref *openapi3.ExampleRef) {
	ref.Ref = ""
}

func (c clearRefs) ProcessHeaderRef(ref *openapi3.HeaderRef) {
	ref.Ref = ""
}

func (c clearRefs) ProcessLinkRef(ref *openapi3.LinkRef) {
	ref.Ref = ""
}

func (c clearRefs) ProcessParameterRef(ref *openapi3.ParameterRef) {
	ref.Ref = ""
}

func (c clearRefs) ProcessRequestBodyRef(ref *openapi3.RequestBodyRef) {
	ref.Ref = ""
}

func (c clearRefs) ProcessResponseRef(ref *openapi3.ResponseRef) {
	ref.Ref = ""
}

func (c clearRefs) ProcessSchemaRef(ref *openapi3.SchemaRef) {
	ref.Ref = ""
}

func (c clearRefs) ProcessSecuritySchemeRef(ref *openapi3.SecuritySchemeRef) {
	ref.Ref = ""
}
