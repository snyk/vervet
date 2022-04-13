// Package mem provides an in-memory implementation of the storage used in
// Vervet Underground. It's not intended for production use, but as a
// functionally complete reference implementation that can be used to validate
// the other parts of the VU system.
package mem

import (
	"sort"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/snyk/vervet/v4"

	"vervet-underground/internal/storage"
)

// Store provides an in-memory implementation of Vervet Underground storage.
type Aggregate struct {
	client *Client
}

type Client struct {
	mu sync.RWMutex

	serviceVersions                   versionedResourceMap
	serviceVersionMappedRevisionSpecs serviceVersionMappedRevisionSpecs

	collatedVersions       vervet.VersionSlice
	collatedVersionedSpecs collatedVersionMappedSpecs
}

// New returns a new Storage instance.
func New() *Aggregate {
	return &Aggregate{
		&Client{
			serviceVersions:                   versionedResourceMap{},
			serviceVersionMappedRevisionSpecs: serviceVersionMappedRevisionSpecs{},

			collatedVersions:       vervet.VersionSlice{},
			collatedVersionedSpecs: collatedVersionMappedSpecs{},
		},
	}
}

// NotifyVersions implements scraper.Storage.
func (s *Aggregate) NotifyVersions(name string, versions []string, scrapeTime time.Time) error {
	for _, version := range versions {
		// TODO: Add method to fetch contents here
		// TODO: implement notify versions; update sunset when versions are removed
		err := s.NotifyVersion(name, version, []byte{}, scrapeTime)
		if err != nil {
			return err
		}
	}
	return nil
}

// HasVersion implements scraper.Storage.
func (s *Aggregate) HasVersion(name string, version string, digest string) (bool, error) {
	s.client.mu.RLock()
	defer s.client.mu.RUnlock()
	revisions, ok := s.client.serviceVersionMappedRevisionSpecs[name][version]

	if !ok {
		return false, nil
	}
	_, ok = revisions[storage.Digest(digest)]
	return ok, nil
}

// NotifyVersion implements scraper.Storage.
func (s *Aggregate) NotifyVersion(name string, version string, contents []byte, scrapeTime time.Time) error {
	s.client.mu.Lock()
	defer s.client.mu.Unlock()

	digest := storage.NewDigest(contents)

	parsedVersion, err := vervet.ParseVersion(version)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to resolve Vervet version for %s : %s", name, version)
		return err
	}

	// Check if service and version structures are initialized
	if _, ok := s.client.serviceVersionMappedRevisionSpecs[name]; !ok {
		s.client.serviceVersionMappedRevisionSpecs[name] = versionMappedRevisionSpecs{}
	}

	revisions, ok := s.client.serviceVersionMappedRevisionSpecs[name][version]
	if ok {
		if _, exist := revisions[digest]; exist {
			return nil
		}
	} else {
		s.client.serviceVersionMappedRevisionSpecs[name][version] = mappedRevisionSpecs{}
		revisions = s.client.serviceVersionMappedRevisionSpecs[name][version]
	}

	// If the version is newly initialized, meaning no revisions exist,
	// create the new vervet.VersionSlice with that version initialized
	// else, append it to the existing service's VersionSlice
	if len(revisions) == 0 {
		if _, ok = s.client.serviceVersions[name]; !ok {
			s.client.serviceVersions[name] = vervet.VersionSlice{parsedVersion}
		} else {
			s.client.serviceVersions[name] = append(s.client.serviceVersions[name], parsedVersion)
			// sort versions when new ones are introduced to maintain BST functionality
			sort.Sort(s.client.serviceVersions[name])
		}
	}
	// End of initializations

	// TODO: we may want to abstract out the storage objects instead of using chained maps.
	// add the new ContentRevision
	s.client.serviceVersionMappedRevisionSpecs[name][version][digest] = storage.ContentRevision{
		Service:   name,
		Timestamp: scrapeTime,
		Digest:    digest,
		Blob:      contents,
		Version:   parsedVersion,
	}

	return nil
}

// Versions implements scraper.Storage.
func (s *Aggregate) Versions() []string {
	s.client.mu.Lock()
	defer s.client.mu.Unlock()
	stringVersions := make([]string, len(s.client.collatedVersions))
	for i, version := range s.client.collatedVersions {
		stringVersions[i] = version.String()
	}

	return stringVersions
}

// Version implements scraper.Storage.
func (s *Aggregate) Version(version string) ([]byte, error) {
	s.client.mu.Lock()
	defer s.client.mu.Unlock()

	parsedVersion, err := vervet.ParseVersion(version)
	if err != nil {
		return nil, err
	}

	spec := s.client.collatedVersionedSpecs[parsedVersion]
	return spec.MarshalJSON()
}

// CollateVersions aggregates versions and revisions from all the services, and produces unified versions and merged specs for all APIs.
func (s *Aggregate) CollateVersions() error {
	// create an aggregate to process collated data from storage data
	aggregate := storage.NewCollator()
	for serv, versions := range s.client.serviceVersionMappedRevisionSpecs {
		for _, revisions := range versions {
			for _, revision := range revisions {
				aggregate.Add(serv, revision)
			}
		}
	}
	versions, specs, err := aggregate.Collate()

	s.client.mu.Lock()
	defer s.client.mu.Unlock()
	s.client.collatedVersions = versions
	s.client.collatedVersionedSpecs = specs

	return err
}

func (s *Aggregate) GetCollatedVersionSpecs() (map[string][]byte, error) {
	s.client.mu.Lock()
	defer s.client.mu.Unlock()

	versionSpecs := map[string][]byte{}
	for key, value := range s.client.collatedVersionedSpecs {
		json, err := value.MarshalJSON()
		if err != nil {
			return nil, err
		}
		versionSpecs[key.String()] = json
	}
	return versionSpecs, nil
}
