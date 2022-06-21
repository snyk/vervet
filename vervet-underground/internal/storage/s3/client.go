// Package s3 provides an implementation of Vervet Underground storage backed
// by Amazon S3.
package s3

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awshttp "github.com/aws/aws-sdk-go-v2/aws/transport/http"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/rs/zerolog/log"
	"github.com/snyk/vervet/v4"
	"go.uber.org/multierr"

	"vervet-underground/internal/storage"
)

// StaticKeyCredentials Defines credential structure used in config.LoadDefaultConfig.
type StaticKeyCredentials struct {
	AccessKey  string
	SecretKey  string
	SessionKey string
}

// Config Defines S3 client target used in config.LoadDefaultConfig.
type Config struct {
	AwsRegion      string
	AwsEndpoint    string
	BucketName     string
	Credentials    StaticKeyCredentials
	IamRoleEnabled bool
}

type Storage struct {
	mu               sync.RWMutex
	client           *s3.Client
	config           Config
	collatedVersions vervet.VersionSlice
	newCollator      func() *storage.Collator
}

func New(ctx context.Context, awsCfg *Config, options ...Option) (storage.Storage, error) {
	if awsCfg == nil || awsCfg.BucketName == "" {
		return nil, fmt.Errorf("missing S3 configuration")
	}

	var loadOptions []func(*config.LoadOptions) error
	/*
		Secrets should come from AWS injected IAM Role volume
		localstack defaults to static credentials for local dev
		awsRegion = os.Getenv("AWS_REGION")
		awsEndpoint = os.Getenv("AWS_ENDPOINT")
		bucketName = os.Getenv("S3_BUCKET")
	*/
	if !awsCfg.IamRoleEnabled {
		customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, loadOptions ...interface{}) (aws.Endpoint, error) {
			if awsCfg.AwsEndpoint != "" {
				return aws.Endpoint{
					PartitionID:   "aws",
					URL:           awsCfg.AwsEndpoint,
					SigningRegion: awsCfg.AwsRegion,
				}, nil
			}

			// returning EndpointNotFoundError will allow the service to fallback to its default resolution
			return aws.Endpoint{}, &aws.EndpointNotFoundError{}
		})

		loadOptions = append(loadOptions, config.WithRegion(awsCfg.AwsRegion),
			config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
				awsCfg.Credentials.AccessKey,
				awsCfg.Credentials.SecretKey,
				awsCfg.Credentials.SessionKey)),
			config.WithEndpointResolverWithOptions(customResolver))
	}

	awsCfgLoader, err := config.LoadDefaultConfig(ctx, loadOptions...)

	if err != nil {
		log.Error().Err(err).Msg("Cannot load the AWS configs")
		return nil, err
	}

	// Create the resource client
	s3Client := s3.NewFromConfig(awsCfgLoader, func(o *s3.Options) {
		o.UsePathStyle = true
	})
	st := &Storage{
		client:           s3Client,
		config:           *awsCfg,
		collatedVersions: vervet.VersionSlice{},
		newCollator:      storage.NewCollator,
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
	key := getServiceVersionRevisionKey(name, version, digest)
	revisions, err := s.ListObjects(ctx, key, "")

	if err != nil {
		return false, err
	}

	// storage.Digest(digest) for the revision present
	return len(revisions.Contents) == 1, nil
}

