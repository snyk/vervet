package vervet

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/getkin/kin-openapi/openapi3"
)

// SpecGlobPattern defines the expected directory structure for the versioned
// OpenAPI specs of a single resource: subdirectories by date, of the form
// YYYY-mm-dd, each containing a spec.yaml file.
const SpecGlobPattern = "**/[0-9][0-9][0-9][0-9]-[0-9][0-9]-[0-9][0-9]/spec.yaml"

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
	resourceMap := map[string][]string{}
	for i := range epPaths {
		resourcePath := filepath.Dir(filepath.Dir(epPaths[i]))
		if resourcePath == "." {
			continue
		}
		resourceMap[resourcePath] = append(resourceMap[resourcePath], epPaths[i])
	}
	var resourceNames []string
	for k := range resourceMap {
		resourceNames = append(resourceNames, k)
	}
	sort.Strings(resourceNames)
	svs := &SpecVersions{}
	for _, resourcePath := range resourceNames {
		specFiles := resourceMap[resourcePath]
		eps, err := LoadResourceVersionsFileset(specFiles)
		if err != nil {
			return nil, fmt.Errorf("failed to load resource at %q: %w", resourcePath, err)
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

// Resources returns a slice of each Resource contained in the spec.
func (s *SpecVersions) Resources() []*ResourceVersions {
	return s.resources
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
		Merge(result, ep.T, false)
	}
	if result == nil {
		return nil, ErrNoMatchingVersion
	}
	// Remove the API stability extension from the merged OpenAPI spec, this
	// extension is only applicable to individual resource version specs.
	delete(result.ExtensionProps.Extensions, ExtSnykApiStability)
	return result, nil
}

func findResources(root string) ([]string, error) {
	var paths []string
	err := doublestar.GlobWalk(os.DirFS(root), SpecGlobPattern,
		func(path string, d fs.DirEntry) error {
			paths = append(paths, filepath.Join(root, path))
			return nil
		})
	if err != nil {
		return nil, err
	}
	return paths, err
}
