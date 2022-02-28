package generator

import (
	"path/filepath"

	"github.com/snyk/vervet/v3"
	"github.com/snyk/vervet/v3/config"
	"github.com/snyk/vervet/v3/internal/compiler"
)

// ResourceKey uniquely identifies an API resource.
type ResourceKey struct {
	API      string
	Resource string
	Path     string
}

// ResourceMap defines a mapping from API resource identity to its versions.
type ResourceMap map[ResourceKey]*vervet.ResourceVersions

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
