package vervet_test

import (
	"context"
	"testing"

	qt "github.com/frankban/quicktest"
	"github.com/ghodss/yaml"

	"github.com/snyk/vervet"
	"github.com/snyk/vervet/testdata"
)

func TestLoadSpecFile(t *testing.T) {
	c := qt.New(t)
	doc, err := vervet.LoadSpecFile(testdata.Path("resources/_examples/hello-world/2021-06-01/spec.yaml"))
	c.Assert(err, qt.IsNil)
	c.Assert(doc.Paths, qt.HasLen, 1)
	c.Assert(doc.Paths["/examples/hello-world/{id}"], qt.Not(qt.IsNil))
	c.Assert(doc.Components.Schemas["HelloWorld"], qt.Not(qt.IsNil))
	c.Assert(doc.Validate(context.TODO()), qt.IsNil)
}

func TestToSpecYAML(t *testing.T) {
	c := qt.New(t)
	doc, err := vervet.LoadSpecFile(testdata.Path("resources/_examples/hello-world/2021-06-01/spec.yaml"))
	c.Assert(err, qt.IsNil)
	yamlBuf, err := vervet.ToSpecYAML(doc)
	c.Assert(err, qt.IsNil)
	doc2 := map[string]interface{}{}
	err = yaml.Unmarshal(yamlBuf, &doc2)
	c.Assert(err, qt.IsNil)
	c.Assert(doc2["openapi"], qt.Equals, "3.0.3")
}
