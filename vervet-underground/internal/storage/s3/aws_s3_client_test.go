package s3

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strconv"
	"testing"

	"github.com/docker/go-connections/nat"
	qt "github.com/frankban/quicktest"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	localstackAccessKey  = "test"
	localstackSecretKey  = "test"
	localstackSessionKey = "test"
	awsEndpoint          = "http://localhost:4566"
	awsRegion            = "us-east-1"
)

var cfg = &AwsConfig{
	awsRegion,
	awsEndpoint,
	bucketName,
	StaticKeyCredentials{
		localstackAccessKey,
		localstackSecretKey,
		localstackSessionKey,
	},
}

func TestPutObject(t *testing.T) {
	// Arrange
	c := qt.New(t)
	if isCIEnabled(t) {
		c.Assert(true, qt.IsTrue)
		return
	}

	ctx := context.Background()

	ctr, err := setupTestContainer(ctx, t)
	defer teardownTestContainer(ctx, t, ctr)
	c.Assert(err, qt.IsNil)

	//testCases := []struct {
	//	name        string
	//	t           string
	//	ID          string
	//	expectedObj []byte
	//	expectedOK  bool
	//}{
	//	{
	//		name:        "non-existent entry",
	//		t:           "undefined",
	//		ID:          "0000",
	//		expectedObj: nil,
	//		expectedOK:  false,
	//	},
	//	{
	//		name:        "existent entry",
	//		t:           "repo",
	//		ID:          "1111",
	//		expectedObj: []byte(`{"key1":"val1"}`),
	//		expectedOK:  true,
	//	},
	//}

	client := NewClient(cfg)
	err = client.CreateBucket()
	c.Assert(err, qt.IsNil)

	data := []byte("this is some data stored as a byte slice in Go Lang!")

	// convert byte slice to io.Reader
	reader := bytes.NewReader(data)
	//
	//for _, tc := range testCases {
	//	t.Run(tc.name, func(t *testing.T) {
	//		// Act
	obj := client.PutObject("dummy", reader)

	// Assert
	c.Assert(obj, qt.IsNotNil)
	c.Assert(obj.ETag, qt.IsNotNil)
	//	})
	//}
}

func isCIEnabled(t *testing.T) bool {
	t.Helper()

	ci, err := strconv.ParseBool(os.Getenv("CI"))
	return err == nil || ci
}

func setupTestContainer(ctx context.Context, t *testing.T) (ctr testcontainers.Container, err error) {
	t.Helper()

	port := "4566/tcp"
	secondPort := "4571/tcp"

	// Do not use `testcontainers` when running in CI.
	// TODO: examine for localstack
	//       https://docs.localstack.cloud/ci/circle-ci/
	if isCIEnabled(t) {
		return nil, fmt.Errorf("CI disables localstack")
	}

	env := map[string]string{
		"MAIN_CONTAINER_NAME": "localstack",
		"EDGE_PORT":           "4566",
	}

	//workingDir, err := os.Getwd()
	//if err != nil {
	//	return ctr, fmt.Errorf("failed to get current working directory: %w", err)
	//}
	//mountPath := workingDir + "/tmp"

	req := testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Env:          env,
			Name:         "localstack",
			ExposedPorts: []string{port, secondPort},
			Image:        "localstack/localstack:0.13.3",
			WaitingFor:   wait.ForHTTP("/health").WithPort(nat.Port(port)),
		},
		Started: true,
	}

	if ctr, err = testcontainers.GenericContainer(ctx, req); err != nil {
		return ctr, fmt.Errorf("failed to start container: %w", err)
	}

	if _, err = ctr.MappedPort(ctx, nat.Port(port)); err != nil {
		return ctr, fmt.Errorf("failed to get container external port: %w", err)
	}

	if _, err = ctr.MappedPort(ctx, nat.Port(secondPort)); err != nil {
		return ctr, fmt.Errorf("failed to get container second external port: %w", err)
	}

	return ctr, nil
}

func teardownTestContainer(ctx context.Context, t *testing.T, ctr testcontainers.Container) {
	t.Helper()

	if ctr == nil {
		return
	}

	if err := ctr.Terminate(ctx); err != nil {
		t.Logf("failed to terminate container: %s", err)
	}
}
