package mem

import (
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/snyk/vervet/v4"
	"vervet-underground/internal/storage"
)

// versionedResourceMap map [service-name] Vervet Version slice array.
type versionedResourceMap map[string]vervet.VersionSlice

// mappedRevisionSpecs map [Sha digest of contents string] --> spec contents and metadata.
type mappedRevisionSpecs map[storage.Digest]storage.ContentRevision

// collatedVersionMappedSpecs Compiled aggregated spec for all services at that given version.
type collatedVersionMappedSpecs map[vervet.Version]openapi3.T

// versionMappedRevisionSpecs map[version-name][digest] --> spec contents and metadata.
type versionMappedRevisionSpecs map[string]mappedRevisionSpecs

// serviceVersionMappedRevisionSpecs map[service-name][version-name][digest] --> spec contents and metadata.
type serviceVersionMappedRevisionSpecs map[string]versionMappedRevisionSpecs
