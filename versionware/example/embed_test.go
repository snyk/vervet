package example

import (
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/snyk/vervet/v6"
	"github.com/snyk/vervet/v6/versionware/example/releases"
)

func TestEmbedding(t *testing.T) {
	c := qt.New(t)

	specs, err := vervet.LoadVersions(releases.Versions)
	c.Assert(err, qt.IsNil)
	c.Assert(specs, qt.HasLen, 4)
	versions := []string{}
	for i := range specs {
		version, err := vervet.ExtensionString(specs[i].Extensions, vervet.ExtSnykApiVersion)
		c.Assert(err, qt.IsNil)
		versions = append(versions, version)
	}
	c.Assert(versions, qt.ContentEquals, []string{
		"2021-11-01~experimental",
		"2021-11-08~experimental",
		"2021-11-20~experimental",
		"2024-10-15", // the pivot date version
	})
}
