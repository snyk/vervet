package gcs

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"cloud.google.com/go/storage"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/rs/zerolog/log"
	"github.com/snyk/vervet/v4"
	"go.uber.org/multierr"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"

	vustorage "vervet-underground/internal/storage"
)

// StaticKeyCredentials structure google.Credentials for
// GCS storage.NewClient API.
type StaticKeyCredentials struct {
	ProjectId string
	Filename  string
}

// Config holds setting up and targeting proper GCS targets.
type Config struct {
	GcsRegion      string
	GcsEndpoint    string
	BucketName     string
	IamRoleEnabled bool
	Credentials    StaticKeyCredentials
}

// Storage implements storage.Storage.
type Storage struct {
	mu               sync.RWMutex
	c                *storage.Client
	config           Config
	collatedVersions vervet.VersionSlice
	newCollator      func() *vustorage.Collator
}

/*
New instantiates a gcs.Storage client to handle
storing and retrieving storage.ContentRevision
and Collated Versions.

Please note that the sample server is running with http. If you
want to test this with https you also need to configure Go to skip
certificate validation.
"http://localhost:8080/storage/v1/"
*/
func New(ctx context.Context, gcsConfig *Config, options ...Option) (vustorage.Storage, error) {
	if gcsConfig == nil || gcsConfig.BucketName == "" {
		return nil, fmt.Errorf("missing GCS configuration")
	}
	var clientOptions []option.ClientOption
	if !gcsConfig.IamRoleEnabled {
		clientOptions = []option.ClientOption{option.WithEndpoint(gcsConfig.GcsEndpoint)}
		if gcsConfig.Credentials.Filename != "" {
			clientOptions = append(clientOptions, option.WithCredentialsFile(gcsConfig.Credentials.Filename))
		}
	}

	client, err := storage.NewClient(ctx, clientOptions...)

	if err != nil {
		log.Error().Err(err).Msg("failed to create client")
		return nil, err
	}

	st := &Storage{
		c:                client,
		config:           *gcsConfig,
		collatedVersions: vervet.VersionSlice{},
		newCollator:      vustorage.NewCollator,
	}
	for _, option := range options {
		option(st)
	}
	err = st.CreateBucket(ctx)
	if err != nil {
		return nil, err
	}
	return st, nil
}

// Option defines a Storage constructor option.
type Option func(*Storage)

// NewCollator configures the Storage instance to use the given constructor
// function for creating collator instances.
func NewCollator(newCollator func() *vustorage.Collator) Option {
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

// CollateVersions iterates over all possible permutations of Service versions
// to create a unified version spec for each unique vervet.Version.
func (s *Storage) CollateVersions(ctx context.Context, serviceFilter map[string]bool) error {
	// create an aggregate to process collated data from storage data
	aggregate := s.newCollator()
	serviceRevisionResults, err := s.ListObjects(ctx, vustorage.ServiceVersionsFolder, "")
	if err != nil {
		return err
	}

	// all specs are stored as key: "service-versions/{service_name}/{version}/{digest}.json"
	for _, revContent := range serviceRevisionResults {
		service, version, digest, err := parseServiceVersionRevisionKey(revContent.Name)
		if err != nil {
			return err
		}
		if _, ok := serviceFilter[service]; !ok {
			continue
		}
		rev, obj, err := s.GetObjectWithMetadata(ctx, vustorage.ServiceVersionsFolder+service+"/"+version+"/"+digest+".json")
		if err != nil {
			return err
		}

		// Assuming version is valid in path uploads
		parsedVersion, err := vervet.ParseVersion(version)
		if err != nil {
			log.Error().Err(err).Msg("unexpected version path in GCS. Validate Service Revision uploads")
			return err
		}

		blob, err := io.ReadAll(rev)
		err = multierr.Append(err, rev.Close())
		if err != nil {
			log.Error().Err(err).Msg("failed to read Service ContentRevision JSON")
			return err
		}

		revision := vustorage.ContentRevision{
			Service:   service,
			Version:   parsedVersion,
			Timestamp: obj.Created,
			Digest:    vustorage.Digest(digest),
			Blob:      blob,
		}
		aggregate.Add(service, revision)
	}
	versions, specs, err := aggregate.Collate()
	if err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.collatedVersions = versions

	objects, err := s.PutCollatedSpecs(ctx, specs)
	if err != nil {
		return err
	}

	if len(objects) == 0 {
		return fmt.Errorf("objects uploaded length unexpectedly zero. upload_time: %v", time.Now().UTC())
	}

	return nil
}

// HasVersion implements scraper.Storage.
func (s *Storage) HasVersion(ctx context.Context, name string, version string, digest string) (bool, error) {
	key := getServiceVersionRevisionKey(name, version, digest)
	revisions, err := s.ListObjects(ctx, key, "")

	if err != nil {
		return false, err
	}

	// storage.Digest(digest) for the revision present
	return len(revisions) == 1, nil
}

// NotifyVersion updates a Service's storage.ContentRevision if storage.Digest has changed.
func (s *Storage) NotifyVersion(ctx context.Context, name string, version string, contents []byte, scrapeTime time.Time) error {
	digest := vustorage.NewDigest(contents)
	key := getServiceVersionRevisionKey(name, version, string(digest))
	parsedVersion, err := vervet.ParseVersion(version)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to resolve Vervet version for %s : %s", name, version)
		return err
	}
	currentRevision := vustorage.ContentRevision{
		Service:   name,
		Timestamp: scrapeTime,
		Digest:    digest,
		Blob:      contents,
		Version:   parsedVersion,
	}

	// will return empty in the event of no keys found without an error
	serviceVersionRevision, err := s.GetObject(ctx, key)
	if err != nil {
		return err
	}

	// if the digest file exists, it counts as a match, no-op, no change
	if len(serviceVersionRevision) > 0 {
		return nil
	}

	// Since the digest doesn't exist, add the whole key path
	reader := bytes.NewReader(currentRevision.Blob)
	_, err = s.PutObject(ctx, key, reader)
	if err != nil {
		return err
	}
	return nil
}

