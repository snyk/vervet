package vervet

import (
	"fmt"

	"github.com/getkin/kin-openapi/openapi3"

	"github.com/snyk/vervet/v8/internal/openapiwalker"
)

type refTracker struct {
	refs map[string]bool
}

func (r refTracker) ProcessCallbackRef(ref *openapi3.CallbackRef) error {
	if ref.Ref == "" {
		return nil
	}
	r.refs[ref.Ref] = true
	return nil
}

func (r refTracker) ProcessExampleRef(ref *openapi3.ExampleRef) error {
	if ref.Ref == "" {
		return nil
	}
	r.refs[ref.Ref] = true
	return nil
}

func (r refTracker) ProcessHeaderRef(ref *openapi3.HeaderRef) error {
	if ref.Ref == "" {
		return nil
	}
	r.refs[ref.Ref] = true
	return nil
}

func (r refTracker) ProcessLinkRef(ref *openapi3.LinkRef) error {
	if ref.Ref == "" {
		return nil
	}
	r.refs[ref.Ref] = true
	return nil
}

func (r refTracker) ProcessParameterRef(ref *openapi3.ParameterRef) error {
	if ref.Ref == "" {
		return nil
	}
	r.refs[ref.Ref] = true
	return nil
}

func (r refTracker) ProcessRequestBodyRef(ref *openapi3.RequestBodyRef) error {
	if ref.Ref == "" {
		return nil
	}
	r.refs[ref.Ref] = true
	return nil
}

func (r refTracker) ProcessResponseRef(ref *openapi3.ResponseRef) error {
	if ref.Ref == "" {
		return nil
	}
	r.refs[ref.Ref] = true
	return nil
}

func (r refTracker) ProcessSchemaRef(ref *openapi3.SchemaRef) error {
	if ref.Ref == "" {
		return nil
	}
	r.refs[ref.Ref] = true
	return nil
}

func (r refTracker) ProcessSecuritySchemeRef(ref *openapi3.SecuritySchemeRef) error {
	if ref.Ref == "" {
		return nil
	}
	r.refs[ref.Ref] = true
	return nil
}

// RemoveUnusedRefs walks the entire document and deletes any components that
// are not referenced anywhere.
func RemoveUnusedRefs(doc *openapi3.T) error {
	if doc.Components == nil {
		return nil
	}

	tracker := refTracker{
		refs: make(map[string]bool),
	}
	err := openapiwalker.ProcessRefs(doc, tracker)
	if err != nil {
		return err
	}

	for componentName := range doc.Components.Schemas {
		ref := fmt.Sprintf("#/components/schemas/%s", componentName)
		if !tracker.refs[ref] {
			fmt.Println("schema", componentName, "not in", tracker.refs)
			delete(doc.Components.Schemas, componentName)
		}
	}
	for componentName := range doc.Components.Parameters {
		ref := fmt.Sprintf("#/components/parameters/%s", componentName)
		if !tracker.refs[ref] {
			delete(doc.Components.Parameters, componentName)
		}
	}
	for componentName := range doc.Components.SecuritySchemes {
		ref := fmt.Sprintf("#/components/securitySchemes/%s", componentName)
		if !tracker.refs[ref] {
			delete(doc.Components.SecuritySchemes, componentName)
		}
	}
	for componentName := range doc.Components.RequestBodies {
		ref := fmt.Sprintf("#/components/requestBodies/%s", componentName)
		if !tracker.refs[ref] {
			delete(doc.Components.RequestBodies, componentName)
		}
	}
	for componentName := range doc.Components.Responses {
		ref := fmt.Sprintf("#/components/responses/%s", componentName)
		if !tracker.refs[ref] {
			delete(doc.Components.Responses, componentName)
		}
	}
	for componentName := range doc.Components.Headers {
		ref := fmt.Sprintf("#/components/headers/%s", componentName)
		if !tracker.refs[ref] {
			delete(doc.Components.Headers, componentName)
		}
	}
	for componentName := range doc.Components.Examples {
		ref := fmt.Sprintf("#/components/examples/%s", componentName)
		if !tracker.refs[ref] {
			delete(doc.Components.Examples, componentName)
		}
	}
	for componentName := range doc.Components.Links {
		ref := fmt.Sprintf("#/components/links/%s", componentName)
		if !tracker.refs[ref] {
			delete(doc.Components.Links, componentName)
		}
	}
	for componentName := range doc.Components.Callbacks {
		ref := fmt.Sprintf("#/components/callbacks/%s", componentName)
		if !tracker.refs[ref] {
			delete(doc.Components.Callbacks, componentName)
		}
	}

	return nil
}
