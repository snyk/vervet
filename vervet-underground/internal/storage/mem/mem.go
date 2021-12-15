// Package mem provides an in-memory implementation of the storage used in
// Vervet Underground. It's not intended for production use, but as a
// functionally complete reference implementation that can be used to validate
// the other parts of the VU system.
package mem

import (
	"sort"
	"sync"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/snyk/vervet"
	"go.uber.org/multierr"

	"vervet-underground/internal/storage"
)

// versionedResourceMap map [service-name] Vervet Version slice array
type versionedResourceMap map[string]vervet.VersionSlice

// mappedRevisionSpecs map [Sha digest of contents string] --> spec contents and metadata
type mappedRevisionSpecs map[storage.Digest]ContentRevision

// collatedVersionMappedSpecs Compiled aggregated spec for all services at that given version
type collatedVersionMappedSpecs map[vervet.Version]openapi3.T

// versionMappedRevisionSpecs map[version-name][digest] --> spec contents and metadata
type versionMappedRevisionSpecs map[string]mappedRevisionSpecs

// serviceVersionMappedRevisionSpecs map[service-name][version-name][digest] --> spec contents and metadata
type serviceVersionMappedRevisionSpecs map[string]versionMappedRevisionSpecs

// ContentRevision is the exact contents and metadata of a service's version at scraping timestamp
type ContentRevision struct {
	serviceVersion string
	timestamp      time.Time
	digest         storage.Digest
	blob           []byte
	// TODO: store the sunset time when a version is removed
	//sunset    *time.Time
}

// Storage provides an in-memory implementation of Vervet Underground storage.
type Storage struct {
	mu sync.RWMutex

	serviceVersions                   versionedResourceMap
	serviceVersionMappedRevisionSpecs serviceVersionMappedRevisionSpecs

	collatedVersions       vervet.VersionSlice
	collatedVersionedSpecs collatedVersionMappedSpecs
}

// New returns a new Storage instance.
func New() *Storage {
	return &Storage{
		serviceVersions:                   versionedResourceMap{},
		serviceVersionMappedRevisionSpecs: serviceVersionMappedRevisionSpecs{},

		collatedVersions:       vervet.VersionSlice{},
		collatedVersionedSpecs: collatedVersionMappedSpecs{},
	}
}

// NotifyVersions implements scraper.Storage.
func (s *Storage) NotifyVersions(name string, versions []string, scrapeTime time.Time) error {
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
func (s *Storage) HasVersion(name string, version string, digest string) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	revisions, ok := s.serviceVersionMappedRevisionSpecs[name][version]

	if !ok {
		return false, nil
	}
	_, ok = revisions[storage.Digest(digest)]
	return ok, nil
}

// NotifyVersion implements scraper.Storage.
func (s *Storage) NotifyVersion(name string, version string, contents []byte, scrapeTime time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()

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
		if _, ok := revisions[digest]; ok {
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
			s.serviceVersions[name] = vervet.VersionSlice{*parsedVersion}
		} else {
			s.serviceVersions[name] = append(s.serviceVersions[name], *parsedVersion)
			// sort versions when new ones are introduced to maintain BST functionality
			sort.Sort(s.serviceVersions[name])
		}
	}
	// End of initializations

	// add the new ContentRevision
	s.serviceVersionMappedRevisionSpecs[name][version][digest] = ContentRevision{
		serviceVersion: name + "_" + version,
		timestamp:      scrapeTime,
		digest:         digest,
		blob:           contents,
	}

	return nil
}

// Versions implements scraper.Storage
func (s *Storage) Versions() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	stringVersions := make([]string, len(s.collatedVersions))
	for i, version := range s.collatedVersions {
		stringVersions[i] = version.String()
	}

	return stringVersions
}

// Version implements scraper.Storage
func (s *Storage) Version(version string) ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	parsedVersion, err := vervet.ParseVersion(version)
	if err != nil {
		return nil, err
	}

	spec := s.collatedVersionedSpecs[*parsedVersion]
	return spec.MarshalJSON()
}

