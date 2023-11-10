package storage_test

import (
	"os"
	"testing"
	"time"

	qt "github.com/frankban/quicktest"
	"github.com/snyk/vervet/v5"
	"github.com/snyk/vervet/v5/testdata"

	"vervet-underground/internal/storage"
)

const serviceASpec = `
openapi: 3.0.0
info:
  title: ServiceA API
  version: 0.0.0
tags:
  - name: example
    description: service a example
paths:
  /test:
    get:
      operation: getTest
      summary: Test endpoint
      responses:
        '204':
          x-internal: its a secret to everybody
          description: An empty response
  /openapi:
    get:
      tags:
        - example
      responses:
        '200':
          description: List OpenAPI versions
  /openapi/{version}:
    get:
      responses:
        '200':
          description: Get OpenAPI at version
`

const serviceBSpec = `
openapi: 3.0.0
info:
  title: ServiceB API
  version: 0.0.0
tags:
  - name: example
    description: service b example
paths:
  /example:
    post:
      tags:
        - example
      x-other-internal: its a secret to everybody else
      operation: postTest
      summary: Example endpoint
      responses:
        '204':
          x-internal: its a secret to everybody
          description: An empty response
  /openapi:
    get:
      responses:
        '200':
          description: List OpenAPI versions
  /openapi/{version}:
    get:
      responses:
        '200':
          description: Get OpenAPI at version
`

const serviceCSpec = `
openapi: 3.0.0
info:
  title: ServiceC API
  version: 0.0.0
tags:
  - name: example
    description: service c example
paths:
  /test:
    get:
      operation: getTest
      summary: Test endpoint
      responses:
        '200':
          x-internal: its a secret to everybody
          description: An empty response
  /openapi:
    get:
      tags:
        - example
      responses:
        '200':
          description: List OpenAPI versions
  /openapi/{version}:
    get:
      responses:
        '200':
          description: Get OpenAPI at version
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

	collator, err := storage.NewCollator()
	c.Assert(err, qt.IsNil)
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
	c.Assert(specs[v20220401_ga].Paths["/example"].Post.Extensions["x-other-internal"], qt.Not(qt.IsNil))
	c.Assert(specs[v20220401_ga].Paths["/example"].Post.Responses["204"].Value.Extensions["x-internal"], qt.Not(qt.IsNil))
}

func TestCollator_Collate_MigratingEndpoints(t *testing.T) {
	c := qt.New(t)

	v20220201_exp := vervet.Version{
		Date:      time.Date(2022, 2, 1, 0, 0, 0, 0, time.UTC),
		Stability: vervet.StabilityExperimental,
	}
	v20230314_exp := vervet.Version{
		Date:      time.Date(2023, 3, 14, 0, 0, 0, 0, time.UTC),
		Stability: vervet.StabilityExperimental,
	}

	collator, err := storage.NewCollator()
	c.Assert(err, qt.IsNil)
	collator.Add("service-a", storage.ContentRevision{
		Version: v20220201_exp,
		Blob:    []byte(serviceASpec),
	})
	collator.Add("service-c", storage.ContentRevision{
		Version: v20230314_exp,
		Blob:    []byte(serviceCSpec),
	})

	versions, specs, err := collator.Collate()
	c.Assert(err, qt.IsNil)
	c.Assert(len(versions), qt.Equals, 2)
	c.Assert(versions[0], qt.Equals, v20220201_exp)
	c.Assert(versions[1], qt.Equals, v20230314_exp)

	c.Assert(specs[v20220201_exp].Paths.Find("/test"), qt.IsNotNil)
	c.Assert(specs[v20230314_exp].Paths.Find("/test"), qt.IsNotNil)

	c.Assert(specs[v20220201_exp].Paths["/test"].Get.Responses["204"], qt.IsNotNil)
	c.Assert(specs[v20220201_exp].Paths["/test"].Get.Responses["200"], qt.IsNil)

	c.Assert(specs[v20230314_exp].Paths["/test"].Get.Responses["200"], qt.IsNotNil)
	c.Assert(specs[v20230314_exp].Paths["/test"].Get.Responses["204"], qt.IsNil)
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

	collator, err := storage.NewCollator(storage.CollatorExcludePattern(vervet.ExcludePatterns{
		ExtensionPatterns: []string{"^x.*-internal$"},
	}))
	c.Assert(err, qt.IsNil)

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

	c.Assert(specs[v20220401_ga].Paths["/example"].Post.Extensions["x-other-internal"], qt.IsNil)
	c.Assert(specs[v20220401_ga].Paths["/example"].Post.Responses["204"].Value.Extensions["x-internal"], qt.IsNil)
}

func TestCollator_Collate_Conflict(t *testing.T) {
	c := qt.New(t)

	v20210615_ga := vervet.Version{
		Date:      time.Date(2021, 6, 15, 0, 0, 0, 0, time.UTC),
		Stability: vervet.StabilityGA,
	}

	collator, err := storage.NewCollator()
	c.Assert(err, qt.IsNil)

	spec1, err := os.ReadFile(testdata.Path("conflict/_examples/2021-06-15/spec.yaml"))
	c.Assert(err, qt.IsNil)
	collator.Add("service-a", storage.ContentRevision{
		Version:   v20210615_ga,
		Blob:      spec1,
		Timestamp: time.Date(2021, 6, 15, 12, 0, 0, 0, time.UTC),
	})

	spec2, err := os.ReadFile(testdata.Path("conflict/_examples2/2021-06-15/spec.yaml"))
	c.Assert(err, qt.IsNil)
	collator.Add("service-b", storage.ContentRevision{
		Version:   v20210615_ga,
		Blob:      spec2,
		Timestamp: time.Date(2021, 6, 15, 0, 0, 0, 0, time.UTC),
	})

	_, specs, err := collator.Collate()
	c.Assert(err, qt.IsNil)
	// First path wins
	c.Assert(specs[vervet.MustParseVersion("2021-06-15")].Paths["/examples/hello-world"].Post.Description,
		qt.Equals, "Create a single result from the hello-world example - from example 1")
}

var testOverlay = `
info:
    title: Snyk Awesome API
    version: REST
servers:
    - url: https://awesome.snyk.io/rest
      description: An awesome API
components:
    securitySchemes:
        BearerAuth:
            type: http
            scheme: bearer
security:
    - BearerAuth: []
`[1:]

func TestCollator_Collate_Overlay(t *testing.T) {
	c := qt.New(t)

	v20220301_ga := vervet.Version{
		Date:      time.Date(2022, 3, 1, 0, 0, 0, 0, time.UTC),
		Stability: vervet.StabilityGA,
	}
	v20220401_ga := vervet.Version{
		Date:      time.Date(2022, 4, 1, 0, 0, 0, 0, time.UTC),
		Stability: vervet.StabilityGA,
	}

	collator, err := storage.NewCollator(
		storage.CollatorOverlay(testOverlay),
	)
	c.Assert(err, qt.IsNil)

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

	c.Assert(specs[v20220401_ga].Servers[0].URL, qt.Equals, "https://awesome.snyk.io/rest")
}
