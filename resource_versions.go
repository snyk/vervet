package vervet

import (
	"fmt"
	"sort"

	"github.com/getkin/kin-openapi/openapi3"
)

type resourceVersionsSlice []*ResourceVersions

func (s resourceVersionsSlice) validate() error {
	for _, v := range s.versions() {
		resourcePaths := map[string]string{}
		for _, eps := range s {
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

func (s resourceVersionsSlice) versions() VersionSlice {
	vset := map[Version]bool{}
	for _, eps := range s {
		for i := range eps.versions {
			vset[eps.versions[i].Version] = true
		}
	}
	var versions VersionSlice
	for v := range vset {
		versions = append(versions, v)
	}
	sort.Sort(versions)
	return versions
}

func (s resourceVersionsSlice) at(v Version) (*openapi3.T, error) {
	dd := NewComponentDeduplicator()
	for _, eps := range s {
		ep, err := eps.At(v.String())
		if err == ErrNoMatchingVersion {
			continue
		} else if err != nil {
			return nil, err
		}
		err = dd.Index(ep.path, ep.T)
		if err != nil {
			return nil, err
		}
	}
	err := dd.Deduplicate()
	if err != nil {
		return nil, err
	}

	var result *openapi3.T
	for _, eps := range s {
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
