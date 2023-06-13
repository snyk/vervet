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

// SpecVersions stores a collection of versioned OpenAPI specs.
type SpecVersions struct {
	versions  VersionSlice
	index     VersionIndex
	documents map[Version]*openapi3.T
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
	resourceNames := []string{}
	for k := range resourceMap {
		resourceNames = append(resourceNames, k)
	}
	sort.Strings(resourceNames)
	var resourceVersions resourceVersionsSlice
	for _, resourcePath := range resourceNames {
		specFiles := resourceMap[resourcePath]
		eps, err := LoadResourceVersionsFileset(specFiles)
		if err != nil {
			return nil, fmt.Errorf("failed to load resource at %q: %w", resourcePath, err)
		}
		resourceVersions = append(resourceVersions, eps)
	}
	if err := resourceVersions.validate(); err != nil {
		return nil, err
	}
	return newSpecVersions(resourceVersions)
}

// Versions returns the distinct API versions in this collection of OpenAPI
// documents.
func (sv *SpecVersions) Versions() VersionSlice {
	return sv.versions
}

// At returns the OpenAPI document that matches the given version. If the
// version is not an exact match for an API release, the OpenAPI document
// effective on the given version date for the version stability level is
// returned. Returns ErrNoMatchingVersion if there is no release matching this
// version.
func (sv *SpecVersions) At(v Version) (*openapi3.T, error) {
	resolvedVersion, err := sv.index.Resolve(v)
	if err != nil {
		return nil, err
	}
	doc, ok := sv.documents[resolvedVersion]
	if !ok {
		panic(fmt.Sprintf("missing expected document for version %v", resolvedVersion))
	}
	return doc, nil
}

func (sv *SpecVersions) resolveOperations() {
	type operationKey struct {
		path, operation string
	}
	type operationVersion struct {
		// src document where the active operation was declared
		src *openapi3.T
		// pathItem where the active operation was declared
		pathItem *openapi3.PathItem
		// operation where the active operation was declared
		operation *openapi3.Operation
		// spec version where the active operation was declared
		version Version
	}
	type operationVersionMap map[operationKey]operationVersion
	activeOpsByStability := map[Stability]operationVersionMap{}
	for _, v := range sv.versions {
		doc := sv.documents[v]
		currentActiveOps, ok := activeOpsByStability[v.Stability]
		if !ok {
			currentActiveOps = operationVersionMap{}
			activeOpsByStability[v.Stability] = currentActiveOps
		}

		// Operations declared in this spec become active for the next version
		// at this stability.
		nextActiveOps := operationVersionMap{}
		for path, pathItem := range doc.Paths {
			for _, opName := range operationNames {
				op := getOperationByName(pathItem, opName)
				if op != nil {
					nextActiveOps[operationKey{path, opName}] = operationVersion{
						doc, pathItem, op, v,
					}
				}
			}
		}

		// Operations currently active for this versions's stability get
		// carried forward and remain active.
		for opKey, opValue := range currentActiveOps {
			currentPathItem := doc.Paths[opKey.path]

			// skip adding sunset operations to current document
			if opValue.operation.Extensions[ExtSnykApiLifecycle] == "sunset" {
				continue
			}

			if currentPathItem == nil {
				currentPathItem = &openapi3.PathItem{
					Extensions:  opValue.pathItem.Extensions,
					Description: opValue.pathItem.Description,
					Summary:     opValue.pathItem.Summary,
					Servers:     opValue.pathItem.Servers,
					Parameters:  opValue.pathItem.Parameters,
				}
				doc.Paths[opKey.path] = currentPathItem
			}
			currentOp := getOperationByName(currentPathItem, opKey.operation)
			if currentOp == nil {
				// The added operation may reference components from its source
				// document; import those that are missing here.
				mergeComponents(doc, opValue.src, false)
				setOperationByName(currentPathItem, opKey.operation, opValue.operation)
			}
		}

		// Update currently active operations from any declared in this version.
		for opKey, nextOpValue := range nextActiveOps {
			currentActiveOps[opKey] = nextOpValue
		}
	}
}

var operationNames = []string{
	"connect", "delete", "get", "head", "options", "patch", "post", "put", "trace",
}

func getOperationByName(path *openapi3.PathItem, op string) *openapi3.Operation {
	switch op {
	case "connect":
		return path.Connect
	case "delete":
		return path.Delete
	case "get":
		return path.Get
	case "head":
		return path.Head
	case "options":
		return path.Options
	case "patch":
		return path.Patch
	case "post":
		return path.Post
	case "put":
		return path.Put
	case "trace":
		return path.Trace
	default:
		return nil
	}
}

func setOperationByName(path *openapi3.PathItem, opName string, op *openapi3.Operation) {
	switch opName {
	case "connect":
		path.Connect = op
	case "delete":
		path.Delete = op
	case "get":
		path.Get = op
	case "head":
		path.Head = op
	case "options":
		path.Options = op
	case "patch":
		path.Patch = op
	case "post":
		path.Post = op
	case "put":
		path.Put = op
	case "trace":
		path.Trace = op
	default:
		panic("unsupported operation: " + opName)
	}
}

var stabilities = []Stability{StabilityExperimental, StabilityBeta, StabilityGA}

func newSpecVersions(specs resourceVersionsSlice) (*SpecVersions, error) {
	versions := specs.versions()
	var versionDates []time.Time
	for _, v := range versions {
		if len(versionDates) == 0 || versionDates[len(versionDates)-1] != v.Date {
			versionDates = append(versionDates, v.Date)
		}
	}

	documentVersions := map[Version]*openapi3.T{}
	for _, date := range versionDates {
		for _, stability := range stabilities {
			v := Version{Date: date, Stability: stability}
			doc, err := specs.at(v)
			if err == ErrNoMatchingVersion {
				continue
			} else if err != nil {
				return nil, err
			}
			if doc.Extensions == nil {
				doc.Extensions = map[string]interface{}{}
			}
			doc.Extensions[ExtSnykApiVersion] = v.String()
			documentVersions[v] = doc
		}
	}
	versions = VersionSlice{}
	for v := range documentVersions {
		versions = append(versions, v)
	}
	sort.Sort(versions)
	sv := &SpecVersions{
		versions:  versions,
		index:     NewVersionIndex(versions),
		documents: documentVersions,
	}
	sv.resolveOperations()
	return sv, nil
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
	return paths, nil
}
