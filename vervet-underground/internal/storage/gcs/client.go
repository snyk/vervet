package gcs

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
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
	c      *storage.Client
	config Config
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
func New(gcsConfig *Config) (vustorage.Storage, error) {
	if gcsConfig == nil || gcsConfig.BucketName == "" {
		return nil, fmt.Errorf("missing GCS configuration")
	}
	var options []option.ClientOption
	if !gcsConfig.IamRoleEnabled {
		options = []option.ClientOption{option.WithEndpoint(gcsConfig.GcsEndpoint)}
		if gcsConfig.Credentials.Filename != "" {
			options = append(options, option.WithCredentialsFile(gcsConfig.Credentials.Filename))
		}
	}

	client, err := storage.NewClient(context.Background(), options...)

	if err != nil {
		log.Error().Err(err).Msg("failed to create client")
		return nil, err
	}

	st := &Storage{client, *gcsConfig}
	err = st.CreateBucket()
	if err != nil {
		return nil, err
	}
	return st, nil
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

// CollateVersions iterates over all possible permutations of Service versions
// to create a unified version spec for each unique vervet.Version.
func (s *Storage) CollateVersions() error {
	// create an aggregate to process collated data from storage data
	aggregate := vustorage.NewCollator()
	serviceRevisionResults, err := s.ListObjects(vustorage.ServiceVersionsFolder, "")
	if err != nil {
		return err
	}

	// all specs are stored as key: "service-versions/{service_name}/{version}/{digest}.json"
	for _, revContent := range serviceRevisionResults {
		service, version, digest, err := parseServiceVersionRevisionKey(revContent.Name)
		if err != nil {
			return err
		}
		rev, obj, err := s.GetObjectWithMetadata(vustorage.ServiceVersionsFolder + service + "/" + version + "/" + digest + ".json")
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
	_, specs, err := aggregate.Collate()
	if err != nil {
		return err
	}

	objects, err := s.PutCollatedSpecs(specs)
	if err != nil {
		return err
	}

	if len(objects) == 0 {
		return fmt.Errorf("objects uploaded length unexpectedly zero. upload_time: %v", time.Now().UTC())
	}

	return nil
}

// HasVersion implements scraper.Storage.
func (s *Storage) HasVersion(name string, version string, digest string) (bool, error) {
	key := getServiceVersionRevisionKey(name, version, digest)
	revisions, err := s.ListObjects(key, "")

	if err != nil {
		return false, err
	}

	// storage.Digest(digest) for the revision present
	return len(revisions) == 1, nil
}

// NotifyVersion updates a Service's storage.ContentRevision if storage.Digest has changed.
func (s *Storage) NotifyVersion(name string, version string, contents []byte, scrapeTime time.Time) error {
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
	serviceVersionRevision, err := s.GetObject(key)
	if err != nil {
		return err
	}

	// if the digest file exists, it counts as a match, no-op, no change
	if len(serviceVersionRevision) > 0 {
		return nil
	}

	// Since the digest doesn't exist, add the whole key path
	reader := bytes.NewReader(currentRevision.Blob)
	_, err = s.PutObject(key, reader)
	if err != nil {
		return err
	}
	return nil
}

// Versions lists all available Collated Versions.
func (s *Storage) Versions() []string {
	prefixes, err := s.ListCollatedVersions()
	if err != nil {
		return nil
	}

	return prefixes
}

// Version retrieves a specific Collated Version.
func (s *Storage) Version(version string) ([]byte, error) {
	_, err := vervet.ParseVersion(version)
	if err != nil {
		return nil, err
	}
	return s.GetCollatedVersionSpec(version)
}

// ListCollatedVersions nice wrapper around the GCS ListObjects request.
// example: key = "collated-versions/"
// result: []string{"2022-02-02~wip", "2022-12-02~beta"}
// Defaults to 1000 results.
func (s *Storage) ListCollatedVersions() ([]string, error) {
	res, err := s.ListObjects(vustorage.CollatedVersionsFolder, "/")

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
func (s *Storage) PutCollatedSpecs(objects map[vervet.Version]openapi3.T) ([]storage.ObjectHandle, error) {
	res := make([]storage.ObjectHandle, 0)
	for key, file := range objects {
		jsonBlob, err := file.MarshalJSON()
		if err != nil {
			return nil, fmt.Errorf("failure to marshal json for collation upload: %w", err)
		}
		reader := bytes.NewReader(jsonBlob)
		r, err := s.PutObject(vustorage.CollatedVersionsFolder+key.String()+"/spec.json", reader)
		if err != nil {
			return nil, err
		}
		res = append(res, *r)
	}

	return res, nil
}

// GetCollatedVersionSpecs retrieves a map of vervet.Version strings
// and their corresponding JSON blobs and returns the result.
func (s *Storage) GetCollatedVersionSpecs() (map[string][]byte, error) {
	versionSpecs := map[string][]byte{}
	versions, err := s.ListCollatedVersions()
	if err != nil {
		return nil, err
	}

	for _, key := range versions {
		jsonBlob, err := s.GetCollatedVersionSpec(key)
		if err != nil {
			return nil, err
		}

		versionSpecs[getCollatedVersionFromKey(key)] = jsonBlob
	}
	return versionSpecs, nil
}

// GetCollatedVersionSpec retrieves a single collated vervet.Version
// and returns the JSON blob.
func (s *Storage) GetCollatedVersionSpec(version string) ([]byte, error) {
	jsonBlob, err := s.GetObject(vustorage.CollatedVersionsFolder + version + "/spec.json")
	if err != nil {
		return nil, err
	}

	return jsonBlob, nil
}

// PutObject nice wrapper around the GCS PutObject request.
func (s *Storage) PutObject(key string, reader io.Reader) (*storage.ObjectHandle, error) {
	ctx := context.Background()
	obj := s.c.Bucket(s.config.BucketName).Object(key)

	ctx, cancel := context.WithTimeout(ctx, time.Second*50)
	defer cancel()

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
func (s *Storage) GetObject(key string) ([]byte, error) {
	reader, err := s.c.Bucket(s.config.BucketName).Object(key).NewReader(context.Background())
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
func (s *Storage) GetObjectWithMetadata(key string) (*storage.Reader, *storage.ObjectAttrs, error) {
	handle := s.c.Bucket(s.config.BucketName).Object(key)
	attrs, err := handle.Attrs(context.Background())
	if err != nil {
		return nil, nil, err
	}
	reader, err := handle.NewReader(context.Background())
	if err != nil {
		return nil, nil, err
	}

	return reader, attrs, err
}

// ListObjects nice wrapper around the GCS storage.BucketHandle Objects request.
// Defaults to 1000 results.
func (s *Storage) ListObjects(key string, delimeter string) ([]storage.ObjectAttrs, error) {
	query := &storage.Query{
		Prefix: key,
	}
	if delimeter == "/" {
		query.Delimiter = "/"
	}
	if query.Prefix == "" {
		query = nil
	}
	it := s.c.Bucket(s.config.BucketName).Objects(context.TODO(), query)
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

	log.Debug().Msgf("GCS Objects response: %+v", r)

	return r, nil
}

// DeleteObject deletes a file if it exists.
func (s *Storage) DeleteObject(key string) error {
	return s.c.Bucket(s.config.BucketName).Object(key).Delete(context.TODO())
}

// CreateBucket idempotently creates an GCS bucket for VU.
func (s *Storage) CreateBucket() error {
	bucket, err := s.getBucketAttrs()
	if err != nil || bucket.Name != s.config.BucketName {
		if !errors.Is(err, storage.ErrBucketNotExist) {
			return err
		}
	}

	err = s.c.Bucket(s.config.BucketName).Create(
		context.Background(),
		s.config.Credentials.ProjectId,
		nil)
	if err != nil {
		return err
	}

	return nil
}

// ListBucketContents lists all available files in a GCS bucket.
func (s *Storage) ListBucketContents() ([]string, error) {
	objects := make([]string, 0)
	it := s.c.Bucket(s.config.BucketName).Objects(context.Background(), &storage.Query{})
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
func (s *Storage) getBucketAttrs() (*storage.BucketAttrs, error) {
	return s.c.Bucket(s.config.BucketName).Attrs(context.TODO())
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
