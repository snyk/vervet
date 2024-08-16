// Package disk provides an implementation of the storage used in Vervet
// Underground that uses a local filesystem. It's not intended for production
// use, but as a functionally complete reference implementation that can be
// used to validate the other parts of the VU system.
package disk

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/getkin/kin-openapi/openapi3"

	"github.com/snyk/vervet/v8"
	"github.com/snyk/vervet/v8/internal/storage"
)

type Storage struct {
	path        string
	newCollator func() (*storage.Collator, error)
}

// Option defines a Storage constructor option.
type Option func(*Storage)

type objectMeta struct {
	blob    []byte
	lastMod time.Time
}

func New(path string, options ...Option) storage.Storage {
	s := &Storage{
		path:        path,
		newCollator: func() (*storage.Collator, error) { return storage.NewCollator() },
	}
	for _, option := range options {
		option(s)
	}
	return s
}

func (s *Storage) Cleanup() error {
	if s.path == "" {
		return fmt.Errorf("not cleaning up invalid path")
	}
	return os.RemoveAll(s.path)
}

// NewCollator configures the Storage instance to use the given constructor
// function for creating collator instances.
func NewCollator(newCollator func() (*storage.Collator, error)) Option {
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

// CollateVersions aggregates versions and revisions from all the services, and
// produces unified versions and merged specs for all APIs.
func (s *Storage) CollateVersions(ctx context.Context, serviceFilter map[string]bool) error {
	// create an aggregate to process collated data from storage data
	aggregate, err := s.newCollator()
	if err != nil {
		return err
	}
	serviceRevisions, err := s.ListObjects(ctx, storage.ServiceVersionsFolder)
	if err != nil {
		return err
	}

	// all specs are stored as: "service-versions/{service_name}/{version}/{digest}.json"
	for _, revKey := range serviceRevisions {
		service, version, digest, err := parseServiceVersionRevisionKey(revKey)
		if err != nil {
			return err
		}
		if _, ok := serviceFilter[service]; !ok {
			continue
		}
		rev, err := s.GetObjectWithMetadata(revKey)
		if err != nil {
			return err
		}

		// Assuming version is valid in path uploads
		parsedVersion, err := vervet.ParseVersion(version)
		if err != nil {
			return err
		}

		revision := storage.ContentRevision{
			Service:   service,
			Version:   parsedVersion,
			Timestamp: rev.lastMod,
			Digest:    storage.Digest(digest),
			Blob:      rev.blob,
		}
		aggregate.Add(service, revision)
	}
	specs, err := aggregate.Collate()
	if err != nil {
		return err
	}

	return s.putCollatedSpecs(specs)
}

// HasVersion implements scraper.Storage.
func (s *Storage) HasVersion(ctx context.Context, name string, version string, digest string) (bool, error) {
	key := s.getServiceVersionRevisionKey(name, version, digest)
	path := path.Join(s.path, key)
	_, err := os.Stat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// NotifyVersion implements scraper.Storage.
func (s *Storage) NotifyVersion(ctx context.Context,
	name string,
	version string,
	contents []byte,
	scrapeTime time.Time,
) error {
	digest := storage.NewDigest(contents)
	key := s.getServiceVersionRevisionKey(name, version, string(digest))
	parsedVersion, err := vervet.ParseVersion(version)
	if err != nil {
		return err
	}

	currentRevision := storage.ContentRevision{
		Service:   name,
		Timestamp: scrapeTime,
		Digest:    digest,
		Blob:      contents,
		Version:   parsedVersion,
	}

	_, err = s.GetObject(key)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			// Since the digest doesn't exist, add the whole key path
			return s.PutObject(key, currentRevision.Blob, &currentRevision.Timestamp)
		}
		return err
	}
	// digest already exists, nothing to do
	return nil
}

