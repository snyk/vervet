package vervet

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/hairyhenderson/go-codeowners"
	"golang.org/x/exp/maps"
)

const (
	// ExtSnykApiStability is used to annotate a top-level resource version
	// spec with its API release stability level.
	ExtSnykApiStability = "x-snyk-api-stability"

	// ExtApiStability is used to annotate a path in a compiled OpenAPI spec
	// with its API release stability level.
	ExtApiStabilityLevel = "x-stability-level"

	// ExtSnykApiLifecycle is used to annotate compiled OpenAPI with lifecycle
	// stage: releases, deprecated or sunset. It is applied at the top-level as
	// well as per-operation.
	ExtSnykApiLifecycle = "x-snyk-api-lifecycle"

	// ExtSnykApiResource is used to annotate a path in a compiled OpenAPI spec
	// with its source resource name.
	ExtSnykApiResource = "x-snyk-api-resource"

	// ExtSnykApiVersion is used to annotate a path in a compiled OpenAPI spec
	// with its resolved release version. It is also used to identify the
	// overall version of the compiled spec at the document level.
	ExtSnykApiVersion = "x-snyk-api-version"

	// ExtSnykApiOwner is used to annotate an operation in a compiled OpenAPI spec
	// with the owners of the operation. This is useful to get to the owning github team.
	ExtSnykApiOwner = "x-snyk-api-owners"

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

// ResourceVersion defines a specific version of a resource, corresponding to a
// standalone OpenAPI specification document that defines its operations,
// schema, etc. While a resource spec may declare multiple paths, they should
// all describe operations on a single conceptual resource.
type ResourceVersion struct {
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

// Validate returns whether the ResourceVersion is valid. The OpenAPI
// specification must be valid, and must declare at least one path.
func (rv *ResourceVersion) Validate(ctx context.Context) error {
	// Validate the OpenAPI spec
	err := rv.Document.Validate(ctx)
	if err != nil {
		return err
	}
	// Resource path checks. There should be at least one path per resource.
	if rv.Paths.Len() < 1 {
		return fmt.Errorf("spec contains no paths")
	}
	return nil
}

// cleanRefs removes any shared pointer references that might exist between
// this resource version document and any others.
func (rv *ResourceVersion) cleanRefs() error {
	buf, err := json.Marshal(rv.Document.T)
	if err != nil {
		return err
	}
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData(buf)
	if err != nil {
		return err
	}
	if err := loader.ResolveRefsIn(rv.T, rv.url); err != nil {
		return err
	}
	rv.T = doc
	return nil
}

// ResourceVersions defines a collection of multiple versions of a resource.
type ResourceVersions struct {
	versions map[Version]*ResourceVersion
	index    VersionIndex
}

// Name returns the resource name for a collection of resource versions.
func (rv *ResourceVersions) Name() string {
	for i := range rv.versions {
		return rv.versions[i].Name
	}
	return ""
}

// Versions returns each Version defined for this resource.
func (rv *ResourceVersions) Versions() VersionSlice {
	return rv.index.Versions()
}

// ErrNoMatchingVersion indicates the requested version cannot be satisfied by
// the declared versions that are available.
var ErrNoMatchingVersion = fmt.Errorf("no matching version")

// At returns the ResourceVersion matching a version string. The version of the
// resource returned will be the latest available version with a stability
// equal to or greater than the requested version, or ErrNoMatchingVersion if
// no matching version is available.
func (rv *ResourceVersions) At(vs string) (*ResourceVersion, error) {
	if vs == "" {
		vs = time.Now().UTC().Format("2006-01-02")
	}
	v, err := ParseVersion(vs)
	if err != nil {
		return nil, fmt.Errorf("invalid version %q: %w", vs, err)
	}
	resolvedVersion, err := rv.index.ResolveForBuild(v)
	if err != nil {
		return nil, err
	}
	r, ok := rv.versions[resolvedVersion]
	if !ok {
		return nil, ErrNoMatchingVersion
	}

	// skip resolving versions for resources that have been marked as sunset in a previous version
	if lifecycle, err := r.Document.Lifecycle(); err == nil && lifecycle == LifecycleSunset &&
		resolvedVersion.DeprecatedBy(v) {
		return nil, ErrNoMatchingVersion
	}

	return r, nil
}

// LoadResourceVersions returns a ResourceVersions slice parsed from a
// directory structure of resource specs. This directory will be of the form:
//
//	resource/
//	+- 2021-01-01
//	   +- spec.yaml
//	+- 2021-06-21
//	   +- spec.yaml
//	+- 2021-07-14
//	   +- spec.yaml
//
// The resource version stability level is defined by the
// ExtSnykApiStability extension value at the top-level of the OpenAPI
// document.
func LoadResourceVersions(epPath string) (*ResourceVersions, error) {
	// Handles case where there is either a spec.yml or spec.yaml file but
	// not edge case where there are both specs for the same API
	// It is assumed that duplicate specs would cause an error elsewhere in vervet
	specs, err := doublestar.FilepathGlob(epPath + "/*/spec.{yaml,yml}")
	if err != nil {
		return nil, err
	}
	specDirs := map[string]struct{}{}
	for _, spec := range specs {
		dir := filepath.Dir(spec)
		if _, ok := specDirs[dir]; ok {
			return nil, fmt.Errorf("duplicate spec found in %s", dir)
		} else {
			specDirs[dir] = struct{}{}
		}
	}
	return LoadResourceVersionsFileset(specs)
}

// LoadResourceVersionFileset returns a ResourceVersions slice parsed from the
// directory structure described above for LoadResourceVersions.
func LoadResourceVersionsFileset(specYamls []string) (*ResourceVersions, error) {
	resourceVersions := ResourceVersions{
		versions: map[Version]*ResourceVersion{},
	}
	var err error
	type operationKey struct {
		path, operation string
	}
	opReleases := map[operationKey]VersionSlice{}
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	ownerFinder, err := codeowners.FromFile(cwd)
	if err != nil {
		return nil, err
	}
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
		for _, path := range rc.Paths.InMatchingOrder() {
			pathItem := rc.Paths.Value(path)
			for _, opName := range operationNames {
				op := getOperationByName(pathItem, opName)
				if op != nil {
					if op.Extensions == nil {
						op.Extensions = make(map[string]any)
					}
					op.Extensions[ExtSnykApiVersion] = rc.Version.String()
					op.Extensions[ExtSnykApiOwner] = ownerFinder.Owners(specYamls[i])
					opKey := operationKey{path, opName}
					opReleases[opKey] = append(opReleases[opKey], rc.Version)
				}
			}
		}
		resourceVersions.versions[rc.Version] = rc
	}
	// Index release versions per path
	opIndexes := make(map[operationKey]VersionIndex, len(opReleases))
	for opKey, releases := range opReleases {
		opIndexes[opKey] = NewVersionIndex(releases)
	}
	// Annotate each path in each resource version with the other change
	// versions affecting the path. This supports navigation across versions.
	for _, rc := range resourceVersions.versions {
		for _, path := range rc.Paths.InMatchingOrder() {
			pathItem := rc.Paths.Value(path)
			for _, opName := range operationNames {
				op := getOperationByName(pathItem, opName)
				if op == nil {
					continue
				}
				// Annotate operation with other release versions available for this path
				releases := opReleases[operationKey{path, opName}]
				index := opIndexes[operationKey{path, opName}]
				op.Extensions[ExtSnykApiReleases] = releases.Strings()
				// Annotate operation with deprecated-by and sunset information
				if deprecatedBy, ok := index.Deprecates(rc.Version); ok {
					op.Extensions[ExtSnykDeprecatedBy] = deprecatedBy.String()
					if sunset, ok := rc.Version.Sunset(deprecatedBy); ok {
						op.Extensions[ExtSnykSunsetEligible] = sunset.Format("2006-01-02")
					}
				}
			}
		}
	}
	resourceVersions.index = NewVersionIndex(maps.Keys(resourceVersions.versions))
	return &resourceVersions, nil
}

// ExtensionString returns the string value of an OpenAPI extension.
func ExtensionString(extensions map[string]interface{}, key string) (string, error) {
	switch m := extensions[key].(type) {
	case json.RawMessage:
		var s string
		err := json.Unmarshal(m, &s)
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

func loadResource(specPath string, versionStr string) (*ResourceVersion, error) {
	name := filepath.Base(filepath.Dir(filepath.Dir(specPath)))
	doc, err := NewDocumentFile(specPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load spec from %q: %w", specPath, err)
	}

	stabilityStr, err := ExtensionString(doc.T.Extensions, ExtSnykApiStability)
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

	if doc.Paths.Len() == 0 {
		return nil, nil //nolint:nilnil //acked
	}

	// Expand x-snyk-include-headers extensions
	err = IncludeHeaders(doc)
	if err != nil {
		return nil, fmt.Errorf("failed to load x-snyk-include-headers extensions: %w", err)
	}

	// Localize all references, so we emit a completely self-contained OpenAPI document.
	// TODO: get context from upstream
	err = Localize(context.Background(), doc)
	if err != nil {
		return nil, fmt.Errorf("failed to localize refs: %w", err)
	}

	ep := &ResourceVersion{Name: name, Document: doc, Version: version}
	for _, path := range doc.T.Paths.InMatchingOrder() {
		if doc.T.Paths.Value(path).Extensions == nil {
			doc.T.Paths.Value(path).Extensions = make(map[string]any)
		}
		doc.T.Paths.Value(path).Extensions[ExtSnykApiResource] = name
	}
	return ep, nil
}

// Localize rewrites all references in an OpenAPI document to local references.
func Localize(ctx context.Context, doc *Document) error {
	doc.InternalizeRefs(ctx, ResolveRefsWithoutSourceName)
	return doc.ResolveRefs()
}

// ResolveRefsWithoutSourceName resolves references without the source url/file name in ref
// background: this was the way kin-openapi used to resolve references, but it was changed
// in the recent versions(v0.127.0) to include the filename in the ref name. Although this
// method prevents conflicts, it causes existing specs to break.
func ResolveRefsWithoutSourceName(t *openapi3.T, componentRef openapi3.ComponentRef) string {
	ref := componentRef.RefString()
	if ref == "" {
		return ""
	}
	split := strings.SplitN(ref, "#", 2)
	if len(split) == 2 {
		return filepath.Base(split[1])
	}
	ref = split[0]
	for ext := filepath.Ext(ref); len(ext) > 0; ext = filepath.Ext(ref) {
		ref = strings.TrimSuffix(ref, ext)
	}
	return filepath.Base(ref)
}