// Versions lists all available Collated Versions.
func (s *Storage) Versions() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	stringVersions := make([]string, len(s.collatedVersions))
	for i, version := range s.collatedVersions {
		stringVersions[i] = version.String()
	}

	return stringVersions
}

// Version implements scraper.Storage.
func (s *Storage) Version(ctx context.Context, version string) ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	parsedVersion, err := vervet.ParseVersion(version)
	if err != nil {
		return nil, err
	}

	blob, err := s.GetCollatedVersionSpec(ctx, version)
	if err != nil {
		resolved, err := s.collatedVersions.Resolve(parsedVersion)
		if err != nil {
			return nil, err
		}

		return s.GetCollatedVersionSpec(ctx, resolved.String())
	}
	return blob, nil
}

// ListCollatedVersions nice wrapper around the GCS ListObjects request.
// example: key = "collated-versions/"
// result: []string{"2022-02-02~wip", "2022-12-02~beta"}
// Defaults to 1000 results.
func (s *Storage) ListCollatedVersions(ctx context.Context) ([]string, error) {
	res, err := s.ListObjects(ctx, vustorage.CollatedVersionsFolder, "/")

	if err != nil {
		return nil, err
	}
	prefixes := make([]string, 0)
	for _, v := range res {
		prefixes = append(prefixes, getCollatedVersionFromKey(v.Prefix))
	}

	return prefixes, nil
}

// PutCollatedSpecs iterative wrapper around the GCS PutObject request.
// TODO: Look for alternative to iteratively uploading.
func (s *Storage) PutCollatedSpecs(ctx context.Context, objects map[vervet.Version]openapi3.T) ([]storage.ObjectHandle, error) {
	res := make([]storage.ObjectHandle, 0)
	for key, file := range objects {
		jsonBlob, err := file.MarshalJSON()
		if err != nil {
			return nil, fmt.Errorf("failure to marshal json for collation upload: %w", err)
		}
		reader := bytes.NewReader(jsonBlob)
		r, err := s.PutObject(ctx, vustorage.CollatedVersionsFolder+key.String()+"/spec.json", reader)
		if err != nil {
			return nil, err
		}
		res = append(res, *r)
	}

	return res, nil
}

// GetCollatedVersionSpecs retrieves a map of vervet.Version strings
// and their corresponding JSON blobs and returns the result.
func (s *Storage) GetCollatedVersionSpecs(ctx context.Context) (map[string][]byte, error) {
	versionSpecs := map[string][]byte{}
	versions, err := s.ListCollatedVersions(ctx)
	if err != nil {
		return nil, err
	}

	for _, key := range versions {
		jsonBlob, err := s.GetCollatedVersionSpec(ctx, key)
		if err != nil {
			return nil, err
		}

		versionSpecs[getCollatedVersionFromKey(key)] = jsonBlob
	}
	return versionSpecs, nil
}

// GetCollatedVersionSpec retrieves a single collated vervet.Version
// and returns the JSON blob.
func (s *Storage) GetCollatedVersionSpec(ctx context.Context, version string) ([]byte, error) {
	jsonBlob, err := s.GetObject(ctx, vustorage.CollatedVersionsFolder+version+"/spec.json")
	if err != nil {
		return nil, err
	}

	return jsonBlob, nil
}

