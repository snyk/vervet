package vervet_test

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/snyk/vervet"
)

func TestLocalize(t *testing.T) {
	c := qt.New(t)
	doc, err := vervet.LoadSpecFile(Testdata("_examples/hello-world/2021-06-01/spec.yaml"))
	c.Assert(err, qt.IsNil)
	err = vervet.NewLocalizer(doc).Localize()
	c.Assert(err, qt.IsNil)
	err = doc.Validate(context.TODO())
	c.Assert(err, qt.IsNil)

	// OpenAPI DOM should be fully localized and relocatable now.
	yamlBuf, err := vervet.ToSpecYAML(doc)
	c.Assert(err, qt.IsNil)
	tmpDir := c.Mkdir()
	err = ioutil.WriteFile(tmpDir+"/spec.yaml", yamlBuf, 0644)
	c.Assert(err, qt.IsNil)

	// This will fail to load if references have not been localized!
	doc2, err := vervet.LoadSpecFile(tmpDir + "/spec.yaml")
	c.Assert(err, qt.IsNil)
	c.Assert(doc2.Validate(context.TODO()), qt.IsNil)

	// Assert round-trip serialization equality
	jsonBuf, err := json.Marshal(doc)
	c.Assert(err, qt.IsNil)
	c.Assert(jsonBuf, qt.JSONEquals, doc2)
}