// NotifyVersion implements scraper.Storage.
func (s *Storage) NotifyVersion(ctx context.Context, name string, version string, contents []byte, scrapeTime time.Time) error {
	digest := storage.NewDigest(contents)
	key := getServiceVersionRevisionKey(name, version, string(digest))
	parsedVersion, err := vervet.ParseVersion(version)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to resolve Vervet version for %s : %s", name, version)
		return err
	}
	currentRevision := storage.ContentRevision{
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

// Versions implements scraper.Storage.
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

// CollateVersions aggregates versions and revisions from all the services, and produces unified versions and merged specs for all APIs.
func (s *Storage) CollateVersions(ctx context.Context) error {
	// create an aggregate to process collated data from storage data
	aggregate := s.newCollator()
	serviceRevisionResults, err := s.ListObjects(ctx, storage.ServiceVersionsFolder, "")
	if err != nil {
		return err
	}

	// all specs are stored as key: "service-versions/{service_name}/{version}/{digest}.json"
	for _, revContent := range serviceRevisionResults.Contents {
		service, version, digest, err := parseServiceVersionRevisionKey(*revContent.Key)
		if err != nil {
			return err
		}
		rev, err := s.GetObjectWithMetadata(ctx, storage.ServiceVersionsFolder+service+"/"+version+"/"+digest+".json")
		if err != nil {
			return err
		}

		// Assuming version is valid in path uploads
		parsedVersion, err := vervet.ParseVersion(version)
		if err != nil {
			log.Error().Err(err).Msg("unexpected version path in S3. Validate Service Revision uploads")
			return err
		}

		blob, err := io.ReadAll(rev.Body)
		err = multierr.Append(err, rev.Body.Close())
		if err != nil {
			log.Error().Err(err).Msg("failed to read Service ContentRevision JSON")
			return err
		}

		revision := storage.ContentRevision{
			Service:   service,
			Version:   parsedVersion,
			Timestamp: *rev.LastModified,
			Digest:    storage.Digest(digest),
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

	return err
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
	jsonBlob, err := s.GetObject(ctx, storage.CollatedVersionsFolder+version+"/spec.json")
	if err != nil {
		return nil, err
	}

	return jsonBlob, nil
}

// PutObject nice wrapper around the S3 PutObject request.
func (s *Storage) PutObject(ctx context.Context, key string, reader io.Reader) (*s3.PutObjectOutput, error) {
	p := s3.PutObjectInput{
		Bucket: aws.String(s.config.BucketName),
		Key:    aws.String(key),
		ACL:    types.ObjectCannedACLPublicRead,
		Body:   reader,
	}

	r, err := s.client.PutObject(ctx, &p)
	log.Trace().Msgf("S3 PutObject response: %+v", r)
	if smith := handleAwsError(err); smith != nil {
		return nil, smith
	}

	return r, err
}

// PutCollatedSpecs iterative wrapper around the S3 PutObject request.
// TODO: Look for alternative to iteratively uploading.
func (s *Storage) PutCollatedSpecs(ctx context.Context, objects map[vervet.Version]openapi3.T) (res []s3.PutObjectOutput, smith error) {
	res = make([]s3.PutObjectOutput, 0)
	for key, file := range objects {
		jsonBlob, err := file.MarshalJSON()
		if err != nil {
			return nil, fmt.Errorf("failure to marshal json for collation upload: %w", err)
		}
		reader := bytes.NewReader(jsonBlob)
		r, err := s.PutObject(ctx, storage.CollatedVersionsFolder+key.String()+"/spec.json", reader)
		if smith = handleAwsError(err); smith != nil {
			return nil, smith
		}
		res = append(res, *r)
	}

	return res, smith
}

// GetObject nice wrapper around the S3 GetObject request.
func (s *Storage) GetObject(ctx context.Context, key string) ([]byte, error) {
	p := s3.GetObjectInput{
		Bucket: aws.String(s.config.BucketName),
		Key:    aws.String(key),
	}

	r, err := s.client.GetObject(ctx, &p)
	if smith := handleAwsError(err); smith != nil {
		return nil, smith
	}

	if r != nil {
		return io.ReadAll(r.Body)
	}
	return nil, nil
}

// GetObjectWithMetadata nice wrapper around the S3 GetObject request.
// Returns metadata as well.
func (s *Storage) GetObjectWithMetadata(ctx context.Context, key string) (*s3.GetObjectOutput, error) {
	p := s3.GetObjectInput{
		Bucket: aws.String(s.config.BucketName),
		Key:    aws.String(key),
	}

	r, err := s.client.GetObject(ctx, &p)
	if smith := handleAwsError(err); smith != nil {
		return nil, smith
	}

	return r, nil
}

// DeleteObject nice wrapper around the S3 DeleteObject request.
func (s *Storage) DeleteObject(ctx context.Context, key string) error {
	p := s3.DeleteObjectInput{
		Bucket: aws.String(s.config.BucketName),
		Key:    aws.String(key),
	}

	r, err := s.client.DeleteObject(ctx, &p)
	log.Trace().Msgf("S3 DeleteObject response: %+v", r)
	if err != nil {
		return err
	}

	return nil
}

// ListCollatedVersions nice wrapper around the S3 ListCommonPrefixes request.
// example: key = "collated-versions/"
// result: []string{"2022-02-02~wip", "2022-12-02~beta"}
// Defaults to 1000 results.
func (s *Storage) ListCollatedVersions(ctx context.Context) ([]string, error) {
	res, err := s.ListCommonPrefixes(ctx, storage.CollatedVersionsFolder)

	if err != nil {
		return nil, err
	}
	var prefixes []string
	for _, v := range res {
		if v.Prefix != nil {
			prefixes = append(prefixes, getCollatedVersionFromKey(*v.Prefix))
		}
	}

	return prefixes, nil
}

// ListCommonPrefixes nice wrapper around the S3 ListCommonPrefixes request.
// example: key = "collated-versions/"
// result: []types.CommonPrefix{"collated-versions/2022-02-02~wip/", "collated-versions/2022-12-02~beta/"}
// Defaults to 1000 results.
func (s *Storage) ListCommonPrefixes(ctx context.Context, key string) ([]types.CommonPrefix, error) {
	r, err := s.ListObjects(ctx, key, "/")
	if err != nil {
		return nil, err
	}
	return r.CommonPrefixes, nil
}

// ListObjects nice wrapper around the S3 ListObjects request.
// "collated-versions" example.
// Defaults to 1000 results.
func (s *Storage) ListObjects(ctx context.Context, key string, delimeter string) (*s3.ListObjectsV2Output, error) {
	p := s3.ListObjectsV2Input{
		Bucket: aws.String(s.config.BucketName),
		Prefix: aws.String(key),
	}

	if delimeter != "" {
		p.Delimiter = aws.String("/")
	}

	r, err := s.client.ListObjectsV2(ctx, &p)
	log.Trace().Msgf("S3 ListObject response: %+v", r)
	if smith := handleAwsError(err); smith != nil {
		return nil, smith
	}

	return r, nil
}

// CreateBucket idempotently creates an S3 bucket for VU.
func (s *Storage) CreateBucket(ctx context.Context) error {
	create := &s3.CreateBucketInput{
		Bucket: aws.String(s.config.BucketName),
	}
	input := &s3.HeadBucketInput{
		Bucket: aws.String(s.config.BucketName),
	}

	_, err := s.client.HeadBucket(ctx, input)
	smith := handleAwsError(err)
	if smith == nil {
		return nil
	}
	log.Info().
		Err(err).
		Msg("Head bucket check failed, continuing to bucket creation")

	bucket, err := s.client.CreateBucket(ctx, create)
	if smith := handleAwsError(err); smith != nil {
		return smith
	}
	if *bucket.Location == "" {
		return fmt.Errorf("invalid bucket output")
	}
	return nil
}

// getCollatedVersionFromKey helper function to clean up S3 keys for
// printing versions. Example: collated-versions/2022-02-02/spec.json --> 2022-02-02.
func getCollatedVersionFromKey(key string) string {
	versionPath := strings.Split(strings.TrimPrefix(key, storage.CollatedVersionsFolder), "/")
	if len(versionPath) != 2 {
		return key
	}
	return versionPath[0]
}

func getServiceVersionRevisionKey(name string, version string, digest string) string {
	host := storage.GetSantizedHost(name)
	return fmt.Sprintf("%v%v/%v/%v.json", storage.ServiceVersionsFolder, host, version, digest)
}

func parseServiceVersionRevisionKey(key string) (string, string, string, error) {
	// digest can have "/" chars, so only split for service and version
	arr := strings.SplitN(strings.TrimPrefix(key, storage.ServiceVersionsFolder), "/", 3)
	if len(arr) != 3 {
		err := fmt.Errorf("service Content Revision not able to be parsed: %v", key)
		log.Error().Err(err).Msg("s3 service path malformed")
		return "", "", "", err
	}
	service, version, digestJson := arr[0], arr[1], arr[2]
	digest := strings.TrimSuffix(digestJson, ".json")
	return service, version, digest, nil
}

/*
handleAwsError parses resulting S3 Operations to view
specific failure types to handle 404s without problems,
and avoid red herring errors during processing.

Casting to the awserr.Error type will allow you to inspect the error
code returned by the service in code. The error code can be used
to switch on context specific functionality. In this case a context
specific error message is printed to the user based on the bucket
and key existing.
For information on other S3 API error codes see:
https://aws.github.io/aws-sdk-go-v2/docs/handling-errors/
*/
func handleAwsError(err error) error {
	var apiErr smithy.APIError
	var re *awshttp.ResponseError
	fault := "client"

	_ = errors.As(err, &apiErr)
	if apiErr != nil {
		fault = apiErr.ErrorFault().String()
	}
	if errors.As(err, &re) {
		log.Error().Err(re).
			Str("service_request_id", re.ServiceRequestID()).
			Int("status_code", re.HTTPStatusCode()).
			Str("smithy_fault", fault).
			Str("request_url", re.HTTPResponse().Request.URL.String()).
			Str("request_method", re.HTTPResponse().Request.Method).
			Msg("S3 call failed")
		switch re.HTTPStatusCode() {
		case 404:
			return nil
		default:
			return re
		}
	}

	return err
}
