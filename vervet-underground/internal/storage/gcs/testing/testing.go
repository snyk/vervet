package testing

import (
	"context"
	"errors"
	"io"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/docker/go-connections/nat"
	qt "github.com/frankban/quicktest"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"vervet-underground/internal/storage/gcs"
)

const (
	GcsRegion  = "US-CENTRAL1" // https://cloud.google.com/storage/docs/locations#location-r
	ProjectId  = "test"
	BucketName = "vervet-underground-specs"
)

// setupMutex protects port acquisition among concurrent tests.
var setupMutex sync.Mutex

// Setup launches a fake GCS server and returns the storage configuration
// needed to connect to it.
//
// Resources are cleaned up automatically when the test completes.
func Setup(c *qt.C) *gcs.Config {
	setupMutex.Lock()
	defer setupMutex.Unlock()
	gcsPort := findOpenPort(c)
	ctx, cancel := context.WithCancel(context.Background())
	c.Cleanup(cancel)
	container := Connect(ctx, c, gcsPort)
	c.Cleanup(func() {
		if container != nil {
			err := container.Terminate(ctx)
			c.Assert(err, qt.IsNil)
		}
	})

	mappedPort, err := container.MappedPort(ctx, "4443/tcp")
	c.Assert(err, qt.IsNil)

	// Proxy localhost:44443 to the mapped port. Ideally we'd just use the
	// mapped port, but fake-gcs-server needs to know its public-url before we
	// start it, so there's a bootstrapping problem.
	ln, err := net.Listen("tcp", "localhost:"+gcsPort)
	c.Assert(err, qt.IsNil)
	copyIO := func(dst, src net.Conn) {
		defer dst.Close()
		defer src.Close()
		if _, err := io.Copy(dst, src); err != nil {
			c.Assert(errors.Is(err, net.ErrClosed), qt.IsTrue)
		}
	}
	go func(ln net.Listener) {
		for {
			conn, err := ln.Accept()
			if errors.Is(err, net.ErrClosed) {
				return
			}
			c.Assert(err, qt.IsNil)
			go func(conn net.Conn) {
				proxy, err := net.Dial("tcp", "localhost:"+mappedPort.Port())
				c.Assert(err, qt.IsNil)
				go copyIO(conn, proxy)
				go copyIO(proxy, conn)
			}(conn)
		}
	}(ln)
	c.Cleanup(func() { ln.Close() })

	return &gcs.Config{
		GcsRegion:   GcsRegion,
		GcsEndpoint: "http://localhost:" + gcsPort + "/storage/v1/",
		BucketName:  BucketName,
		Credentials: gcs.StaticKeyCredentials{
			ProjectId: ProjectId,
		},
		WithoutAuthentication: true,
	}
}

// Connect returns a newly launched fake GCS server container.
func Connect(ctx context.Context, c *qt.C, gcsPort string) testcontainers.Container {
	dataDir := c.TempDir()
	c.Assert(os.MkdirAll(filepath.Join(dataDir, BucketName), 0777), qt.IsNil)
	c.Assert(os.WriteFile(filepath.Join(dataDir, BucketName, "test.txt"), []byte("test"), 0600), qt.IsNil)
	req := testcontainers.ContainerRequest{
		Image:        "fsouza/fake-gcs-server:latest",
		ExposedPorts: []string{"4443/tcp"},
		WaitingFor:   wait.ForListeningPort(nat.Port("4443/tcp")).WithStartupTimeout(30 * time.Second),
		Cmd: []string{
			"-scheme", "http",
			"-backend", "memory",
			"-port", "4443",
			"-public-host", "localhost:" + gcsPort,
			"-data", "/data",
		},
		Mounts: []testcontainers.ContainerMount{
			testcontainers.BindMount(dataDir, "/data"),
		},
	}
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	c.Assert(err, qt.IsNil)
	return container
}

// findOpenPort returns an available open port on the host machine.
func findOpenPort(c *qt.C) string {
	ln, err := net.Listen("tcp", "localhost:0")
	c.Assert(err, qt.IsNil)
	defer ln.Close()
	return strconv.Itoa(ln.Addr().(*net.TCPAddr).Port) //nolint:forcetypeassert // acked
}
