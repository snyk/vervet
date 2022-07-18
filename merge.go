package vervet

import (
	"sort"

	"github.com/getkin/kin-openapi/openapi3"
)

// Merge adds the paths and components from a source OpenAPI document root,
// to a destination document root.
//
// TODO: This is a naive implementation that should be improved to detect and
// resolve conflicts better. For example, distinct resources might have
// localized references with the same URIs but different content.
// Content-addressible resource versions may further facilitate governance;
// this also would facilitate detecting and relocating such conflicts.
//
// TODO(next-release):
// - This function is suitable for overlay merging scenarios only.
// - Component merging should be removed. Use Collator for safe component
//   merging.
func Merge(dst, src *openapi3.T, replace bool) {
	mergeComponents(dst, src, replace)
	mergeExtensions(dst, src, replace)
	mergeInfo(dst, src, replace)
	mergeOpenAPIVersion(dst, src, replace)
	mergePaths(dst, src, replace)
	mergeSecurityRequirements(dst, src, replace)
	mergeServers(dst, src, replace)
	mergeTags(dst, src, replace)
}

func mergeOpenAPIVersion(dst, src *openapi3.T, replace bool) {
	if dst.OpenAPI == "" || (src.OpenAPI != "" && replace) {
		dst.OpenAPI = src.OpenAPI
	}
}

func mergeTags(dst, src *openapi3.T, replace bool) {
	m := map[string]*openapi3.Tag{}
	for _, t := range dst.Tags {
		m[t.Name] = t
	}
	for _, t := range src.Tags {
		if _, ok := m[t.Name]; !ok || replace {
			m[t.Name] = t
		}
	}
	dst.Tags = openapi3.Tags{}
	var tagNames []string
	for tagName := range m {
		tagNames = append(tagNames, tagName)
	}
	sort.Strings(tagNames)
	for _, tagName := range tagNames {
		dst.Tags = append(dst.Tags, m[tagName])
	}
}

func initDestinationComponents(dst, src *openapi3.T) {
	if src.Components.Schemas != nil && dst.Components.Schemas == nil {
		dst.Components.Schemas = make(map[string]*openapi3.SchemaRef)
	}
	if src.Components.Parameters != nil && dst.Components.Parameters == nil {
		dst.Components.Parameters = make(map[string]*openapi3.ParameterRef)
	}
	if src.Components.Headers != nil && dst.Components.Headers == nil {
		dst.Components.Headers = make(map[string]*openapi3.HeaderRef)
	}
	if src.Components.RequestBodies != nil && dst.Components.RequestBodies == nil {
		dst.Components.RequestBodies = make(map[string]*openapi3.RequestBodyRef)
	}
	if src.Components.Responses != nil && dst.Components.Responses == nil {
		dst.Components.Responses = make(map[string]*openapi3.ResponseRef)
	}
	if src.Components.SecuritySchemes != nil && dst.Components.SecuritySchemes == nil {
		dst.Components.SecuritySchemes = make(map[string]*openapi3.SecuritySchemeRef)
	}
	if src.Components.Examples != nil && dst.Components.Examples == nil {
		dst.Components.Examples = make(map[string]*openapi3.ExampleRef)
	}
	if src.Components.Links != nil && dst.Components.Links == nil {
		dst.Components.Links = make(map[string]*openapi3.LinkRef)
	}
	if src.Components.Callbacks != nil && dst.Components.Callbacks == nil {
		dst.Components.Callbacks = make(map[string]*openapi3.CallbackRef)
	}
}

func mergeComponents(dst, src *openapi3.T, replace bool) {
	initDestinationComponents(dst, src)
	for k, v := range src.Components.Schemas {
		if _, ok := dst.Components.Schemas[k]; !ok || replace {
			dst.Components.Schemas[k] = v
		}
	}
	for k, v := range src.Components.Parameters {
		if _, ok := dst.Components.Parameters[k]; !ok || replace {
			dst.Components.Parameters[k] = v
		}
	}
	for k, v := range src.Components.Headers {
		if _, ok := dst.Components.Headers[k]; !ok || replace {
			dst.Components.Headers[k] = v
		}
	}
	for k, v := range src.Components.RequestBodies {
		if _, ok := dst.Components.RequestBodies[k]; !ok || replace {
			dst.Components.RequestBodies[k] = v
		}
	}
	for k, v := range src.Components.Responses {
		if _, ok := dst.Components.Responses[k]; !ok || replace {
			dst.Components.Responses[k] = v
		}
	}
	for k, v := range src.Components.SecuritySchemes {
		if _, ok := dst.Components.SecuritySchemes[k]; !ok || replace {
			dst.Components.SecuritySchemes[k] = v
		}
	}
	for k, v := range src.Components.Examples {
		if _, ok := dst.Components.Examples[k]; !ok || replace {
			dst.Components.Examples[k] = v
		}
	}
	for k, v := range src.Components.Links {
		if _, ok := dst.Components.Links[k]; !ok || replace {
			dst.Components.Links[k] = v
		}
	}
	for k, v := range src.Components.Callbacks {
		if _, ok := dst.Components.Callbacks[k]; !ok || replace {
			dst.Components.Callbacks[k] = v
		}
	}
}

func mergeExtensions(dst, src *openapi3.T, replace bool) {
	if src.Extensions != nil && dst.Extensions == nil {
		dst.Extensions = make(map[string]interface{}, len(src.Extensions))
	}

	for k, v := range src.Extensions {
		// It's possible for specs to be merged from multiple stabilities and
		// we don't want these different stability inputs to override
		// the declared stability of the output we're building.
		if k == ExtSnykApiStability {
			continue
		}

		if _, ok := dst.Extensions[k]; !ok || replace {
			dst.Extensions[k] = v
		}
	}
}

func mergeInfo(dst, src *openapi3.T, replace bool) {
	if src.Info != nil && (dst.Info == nil || replace) {
		dst.Info = src.Info
	}
}

func mergePaths(dst, src *openapi3.T, replace bool) {
	if src.Paths != nil && dst.Paths == nil {
		dst.Paths = make(openapi3.Paths, len(src.Paths))
	}
	for k, v := range src.Paths {
		if _, ok := dst.Paths[k]; !ok || replace {
			dst.Paths[k] = v
		}
	}
}

func mergeSecurityRequirements(dst, src *openapi3.T, replace bool) {
	if len(src.Security) > 0 && (len(dst.Security) == 0 || replace) {
		dst.Security = src.Security
	}
}

func mergeServers(dst, src *openapi3.T, replace bool) {
	if len(src.Servers) > 0 && (len(dst.Security) == 0 || replace) {
		dst.Servers = src.Servers
	}
}
