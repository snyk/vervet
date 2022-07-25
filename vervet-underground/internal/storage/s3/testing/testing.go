package testing

import (
	"context"

	"github.com/elgohr/go-localstack"
	qt "github.com/frankban/quicktest"

	"vervet-underground/internal/storage/s3"
)

const (
	bucketName           = "vervet-underground-specs"
	localstackAccessKey  = "test"
	localstackSecretKey  = "test"
	localstackSessionKey = "test"
	awsRegion            = "us-east-1"
)

// Setup launches a localstack S3 server and returns the storage configuration
// needed to connect to it.
//
// Resources are cleaned up automatically on context cancellation when the test
// completes.
func Setup(c *qt.C) *s3.Config {
	ctx, cancel := context.WithCancel(context.Background())
	c.Cleanup(cancel)
	l, err := localstack.NewInstance()
	c.Assert(err, qt.IsNil)
	err = l.StartWithContext(ctx)
	c.Assert(err, qt.IsNil)

	return &s3.Config{
		AwsRegion:   awsRegion,
		AwsEndpoint: l.EndpointV2(localstack.S3),
		BucketName:  bucketName,
		Credentials: s3.StaticKeyCredentials{
			AccessKey:  localstackAccessKey,
			SecretKey:  localstackSecretKey,
			SessionKey: localstackSessionKey,
		},
	}
}
