// Package s3 provides an implementation of Vervet Underground storage backed
// by Amazon S3.
package s3

import (
	"context"
	"fmt"
	"io"
	"log"
	"vervet-underground/internal/storage"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

type Client struct {
	c *s3.Client
}

func NewClient(awsCfg *storage.Config) *Client {
	/*
		TODO: Really should come from secrets volume
			  awsRegion = os.Getenv("AWS_REGION")
			  awsEndpoint = os.Getenv("AWS_ENDPOINT")
			  bucketName = os.Getenv("S3_BUCKET")

		localstack default, will make configurable
	*/
	if awsCfg == nil {
		return nil
	}

	customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		if awsCfg.Endpoint != "" {
			return aws.Endpoint{
				PartitionID:   "aws",
				URL:           awsCfg.Endpoint,
				SigningRegion: awsCfg.Region,
			}, nil
		}

		// returning EndpointNotFoundError will allow the service to fallback to its default resolution
		return aws.Endpoint{}, &aws.EndpointNotFoundError{}
	})

	awsCfgLoader, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(awsCfg.Region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			awsCfg.Credentials.AccessKey,
			awsCfg.Credentials.SecretKey,
			awsCfg.Credentials.SessionKey)),
		config.WithEndpointResolverWithOptions(customResolver),
	)
	if err != nil {
		log.Fatalf("Cannot load the AWS configs: %s", err)
	}

	// Create the resource client
	s3Client := s3.NewFromConfig(awsCfgLoader, func(o *s3.Options) {
		o.UsePathStyle = true
	})
	return &Client{s3Client}
}

// PutObject nice wrapper around the S3 PutObject request.
func (s3Client *Client) PutObject(key string, reader io.Reader) (*s3.PutObjectOutput, error) {
	p := s3.PutObjectInput{
		Bucket: aws.String(storage.BucketName),
		Key:    aws.String(key),
		ACL:    types.ObjectCannedACLPublicRead,
		Body:   reader,
	}

	r, err := s3Client.c.PutObject(context.Background(), &p)
	log.Printf("S3 PutObject response: %+v", r)
	if err != nil {
		return nil, err
	}

	return r, nil
}

// GetObject nice wrapper around the S3 GetObject request.
func (s3Client *Client) GetObject(key string) ([]byte, error) {

	p := s3.GetObjectInput{
		Bucket: aws.String(storage.BucketName),
		Key:    aws.String(key),
	}

	r, err := s3Client.c.GetObject(context.Background(), &p)
	log.Printf("S3 GetObject response: %+v", r)
	if err != nil {
		return nil, err
	}

	bytes, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

// CreateBucket idempotently creates an S3 bucket for VU.
func (s3Client *Client) CreateBucket() error {
	create := &s3.CreateBucketInput{
		Bucket: aws.String(storage.BucketName),
	}

	bucketOutput, err := s3Client.c.ListBuckets(context.TODO(), &s3.ListBucketsInput{})
	if err != nil {
		return err
	}

	exists := false
	for _, bucket := range bucketOutput.Buckets {
		if *bucket.Name == storage.BucketName {
			exists = true
			break
		}
	}

	if !exists {
		bucket, err := s3Client.c.CreateBucket(context.TODO(), create)
		if err != nil {
			return err
		}
		if *bucket.Location == "" {
			return fmt.Errorf("invalid bucket output")
		}
	}
	return nil
}
