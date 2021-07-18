package vervet

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"sort"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
)

const (
	// ExtSnykApiStability is used to annotate a top-level endpoint version spec with its API release stability level.
	ExtSnykApiStability = "x-snyk-api-stability"

	// ExtSnykApiVersion is used to annotate a path in a compiled OpenAPI spec with its resolved release version.
	ExtSnykApiVersion = "x-snyk-api-version"
)

// Endpoint defines a specific version of an endpoint, having a standalone
// OpenAPI specification document defining a single endpoint.
type Endpoint struct {
	*openapi3.T
	Version      *Version
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
func (e *EndpointVersions) Versions() []*Version {
	result := make([]*Version, len(e.versions))
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
		ev := e.versions[i].Version
		if (ev.Date.Before(v.Date) || ev.Date.Equal(v.Date)) && v.Stability.Compare(ev.Stability) <= 0 {
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

type versionSlice []*Version

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
//     +- 2021-07-14
//        +- spec.yaml
//
// The endpoint version stability level is defined by the
// ExtSnykApiStability extension value at the top-level of the OpenAPI
// document.
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
		versionBase := filepath.Base(versionDir)
		ep, err := loadEndpoint(specYamls[i], versionBase)
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

// ExtensionString returns the string value of an OpenAPI extension.
func ExtensionString(extProps openapi3.ExtensionProps, key string) (string, error) {
	switch m := extProps.Extensions[key].(type) {
	case json.RawMessage:
		var s string
		err := json.Unmarshal(extProps.Extensions[key].(json.RawMessage), &s)
		return s, err
	case string:
		return m, nil
	default:
		if m == nil {
			return "", fmt.Errorf("extension %q not found", key)
		}
		return "", fmt.Errorf("unexpected extension %v type %T", m, m)
	}
}

func loadEndpoint(specPath string, versionStr string) (*Endpoint, error) {
	t, err := LoadSpecFile(specPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load spec from %q: %w", specPath, err)
	}

	stabilityStr, err := ExtensionString(t.ExtensionProps, ExtSnykApiStability)
	if err != nil {
		return nil, err
	}
	if stabilityStr != "ga" {
		versionStr = versionStr + "~" + stabilityStr
	}
	version, err := ParseVersion(versionStr)
	if err != nil {
		return nil, fmt.Errorf("invalid version %q", versionStr)
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
		t.Paths[path].ExtensionProps.Extensions[ExtSnykApiVersion] = version.String()
	}
	return ep, nil
}
