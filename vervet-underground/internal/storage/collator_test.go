package storage_test

import (
	"os"
	"testing"
	"time"

	qt "github.com/frankban/quicktest"
	"github.com/snyk/vervet/v4"
	"github.com/snyk/vervet/v4/testdata"

	"vervet-underground/internal/storage"
)

const serviceASpec = `
openapi: 3.0.0
info:
  title: ServiceA API
  version: 0.0.0
paths:
  /test:
    get:
      operation: getTest
      summary: Test endpoint
      responses:
        '204':
          x-internal: its a secret to everybody
          description: An empty response
`

const serviceBSpec = `
openapi: 3.0.0
info:
  title: ServiceB API
  version: 0.0.0
paths:
  /example:
    post:
      x-other-internal: its a secret to everybody else
      operation: postTest
      summary: Example endpoint
      responses:
        '204':
          x-internal: its a secret to everybody
          description: An empty response
`

func TestCollator_Collate(t *testing.T) {
	c := qt.New(t)

	v20220201_beta := vervet.Version{
		Date:      time.Date(2022, 2, 1, 0, 0, 0, 0, time.UTC),
		Stability: vervet.StabilityBeta,
	}
	v20220301_ga := vervet.Version{
		Date:      time.Date(2022, 3, 1, 0, 0, 0, 0, time.UTC),
		Stability: vervet.StabilityGA,
	}
	v20220401_ga := vervet.Version{
		Date:      time.Date(2022, 4, 1, 0, 0, 0, 0, time.UTC),
		Stability: vervet.StabilityGA,
	}

	collator := storage.NewCollator()
	collator.Add("service-a", storage.ContentRevision{
		Version: v20220201_beta,
		Blob:    []byte(serviceASpec),
	})
	collator.Add("service-a", storage.ContentRevision{
		Version: v20220301_ga,
		Blob:    []byte(serviceASpec),
	})
	collator.Add("service-b", storage.ContentRevision{
		Version: v20220401_ga,
		Blob:    []byte(serviceBSpec),
	})

	versions, specs, err := collator.Collate()
	c.Assert(err, qt.IsNil)
	c.Assert(len(versions), qt.Equals, 3)
	c.Assert(versions[0], qt.Equals, v20220201_beta)
	c.Assert(versions[1], qt.Equals, v20220301_ga)
	c.Assert(versions[2], qt.Equals, v20220401_ga)

	c.Assert(specs[v20220201_beta].Paths.Find("/test"), qt.IsNotNil)
	c.Assert(specs[v20220201_beta].Paths.Find("/example"), qt.IsNil)

	c.Assert(specs[v20220301_ga].Paths.Find("/test"), qt.IsNotNil)
	c.Assert(specs[v20220301_ga].Paths.Find("/example"), qt.IsNil)

	c.Assert(specs[v20220401_ga].Paths.Find("/test"), qt.IsNotNil)
	c.Assert(specs[v20220401_ga].Paths.Find("/example"), qt.IsNotNil)

	// No filtering, so extensions are all present
	c.Assert(specs[v20220401_ga].Paths["/example"].Post.ExtensionProps.Extensions["x-other-internal"], qt.Not(qt.IsNil))
	c.Assert(specs[v20220401_ga].Paths["/example"].Post.Responses["204"].Value.ExtensionProps.Extensions["x-internal"], qt.Not(qt.IsNil))
}

func TestCollator_Collate_ExcludePatterns(t *testing.T) {
	c := qt.New(t)

	v20220301_ga := vervet.Version{
		Date:      time.Date(2022, 3, 1, 0, 0, 0, 0, time.UTC),
		Stability: vervet.StabilityGA,
	}
	v20220401_ga := vervet.Version{
		Date:      time.Date(2022, 4, 1, 0, 0, 0, 0, time.UTC),
		Stability: vervet.StabilityGA,
	}

	collator := storage.NewCollatorExcludePatterns(vervet.ExcludePatterns{
		ExtensionPatterns: []string{"^x.*-internal$"},
	})

	collator.Add("service-a", storage.ContentRevision{
		Version: v20220301_ga,
		Blob:    []byte(serviceASpec),
	})
	collator.Add("service-b", storage.ContentRevision{
		Version: v20220401_ga,
		Blob:    []byte(serviceBSpec),
	})
	_, specs, err := collator.Collate()
	c.Assert(err, qt.IsNil)

	c.Assert(specs[v20220401_ga].Paths["/example"].Post.ExtensionProps.Extensions["x-other-internal"], qt.IsNil)
	c.Assert(specs[v20220401_ga].Paths["/example"].Post.Responses["204"].Value.ExtensionProps.Extensions["x-internal"], qt.IsNil)
}

func TestCollator_Collate_Conflict(t *testing.T) {
	c := qt.New(t)

	v20210615_ga := vervet.Version{
		Date:      time.Date(2021, 6, 15, 0, 0, 0, 0, time.UTC),
		Stability: vervet.StabilityGA,
	}

	collator := storage.NewCollator()

	spec1, err := os.ReadFile(testdata.Path("conflict/_examples/2021-06-15/spec.yaml"))
	c.Assert(err, qt.IsNil)
	collator.Add("service-a", storage.ContentRevision{
		Version: v20210615_ga,
		Blob:    spec1,
	})

	spec2, err := os.ReadFile(testdata.Path("conflict/_examples2/2021-06-15/spec.yaml"))
	c.Assert(err, qt.IsNil)
	collator.Add("service-b", storage.ContentRevision{
		Version: v20210615_ga,
		Blob:    spec2,
	})

	_, _, err = collator.Collate()
	c.Assert(err.Error(), qt.Contains, "conflict in #/paths /examples/hello-world")
}
