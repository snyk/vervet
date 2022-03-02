package vervet

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"sort"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
)

const (
	// ExtSnykApiStability is used to annotate a top-level endpoint version spec with its API release stability level.
	ExtSnykApiStability = "x-snyk-api-stability"

	// ExtSnykApiResource is used to annotate a path in a compiled OpenAPI spec with its source resource name.
	ExtSnykApiResource = "x-snyk-api-resource"

	// ExtSnykApiVersion is used to annotate a path in a compiled OpenAPI spec with its resolved release version.
	ExtSnykApiVersion = "x-snyk-api-version"

	// ExtSnykApiReleases is used to annotate a path in a compiled OpenAPI spec
	// with all the release versions containing a change in the path info. This
	// is useful for navigating changes in a particular path across versions.
	ExtSnykApiReleases = "x-snyk-api-releases"

	// ExtSnykDeprecatedBy is used to annotate a path in a resource version
	// spec with the subsequent version that deprecates it. This may be used
	// by linters, service middleware and API documentation to indicate which
	// version deprecates a given version.
	ExtSnykDeprecatedBy = "x-snyk-deprecated-by"

	// ExtSnykSunsetEligible is used to annotate a path in a resource version
	// spec which is deprecated, with the sunset eligible date: the date after
	// which the resource version may be removed and no longer available.
	ExtSnykSunsetEligible = "x-snyk-sunset-eligible"
)

// Resource defines a specific version of a resource, corresponding to a
// standalone OpenAPI specification document that defines its operations,
// schema, etc. While a resource spec may declare multiple paths, they should
// all describe operations on a single conceptual resource.
type Resource struct {
	*Document
	Name         string
	Version      Version
	sourcePrefix string
}

type extensionNotFoundError struct {
	extension string
}

// Error implements error.
func (e *extensionNotFoundError) Error() string {
	return fmt.Sprintf("extension \"%s\" not found", e.extension)
}

// Is returns whether an error matches this error instance.
func (e *extensionNotFoundError) Is(err error) bool {
	_, ok := err.(*extensionNotFoundError)
	return ok
}

// Validate returns whether the Resource is valid. The OpenAPI specification
// must be valid, and must declare at least one path.
func (e *Resource) Validate(ctx context.Context) error {
	// Validate the OpenAPI spec
	err := e.Document.Validate(ctx)
	if err != nil {
		return err
	}
	// Resource path checks. There should be at least one path per resource.
	if len(e.Paths) < 1 {
		return fmt.Errorf("spec contains no paths")
	}
	return nil
}

// ResourceVersions defines a collection of multiple versions of an Resource.
type ResourceVersions struct {
	versions resourceVersionSlice
}

// Name returns the resource name for a collection of resource versions.
func (e *ResourceVersions) Name() string {
	for i := range e.versions {
		return e.versions[i].Name
	}
	return ""
}

// Versions returns a slice containing each Version defined for this endpoint.
func (e *ResourceVersions) Versions() []Version {
	result := make([]Version, len(e.versions))
	for i := range e.versions {
		result[i] = e.versions[i].Version
	}
	return result
}

// ErrNoMatchingVersion indicates the requested endpoint version cannot be
// satisfied by the declared Resource versions that are available.
var ErrNoMatchingVersion = fmt.Errorf("no matching version")

// At returns the Resource matching a version string. The endpoint returned
// will be the latest available version with a stability equal to or greater
// than the requested version, or ErrNoMatchingVersion if no matching version
// is available.
func (e *ResourceVersions) At(vs string) (*Resource, error) {
	if vs == "" {
		vs = time.Now().UTC().Format("2006-01-02")
	}
	v, err := ParseVersion(vs)
	if err != nil {
		return nil, fmt.Errorf("invalid version %q: %w", vs, err)
	}
	for i := len(e.versions) - 1; i >= 0; i-- {
		ev := e.versions[i].Version
		if dateCmp, stabilityCmp := ev.compareDateStability(&v); dateCmp <= 0 && stabilityCmp >= 0 {
			return e.versions[i], nil
		}
	}
	return nil, ErrNoMatchingVersion
}

type resourceVersionSlice []*Resource

// Less implements sort.Interface.
func (e resourceVersionSlice) Less(i, j int) bool {
	return e[i].Version.Compare(e[j].Version) < 0
}

// Len implements sort.Interface.
func (e resourceVersionSlice) Len() int { return len(e) }

// Swap implements sort.Interface.
func (e resourceVersionSlice) Swap(i, j int) { e[i], e[j] = e[j], e[i] }