// PutObject nice wrapper around the GCS PutObject request.
func (s *Storage) PutObject(ctx context.Context, key string, reader io.Reader) (*storage.ObjectHandle, error) {
	obj := s.c.Bucket(s.config.BucketName).Object(key)
	wc := obj.NewWriter(ctx)
	defer wc.Close()

	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(reader)
	if err != nil {
		return nil, err
	}

	// example []byte("top secret")
	if _, err = wc.Write(buf.Bytes()); err != nil {
		return nil, err
	}

	return obj, nil
}

// GetObject actually retrieves the json blob form GCS.
func (s *Storage) GetObject(ctx context.Context, key string) ([]byte, error) {
	reader, err := s.c.Bucket(s.config.BucketName).Object(key).NewReader(ctx)
	if err != nil {
		if errors.Is(err, storage.ErrObjectNotExist) {
			return nil, nil
		}
		return nil, err
	}

	defer reader.Close()
	return io.ReadAll(reader)
}

// GetObjectWithMetadata actually retrieves the json blob form GCS
// with metadata around the storage in GCS.
func (s *Storage) GetObjectWithMetadata(ctx context.Context, key string) (*storage.Reader, *storage.ObjectAttrs, error) {
	handle := s.c.Bucket(s.config.BucketName).Object(key)
	attrs, err := handle.Attrs(ctx)
	if err != nil {
		return nil, nil, err
	}
	reader, err := handle.NewReader(ctx)
	if err != nil {
		return nil, nil, err
	}

	return reader, attrs, err
}

// ListObjects nice wrapper around the GCS storage.BucketHandle Objects request.
// Defaults to 1000 results.
func (s *Storage) ListObjects(ctx context.Context, key string, delimeter string) ([]storage.ObjectAttrs, error) {
	query := &storage.Query{
		Prefix: key,
	}
	if delimeter == "/" {
		query.Delimiter = "/"
	}
	if query.Prefix == "" {
		query = nil
	}
	it := s.c.Bucket(s.config.BucketName).Objects(ctx, query)
	r := make([]storage.ObjectAttrs, 0)
	for {
		obj, err := it.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return nil, err
		}
		if obj == nil {
			break
		}
		r = append(r, *obj)
	}

	log.Trace().Msgf("GCS Objects response: %+v", r)

	return r, nil
}

// DeleteObject deletes a file if it exists.
func (s *Storage) DeleteObject(ctx context.Context, key string) error {
	return s.c.Bucket(s.config.BucketName).Object(key).Delete(ctx)
}

// CreateBucket idempotently creates an GCS bucket for VU.
func (s *Storage) CreateBucket(ctx context.Context) error {
	bucket, err := s.getBucketAttrs(ctx)
	if err != nil && !errors.Is(err, storage.ErrBucketNotExist) {
		return err
	}
	// if bucket exists, idempotent return
	if bucket != nil {
		return nil
	}

	err = s.c.Bucket(s.config.BucketName).Create(
		ctx,
		s.config.Credentials.ProjectId,
		nil)
	if err != nil {
		return err
	}

	return nil
}

// ListBucketContents lists all available files in a GCS bucket.
func (s *Storage) ListBucketContents(ctx context.Context) ([]string, error) {
	objects := make([]string, 0)
	it := s.c.Bucket(s.config.BucketName).Objects(ctx, &storage.Query{})
	for {
		attrs, err := it.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return nil, err
		}
		objects = append(objects, attrs.Name)
	}
	return objects, nil
}

// getBucketAttrs gets the metadata around a bucket if it exists.
func (s *Storage) getBucketAttrs(ctx context.Context) (*storage.BucketAttrs, error) {
	return s.c.Bucket(s.config.BucketName).Attrs(ctx)
}

// getCollatedVersionFromKey helper function to clean up GCS keys for
// printing versions. Example: collated-versions/2022-02-02/spec.json --> 2022-02-02.
func getCollatedVersionFromKey(key string) string {
	versionPath := strings.Split(strings.TrimPrefix(key, vustorage.CollatedVersionsFolder), "/")
	if len(versionPath) != 2 {
		return key
	}
	return versionPath[0]
}

func getServiceVersionRevisionKey(name string, version string, digest string) string {
	host := vustorage.GetSantizedHost(name)
	return fmt.Sprintf("%v%v/%v/%v.json", vustorage.ServiceVersionsFolder, host, version, digest)
}

func parseServiceVersionRevisionKey(key string) (string, string, string, error) {
	// digest can have "/" chars, so only split for service and version
	arr := strings.SplitN(strings.TrimPrefix(key, vustorage.ServiceVersionsFolder), "/", 3)
	if len(arr) != 3 {
		err := fmt.Errorf("service Content Revision not able to be parsed: %v", key)
		log.Error().Err(err).Msg("GCS service path malformed")
		return "", "", "", err
	}
	service, version, digestJson := arr[0], arr[1], arr[2]
	digest := strings.TrimSuffix(digestJson, ".json")
	return service, version, digest, nil
}
