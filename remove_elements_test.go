package vervet_test

import (
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/snyk/vervet/v5"
	"github.com/snyk/vervet/v5/testdata"
)

func TestRemoveElementsExact(t *testing.T) {
	c := qt.New(t)
	doc, err := vervet.NewDocumentFile(testdata.Path("output/2021-08-20~experimental/spec.yaml"))
	c.Assert(err, qt.IsNil)

	// Establish that the OpenAPI document has these expected features

	c.Assert(doc.Paths["/examples/hello-world"], qt.Not(qt.IsNil))
	c.Assert(doc.Paths["/examples/hello-world/{id}"], qt.Not(qt.IsNil))
	c.Assert(doc.Paths["/orgs/{orgId}/projects"].Get.Responses["200"].Value.Headers["snyk-request-id"], qt.Not(qt.IsNil))
	c.Assert(doc.Paths["/orgs/{orgId}/projects"].Get.Responses["200"].Value.Headers["snyk-version-served"], qt.Not(qt.IsNil))
	c.Assert(doc.Paths["/orgs/{org_id}/projects/{project_id}"].Delete.Parameters, qt.HasLen, 4)
	c.Assert(doc.Paths["/orgs/{org_id}/projects/{project_id}"].Delete.Parameters[3].Value.Name, qt.Equals, "x-private-matter")

	c.Assert(doc.Paths["/orgs/{orgId}/projects"].Extensions["x-snyk-api-resource"], qt.Not(qt.IsNil))
	c.Assert(doc.Extensions["x-snyk-api-lifecycle"], qt.Not(qt.IsNil))

	// Remove some of them

	err = vervet.RemoveElements(doc.T, vervet.ExcludePatterns{
		ExtensionPatterns: []string{"x-snyk-api-releases", "x-snyk-api-resource"},
		HeaderPatterns:    []string{"snyk-request-id", "x-private-matter"},
		Paths:             []string{"/examples/hello-world", "/examples/hello-world/{id}"},
	})
	c.Assert(err, qt.IsNil)

	// Assert their removal

	c.Assert(doc.Paths["/examples/hello-world"], qt.IsNil)
	c.Assert(doc.Paths["/examples/hello-world/{id}"], qt.IsNil)
	c.Assert(doc.Paths["/orgs/{orgId}/projects"].Get.Responses["200"].Value.Headers["snyk-request-id"], qt.IsNil)             // now removed
	c.Assert(doc.Paths["/orgs/{orgId}/projects"].Get.Responses["200"].Value.Headers["snyk-version-served"], qt.Not(qt.IsNil)) // still there
	c.Assert(doc.Paths["/orgs/{org_id}/projects/{project_id}"].Delete.Parameters, qt.HasLen, 3)                               // x-private-matter removed

	c.Assert(doc.Paths["/orgs/{orgId}/projects"].Extensions["x-snyk-api-resource"], qt.IsNil) // now removed
	c.Assert(doc.Extensions["x-snyk-api-lifecycle"], qt.Not(qt.IsNil))                        // still there
}

func TestRemoveElementsRegex(t *testing.T) {
	c := qt.New(t)
	doc, err := vervet.NewDocumentFile(testdata.Path("output/2021-08-20~experimental/spec.yaml"))
	c.Assert(err, qt.IsNil)

	// Establish that the OpenAPI document has these expected features

	c.Assert(doc.Paths["/orgs/{orgId}/projects"].Get.Responses["200"].Value.Headers["snyk-request-id"], qt.Not(qt.IsNil))
	c.Assert(doc.Paths["/orgs/{orgId}/projects"].Get.Responses["200"].Value.Headers["snyk-version-served"], qt.Not(qt.IsNil))
	c.Assert(doc.Paths["/orgs/{org_id}/projects/{project_id}"].Delete.Parameters, qt.HasLen, 4)
	c.Assert(doc.Paths["/orgs/{org_id}/projects/{project_id}"].Delete.Parameters[3].Value.Name, qt.Equals, "x-private-matter")

	c.Assert(doc.Paths["/orgs/{orgId}/projects"].Extensions["x-snyk-api-resource"], qt.Not(qt.IsNil))
	c.Assert(doc.Extensions["x-snyk-api-lifecycle"], qt.Not(qt.IsNil))

	// Remove some of them

	err = vervet.RemoveElements(doc.T, vervet.ExcludePatterns{
		ExtensionPatterns: []string{"x-snyk-api-r.*"},
		HeaderPatterns:    []string{"snyk-version-.*", "x-private-.*"},
	})
	c.Assert(err, qt.IsNil)

	// Assert their removal

	c.Assert(doc.Paths["/orgs/{orgId}/projects"].Get.Responses["200"].Value.Headers["snyk-request-id"], qt.Not(qt.IsNil)) // still there
	c.Assert(doc.Paths["/orgs/{orgId}/projects"].Get.Responses["200"].Value.Headers["snyk-version-served"], qt.IsNil)     // now removed
	c.Assert(doc.Paths["/orgs/{org_id}/projects/{project_id}"].Delete.Parameters, qt.HasLen, 3)                           // x-private-matter removed

	c.Assert(doc.Paths["/orgs/{orgId}/projects"].Extensions["x-snyk-api-resource"], qt.IsNil) // now removed
	c.Assert(doc.Extensions["x-snyk-api-lifecycle"], qt.Not(qt.IsNil))                        // still there
}
