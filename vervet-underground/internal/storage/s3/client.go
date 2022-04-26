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
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
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
	client *s3.Client
	config Config
}

func New(awsCfg *Config) (storage.Storage, error) {
	if awsCfg == nil || awsCfg.BucketName == "" {
		return nil, fmt.Errorf("missing S3 configuration")
	}

	var options []func(*config.LoadOptions) error
	/*
		Secrets should come from volume or IamRole
		localstack defaults to static credentials for local dev
		awsRegion = os.Getenv("AWS_REGION")
		awsEndpoint = os.Getenv("AWS_ENDPOINT")
		bucketName = os.Getenv("S3_BUCKET")
	*/
	if !awsCfg.IamRoleEnabled {
		customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
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

		options = []func(*config.LoadOptions) error{config.WithRegion(awsCfg.AwsRegion),
			config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
				awsCfg.Credentials.AccessKey,
				awsCfg.Credentials.SecretKey,
				awsCfg.Credentials.SessionKey)),
			config.WithEndpointResolverWithOptions(customResolver),
		}
	}
	awsCfgLoader, err := config.LoadDefaultConfig(context.Background(), options...)

	if err != nil {
		log.Error().Err(err).Msg("Cannot load the AWS configs")
		return nil, err
	}

	// Create the resource client
	s3Client := s3.NewFromConfig(awsCfgLoader, func(o *s3.Options) {
		o.UsePathStyle = true
	})
	st := &Storage{client: s3Client, config: *awsCfg}
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

// HasVersion implements scraper.Storage.
func (s *Storage) HasVersion(name string, version string, digest string) (bool, error) {
	key := getServiceVersionRevisionKey(name, version, digest)
	revisions, err := s.ListObjects(key, "")

	if err != nil {
		return false, err
	}

	// storage.Digest(digest) for the revision present
	return len(revisions.Contents) == 1, nil
}

