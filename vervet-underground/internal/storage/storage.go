package storage

import (
	"net/url"
	"strings"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/rs/zerolog/log"
	"github.com/snyk/vervet/v4"
)

// Storage defines the storage functionality needed in order to store service
// API version spec snapshots.
type Storage interface {
	// NotifyVersions tells the storage which versions are currently available.
	// This is the primary mechanism by which the storage layer discovers and
	// processes versions which are removed post-sunset.
	NotifyVersions(name string, versions []string, scrapeTime time.Time) error

	// CollateVersions tells the storage to execute the compilation and
	// update all VU-formatted specs from all services and their
	// respective versions gathered.
	CollateVersions() error

	// HasVersion returns whether the storage has already stored the service
	// API spec version at the given content digest.
	HasVersion(name string, version string, digest string) (bool, error)

	// NotifyVersion tells the storage to store the given version contents at
	// the scrapeTime. The storage implementation must detect and ignore
	// duplicate version contents, as some services may not provide content
	// digest headers in their responses.
	NotifyVersion(name string, version string, contents []byte, scrapeTime time.Time) error

	// Versions fetches the Storage Versions compiled by VU
	Versions() []string

	// Version fetches the Storage Version spec compiled by VU
	Version(version string) ([]byte, error)
}

// CollatedVersionMappedSpecs Compiled aggregated spec for all services at that given version.
type CollatedVersionMappedSpecs map[vervet.Version]openapi3.T

const (
	BucketName             = "vervet-underground-specs"
	CollatedVersionsFolder = "collated-versions/"
	ServiceVersionsFolder  = "service-versions/"
)

func GetSantizedHost(name string) string {
	host := name
	if strings.HasPrefix(name, "http") {
		parsedUrl, err := url.Parse(name)
		if err != nil {
			log.Warn().Err(err).Msgf("service.base url misconfigured. Falling back %v", name)
		} else {
			host = parsedUrl.Host
		}
	}
	return host
}
