package vervet

import (
	"fmt"
	"sort"
	"time"

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
			for _, path := range ep.Paths.InMatchingOrder() {
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
	coll := NewCollator()
	for _, eps := range s {
		ep, err := eps.At(v.String())
		if err == ErrNoMatchingVersion {
			continue
		} else if err != nil {
			return nil, err
		}
		err = coll.Collate(ep)
		if err != nil {
			return nil, err
		}
	}
	result := coll.Result()
	if result == nil {
		return nil, ErrNoMatchingVersion
	}
	if result.Extensions == nil {
		result.Extensions = map[string]any{}
	}
	result.Extensions[ExtSnykApiLifecycle] = v.LifecycleAt(time.Time{}).String()
	return result, nil
}