// NotifyVersion implements scraper.Storage.
func (s *Storage) NotifyVersion(name string, version string, contents []byte, scrapeTime time.Time) error {
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

// Versions implements scraper.Storage.
func (s *Storage) Versions() []string {
	prefixes, err := s.ListCollatedVersions()
	if err != nil {
		return nil
	}

	return prefixes
}

// Version implements scraper.Storage.
func (s *Storage) Version(version string) ([]byte, error) {
	_, err := vervet.ParseVersion(version)
	if err != nil {
		return nil, err
	}
	return s.GetCollatedVersionSpec(version)
}

// CollateVersions aggregates versions and revisions from all the services, and produces unified versions and merged specs for all APIs.
func (s *Storage) CollateVersions() error {
	// create an aggregate to process collated data from storage data
	aggregate := storage.NewCollator()
	serviceRevisionResults, err := s.ListObjects(storage.ServiceVersionsFolder, "")
	if err != nil {
		return err
	}

	// all specs are stored as key: "service-versions/{service_name}/{version}/{digest}.json"
	for _, revContent := range serviceRevisionResults.Contents {
		service, version, digest, err := parseServiceVersionRevisionKey(*revContent.Key)
		if err != nil {
			return err
		}
		rev, err := s.GetObjectWithMetadata(storage.ServiceVersionsFolder + service + "/" + version + "/" + digest + ".json")
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

	return err
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
	jsonBlob, err := s.GetObject(storage.CollatedVersionsFolder + version + "/spec.json")
	if err != nil {
		return nil, err
	}

	return jsonBlob, nil
}

// PutObject nice wrapper around the S3 PutObject request.
func (s *Storage) PutObject(key string, reader io.Reader) (*s3.PutObjectOutput, error) {
	p := s3.PutObjectInput{
		Bucket: aws.String(s.config.BucketName),
		Key:    aws.String(key),
		ACL:    types.ObjectCannedACLPublicRead,
		Body:   reader,
	}

	r, err := s.client.PutObject(context.Background(), &p)
	log.Trace().Msgf("S3 PutObject response: %+v", r)
	if smith := handleAwsError(err); smith != nil {
		return nil, smith
	}

	return r, err
}

// PutCollatedSpecs iterative wrapper around the S3 PutObject request.
// TODO: Look for alternative to iteratively uploading.
func (s *Storage) PutCollatedSpecs(objects map[vervet.Version]openapi3.T) (res []s3.PutObjectOutput, smith error) {
	res = make([]s3.PutObjectOutput, 0)
	for key, file := range objects {
		jsonBlob, err := file.MarshalJSON()
		if err != nil {
			return nil, fmt.Errorf("failure to marshal json for collation upload: %w", err)
		}
		reader := bytes.NewReader(jsonBlob)
		r, err := s.PutObject(storage.CollatedVersionsFolder+key.String()+"/spec.json", reader)
		if smith = handleAwsError(err); smith != nil {
			return nil, smith
		}
		res = append(res, *r)
	}

	return res, smith
}

// GetObject nice wrapper around the S3 GetObject request.
func (s *Storage) GetObject(key string) ([]byte, error) {
	p := s3.GetObjectInput{
		Bucket: aws.String(s.config.BucketName),
		Key:    aws.String(key),
	}

	r, err := s.client.GetObject(context.Background(), &p)
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
func (s *Storage) GetObjectWithMetadata(key string) (*s3.GetObjectOutput, error) {
	p := s3.GetObjectInput{
		Bucket: aws.String(s.config.BucketName),
		Key:    aws.String(key),
	}

	r, err := s.client.GetObject(context.Background(), &p)
	if smith := handleAwsError(err); smith != nil {
		return nil, smith
	}

	return r, nil
}

// DeleteObject nice wrapper around the S3 DeleteObject request.
func (s *Storage) DeleteObject(key string) error {
	p := s3.DeleteObjectInput{
		Bucket: aws.String(s.config.BucketName),
		Key:    aws.String(key),
	}

	r, err := s.client.DeleteObject(context.Background(), &p)
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
func (s *Storage) ListCollatedVersions() ([]string, error) {
	res, err := s.ListCommonPrefixes(storage.CollatedVersionsFolder)

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
func (s *Storage) ListCommonPrefixes(key string) ([]types.CommonPrefix, error) {
	r, err := s.ListObjects(key, "/")
	if err != nil {
		return nil, err
	}
	return r.CommonPrefixes, nil
}

// ListObjects nice wrapper around the S3 ListObjects request.
// "collated-versions" example.
// Defaults to 1000 results.
func (s *Storage) ListObjects(key string, delimeter string) (*s3.ListObjectsV2Output, error) {
	p := s3.ListObjectsV2Input{
		Bucket: aws.String(s.config.BucketName),
		Prefix: aws.String(key),
	}

	if delimeter != "" {
		p.Delimiter = aws.String("/")
	}

	r, err := s.client.ListObjectsV2(context.Background(), &p)
	log.Trace().Msgf("S3 ListObject response: %+v", r)
	if smith := handleAwsError(err); smith != nil {
		return nil, smith
	}

	return r, nil
}

// CreateBucket idempotently creates an S3 bucket for VU.
func (s *Storage) CreateBucket() error {
	create := &s3.CreateBucketInput{
		Bucket: aws.String(s.config.BucketName),
	}

	bucketOutput, err := s.client.ListBuckets(context.TODO(), &s3.ListBucketsInput{})
	if smith := handleAwsError(err); smith != nil {
		return smith
	}

	exists := false
	for _, bucket := range bucketOutput.Buckets {
		if *bucket.Name == s.config.BucketName {
			exists = true
			break
		}
	}

	if !exists {
		bucket, err := s.client.CreateBucket(context.TODO(), create)
		if smith := handleAwsError(err); smith != nil {
			return smith
		}
		if *bucket.Location == "" {
			return fmt.Errorf("invalid bucket output")
		}
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
	var opErr *smithy.OperationError
	var apiErr smithy.APIError
	if errors.As(err, &opErr) {
		log.Error().Err(err).Msgf("failed to call service: %s, operation: %s, error: %v",
			opErr.Service(),
			opErr.Operation(),
			opErr.Unwrap())
	}

	if opErr != nil {
		err := opErr.Unwrap()
		if errors.As(err, &apiErr) {
			switch apiErr.ErrorCode() {
			case "NoSuchKey":
				return nil
			default:
				return apiErr
			}
		}
	}

	return err
}