// CollateVersions does the following:
//   - calls updateCollatedVersions for a slice of unique vervet.Version entries
//   - for each unique vervet.Version, run collateVersion to create a compiled VU openapi doc
func (s *Storage) CollateVersions() error {
	var errs error
	s.updateCollatedVersions()
	for _, version := range s.collatedVersions {
		err := s.collateVersion(version)
		if err != nil {
			errs = multierr.Append(errs, errors.Wrapf(err, "failed to collate version %s", version.String()))
		}
	}
	return errs
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

// updateCollatedVersions collects all unique vervet.Versions detected for every service published using
func (s *Storage) updateCollatedVersions() {
	uniqueVersions := make(map[string]vervet.Version)

	// map keys automatically make versions unique rather than iterate through array for each entry
	// service, versionSlice
	for _, versionSlice := range s.serviceVersions {
		// index, version
		for _, version := range versionSlice {
			uniqueVersions[version.String()] = version
		}
	}

	s.mu.Lock()
	s.collatedVersions = make([]vervet.Version, 0)
	for _, value := range uniqueVersions {
		s.collatedVersions = append(s.collatedVersions, value)
	}

	// must sort at end
	sort.Sort(s.collatedVersions)
	s.mu.Unlock()
}

// collateVersion fuzzy matches each service's closest version to the target vervet.Version,
// collects the latest ContentRevision, if one exists and matches the target.
// Once all services have been searched, call mergeContentRevisions
func (s *Storage) collateVersion(version vervet.Version) error {
	// number of services maximum needed
	contentRevisions := make([]ContentRevision, 0)

	s.mu.RLock()
	// preprocessing all relevant docs in byte format before combining
	for service, versionSlice := range s.serviceVersions {
		// If there is an exact match on versions 1-to-1
		var currentRevision ContentRevision
		revisions, ok := s.serviceVersionMappedRevisionSpecs[service][version.String()]
		if ok {
			// TODO: iterate through and take last contentRevision.
			//       Could change to []ContentRevision in struct later
			for _, contentRevision := range revisions {
				currentRevision = contentRevision
			}
			contentRevisions = append(contentRevisions, currentRevision)
		} else {
			// If there is a fuzzy match on version supplied at collation for this service
			// aka execute binarySearch on this service's available versions
			resolvedVersion, err := versionSlice.Resolve(version)
			if err != nil {
				log.Error().Err(err).Msgf("Could not resolve for service %s version %s", service, version.String())
				continue
			}

			revisions, ok = s.serviceVersionMappedRevisionSpecs[service][resolvedVersion.String()]
			if ok {
				// TODO: iterate through and take last contentRevision.
				//       Could change to []ContentRevision in struct later
				for _, contentRevision := range revisions {
					currentRevision = contentRevision
				}
				contentRevisions = append(contentRevisions, currentRevision)
			}
		}
	}
	s.mu.RUnlock()

	err := s.mergeContentRevisions(version, contentRevisions)
	if err != nil {
		log.Error().Err(err).Msg("Could not Merge specs for all services %s")
		return err
	}
	return nil
}

// mergeContentRevisions takes all ContentRevision objects fuzzy matching
// the vervet.Version and combines them into one collated openapi.T doc
// then saves to the shared memory area at the end
func (s *Storage) mergeContentRevisions(version vervet.Version, serviceRevisionCollection []ContentRevision) error {
	loader := openapi3.NewLoader()
	var dst *openapi3.T
	for _, serviceRevision := range serviceRevisionCollection {
		// JSON will deserialize here correctly
		src, err := loader.LoadFromData(serviceRevision.blob)
		if err != nil {
			log.Error().
				Err(err).
				Msgf("Could not merge ServiceRevision %s:%s",
					serviceRevision.serviceVersion,
					serviceRevision.digest)
			return err
		}

		if dst == nil {
			dst = src
		} else {
			// TODO: evaluate whether to use replace bool or not during merging
			vervet.Merge(dst, src, true)
		}
	}

	// lock memory safely here
	s.mu.Lock()
	s.collatedVersionedSpecs[version] = *dst
	s.mu.Unlock()
	return nil
}
