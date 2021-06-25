package vervet_test

import (
	"context"
	"fmt"
	"path/filepath"
	"runtime"
	"testing"

	qt "github.com/frankban/quicktest"
	"github.com/ghodss/yaml"

	"github.com/snyk/vervet"
)

func Testdata(path string) string {
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		panic(fmt.Errorf("cannot locate caller"))
	}
	return filepath.Dir(thisFile) + "/testdata/" + path
}

func TestLoadSpecFile(t *testing.T) {
	c := qt.New(t)
	doc, err := vervet.LoadSpecFile(Testdata("_examples/hello-world/2021-06-01/spec.yaml"))
	c.Assert(err, qt.IsNil)
	c.Assert(doc.Paths, qt.HasLen, 1)
	c.Assert(doc.Paths["/examples/hello-world/{id}"], qt.Not(qt.IsNil))
	c.Assert(doc.Components.Schemas["HelloWorld"], qt.Not(qt.IsNil))
	c.Assert(doc.Validate(context.TODO()), qt.IsNil)
}

func TestToSpecYAML(t *testing.T) {
	c := qt.New(t)
	doc, err := vervet.LoadSpecFile(Testdata("_examples/hello-world/2021-06-01/spec.yaml"))
	c.Assert(err, qt.IsNil)
	yamlBuf, err := vervet.ToSpecYAML(doc)
	c.Assert(err, qt.IsNil)
	doc2 := map[string]interface{}{}
	err = yaml.Unmarshal(yamlBuf, &doc2)
	c.Assert(err, qt.IsNil)
	c.Assert(doc2["openapi"], qt.Equals, "3.0.3")
}
