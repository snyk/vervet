package generator

import (
	"path/filepath"

	"github.com/getkin/kin-openapi/openapi3"

	"github.com/snyk/vervet/v7"
	"github.com/snyk/vervet/v7/config"
	"github.com/snyk/vervet/v7/internal/compiler"
)

// ResourceKey uniquely identifies an API resource.
type ResourceKey struct {
	API      string
	Resource string
	Path     string
}

// ResourceMap defines a mapping from API resource identity to its versions.
type ResourceMap map[ResourceKey]*vervet.ResourceVersions

// OperationMap defines a mapping from operation name to all versions
// of that operation within a resource.
type OperationMap map[string][]OperationVersion

// OperationVersion represents a version of an operation within a collection of
// resource versions.
type OperationVersion struct {
	*vervet.ResourceVersion
	Path      string
	Method    string
	Operation *openapi3.Operation
}

// MapResourceOperations returns a mapping from operation ID to all versions of that
// operation.
func MapResourceOperations(resourceVersions *vervet.ResourceVersions) (OperationMap, error) {
	result := OperationMap{}
	versions := resourceVersions.Versions()
	for i := range versions {
		r, err := resourceVersions.At(versions[i].String())
		if err != nil {
			return nil, err
		}
		for _, path := range r.Document.Paths.InMatchingOrder() {
			pathItem := r.Document.Paths.Value(path)
			ops := MapPathOperations(pathItem)
			for method, op := range ops {
				opVersion := OperationVersion{
					ResourceVersion: r,
					Path:            path,
					Method:          method,
					Operation:       op,
				}
				result[op.OperationID] = append(result[op.OperationID], opVersion)
			}
		}
	}
	return result, nil
}

// MapPathOperations returns a mapping from HTTP method to *openapi3.Operation
// for a given *openapi3.PathItem.
func MapPathOperations(p *openapi3.PathItem) map[string]*openapi3.Operation {
	result := map[string]*openapi3.Operation{}
	if p.Connect != nil {
		result["connect"] = p.Connect
	}
	if p.Delete != nil {
		result["delete"] = p.Delete
	}
	if p.Get != nil {
		result["get"] = p.Get
	}
	if p.Head != nil {
		result["head"] = p.Head
	}
	if p.Options != nil {
		result["options"] = p.Options
	}
	if p.Patch != nil {
		result["patch"] = p.Patch
	}
	if p.Post != nil {
		result["post"] = p.Post
	}
	if p.Put != nil {
		result["put"] = p.Put
	}
	if p.Trace != nil {
		result["trace"] = p.Trace
	}
	return result
}

// MapResources returns a mapping of all resources managed within a Vervet
// project.
func MapResources(proj *config.Project) (ResourceMap, error) {
	resources := ResourceMap{}
	for apiName, apiConfig := range proj.APIs {
		for _, rcConfig := range apiConfig.Resources {
			specFiles, err := compiler.ResourceSpecFiles(rcConfig)
			if err != nil {
				return nil, err
			}
			for i := range specFiles {
				versionDir := filepath.Dir(specFiles[i])
				resourceDir := filepath.Dir(versionDir)
				resourceKey := ResourceKey{API: apiName, Resource: filepath.Base(resourceDir), Path: resourceDir}
				if _, ok := resources[resourceKey]; !ok {
					rcVersions, err := vervet.LoadResourceVersions(resourceDir)
					if err != nil {
						return nil, err
					}
					resources[resourceKey] = rcVersions
				}
			}
		}
	}
	return resources, nil
}
