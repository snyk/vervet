package vervet

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
)

// Spec defines a specific version of an OpenAPI document.
type Spec struct {
	*openapi3.T
}

// SpecVersions defines an OpenAPI specification consisting of one or more
// versioned endpoints.
type SpecVersions struct {
	paths map[string]*EndpointVersions
}

// LoadSpecVersions returns a SpecVersions loaded from a directory structure
// containing one or more Endpoint subdirectories.
func LoadSpecVersions(root string) (*SpecVersions, error) {
	epPaths, err := findEndpoints(root)
	if err != nil {
		return nil, err
	}
	svs := &SpecVersions{paths: map[string]*EndpointVersions{}}
	for i := range epPaths {
		eps, err := LoadEndpointVersions(epPaths[i])
		if err != nil {
			return nil, fmt.Errorf("failed to load endpoint at %q: %w", epPaths[i], err)
		}
		path := eps.Path()
		if path == "" {
			continue
		}
		if _, ok := svs.paths[path]; ok {
			return nil, fmt.Errorf("multiple conflicting endpoints found for path %q", path)
		}
		svs.paths[path] = eps
	}
	return svs, nil
}

// Versions returns a slice containing each Version defined by an Endpoint in this specification.
func (s *SpecVersions) Versions() []Version {
	vset := map[Version]bool{}
	for _, eps := range s.paths {
		for i := range eps.versions {
			vset[eps.versions[i].Version] = true
		}
	}
	versions := make([]Version, len(vset))
	i := 0
	for k := range vset {
		versions[i] = k
		i++
	}
	sort.Sort(versionSlice(versions))
	return versions
}

// At returns the Spec matching a version string.
func (s *SpecVersions) At(vs string) (*Spec, error) {
	if vs == "" {
		vs = time.Now().UTC().Format("2006-01-02")
	}
	v, err := ParseVersion(vs)
	if err != nil {
		return nil, err
	}
	var result *openapi3.T
	for _, eps := range s.paths {
		ep, err := eps.At(string(v))
		if err == ErrNoMatchingVersion {
			continue
		} else if err != nil {
			return nil, err
		}
		if result == nil {
			result = ep.T
		} else {
			mergeSpec(result, ep.T)
		}
	}
	if result == nil {
		return nil, ErrNoMatchingVersion
	}
	return &Spec{T: result}, nil
}

// mergeSpec adds the paths and components from a source OpenAPI document root,
// to a destination document root.
//
// TODO: This is a naive implementation that should be improved to detect and
// resolve conflicts better. For example, distinct endpoints might have
// localized references with the same URIs but different content.
// Content-addressible endpoint versions may further facilitate governance;
// this also would facilitate detecting and relocating such conflicts.
func mergeSpec(dst, src *openapi3.T) {
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

func findEndpoints(root string) ([]string, error) {
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
