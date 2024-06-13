package vervet_test

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	qt "github.com/frankban/quicktest"

	"github.com/snyk/vervet/v7"
	"github.com/snyk/vervet/v7/testdata"
)

func TestLocalize(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	c := qt.New(t)
	doc, err := vervet.NewDocumentFile(testdata.Path("resources/_examples/hello-world/2021-06-01/spec.yaml"))
	c.Assert(err, qt.IsNil)
	err = vervet.Localize(ctx, doc)
	c.Assert(err, qt.IsNil)
	err = doc.Validate(context.TODO())
	c.Assert(err, qt.IsNil)

	// OpenAPI DOM should be fully localized and relocatable now.
	yamlBuf, err := vervet.ToSpecYAML(doc)
	c.Assert(err, qt.IsNil)
	tmpDir := c.TempDir()
	err = os.WriteFile(tmpDir+"/spec.yaml", yamlBuf, 0644)
	c.Assert(err, qt.IsNil)

	// This will fail to load if references have not been localized!
	doc2, err := vervet.NewDocumentFile(tmpDir + "/spec.yaml")
	c.Assert(err, qt.IsNil)
	c.Assert(doc2.Validate(context.TODO()), qt.IsNil)

	// Assert round-trip serialization equality
	jsonBuf, err := json.Marshal(doc)
	c.Assert(err, qt.IsNil)
	c.Assert(jsonBuf, qt.JSONEquals, doc2)
}
