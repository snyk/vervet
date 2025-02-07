package config

import (
	"encoding/json"
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
	Paths []string
}

// outputJSON exists for historical purposes, we allowed both Path and Paths to
// be specified in the api spec config. This makes handling internally more
// complex as we have to deal with both, instead we can deal with it at
// deserialisation time.
type outputJSON struct {
	Path  string   `json:"path,omitempty"`
	Paths []string `json:"paths,omitempty"`
}

func (o *Output) UnmarshalJSON(data []byte) error {
	oj := outputJSON{}
	err := json.Unmarshal(data, &oj)
	if err != nil {
		return err
	}
	if oj.Path != "" {
		if len(oj.Paths) > 0 {
			return fmt.Errorf("output should specify one of 'path' or 'paths', not both")
		}
		o.Paths = []string{oj.Path}
		return nil
	}
	o.Paths = oj.Paths
	return nil
}

func (a APIs) init() error {
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
	}
	return nil
}