// LoadResourceVersions returns a ResourceVersions slice parsed from a
// directory structure of resource specs. This directory will be of the form:
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
func LoadResourceVersions(epPath string) (*ResourceVersions, error) {
	specYamls, err := filepath.Glob(epPath + "/*/spec.yaml")
	if err != nil {
		return nil, err
	}
	return LoadResourceVersionsFileset(specYamls)
}

// LoadResourceVersionFileset returns a ResourceVersions slice parsed from the
// directory structure described above for LoadResourceVersions.
func LoadResourceVersionsFileset(specYamls []string) (*ResourceVersions, error) {
	var resourceVersions ResourceVersions
	var err error
	type operationKey struct {
		path, operation string
	}
	opReleases := map[operationKey]VersionSlice{}

	for i := range specYamls {
		specYamls[i], err = filepath.Abs(specYamls[i])
		if err != nil {
			return nil, fmt.Errorf("failed to canonicalize %q: %w", specYamls[i], err)
		}
		versionDir := filepath.Dir(specYamls[i])
		versionBase := filepath.Base(versionDir)
		rc, err := loadResource(specYamls[i], versionBase)
		if err != nil {
			return nil, err
		}
		if rc == nil {
			continue
		}
		rc.sourcePrefix = specYamls[i]
		err = rc.Validate(context.TODO())
		if err != nil {
			return nil, err
		}
		// Map release versions per operation
		for path, pathItem := range rc.Paths {
			for _, opName := range operationNames {
				op := getOperationByName(pathItem, opName)
				if op != nil {
					op.ExtensionProps.Extensions[ExtSnykApiVersion] = rc.Version.String()
					opKey := operationKey{path, opName}
					opReleases[opKey] = append(opReleases[opKey], rc.Version)
				}
			}
		}
		resourceVersions.versions = append(resourceVersions.versions, rc)
	}
	// Sort release versions per path
	for _, releases := range opReleases {
		sort.Sort(releases)
	}
	// Sort the resources themselves by version
	sort.Sort(resourceVersionSlice(resourceVersions.versions))
	// Annotate each path in each resource version with the other change
	// versions affecting the path. This supports navigation across versions.
	for _, rc := range resourceVersions.versions {
		for path, pathItem := range rc.Paths {
			for _, opName := range operationNames {
				op := getOperationByName(pathItem, opName)
				if op == nil {
					continue
				}
				// Annotate operation with other release versions available for this path
				releases := opReleases[operationKey{path, opName}]
				op.ExtensionProps.Extensions[ExtSnykApiReleases] = releases.Strings()
				// Annotate operation with deprecated-by and sunset information
				if deprecatedBy, ok := releases.Deprecates(rc.Version); ok {
					op.ExtensionProps.Extensions[ExtSnykDeprecatedBy] = deprecatedBy.String()
					if sunset, ok := rc.Version.Sunset(deprecatedBy); ok {
						op.ExtensionProps.Extensions[ExtSnykSunsetEligible] = sunset.Format("2006-01-02")
					}
				}
			}
		}
	}
	return &resourceVersions, nil
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
			return "", &extensionNotFoundError{key}
		}
		return "", fmt.Errorf("unexpected extension %v type %T", m, m)
	}
}

// IsExtensionNotFound returns bool whether error from ExtensionString is not found versus unexpected.
func IsExtensionNotFound(err error) bool {
	return errors.Is(err, &extensionNotFoundError{})
}

func loadResource(specPath string, versionStr string) (*Resource, error) {
	name := filepath.Base(filepath.Dir(filepath.Dir(specPath)))
	doc, err := NewDocumentFile(specPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load spec from %q: %w", specPath, err)
	}

	stabilityStr, err := ExtensionString(doc.T.ExtensionProps, ExtSnykApiStability)
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

	if len(doc.Paths) == 0 {
		return nil, nil
	}

	// Expand x-snyk-include-headers extensions
	err = IncludeHeaders(doc)
	if err != nil {
		return nil, fmt.Errorf("failed to load x-snyk-include-headers extensions: %w", err)
	}

	// Localize all references, so we emit a completely self-contained OpenAPI document.
	err = Localize(doc)
	if err != nil {
		return nil, fmt.Errorf("failed to localize refs: %w", err)
	}

	ep := &Resource{Name: name, Document: doc, Version: version}
	for path := range doc.T.Paths {
		doc.T.Paths[path].ExtensionProps.Extensions[ExtSnykApiResource] = name
	}
	return ep, nil
}
