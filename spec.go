package vervet

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
)

// SpecVersions defines an OpenAPI specification consisting of one or more
// versioned resources.
type SpecVersions struct {
	resources []*ResourceVersions
}

// LoadSpecVersions returns SpecVersions loaded from a directory structure
// containing one or more Resource subdirectories.
func LoadSpecVersions(root string) (*SpecVersions, error) {
	epPaths, err := findResources(root)
	if err != nil {
		return nil, err
	}
	return LoadSpecVersionsFileset(epPaths)
}

// LoadSpecVersionsFileset returns SpecVersions loaded from a set of spec
// files.
func LoadSpecVersionsFileset(epPaths []string) (*SpecVersions, error) {
	svs := &SpecVersions{}
	for i := range epPaths {
		eps, err := LoadResourceVersions(epPaths[i])
		if err != nil {
			return nil, fmt.Errorf("failed to load resource at %q: %w", epPaths[i], err)
		}
		svs.resources = append(svs.resources, eps)
	}
	if err := svs.Validate(); err != nil {
		return nil, err
	}
	return svs, nil
}

// Validate returns an error if there are conflicting resources at a spec version.
func (s *SpecVersions) Validate() error {
	for _, v := range s.Versions() {
		resourcePaths := map[string]string{}
		for _, eps := range s.resources {
			ep, err := eps.At(v.String())
			if err == ErrNoMatchingVersion {
				continue
			} else if err != nil {
				return fmt.Errorf("validation failed: %w", err)
			}
			for path := range ep.Paths {
				if conflict, ok := resourcePaths[path]; ok {
					return fmt.Errorf("conflict: %q %q", conflict, ep.sourcePrefix)
				}
				resourcePaths[path] = ep.sourcePrefix
			}
		}
	}
	return nil
}

// Versions returns a slice containing each Version defined by an Resource in
// this specification. Versions are sorted in ascending order.
func (s *SpecVersions) Versions() []*Version {
	vset := map[Version]bool{}
	for _, eps := range s.resources {
		for i := range eps.versions {
			vset[*eps.versions[i].Version] = true
		}
	}
	versions := make([]*Version, len(vset))
	i := 0
	for k := range vset {
		v := k
		versions[i] = &v
		i++
	}
	sort.Sort(versionSlice(versions))
	return versions
}

// At returns the OpenAPI document matching a version string.
func (s *SpecVersions) At(vs string) (*openapi3.T, error) {
	if vs == "" {
		vs = time.Now().UTC().Format("2006-01-02")
	}
	v, err := ParseVersion(vs)
	if err != nil {
		return nil, err
	}
	var result *openapi3.T
	for _, eps := range s.resources {
		ep, err := eps.At(v.String())
		if err == ErrNoMatchingVersion {
			continue
		} else if err != nil {
			return nil, err
		}
		if result == nil {
			// Assign a clean copy of the contents of the first resource to the
			// resulting spec. Marshaling is used to ensure that references in
			// the source resource are dropped from the result, which could be
			// modified on subsequent merges.
			buf, err := ep.T.MarshalJSON()
			if err != nil {
				return nil, err
			}
			result = &openapi3.T{}
			err = result.UnmarshalJSON(buf)
			if err != nil {
				return nil, err
			}
		}
		MergeSpec(result, ep.T)
	}
	if result == nil {
		return nil, ErrNoMatchingVersion
	}
	// Remove the API stability extension from the merged OpenAPI spec, this
	// extension is only applicable to individual resource version specs.
	delete(result.ExtensionProps.Extensions, ExtSnykApiStability)
	return result, nil
}

// MergeSpec adds the paths and components from a source OpenAPI document root,
// to a destination document root.
//
// TODO: This is a naive implementation that should be improved to detect and
// resolve conflicts better. For example, distinct resources might have
// localized references with the same URIs but different content.
// Content-addressible resource versions may further facilitate governance;
// this also would facilitate detecting and relocating such conflicts.
func MergeSpec(dst, src *openapi3.T) {
	for k, v := range src.Paths {
		if _, ok := dst.Paths[k]; !ok {
			dst.Paths[k] = v
		}
	}
	for k, v := range src.Components.Schemas {
		if _, ok := dst.Components.Schemas[k]; !ok {
			dst.Components.Schemas[k] = v
		}
	}
	for k, v := range src.Components.Parameters {
		if _, ok := dst.Components.Parameters[k]; !ok {
			dst.Components.Parameters[k] = v
		}
	}
	for k, v := range src.Components.Headers {
		if _, ok := dst.Components.Headers[k]; !ok {
			dst.Components.Headers[k] = v
		}
	}
	for k, v := range src.Components.RequestBodies {
		if _, ok := dst.Components.RequestBodies[k]; !ok {
			dst.Components.RequestBodies[k] = v
		}
	}
	for k, v := range src.Components.Responses {
		if _, ok := dst.Components.Responses[k]; !ok {
			dst.Components.Responses[k] = v
		}
	}
	for k, v := range src.Components.SecuritySchemes {
		if _, ok := dst.Components.SecuritySchemes[k]; !ok {
			dst.Components.SecuritySchemes[k] = v
		}
	}
	for k, v := range src.Components.Examples {
		if _, ok := dst.Components.Examples[k]; !ok {
			dst.Components.Examples[k] = v
		}
	}
	for k, v := range src.Components.Links {
		if _, ok := dst.Components.Links[k]; !ok {
			dst.Components.Links[k] = v
		}
	}
	for k, v := range src.Components.Callbacks {
		if _, ok := dst.Components.Callbacks[k]; !ok {
			dst.Components.Callbacks[k] = v
		}
	}
}

func findResources(root string) ([]string, error) {
	var paths []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		specYamls, err := filepath.Glob(path + "/*/spec.yaml")
		if err != nil {
			return err
		}
		if len(specYamls) > 0 {
			paths = append(paths, path)
			return fs.SkipDir
		}
		return nil
	})
	return paths, err
}
