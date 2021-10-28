package testdata

import (
	"testing"

	qt "github.com/frankban/quicktest"
	"github.com/getkin/kin-openapi/openapi3"

	"github.com/snyk/vervet/testdata/output"
)

func TestEmbedding(t *testing.T) {
	c := qt.New(t)

	specYAML, err := output.Versions.ReadFile("2021-06-13~experimental/spec.yaml")
	c.Assert(err, qt.IsNil)
	l := openapi3.NewLoader()
	_, err = l.LoadFromData(specYAML)
	c.Assert(err, qt.IsNil)
}
