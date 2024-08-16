package openapiwalker

import (
	"fmt"

	"github.com/getkin/kin-openapi/openapi3"
)

type RefProcessor interface {
	ProcessCallbackRef(ref *openapi3.CallbackRef)
	ProcessExampleRef(ref *openapi3.ExampleRef)
	ProcessHeaderRef(ref *openapi3.HeaderRef)
	ProcessLinkRef(ref *openapi3.LinkRef)
	ProcessParameterRef(ref *openapi3.ParameterRef)
	ProcessRequestBodyRef(ref *openapi3.RequestBodyRef)
	ProcessResponseRef(ref *openapi3.ResponseRef)
	ProcessSchemaRef(ref *openapi3.SchemaRef)
	ProcessSecuritySchemeRef(ref *openapi3.SecuritySchemeRef)
}

// ProcessRefs visits all the documents and calls the RefProcessor for each ref encountered.
//
//nolint:gocyclo // needs to check each type in the kinopneapi lib
func ProcessRefs(data any, p RefProcessor) {
	switch v := data.(type) {
	case nil:
		return

	case *openapi3.T:
		if v != nil {
			ProcessRefs(*v, p)
		}
	case *openapi3.Components:
		if v != nil {
			ProcessRefs(*v, p)
		}
	case *openapi3.MediaType:
		if v != nil {
			ProcessRefs(*v, p)
		}
	case *openapi3.Response:
		if v != nil {
			ProcessRefs(*v, p)
		}

	case *openapi3.Parameter:
		if v != nil {
			ProcessRefs(*v, p)
		}
	case *openapi3.RequestBody:
		if v != nil {
			ProcessRefs(*v, p)
		}
	case openapi3.RequestBody:
		ProcessRefs(v.Content, p)

	case openapi3.T:
		ProcessRefs(v.Components, p)
		ProcessRefs(v.Info, p)
		ProcessRefs(v.Paths, p)
		ProcessRefs(v.Security, p)
		ProcessRefs(v.Servers, p)
		ProcessRefs(v.Tags, p)
		ProcessRefs(v.ExternalDocs, p)
	case openapi3.Components:
		ProcessRefs(v.Schemas, p)
		ProcessRefs(v.Parameters, p)
		ProcessRefs(v.Headers, p)
		ProcessRefs(v.RequestBodies, p)
		ProcessRefs(v.Responses, p)
		ProcessRefs(v.SecuritySchemes, p)
		ProcessRefs(v.Examples, p)
		ProcessRefs(v.Links, p)
		ProcessRefs(v.Callbacks, p)

	case openapi3.ResponseBodies:
		for _, ref := range v {
			ProcessRefs(ref, p)
		}

	case openapi3.RequestBodies:
		for _, ref := range v {
			ProcessRefs(ref, p)
		}

	case openapi3.SecurityRequirements:
		for _, requirement := range v {
			ProcessRefs(requirement, p)
		}

	case openapi3.Response:
		ProcessRefs(v.Headers, p)
		ProcessRefs(v.Content, p)
		ProcessRefs(v.Links, p)

	case openapi3.Links:
		for _, link := range v {
			ProcessRefs(link, p)
		}
	case openapi3.Content:
		for _, mediaType := range v {
			ProcessRefs(mediaType, p)
		}
	case openapi3.ParametersMap:
		for _, ref := range v {
			ProcessRefs(ref, p)
		}
	case openapi3.Schemas:
		for _, schema := range v {
			ProcessRefs(schema, p)
		}
	case openapi3.SchemaRefs:
		for _, schema := range v {
			ProcessRefs(schema, p)
		}
	case openapi3.Headers:
		for _, header := range v {
			ProcessRefs(header, p)
		}

	case openapi3.MediaType:
		ProcessRefs(v.Schema, p)
		ProcessRefs(v.Examples, p)

	case openapi3.Parameter:
		ProcessRefs(v.Schema, p)
		ProcessRefs(v.Content, p)
		ProcessRefs(v.Examples, p)

	case openapi3.Examples:
		for _, example := range v {
			ProcessRefs(example, p)
		}
	case *openapi3.Schema:
		if v != nil {
			ProcessRefs(*v, p)
		}
	case openapi3.SecuritySchemes:
		for _, ref := range v {
			ProcessRefs(ref, p)
		}
	case openapi3.Callbacks:
		for _, ref := range v {
			ProcessRefs(ref, p)
		}
	case *openapi3.Paths:
		if v != nil {
			ProcessRefs(*v, p)
		}
	case openapi3.Paths:
		for _, path := range v.Map() {
			ProcessRefs(path, p)
		}

	case openapi3.Schema:
		ProcessRefs(v.Properties, p)
		ProcessRefs(v.Items, p)
		ProcessRefs(v.AllOf, p)
		ProcessRefs(v.AnyOf, p)
		ProcessRefs(v.OneOf, p)
		ProcessRefs(v.Not, p)

	case *openapi3.PathItem:
		if v != nil {
			ProcessRefs(*v, p)
		}
	case openapi3.PathItem:
		ProcessRefs(v.Connect, p)
		ProcessRefs(v.Delete, p)
		ProcessRefs(v.Get, p)
		ProcessRefs(v.Head, p)
		ProcessRefs(v.Options, p)
		ProcessRefs(v.Patch, p)
		ProcessRefs(v.Post, p)
		ProcessRefs(v.Put, p)
		ProcessRefs(v.Trace, p)
		ProcessRefs(v.Servers, p)
		ProcessRefs(v.Parameters, p)
	case *openapi3.Operation:
		if v != nil {
			ProcessRefs(*v, p)
		}
	case openapi3.Operation:
		ProcessRefs(v.Parameters, p)
		ProcessRefs(v.RequestBody, p)
		ProcessRefs(v.Responses, p)
		ProcessRefs(v.Callbacks, p)
		ProcessRefs(v.Security, p)
		ProcessRefs(v.Servers, p)
		ProcessRefs(v.ExternalDocs, p)
	case *openapi3.Responses:
		if v != nil {
			ProcessRefs(*v, p)
		}
	case openapi3.Responses:
		for _, ref := range v.Map() {
			ProcessRefs(ref, p)
		}
	case openapi3.Parameters:
		for _, parameter := range v {
			ProcessRefs(parameter, p)
		}
	case *openapi3.Callback:
		if v != nil {
			ProcessRefs(*v, p)
		}
	case openapi3.Callback:
		for _, pathItem := range v.Map() {
			ProcessRefs(pathItem, p)
		}
	case *openapi3.Example:
		if v != nil {
			ProcessRefs(*v, p)
		}
	case *openapi3.Header:
		if v != nil {
			ProcessRefs(*v, p)
		}
	case openapi3.Header:
		ProcessRefs(v.Parameter, p)

	case *openapi3.CallbackRef:
		if v != nil {
			p.ProcessCallbackRef(v)
			ProcessRefs(v.Value, p)
		}

	case *openapi3.ExampleRef:
		if v != nil {
			p.ProcessExampleRef(v)
			ProcessRefs(v.Value, p)
		}

	case *openapi3.HeaderRef:
		if v != nil {
			p.ProcessHeaderRef(v)
			ProcessRefs(v.Value, p)
		}
	case *openapi3.LinkRef:
		if v != nil {
			p.ProcessLinkRef(v)
			ProcessRefs(v.Value, p)
		}
	case *openapi3.ParameterRef:
		if v != nil {
			p.ProcessParameterRef(v)
			ProcessRefs(v.Value, p)
		}
	case *openapi3.RequestBodyRef:
		if v != nil {
			p.ProcessRequestBodyRef(v)
			ProcessRefs(v.Value, p)
		}
	case *openapi3.ResponseRef:
		if v != nil {
			p.ProcessResponseRef(v)
			ProcessRefs(v.Value, p)
		}

	case *openapi3.SchemaRef:
		if v != nil {
			p.ProcessSchemaRef(v)
			ProcessRefs(v.Value, p)
		}
	case *openapi3.SecuritySchemeRef:
		if v != nil {
			p.ProcessSecuritySchemeRef(v)
			ProcessRefs(v.Value, p)
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
}
