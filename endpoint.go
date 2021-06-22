package apiutil

import (
	"context"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
)

type Version string

const (
	VersionExperimental = "experimental"
	VersionBeta         = "beta"
)

func ParseVersion(s string) (Version, error) {
	switch s {
	case string(VersionExperimental):
		return VersionExperimental, nil
	case string(VersionBeta):
		return VersionBeta, nil
	default:
		_, err := time.Parse("2006-01-02", s)
		return Version(s), err
	}
}

type EndpointVersion struct {
	*openapi3.T
	path    string
	version Version
}

func (e *EndpointVersion) Validate() error {
	// Validate the OpenAPI spec
	err := e.T.Validate(context.TODO())
	if err != nil {
		return err
	}
	// Endpoint path checks. Should be one and only one path per endpoint.
	if len(e.Paths) < 1 {
		return fmt.Errorf("spec contains no paths")
	}
	if len(e.Paths) > 1 {
		return fmt.Errorf("spec contains more than one path; this is not allowed for a single endpoint version")
	}
	return nil
}

type EndpointVersions struct {
	versions endpointVersionSlice
	path     string
}

func (e *EndpointVersions) Versions() map[Version]*EndpointVersion {
	m := map[Version]*EndpointVersion{}
	for i := range e.versions {
		m[e.versions[i].version] = e.versions[i]
	}
	return m
}

type endpointVersionSlice []*EndpointVersion

func (e endpointVersionSlice) Less(i, j int) bool {
	// Lexicographical compare actually works fine for this:
	// YYYY-mm-dd < beta < experimental
	// TODO: mere coincidence, fix this
	return strings.Compare(string(e[i].version), string(e[j].version)) < 0
}
func (e endpointVersionSlice) Len() int      { return len(e) }
func (e endpointVersionSlice) Swap(i, j int) { e[i], e[j] = e[j], e[i] }

func LoadEndpointVersions(epPath string) (*EndpointVersions, error) {
	specYamls, err := filepath.Glob(epPath + "/*/spec.yaml")
	if err != nil {
		return nil, err
	}
	if len(specYamls) == 0 {
		return nil, fmt.Errorf("no endpoint versions found at %q", epPath)
	}
	var eps EndpointVersions
	for i := range specYamls {
		specYamls[i], err = filepath.Abs(specYamls[i])
		if err != nil {
			return nil, fmt.Errorf("failed to canonicalize %q: %w", specYamls[i], err)
		}
		versionDir := filepath.Dir(specYamls[i])
		versionName := filepath.Base(versionDir)
		version, err := ParseVersion(versionName)
		if err != nil {
			return nil, fmt.Errorf("invalid version %q at %q: %w", versionName, versionDir, err)
		}
		ep, err := loadEndpointVersion(specYamls[i], version)
		if err != nil {
			return nil, err
		}
		err = ep.Validate()
		if err != nil {
			return nil, err
		}
		if len(eps.versions) > 0 && eps.versions[0].path != ep.path {
			return nil, fmt.Errorf("multiple conflicting paths (%q, %q) for endpoint versions at %q", eps.versions[0].path, ep.path, versionDir)
		}
		eps.versions = append(eps.versions, ep)
	}
	sort.Sort(endpointVersionSlice(eps.versions))
	return &eps, nil
}

func loadEndpointVersion(specPath string, version Version) (*EndpointVersion, error) {
	t, err := LoadSpecFile(specPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load spec from %q: %w", specPath, err)
	}

	// Localize all references, so we emit a completely self-contained OpenAPI document.
	err = NewLocalizer(t).Localize()
	if err != nil {
		return nil, fmt.Errorf("failed to localize refs: %w", err)
	}

	ep := &EndpointVersion{T: t, version: version}
	for path := range t.Paths {
		ep.path = path
		break
	}
	return ep, nil
}
