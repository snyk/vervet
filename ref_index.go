package vervet

import (
	"github.com/getkin/kin-openapi/openapi3"
)

// RefIndex indexes the distinct references used in an OpenAPI document.
type RefIndex struct {
	refs map[string]struct{}
}

// NewRefIndex returns a new reference index on an OpenAPI document.
func NewRefIndex(doc *openapi3.T) (*RefIndex, error) {
	ix := &RefIndex{refs: map[string]struct{}{}}
	ix.index(doc)
	return ix, nil
}

func (ix *RefIndex) index(doc *openapi3.T) {
	for _, schemaRef := range doc.Components.Schemas {
		ix.refs[extractRef(schemaRef)] = struct{}{}
	}
	for _, parameterRef := range doc.Components.Parameters {
		ix.refs[extractRef(parameterRef)] = struct{}{}
	}
	for _, headerRef := range doc.Components.Headers {
		ix.refs[extractRef(headerRef)] = struct{}{}
	}
	for _, requestBodyRef := range doc.Components.RequestBodies {
		ix.refs[extractRef(requestBodyRef)] = struct{}{}
	}
	for _, responseRef := range doc.Components.Responses {
		ix.refs[extractRef(responseRef)] = struct{}{}
		for _, content := range responseRef.Value.Content {
			ix.refs[extractRef(content.Schema)] = struct{}{}
			for _, propertyRef := range content.Schema.Value.Properties {
				ix.refs[extractRef(propertyRef)] = struct{}{}
			}
		}
	}
	for _, securitySchemesRef := range doc.Components.SecuritySchemes {
		ix.refs[extractRef(securitySchemesRef)] = struct{}{}
	}
	for _, exampleRef := range doc.Components.Examples {
		ix.refs[extractRef(exampleRef)] = struct{}{}
	}
	for _, linkRef := range doc.Components.Links {
		ix.refs[extractRef(linkRef)] = struct{}{}
	}
	for _, callbackRef := range doc.Components.Callbacks {
		ix.refs[extractRef(callbackRef)] = struct{}{}
	}
}

func extractRef(componentRef openapi3.ComponentRef) string {
	if componentRef == nil || componentRef.RefPath() == nil {
		return ""
	}
	return "#" + componentRef.RefPath().Fragment
}

// HasRef returns whether the indexed document contains the given ref.
func (ix *RefIndex) HasRef(ref string) bool {
	_, ok := ix.refs[ref]
	return ok
}
