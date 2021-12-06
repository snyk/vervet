// Package mem provides an in-memory implementation of the storage used in
// Vervet Underground. It's not intended for production use, but as a
// functionally complete reference implementation that can be used to validate
// the other parts of the VU system.
package mem

import (
	"context"
	"sync"
	"time"

	"vervet-underground/internal/storage"
)

type serviceVersion struct {
	service string
	version string
}

type contentRevision struct {
	timestamp time.Time
	digest    string

	// TODO: store the sunset time when a version is removed
	//sunset    *time.Time
}

type serviceVersions map[serviceVersion][]contentRevision

type contents map[serviceVersion]map[string][]byte

// Storage provides an in-memory implementation of Vervet Underground storage.
type Storage struct {
	mu              sync.RWMutex
	serviceVersions serviceVersions
	contents        contents
}

// New returns a new Storage instance.
func New() *Storage {
	return &Storage{
		serviceVersions: serviceVersions{},
		contents:        contents{},
	}
}

// NotifyVersions implements scraper.Storage.
func (s *Storage) NotifyVersions(ctx context.Context, name string, versions []string, scrapeTime time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	// TODO: implement notify versions; update sunset when versions are removed
	return nil
}

// HasVersion implements scraper.Storage.
func (s *Storage) HasVersion(ctx context.Context, name string, version string, digest string) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	digests, ok := s.contents[serviceVersion{service: name, version: version}]
	if !ok {
		return false, nil
	}
	_, ok = digests[digest]
	return ok, nil
}

// NotifyVersion implements scraper.Storage.
func (s *Storage) NotifyVersion(ctx context.Context, name string, version string, contents []byte, scrapeTime time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	k := serviceVersion{service: name, version: version}
	digest := storage.Digest(contents)
	if digests, ok := s.contents[k]; ok {
		if _, ok := digests[digest]; ok {
			return nil
		}
	} else {
		s.contents[k] = map[string][]byte{}
	}
	s.contents[k][digest] = contents
	s.serviceVersions[k] = append(s.serviceVersions[k], contentRevision{
		timestamp: scrapeTime,
		digest:    digest,
	})
	return nil
}
