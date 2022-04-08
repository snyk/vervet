// Package s3 provides an implementation of Vervet Underground storage backed
// by Amazon S3.
package s3

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

type StaticKeyCredentials struct {
	AccessKey  string
	SecretKey  string
	SessionKey string
}

type AwsConfig struct {
	AwsRegion   string
	AwsEndpoint string
	BucketName  string
	Credentials StaticKeyCredentials
}

type AwsS3Client struct {
	c *s3.Client
}

const bucketName = "vervet-underground-specs"

func NewClient(awsCfg *AwsConfig) *AwsS3Client {
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

	awsCfgLoader, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(awsCfg.AwsRegion),
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
	return &AwsS3Client{s3Client}
}

// PutObject nice wrapper around the S3 PutObject request
func (s3Client *AwsS3Client) PutObject(key string, reader io.Reader) *s3.PutObjectOutput {
	p := s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(key),
		ACL:    types.ObjectCannedACLPublicRead,
		Body:   reader,
	}

	r, err := s3Client.c.PutObject(context.Background(), &p)
	log.Printf("S3 PutObject response: %+v", r)
	if err != nil {
		panic(err)
	}

	return r
}

// CreateBucket idempotently creates an S3 bucket for VU.
func (s3Client *AwsS3Client) CreateBucket() error {
	create := &s3.CreateBucketInput{
		Bucket: aws.String(bucketName),
	}

	bucketOutput, err := s3Client.c.ListBuckets(context.TODO(), &s3.ListBucketsInput{})
	if err != nil {
		return err
	}

	exists := false
	for _, bucket := range bucketOutput.Buckets {
		if *bucket.Name == bucketName {
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

func main() {
	localstackAccessKey := "test"
	localstackSecretKey := "test"
	localstackSessionKey := ""
	awsEndpoint := "http://localhost:4566"
	awsRegion := "us-east-1"
	cfg := &AwsConfig{
		awsRegion,
		awsEndpoint,
		bucketName,
		StaticKeyCredentials{
			localstackAccessKey,
			localstackSecretKey,
			localstackSessionKey,
		},
	}

	client := NewClient(cfg)
	err := client.CreateBucket()
	if err != nil {
		fmt.Printf("Error creating bucket %v", err)
		return
	}

	data := []byte("this is some data stored as a byte slice in Go Lang!")
	reader := bytes.NewReader(data)
	obj := client.PutObject("dummy", reader)
	fmt.Printf("Resulting putObject response: %v", obj)

	return
}