// VersionIndex implements scraper.Storage.
func (s *Storage) VersionIndex(ctx context.Context) (vervet.VersionIndex, error) {
	objects, err := s.ListObjects(ctx, storage.ServiceVersionsFolder)
	if err != nil {
		return vervet.VersionIndex{}, err
	}
	vs := make(vervet.VersionSlice, len(objects))
	for idx, obj := range objects {
		_, versionStr, _, err := parseServiceVersionRevisionKey(obj)
		if err != nil {
			return vervet.VersionIndex{}, err
		}
		vs[idx], err = vervet.ParseVersion(versionStr)
		if err != nil {
			return vervet.VersionIndex{}, err
		}
	}

	return vervet.NewVersionIndex(vs), nil
}

// Version implements scraper.Storage.
func (s *Storage) Version(ctx context.Context, version string) ([]byte, error) {
	parsedVersion, err := vervet.ParseVersion(version)
	if err != nil {
		return nil, err
	}

	blob, err := s.GetCollatedVersionSpec(version)
	if err != nil {
		index, err := s.VersionIndex(ctx)
		if err != nil {
			return nil, err
		}
		resolved, err := index.Resolve(parsedVersion)
		if err != nil {
			return nil, err
		}
		return s.GetCollatedVersionSpec(resolved.String())
	}
	return blob, nil
}

func (s *Storage) getServiceVersionRevisionKey(name string, version string, digest string) string {
	// digest could contain slashes
	b64 := base64.StdEncoding.EncodeToString([]byte(digest))
	return path.Join(storage.ServiceVersionsFolder, name, version, b64) + ".json"
}

func (s *Storage) GetObject(key string) ([]byte, error) {
	path := path.Join(s.path, key)
	return os.ReadFile(path)
}

func (s *Storage) PutObject(key string, body []byte, timestamp *time.Time) error {
	path := path.Join(s.path, key)
	err := os.MkdirAll(filepath.Dir(path), os.ModePerm)
	if err != nil {
		return err
	}
	err = os.WriteFile(path, body, 0600)
	if err != nil {
		return err
	}
	if timestamp == nil {
		return nil
	}
	// Preserve specified timestamp, mostly for testing using a monotonic clock
	// on fast file systems.
	return os.Chtimes(path, *timestamp, *timestamp)
}

// GetCollatedVersionSpec retrieves a single collated vervet.Version
// and returns the JSON blob.
func (s *Storage) GetCollatedVersionSpec(version string) ([]byte, error) {
	path := path.Join(storage.CollatedVersionsFolder, version, "spec.json")
	return s.GetObject(path)
}

// ListObjects gets all objects under a given directory.
func (s *Storage) ListObjects(ctx context.Context, key string) ([]string, error) {
	path := path.Join(s.path, key)
	objects := make([]string, 0)
	err := filepath.Walk(path, func(obj string, info os.FileInfo, err error) error {
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return nil
			}
			return err
		}
		if !info.IsDir() {
			objects = append(objects, obj)
		}
		return nil
	})
	return objects, err
}

func parseServiceVersionRevisionKey(key string) (string, string, string, error) {
	digestB64 := filepath.Base(key)
	digestB64 = strings.TrimSuffix(digestB64, ".json")
	digest, err := base64.StdEncoding.DecodeString(digestB64)
	if err != nil {
		return "", "", "", err
	}
	rest := filepath.Dir(key)
	version := filepath.Base(rest)
	rest = filepath.Dir(rest)
	service := filepath.Base(rest)

	return service, version, string(digest), nil
}

func (s *Storage) GetObjectWithMetadata(key string) (*objectMeta, error) {
	info, err := os.Stat(key)
	if err != nil {
		return nil, err
	}
	lastMod := info.ModTime()
	body, err := os.ReadFile(key)
	return &objectMeta{
		lastMod: lastMod,
		blob:    body,
	}, err
}

// putCollatedSpecs stores the given collated OpenAPI document objects.
func (s *Storage) putCollatedSpecs(objects map[vervet.Version]openapi3.T) error {
	for key, file := range objects {
		jsonBlob, err := file.MarshalJSON()
		if err != nil {
			return fmt.Errorf("failed to marshal json for collation upload: %w", err)
		}
		err = s.PutObject(storage.CollatedVersionsFolder+key.String()+"/spec.json", jsonBlob, nil)
		if err != nil {
			return err
		}
	}
	return nil
}
