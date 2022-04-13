package storage

import (
	"io"
	"time"
)

type StaticKeyCredentials struct {
	AccessKey  string
	SecretKey  string
	SessionKey string
}

type Config struct {
	Region      string
	Endpoint    string
	BucketName  string
	Credentials StaticKeyCredentials
}

type Aggregate interface {
	HasVersion(name string, version string, digest string) (bool, error)
	NotifyVersions(name string, versions []string, scrapeTime time.Time) error
	NotifyVersion(name string, version string, contents []byte, scrapeTime time.Time) error
	Version(version string) ([]byte, error)
	CollateVersions() error
	GetCollatedVersionSpecs() (map[string][]byte, error)
}

type Client interface {
	PutObject(key string, reader io.Reader) // TODO: add result type to validate
	GetObject(key string) ([]byte, error)
	CreateBucket() error
	NewClient(cfg Config) *Client
}

type Storage struct {
	aggregator Aggregate
}
