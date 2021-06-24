package vervet

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"

	"github.com/getkin/kin-openapi/openapi3"
)

type SpecVersion struct {
	*openapi3.T
}

type SpecVersions struct {
	paths map[string]*EndpointVersions
}

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
		svs.paths[eps.path] = eps
	}
	return svs, nil
}

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

func (s *SpecVersions) At(v Version) (*SpecVersion, error) {
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
	return &SpecVersion{T: result}, nil
}

func mergeSpec(dst, src *openapi3.T) {
	panic("TODO")
}

func findEndpoints(root string) ([]string, error) {
	var paths []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
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
