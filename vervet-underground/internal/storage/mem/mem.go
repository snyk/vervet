// Package mem provides an in-memory implementation of the storage used in
// Vervet Underground. It's not intended for production use, but as a
// functionally complete reference implementation that can be used to validate
// the other parts of the VU system.
package mem

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/snyk/vervet/v4"

	"vervet-underground/internal/storage"
)

// versionedResourceMap map [service-name] Vervet Version slice array.
type versionedResourceMap map[string]vervet.VersionSlice

// mappedRevisionSpecs map [Sha digest of contents string] --> spec contents and metadata.
type mappedRevisionSpecs map[storage.Digest]storage.ContentRevision

// versionMappedRevisionSpecs map[version-name][digest] --> spec contents and metadata.
type versionMappedRevisionSpecs map[string]mappedRevisionSpecs

// serviceVersionMappedRevisionSpecs map[service-name][version-name][digest] --> spec contents and metadata.
type serviceVersionMappedRevisionSpecs map[string]versionMappedRevisionSpecs

// Storage provides an in-memory implementation of Vervet Underground storage.
type Storage struct {
	mu sync.RWMutex

	serviceVersions                   versionedResourceMap
	serviceVersionMappedRevisionSpecs serviceVersionMappedRevisionSpecs

	collatedVersions       vervet.VersionSlice
	collatedVersionedSpecs storage.CollatedVersionMappedSpecs

	newCollator func() *storage.Collator
}

// New returns a new Storage instance.
func New(options ...Option) storage.Storage {
	s := &Storage{
		serviceVersions:                   versionedResourceMap{},
		serviceVersionMappedRevisionSpecs: serviceVersionMappedRevisionSpecs{},

		collatedVersions:       vervet.VersionSlice{},
		collatedVersionedSpecs: storage.CollatedVersionMappedSpecs{},

		newCollator: storage.NewCollator,
	}
	for _, option := range options {
		option(s)
	}
	return s
}

// Option defines a Storage constructor option.
type Option func(*Storage)

// NewCollator configures the Storage instance to use the given constructor
// function for creating collator instances.
func NewCollator(newCollator func() *storage.Collator) Option {
	return func(s *Storage) {
		s.newCollator = newCollator
	}
}

// NotifyVersions implements scraper.Storage.
func (s *Storage) NotifyVersions(ctx context.Context, name string, versions []string, scrapeTime time.Time) error {
	for _, version := range versions {
		// TODO: Add method to fetch contents here
		// TODO: implement notify versions; update sunset when versions are removed
		err := s.NotifyVersion(ctx, name, version, []byte{}, scrapeTime)
		if err != nil {
			return err
		}
	}
	return nil
}

// HasVersion implements scraper.Storage.
func (s *Storage) HasVersion(ctx context.Context, name string, version string, digest string) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	name = storage.GetSantizedHost(name)
	revisions, ok := s.serviceVersionMappedRevisionSpecs[name][version]

	if !ok {
		return false, nil
	}
	_, ok = revisions[storage.Digest(digest)]
	return ok, nil
}

// NotifyVersion implements scraper.Storage.
func (s *Storage) NotifyVersion(ctx context.Context, name string, version string, contents []byte, scrapeTime time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	name = storage.GetSantizedHost(name)
	digest := storage.NewDigest(contents)

	parsedVersion, err := vervet.ParseVersion(version)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to resolve Vervet version for %s : %s", name, version)
		return err
	}

	// Check if service and version structures are initialized
	if _, ok := s.serviceVersionMappedRevisionSpecs[name]; !ok {
		s.serviceVersionMappedRevisionSpecs[name] = versionMappedRevisionSpecs{}
	}

	revisions, ok := s.serviceVersionMappedRevisionSpecs[name][version]
	if ok {
		if _, exist := revisions[digest]; exist {
			return nil
		}
	} else {
		s.serviceVersionMappedRevisionSpecs[name][version] = mappedRevisionSpecs{}
		revisions = s.serviceVersionMappedRevisionSpecs[name][version]
	}

	// If the version is newly initialized, meaning no revisions exist,
	// create the new vervet.VersionSlice with that version initialized
	// else, append it to the existing service's VersionSlice
	if len(revisions) == 0 {
		if _, ok = s.serviceVersions[name]; !ok {
			s.serviceVersions[name] = vervet.VersionSlice{parsedVersion}
		} else {
			s.serviceVersions[name] = append(s.serviceVersions[name], parsedVersion)
			// sort versions when new ones are introduced to maintain BST functionality
			sort.Sort(s.serviceVersions[name])
		}
	}
	// End of initializations

	// TODO: we may want to abstract out the storage objects instead of using chained maps.
	// add the new ContentRevision
	s.serviceVersionMappedRevisionSpecs[name][version][digest] = storage.ContentRevision{
		Service:   name,
		Timestamp: scrapeTime,
		Digest:    digest,
		Blob:      contents,
		Version:   parsedVersion,
	}

	return nil
}

// Versions implements scraper.Storage.
func (s *Storage) Versions() vervet.VersionSlice {
	s.mu.RLock()
	result := make(vervet.VersionSlice, len(s.collatedVersions))
	copy(result, s.collatedVersions)
	s.mu.RUnlock()
	return result
}

// Version implements scraper.Storage.
func (s *Storage) Version(ctx context.Context, version string) ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	parsedVersion, err := vervet.ParseVersion(version)
	if err != nil {
		return nil, err
	}

	spec, ok := s.collatedVersionedSpecs[parsedVersion]
	if !ok {
		resolved, err := s.collatedVersions.Resolve(parsedVersion)
		if err != nil {
			return nil, err
		}
		resolvedSpec := s.collatedVersionedSpecs[resolved]
		return resolvedSpec.MarshalJSON()
	}
	return spec.MarshalJSON()
}

// CollateVersions aggregates versions and revisions from all the services, and produces unified versions and merged specs for all APIs.
func (s *Storage) CollateVersions(ctx context.Context, serviceFilter map[string]bool) error {
	// create an aggregate to process collated data from storage data
	aggregate := s.newCollator()
	for serv, versions := range s.serviceVersionMappedRevisionSpecs {
		if _, ok := serviceFilter[serv]; !ok {
			continue
		}
		for _, revisions := range versions {
			for _, revision := range revisions {
				aggregate.Add(serv, revision)
			}
		}
	}
	versions, specs, err := aggregate.Collate()

	s.mu.Lock()
	defer s.mu.Unlock()
	s.collatedVersions = versions
	s.collatedVersionedSpecs = specs

	return err
}

func (s *Storage) GetCollatedVersionSpecs() (map[string][]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	versionSpecs := map[string][]byte{}
	for key, value := range s.collatedVersionedSpecs {
		json, err := value.MarshalJSON()
		if err != nil {
			return nil, err
		}
		versionSpecs[key.String()] = json
	}
	return versionSpecs, nil
}
