package apiutil

import (
	"strings"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
)

type Version string

const (
	VersionExperimental = "experimental"
	VersionBeta         = "beta"
)

func ParseVersion(s string) (Version, error) {
	switch s {
	case string(VersionExperimental):
		return VersionExperimental, nil
	case string(VersionBeta):
		return VersionBeta, nil
	default:
		_, err := time.Parse("2006-01-02", s)
		return Version(s), err
	}
}

type EndpointVersion struct {
	Version Version
	Spec    *openapi3.T
}

type Endpoint []EndpointVersion

func (e Endpoint) Less(i, j int) bool {
	// Lexicographical compare actually works fine for this:
	// YYYY-mm-dd < beta < experimental
	// TODO: mere coincidence
	return strings.Compare(string(e[i].Version), string(e[j].Version)) < 0
}
