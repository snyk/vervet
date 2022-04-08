// Package s3 provides an implementation of Vervet Underground storage backed
// by Amazon S3.
package s3

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"io"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

var (
	awsRegion   string
	awsEndpoint string

	s3Client *s3.Client
)

const bucketName = "vervet-underground-specs"

func init() {
	/*
		TODO: Really should come from secrets volume
			  awsRegion = os.Getenv("AWS_REGION")
			  awsEndpoint = os.Getenv("AWS_ENDPOINT")
			  bucketName = os.Getenv("S3_BUCKET")
	*/
	localstackAccessKey := "test"
	localstackSecretKey := "test"

	// localstack default, will make configurable
	awsEndpoint = "http://localhost:4566"
	awsRegion = "us-east-1"

	customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		if awsEndpoint != "" {
			return aws.Endpoint{
				PartitionID:   "aws",
				URL:           awsEndpoint,
				SigningRegion: awsRegion,
			}, nil
		}

		// returning EndpointNotFoundError will allow the service to fallback to its default resolution
		return aws.Endpoint{}, &aws.EndpointNotFoundError{}
	})

	awsCfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(awsRegion),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			localstackAccessKey,
			localstackSecretKey,
			"dummy")),
		config.WithEndpointResolverWithOptions(customResolver),
	)
	if err != nil {
		log.Fatalf("Cannot load the AWS configs: %s", err)
	}

	// Create the resource client
	s3Client = s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.UsePathStyle = true
	})
}

func PutObject(key string, reader io.Reader) *s3.PutObjectOutput {
	p := s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(key),
		ACL:    types.ObjectCannedACLPublicRead,
		Body:   reader,
	}

	r, err := s3Client.PutObject(context.Background(), &p)
	log.Printf("S3 PutObject response: %+v", r)
	if err != nil {
		panic(err)
	}

	return r
}

// CreateBucket idempotently creates an S3 bucket for VU
//
func CreateBucket() error {
	create := &s3.CreateBucketInput{
		Bucket: aws.String(bucketName),
	}

	bucketOutput, err := s3Client.ListBuckets(context.TODO(), &s3.ListBucketsInput{})
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
		bucket, err := s3Client.CreateBucket(context.TODO(), create)
		if err != nil {
			return err
		}
		if *bucket.Location == "" {
			return fmt.Errorf("invalid bucket output")
		}
	}
	return nil
}

//func main() {
//	err := CreateBucket()
//	if err != nil {
//		fmt.Printf("Error creating bucket %v", err)
//		return
//	}
//
//	data := []byte("this is some data stored as a byte slice in Go Lang!")
//	reader := bytes.NewReader(data)
//	obj := PutObject("dummy", reader)
//	fmt.Printf("Resulting putObject response: %v", obj)
//
//	return
//}
