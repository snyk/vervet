package openapiwalker

import (
	"fmt"

	"github.com/getkin/kin-openapi/openapi3"
)

type RefProcessor interface {
	ProcessCallbackRef(ref *openapi3.CallbackRef) error
	ProcessExampleRef(ref *openapi3.ExampleRef) error
	ProcessHeaderRef(ref *openapi3.HeaderRef) error
	ProcessLinkRef(ref *openapi3.LinkRef) error
	ProcessParameterRef(ref *openapi3.ParameterRef) error
	ProcessRequestBodyRef(ref *openapi3.RequestBodyRef) error
	ProcessResponseRef(ref *openapi3.ResponseRef) error
	ProcessSchemaRef(ref *openapi3.SchemaRef) error
	ProcessSecuritySchemeRef(ref *openapi3.SecuritySchemeRef) error
}

// ProcessRefs visits all the documents and calls the RefProcessor for each ref encountered.
//
//nolint:gocyclo // needs to check each type in the kinopneapi lib
func ProcessRefs(data any, p RefProcessor) error {
	switch v := data.(type) {
	case nil:
		return nil

	case *openapi3.T:
		if v != nil {
			return ProcessRefs(*v, p)
		}
	case *openapi3.Components:
		if v != nil {
			return ProcessRefs(*v, p)
		}
	case *openapi3.MediaType:
		if v != nil {
			return ProcessRefs(*v, p)
		}
	case *openapi3.Response:
		if v != nil {
			return ProcessRefs(*v, p)
		}

	case *openapi3.Parameter:
		if v != nil {
			return ProcessRefs(*v, p)
		}
	case *openapi3.RequestBody:
		if v != nil {
			return ProcessRefs(*v, p)
		}
	case openapi3.RequestBody:
		return ProcessRefs(v.Content, p)

	case openapi3.T:
		if err := ProcessRefs(v.Components, p); err != nil {
			return err
		}
		if err := ProcessRefs(v.Info, p); err != nil {
			return err
		}
		if err := ProcessRefs(v.Paths, p); err != nil {
			return err
		}
		if err := ProcessRefs(v.Security, p); err != nil {
			return err
		}
		if err := ProcessRefs(v.Servers, p); err != nil {
			return err
		}
		if err := ProcessRefs(v.Tags, p); err != nil {
			return err
		}
		if err := ProcessRefs(v.ExternalDocs, p); err != nil {
			return err
		}
	case openapi3.Components:
		if err := ProcessRefs(v.Schemas, p); err != nil {
			return err
		}
		if err := ProcessRefs(v.Parameters, p); err != nil {
			return err
		}
		if err := ProcessRefs(v.Headers, p); err != nil {
			return err
		}
		if err := ProcessRefs(v.RequestBodies, p); err != nil {
			return err
		}
		if err := ProcessRefs(v.Responses, p); err != nil {
			return err
		}
		if err := ProcessRefs(v.SecuritySchemes, p); err != nil {
			return err
		}
		if err := ProcessRefs(v.Examples, p); err != nil {
			return err
		}
		if err := ProcessRefs(v.Links, p); err != nil {
			return err
		}
		if err := ProcessRefs(v.Callbacks, p); err != nil {
			return err
		}

	case openapi3.ResponseBodies:
		for _, ref := range v {
			if err := ProcessRefs(ref, p); err != nil {
				return err
			}
		}

	case openapi3.RequestBodies:
		for _, ref := range v {
			if err := ProcessRefs(ref, p); err != nil {
				return err
			}
		}

	case openapi3.SecurityRequirements:
		for _, requirement := range v {
			if err := ProcessRefs(requirement, p); err != nil {
				return err
			}
		}

	case openapi3.Response:
		if err := ProcessRefs(v.Headers, p); err != nil {
			return err
		}
		if err := ProcessRefs(v.Content, p); err != nil {
			return err
		}
		if err := ProcessRefs(v.Links, p); err != nil {
			return err
		}

	case openapi3.Links:
		for _, link := range v {
			if err := ProcessRefs(link, p); err != nil {
				return err
			}
		}
	case openapi3.Content:
		for _, mediaType := range v {
			if err := ProcessRefs(mediaType, p); err != nil {
				return err
			}
		}
	case openapi3.ParametersMap:
		for _, ref := range v {
			if err := ProcessRefs(ref, p); err != nil {
				return err
			}
		}
	case openapi3.Schemas:
		for _, schema := range v {
			if err := ProcessRefs(schema, p); err != nil {
				return err
			}
		}
	case openapi3.SchemaRefs:
		for _, schema := range v {
			if err := ProcessRefs(schema, p); err != nil {
				return err
			}
		}
	case openapi3.Headers:
		for _, header := range v {
			if err := ProcessRefs(header, p); err != nil {
				return err
			}
		}

	case openapi3.MediaType:
		if err := ProcessRefs(v.Schema, p); err != nil {
			return err
		}
		if err := ProcessRefs(v.Examples, p); err != nil {
			return err
		}

	case openapi3.Parameter:
		if err := ProcessRefs(v.Schema, p); err != nil {
			return err
		}
		if err := ProcessRefs(v.Content, p); err != nil {
			return err
		}
		if err := ProcessRefs(v.Examples, p); err != nil {
			return err
		}

	case openapi3.Examples:
		for _, example := range v {
			if err := ProcessRefs(example, p); err != nil {
				return err
			}
		}
	case *openapi3.Schema:
		if v != nil {
			if err := ProcessRefs(*v, p); err != nil {
				return err
			}
		}
	case openapi3.SecuritySchemes:
		for _, ref := range v {
			if err := ProcessRefs(ref, p); err != nil {
				return err
			}
		}
	case openapi3.Callbacks:
		for _, ref := range v {
			if err := ProcessRefs(ref, p); err != nil {
				return err
			}
		}
	case *openapi3.Paths:
		if v != nil {
			if err := ProcessRefs(*v, p); err != nil {
				return err
			}
		}
	case openapi3.Paths:
		for _, path := range v.Map() {
			if err := ProcessRefs(path, p); err != nil {
				return err
			}
		}

	case openapi3.Schema:
		if err := ProcessRefs(v.Properties, p); err != nil {
			return err
		}
		if err := ProcessRefs(v.Items, p); err != nil {
			return err
		}
		if err := ProcessRefs(v.AllOf, p); err != nil {
			return err
		}
		if err := ProcessRefs(v.AnyOf, p); err != nil {
			return err
		}
		if err := ProcessRefs(v.OneOf, p); err != nil {
			return err
		}
		if err := ProcessRefs(v.Not, p); err != nil {
			return err
		}

	case *openapi3.PathItem:
		if v != nil {
			return ProcessRefs(*v, p)
		}
	case openapi3.PathItem:
		if err := ProcessRefs(v.Connect, p); err != nil {
			return err
		}
		if err := ProcessRefs(v.Delete, p); err != nil {
			return err
		}
		if err := ProcessRefs(v.Get, p); err != nil {
			return err
		}
		if err := ProcessRefs(v.Head, p); err != nil {
			return err
		}
		if err := ProcessRefs(v.Options, p); err != nil {
			return err
		}
		if err := ProcessRefs(v.Patch, p); err != nil {
			return err
		}
		if err := ProcessRefs(v.Post, p); err != nil {
			return err
		}
		if err := ProcessRefs(v.Put, p); err != nil {
			return err
		}
		if err := ProcessRefs(v.Trace, p); err != nil {
			return err
		}
		if err := ProcessRefs(v.Servers, p); err != nil {
			return err
		}
		if err := ProcessRefs(v.Parameters, p); err != nil {
			return err
		}
	case *openapi3.Operation:
		if v != nil {
			return ProcessRefs(*v, p)
		}
	case openapi3.Operation:
		if err := ProcessRefs(v.Parameters, p); err != nil {
			return err
		}
		if err := ProcessRefs(v.RequestBody, p); err != nil {
			return err
		}
		if err := ProcessRefs(v.Responses, p); err != nil {
			return err
		}
		if err := ProcessRefs(v.Callbacks, p); err != nil {
			return err
		}
		if err := ProcessRefs(v.Security, p); err != nil {
			return err
		}
		if err := ProcessRefs(v.Servers, p); err != nil {
			return err
		}
		if err := ProcessRefs(v.ExternalDocs, p); err != nil {
			return err
		}
	case *openapi3.Responses:
		if v != nil {
			return ProcessRefs(*v, p)
		}
	case openapi3.Responses:
		for _, ref := range v.Map() {
			if err := ProcessRefs(ref, p); err != nil {
				return err
			}
		}
	case openapi3.Parameters:
		for _, parameter := range v {
			if err := ProcessRefs(parameter, p); err != nil {
				return err
			}
		}
	case *openapi3.Callback:
		if v != nil {
			return ProcessRefs(*v, p)
		}
	case openapi3.Callback:
		for _, pathItem := range v.Map() {
			if err := ProcessRefs(pathItem, p); err != nil {
				return err
			}
		}
	case *openapi3.Example:
		if v != nil {
			return ProcessRefs(*v, p)
		}
	case *openapi3.Header:
		if v != nil {
			return ProcessRefs(*v, p)
		}
	case openapi3.Header:
		return ProcessRefs(v.Parameter, p)

	case *openapi3.CallbackRef:
		if v != nil {
			if err := p.ProcessCallbackRef(v); err != nil {
				return err
			}
			return ProcessRefs(v.Value, p)
		}

	case *openapi3.ExampleRef:
		if v != nil {
			if err := p.ProcessExampleRef(v); err != nil {
				return err
			}
			return ProcessRefs(v.Value, p)
		}

	case *openapi3.HeaderRef:
		if v != nil {
			if err := p.ProcessHeaderRef(v); err != nil {
				return err
			}
			return ProcessRefs(v.Value, p)
		}
	case *openapi3.LinkRef:
		if v != nil {
			if err := p.ProcessLinkRef(v); err != nil {
				return err
			}
			return ProcessRefs(v.Value, p)
		}
	case *openapi3.ParameterRef:
		if v != nil {
			if err := p.ProcessParameterRef(v); err != nil {
				return err
			}
			return ProcessRefs(v.Value, p)
		}
	case *openapi3.RequestBodyRef:
		if v != nil {
			if err := p.ProcessRequestBodyRef(v); err != nil {
				return err
			}
			return ProcessRefs(v.Value, p)
		}
	case *openapi3.ResponseRef:
		if v != nil {
			if err := p.ProcessResponseRef(v); err != nil {
				return err
			}
			return ProcessRefs(v.Value, p)
		}

	case *openapi3.SchemaRef:
		if v != nil {
			if err := p.ProcessSchemaRef(v); err != nil {
				return err
			}
			return ProcessRefs(v.Value, p)
		}
	case *openapi3.SecuritySchemeRef:
		if v != nil {
			if err := p.ProcessSecuritySchemeRef(v); err != nil {
				return err
			}
			return ProcessRefs(v.Value, p)
		}
	// no interesting nested fields
	case *openapi3.Info:
	case openapi3.Info:
	case *openapi3.SecurityRequirements:
	case *openapi3.SecurityRequirement:
	case openapi3.SecurityRequirement:
	case *openapi3.Servers:
	case openapi3.Servers:
	case openapi3.Server:
	case *openapi3.ExternalDocs:
	case openapi3.ExternalDocs:
	case openapi3.Tags:
	case openapi3.Tag:
	case *openapi3.SecurityScheme:
	case openapi3.SecurityScheme:
	case openapi3.Example:

	default:
		// intentional panic, have covered all the types in kin-openapi v0.127.0
		// might fail in the future if new types are added/types changed, should
		// be caught in tests
		panic(fmt.Sprintf("unhandled type %#v", v))
	}
	return nil
}
