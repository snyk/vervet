package backstage

import (
	"bytes"
	"os"
	"testing"
	"time"

	qt "github.com/frankban/quicktest"

	"github.com/snyk/vervet/v8/testdata"
)

func TestBackstageName(t *testing.T) {
	c := qt.New(t)
	tests := []struct {
		in, out string
	}{{
		"foo bar", "foo-bar",
	}, {
		"foo_bar", "foo-bar",
	}, {
		"foo-bar", "foo-bar",
	}, {
		"foo~bar", "foobar",
	}, {
		"Foo1Bar_Baz@#$%^&*()", "Foo1Bar-Baz",
	}}
	for _, test := range tests {
		c.Check(test.out, qt.Equals, toBackstageName(test.in))
	}
}

func TestRoundTripCatalog(t *testing.T) {
	catalogSrc := `
# Important user-authored comment
apiVersion: backstage.io/v1alpha1
kind: Component
metadata:
  name: some-component # inline comment
---
apiVersion: backstage.io/v1alpha1
kind: API
# special instructions
metadata:
  name: some-api
`[1:]
	vervetAPI := `
---
# Generated by vervet, DO NOT EDIT
apiVersion: backstage.io/v1alpha1
kind: API
metadata:
  name: vervet-api
  annotations:
    api.snyk.io/generated-by: vervet
  labels:
    api.snyk.io/version-date: "2021-06-04"
    api.snyk.io/version-lifecycle: sunset
    api.snyk.io/version-stability: experimental
`[1:]
	c := qt.New(t)
	catalog, err := LoadCatalogInfo(bytes.NewBufferString(catalogSrc + vervetAPI))
	c.Assert(err, qt.IsNil)
	c.Assert(catalog.service, qt.Not(qt.IsNil))
	c.Assert(catalog.components, qt.HasLen, 1)
	var saveOutput bytes.Buffer
	c.Assert(catalog.Save(&saveOutput), qt.IsNil)
	c.Assert(saveOutput.String(), qt.Equals, catalogSrc)
}

func TestLoadCatalogEmpty(t *testing.T) {
	c := qt.New(t)
	catalog, err := LoadCatalogInfo(bytes.NewBufferString(``))
	c.Assert(err, qt.IsNil)
	c.Assert(catalog.service, qt.IsNil)
	c.Assert(catalog.components, qt.HasLen, 0)
}

func TestLoadCatalogNoService(t *testing.T) {
	c := qt.New(t)
	catalogSrc := `
apiVersion: backstage.io/v1alpha1
kind: Location
metadata:
  name: some-place
  tags:
    - things
spec:
  type: url
`[1:]
	catalog, err := LoadCatalogInfo(bytes.NewBufferString(catalogSrc))
	c.Assert(err, qt.IsNil)
	c.Assert(catalog.service, qt.IsNil)
	c.Assert(catalog.components, qt.HasLen, 1)
	var saveOutput bytes.Buffer
	c.Assert(catalog.Save(&saveOutput), qt.IsNil)
	c.Assert(saveOutput.String(), qt.Equals, catalogSrc)
}

var catalogSrc = `
# Important user-authored comment
apiVersion: backstage.io/v1alpha1
kind: Component
metadata:
  name: some-component # inline comment
spec:
  owner: "someone-else"
`[1:]

func TestLoadVersionsNoApis(t *testing.T) {
	c := qt.New(t)
	vervetAPIs, err := os.ReadFile(testdata.Path("catalog-vervet-apis.yaml"))
	c.Assert(err, qt.IsNil)
	catalog, err := LoadCatalogInfo(bytes.NewBufferString(catalogSrc))
	c.Assert(err, qt.IsNil)
	versionsRoot := testdata.Path("output")
	err = catalog.LoadVervetAPIs(testdata.Path("."), versionsRoot, time.Date(3000, 1, 1, 0, 0, 0, 0, time.UTC))
	c.Assert(err, qt.IsNil)

	var saveOutput bytes.Buffer
	err = catalog.Save(&saveOutput)
	c.Assert(err, qt.IsNil)
	c.Assert(saveOutput.String(), qt.Equals, catalogSrc+`
  providesApis:
    - Registry_2021-06-01_experimental
    - Registry_2021-06-04_experimental
    - Registry_2021-06-07_experimental
    - Registry_2021-06-13_beta
    - Registry_2021-06-13_experimental
    - Registry_2021-08-20_experimental
    - Registry_2023-06-01_experimental
    - Registry_2023-06-02_experimental
    - Registry_2023-06-03_experimental
    - Registry_2024-10-15_ga
---
`[1:]+string(vervetAPIs))
}

