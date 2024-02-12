package config

import (
	"fmt"

	"github.com/bmatcuk/doublestar/v4"
)

// APIs defines a named map of API instances.
type APIs map[string]*API

// An API defines how and where to build versioned OpenAPI documents from a
// source collection of individual resource specifications and additional
// overlay content to merge.
type API struct {
	Name      string         `json:"-"`
	Resources []*ResourceSet `json:"resources"`
	Overlays  []*Overlay     `json:"overlays"`
	Output    *Output        `json:"output"`
}

// A ResourceSet defines a set of versioned resources that adhere to the same
// standards.
//
// Versioned resources are expressed as individual OpenAPI documents in a
// directory structure:
//
// +-resource
//
//	|
//	+-2021-08-01
//	| |
//	| +-spec.yaml
//	| +-<implementation code, etc. can go here>
//	|
//	+-2021-08-15
//	| |
//	| +-spec.yaml
//	| +-<implementation code, etc. can go here>
//	...
//
// Each YYYY-mm-dd directory under a resource is a version.  The spec.yaml
// in each version is a complete OpenAPI document describing the resource
// at that version.
type ResourceSet struct {
	Description string   `json:"description"`
	Path        string   `json:"path"`
	Excludes    []string `json:"excludes"`
}

func (r *ResourceSet) validate() error {
	for _, exclude := range r.Excludes {
		if !doublestar.ValidatePattern(exclude) {
			return fmt.Errorf("invalid exclude pattern %q", exclude)
		}
	}
	return nil
}

// An Overlay defines additional OpenAPI documents to merge into the aggregate
// OpenAPI spec when compiling an API. These might include special endpoints
// that should be included in the aggregate API but are not versioned, or
// top-level descriptions of the API itself.
type Overlay struct {
	Include string `json:"include"`
	Inline  string `json:"inline"`
}

// Output defines where the aggregate versioned OpenAPI specs should be created
// during compilation.
type Output struct {
	Path  string   `json:"path,omitempty"`
	Paths []string `json:"paths,omitempty"`
}

// EffectivePaths returns a slice of effective configured output paths, whether
// a single or multiple output paths have been configured.
func (o *Output) ResolvePaths() []string {
	if o.Path != "" {
		return []string{o.Path}
	}
	return o.Paths
}

func (a APIs) init(p *Project) error {
	if len(a) == 0 {
		return fmt.Errorf("no apis defined")
	}
	// Referenced generators all exist
	for name, api := range a {
		api.Name = name
		if len(api.Resources) == 0 {
			return fmt.Errorf("no resources defined (apis.%s.resources)", api.Name)
		}
		for rcIndex, resource := range api.Resources {
			if err := resource.validate(); err != nil {
				return fmt.Errorf("%w (apis.%s.resources[%d])", err, api.Name, rcIndex)
			}
		}
		if api.Output != nil {
			if len(api.Output.Paths) > 0 && api.Output.Path != "" {
				return fmt.Errorf("output should specify one of 'path' or 'paths', not both (apis.%s.output)",
					api.Name)
			}
		}
	}
	return nil
}
