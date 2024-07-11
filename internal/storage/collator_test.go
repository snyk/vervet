package storage_test

import (
	"os"
	"testing"
	"time"

	qt "github.com/frankban/quicktest"

	"github.com/snyk/vervet/v7"
	"github.com/snyk/vervet/v7/internal/storage"
	"github.com/snyk/vervet/v7/testdata"
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

const serviceDSpecWithMixedSunset = `
openapi: 3.0.0
info:
  title: ServiceD API
  version: 0.0.0
tags:
  - name: example
    description: service d example
paths:
  /sunset:
    get:
      x-snyk-sunset-eligible: 2023-01-01
      summary: Sunset endpoint
      responses:
        '200':
          description: An empty response
  /notsunset:
    get:
      summary: Not Sunset endpoint
      responses:
        '200':
          description: An empty response
  /latersunset:
    get:
      x-snyk-sunset-eligible: 2025-01-01
      summary: Later Sunset endpoint
      responses:
        '200':
          description: An empty response
`

func TestCollator_Collate_MixedSunset(t *testing.T) {
	c := qt.New(t)

	v20230101_ga := vervet.Version{
		Date:      time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
		Stability: vervet.StabilityGA,
	}

	collator, err := storage.NewCollator()
	c.Assert(err, qt.IsNil)

	collator.Add("service-d", storage.ContentRevision{
		Version: v20230101_ga,
		Blob:    []byte(serviceDSpecWithMixedSunset),
	})

	specs, err := collator.Collate()
	c.Assert(err, qt.IsNil)

	// Assert that the sunset endpoint is not present in the collated specs
	_, exists := specs[v20230101_ga].Paths["/sunset"]
	c.Assert(exists, qt.IsFalse)

	// Assert that the not sunset endpoint is still present
	_, exists = specs[v20230101_ga].Paths["/notsunset"]
	c.Assert(exists, qt.IsTrue)

	// Assert that the later sunset endpoint is still present
	_, exists = specs[v20230101_ga].Paths["/latersunset"]
	c.Assert(exists, qt.IsTrue)
}

const serviceDSpecAllSunset = `
openapi: 3.0.0
info:
  title: ServiceD API
  version: 0.0.0
tags:
  - name: example
    description: service d example
paths:
  /sunset1:
    get:
      x-snyk-sunset-eligible: 2023-01-01
      summary: Sunset endpoint 1
      responses:
        '200':
          description: An empty response
  /sunset2:
    get:
      x-snyk-sunset-eligible: 2022-12-31
      summary: Sunset endpoint 2
      responses:
        '200':
          description: An empty response
`

func TestCollator_Collate_AllSunset(t *testing.T) {
	c := qt.New(t)

	v20230101_ga := vervet.Version{
		Date:      time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
		Stability: vervet.StabilityGA,
	}

	collator, err := storage.NewCollator()
	c.Assert(err, qt.IsNil)

	collator.Add("service-d", storage.ContentRevision{
		Version: v20230101_ga,
		Blob:    []byte(serviceDSpecAllSunset),
	})

	specs, err := collator.Collate()
	c.Assert(err, qt.IsNil)

	// Assert that all sunset endpoints are not present in the collated specs
	_, exists := specs[v20230101_ga].Paths["/sunset1"]
	c.Assert(exists, qt.IsFalse)

	_, exists = specs[v20230101_ga].Paths["/sunset2"]
	c.Assert(exists, qt.IsFalse)
}

const serviceDSpecNoSunset = `
openapi: 3.0.0
info:
  title: ServiceD API
  version: 0.0.0
tags:
  - name: example
    description: service d example
paths:
  /active1:
    get:
      summary: Active endpoint 1
      responses:
        '200':
          description: An empty response
  /active2:
    get:
      summary: Active endpoint 2
      responses:
        '200':
          description: An empty response
`

func TestCollator_Collate_NoSunset(t *testing.T) {
	c := qt.New(t)

	v20230101_ga := vervet.Version{
		Date:      time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
		Stability: vervet.StabilityGA,
	}

	collator, err := storage.NewCollator()
	c.Assert(err, qt.IsNil)

	collator.Add("service-d", storage.ContentRevision{
		Version: v20230101_ga,
		Blob:    []byte(serviceDSpecNoSunset),
	})

	specs, err := collator.Collate()
	c.Assert(err, qt.IsNil)

	// Assert that all active endpoints are still present in the collated specs
	_, exists := specs[v20230101_ga].Paths["/active1"]
	c.Assert(exists, qt.IsTrue)

	_, exists = specs[v20230101_ga].Paths["/active2"]
	c.Assert(exists, qt.IsTrue)
}

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

	specs, err := collator.Collate()
	c.Assert(err, qt.IsNil)

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

	specs, err := collator.Collate()
	c.Assert(err, qt.IsNil)

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
	specs, err := collator.Collate()
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

	specs, err := collator.Collate()
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
	specs, err := collator.Collate()
	c.Assert(err, qt.IsNil)

	c.Assert(specs[v20220401_ga].Servers[0].URL, qt.Equals, "https://awesome.snyk.io/rest")
}
