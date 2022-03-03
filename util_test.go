package vervet_test

import (
	"testing"

	qt "github.com/frankban/quicktest"
	"github.com/ghodss/yaml"

	"github.com/snyk/vervet/v4"
	"github.com/snyk/vervet/v4/testdata"
)

func TestToSpecYAML(t *testing.T) {
	c := qt.New(t)
	doc, err := vervet.NewDocumentFile(testdata.Path("resources/_examples/hello-world/2021-06-01/spec.yaml"))
	c.Assert(err, qt.IsNil)
	yamlBuf, err := vervet.ToSpecYAML(doc)
	c.Assert(err, qt.IsNil)
	doc2 := map[string]interface{}{}
	err = yaml.Unmarshal(yamlBuf, &doc2)
	c.Assert(err, qt.IsNil)
	c.Assert(doc2["openapi"], qt.Equals, "3.0.3")
}
