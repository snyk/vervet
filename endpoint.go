package vervet

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
	Version Version
	path    string
}

func (e *EndpointVersion) Validate(ctx context.Context) error {
	// Validate the OpenAPI spec
	err := e.T.Validate(ctx)
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
		m[e.versions[i].Version] = e.versions[i]
	}
	return m
}

var ErrNoMatchingVersion = fmt.Errorf("no matching version")

func (e *EndpointVersions) At(vs string) (*EndpointVersion, error) {
	if vs == "" {
		vs = time.Now().UTC().Format("2006-01-02")
	}
	v, err := ParseVersion(vs)
	if err != nil {
		return nil, fmt.Errorf("invalid version %q: %w", vs, err)
	}
	for i := len(e.versions) - 1; i >= 0; i-- {
		if e.versions[i].Version.Compare(v) <= 0 {
			return e.versions[i], nil
		}
	}
	return nil, ErrNoMatchingVersion
}

func (v Version) Compare(vr Version) int {
	// Lexicographical compare actually works fine for this:
	// YYYY-mm-dd < beta < experimental
	// FIXME: mere coincidence!
	return strings.Compare(string(v), string(vr))
}

type endpointVersionSlice []*EndpointVersion

func (e endpointVersionSlice) Less(i, j int) bool {
	return e[i].Version.Compare(e[j].Version) < 0
}
func (e endpointVersionSlice) Len() int      { return len(e) }
func (e endpointVersionSlice) Swap(i, j int) { e[i], e[j] = e[j], e[i] }

type versionSlice []Version

func (vs versionSlice) Less(i, j int) bool {
	return vs[i].Compare(vs[j]) < 0
}
func (vs versionSlice) Len() int      { return len(vs) }
func (vs versionSlice) Swap(i, j int) { vs[i], vs[j] = vs[j], vs[i] }

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
		err = ep.Validate(context.TODO())
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

	t.ExtensionProps.Extensions["x-snyk-api-version"] = string(version)

	ep := &EndpointVersion{T: t, Version: version}
	for path := range t.Paths {
		ep.path = path
		break
	}
	return ep, nil
}
