package storage

import (
	"fmt"
	"sort"
	"time"

	"github.com/snyk/vervet/v4"
)

// ContentRevision is the exact contents and metadata of a service's version at scraping timestamp.
type ContentRevision struct {
	// Service is the name of the service.
	Service string
	// Version is the API version number of this content.
	Version vervet.Version
	// Timestamp represents when this revision was scraped.
	Timestamp time.Time
	// Digest is the sha of the revision derived from its content.
	Digest Digest
	// Blob is the actual content of this revision, the unmarshalled API spec.
	Blob []byte
	// TODO: store the sunset time when a version is removed
	//sunset    *time.Time
}

// ServiceRevisions tracks a collection of ContentRevisions and API uniqueVersions for a single service.
type ServiceRevisions struct {
	// Revisions is a map of version to a collection of revisions.  During collation, content revision with the latest scraping timestamp is used.
	Revisions map[vervet.Version][]ContentRevision
	// Versions is a collection of API uniqueVersions that this service serves.
	Versions vervet.VersionSlice
}

// NewServiceRevisions returns a new instance of ServiceRevisions.
func NewServiceRevisions() *ServiceRevisions {
	return &ServiceRevisions{
		Revisions: make(map[vervet.Version][]ContentRevision),
		Versions:  make(vervet.VersionSlice, 0),
	}
}

// Add registers a new ContentRevision for the service.
func (s *ServiceRevisions) Add(revision ContentRevision) {
	version := revision.Version
	if _, ok := s.Revisions[version]; !ok {
		s.Versions = append(s.Versions, version)
		sort.Sort(s.Versions)
	}
	s.Revisions[version] = append(s.Revisions[version], revision)
}

// ResolveLatestRevision returns the latest revision that matches the given version number. If no exact version is found,
// it uses vervet to resolve the most recent version. When multiple revisions are found for a given version,
// the content revision with the latest scrape timestamp is returned.
func (s ServiceRevisions) ResolveLatestRevision(version vervet.Version) (ContentRevision, error) {
	var revision ContentRevision
	revisions, ok := s.Revisions[version]
	if !ok {
		resolvedVersion, err := s.Versions.Resolve(version)
		if err != nil {
			return revision, err
		}

		revisions, ok = s.Revisions[resolvedVersion]
		if !ok {
			return revision, fmt.Errorf("no revision found for resolved version: %s", resolvedVersion)
		}
	}

	for _, r := range revisions {
		if revision.Timestamp.IsZero() || r.Timestamp.After(revision.Timestamp) {
			revision = r
		}
	}
	return revision, nil
}
