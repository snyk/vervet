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

// Version defines an API version. API versions may be dates of the form
// "YYYY-mm-dd", or stability tags "beta", "experimental".
type Version string

const (
	// VersionExperimental means the API is experimental and still subject to
	// drastic change.
	VersionExperimental = "experimental"

	// VersionBeta means the API is becoming more stable, but may undergo some
	// final changes before being released.
	VersionBeta = "beta"
)

// ParseVersion parses a version string into a Version type, returning an error
// if the string is invalid.
func ParseVersion(s string) (Version, error) {
	switch s {
	case string(VersionExperimental):
		return VersionExperimental, nil
	case string(VersionBeta):
		return VersionBeta, nil
	default:
		_, err := time.Parse("2006-01-02", s)
		if err != nil {
			err = fmt.Errorf("invalid version %q", s)
		}
		return Version(s), err
	}
}

// Compare returns -1 if the given version is less than, 0 if equal to, and 1
// if greater than the caller target version.
func (v Version) Compare(vr Version) int {
	// Lexicographical compare actually works fine for this:
	// YYYY-mm-dd < beta < experimental
	// FIXME: mere coincidence!
	return strings.Compare(string(v), string(vr))
}

// Endpoint defines a specific version of an endpoint, having a standalone
// OpenAPI specification document defining a single endpoint.
type Endpoint struct {
	*openapi3.T
	Version      Version
	sourcePrefix string
}

// Validate returns whether the Endpoint is valid. The OpenAPI specification
// must be valid, and must declare at least one path.
func (e *Endpoint) Validate(ctx context.Context) error {
	// Validate the OpenAPI spec
	err := e.T.Validate(ctx)
	if err != nil {
		return err
	}
	// Endpoint path checks. Should be one and only one path per endpoint.
	if len(e.Paths) < 1 {
		return fmt.Errorf("spec contains no paths")
	}
	return nil
}

// EndpointVersions defines a collection of Endpoint versions sharing the same
// path.
type EndpointVersions struct {
	versions endpointVersionSlice
}

// Versions returns a slice containing each Version defined for this endpoint.
func (e *EndpointVersions) Versions() []Version {
	result := make([]Version, len(e.versions))
	for i := range e.versions {
		result[i] = e.versions[i].Version
	}
	return result
}

// ErrNoMatchingVersion indicates the requested endpoint version cannot be
// satisfied by the declared Endpoint versions that are available.
var ErrNoMatchingVersion = fmt.Errorf("no matching version")

// At returns the Endpoint matching a version string. The endpoint returned
// will be the latest available version less-than or equal to the requested
// version, or ErrNoMatchingVersion if no matching version is available.
func (e *EndpointVersions) At(vs string) (*Endpoint, error) {
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

type endpointVersionSlice []*Endpoint

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

// LoadEndpointVersions returns an EndpointVersions parsed from a directory
// structure of endpoint specs. This directory will be of the form:
//
//     endpoint/
//     +- 2021-01-01
//        +- spec.yaml
//     +- 2021-06-21
//        +- spec.yaml
//     +- beta
//        +- spec.yaml
//
func LoadEndpointVersions(epPath string) (*EndpointVersions, error) {
	specYamls, err := filepath.Glob(epPath + "/*/spec.yaml")
	if err != nil {
		return nil, err
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
		ep, err := loadEndpoint(specYamls[i], version)
		if err != nil {
			return nil, err
		}
		if ep == nil {
			continue
		}
		ep.sourcePrefix = specYamls[i]
		err = ep.Validate(context.TODO())
		if err != nil {
			return nil, err
		}
		eps.versions = append(eps.versions, ep)
	}
	sort.Sort(endpointVersionSlice(eps.versions))
	return &eps, nil
}

func loadEndpoint(specPath string, version Version) (*Endpoint, error) {
	t, err := LoadSpecFile(specPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load spec from %q: %w", specPath, err)
	}

	if len(t.Paths) == 0 {
		return nil, nil
	}

	// Localize all references, so we emit a completely self-contained OpenAPI document.
	err = NewLocalizer(t).Localize()
	if err != nil {
		return nil, fmt.Errorf("failed to localize refs: %w", err)
	}

	ep := &Endpoint{T: t, Version: version}
	for path := range t.Paths {
		t.Paths[path].ExtensionProps.Extensions["x-snyk-api-version"] = string(version)
	}
	return ep, nil
}