func TestLoadVersionsSomeApis(t *testing.T) {
	c := qt.New(t)
	vervetAPIs, err := os.ReadFile(testdata.Path("catalog-vervet-apis.yaml"))
	c.Assert(err, qt.IsNil)
	catalog, err := LoadCatalogInfo(bytes.NewBufferString(catalogSrc + `
  providesApis:
    - someOtherApi
    - someOtherApi
`[1:]))
	c.Assert(err, qt.IsNil)
	versionsRoot := testdata.Path("output")
	err = catalog.LoadVervetAPIs(testdata.Path("."), versionsRoot, time.Date(3000, 1, 1, 0, 0, 0, 0, time.UTC))
	c.Assert(err, qt.IsNil)

	var saveOutput bytes.Buffer
	err = catalog.Save(&saveOutput)
	c.Assert(err, qt.IsNil)
	c.Assert(saveOutput.String(), qt.Equals, catalogSrc+`
  providesApis:
    - Registry_2021-06-01_experimental
    - Registry_2021-06-04_experimental
    - Registry_2021-06-07_experimental
    - Registry_2021-06-13_beta
    - Registry_2021-06-13_experimental
    - Registry_2021-08-20_experimental
    - Registry_2023-06-01_experimental
    - Registry_2023-06-02_experimental
    - Registry_2023-06-03_experimental
    - Registry_2024-10-15_ga
    - someOtherApi
---
`[1:]+string(vervetAPIs))
}

func TestDoesNotOutputStabilitiesAfterPivotDate(t *testing.T) {
	c := qt.New(t)
	vervetAPIs, err := os.ReadFile(testdata.Path("catalog-vervet-apis-with-pivot.yaml"))
	c.Assert(err, qt.IsNil)
	catalog, err := LoadCatalogInfo(bytes.NewBufferString(catalogSrc))
	c.Assert(err, qt.IsNil)
	versionsRoot := testdata.Path("output")
	err = catalog.LoadVervetAPIs(testdata.Path("."), versionsRoot, time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC))
	c.Assert(err, qt.IsNil)

	var saveOutput bytes.Buffer
	err = catalog.Save(&saveOutput)
	c.Assert(err, qt.IsNil)
	c.Assert(saveOutput.String(), qt.Equals, catalogSrc+`
  providesApis:
    - Registry_2021-06-01_experimental
    - Registry_2021-06-04_experimental
    - Registry_2021-06-07_experimental
    - Registry_2021-06-13_beta
    - Registry_2021-06-13_experimental
    - Registry_2021-08-20_experimental
    - Registry_2023-06-01
    - Registry_2023-06-02
    - Registry_2023-06-03
    - Registry_2024-10-15
---
`[1:]+string(vervetAPIs))
}

func TestWarnsAboutConflictingNames(t *testing.T) {
	c := qt.New(t)
	catalog, err := LoadCatalogInfo(bytes.NewBufferString(catalogSrc))
	c.Assert(err, qt.IsNil)
	versionsRoot := testdata.Path("output")
	err = catalog.LoadVervetAPIs(testdata.Path("."), versionsRoot, time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC))
	c.Assert(err, qt.IsNil)

	err = catalog.LoadVervetAPIs(testdata.Path("."), versionsRoot, time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC))
	c.Assert(err, qt.IsNotNil)
}
