package vervet_test

import (
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/snyk/vervet/v6"
	"github.com/snyk/vervet/v6/testdata"
)

func TestRefRemover(t *testing.T) {
	c := qt.New(t)
	doc, err := vervet.NewDocumentFile(testdata.Path("resources/projects/2021-08-20/spec.yaml"))
	c.Assert(err, qt.IsNil)
	resp400 := doc.Paths["/orgs/{org_id}/projects/{project_id}"].Delete.Responses["400"]
	errDoc := resp400.Value.Content["application/vnd.api+json"].Schema
	c.Assert(err, qt.IsNil)
	c.Assert("{\"$ref\":\"../errors.yaml#/ErrorDocument\"}", qt.JSONEquals, errDoc)
	in := vervet.NewRefRemover(errDoc)
	err = in.RemoveRef()
	c.Assert(err, qt.IsNil)
	c.Assert(err, qt.IsNil)
	//nolint:lll // acked
	c.Assert("{\"additionalProperties\":false,\"example\":{\"errors\":[{\"detail\":\"Permission denied for this "+
		"resource\",\"status\":\"403\"}],\"jsonapi\":{\"version\":\"1.0\"}},\"properties\":{\"errors\":{\"example\":"+
		"[{\"detail\":\"Permission denied for this resource\",\"status\":\"403\"}],\"items\":{\"additionalProperties\""+
		":false,\"example\":{\"detail\":\"Not Found\",\"status\":\"404\"},\"properties\":{\"detail\":{\"description\":"+
		"\"A human-readable explanation specific to this occurrence of the problem.\",\"example\":\"The request was "+
		"missing these required fields: ...\",\"type\":\"string\"},\"id\":{\"description\":\"A unique identifier for "+
		"this particular occurrence of the problem.\",\"example\":\"f16c31b5-6129-4571-add8-d589da9be524\",\"format\""+
		":\"uuid\",\"type\":\"string\"},\"meta\":{\"additionalProperties\":true,\"example\":{\"key\":\"value\"},\"type\""+
		":\"object\"},\"source\":{\"additionalProperties\":false,\"example\":{\"pointer\":\"/data/attributes\"},\"properties\""+
		":{\"parameter\":{\"description\":\"A string indicating which URI query parameter caused the error.\",\"example\""+
		":\"param1\",\"type\":\"string\"},\"pointer\":{\"description\":\"A JSON Pointer [RFC6901] to the associated entity "+
		"in the request document.\",\"example\":\"/data/attributes\",\"type\":\"string\"}},\"type\":\"object\"},\"status\""+
		":{\"description\":\"The HTTP status code applicable to this problem, expressed as a string value.\",\"example\""+
		":\"400\",\"pattern\":\"^[45]\\\\d\\\\d$\",\"type\":\"string\"}},\"required\":[\"status\",\"detail\"],\"type\""+
		":\"object\"},\"minItems\":1,\"type\":\"array\"},\"jsonapi\":{\"additionalProperties\":false,\"example\":"+
		"{\"version\":\"1.0\"},\"properties\":{\"version\":{\"description\":\"Version of the JSON API specification "+
		"this server supports.\",\"example\":\"1.0\",\"pattern\":\"^(0|[1-9]\\\\d*)\\\\.(0|[1-9]\\\\d*)$\",\"type\""+
		":\"string\"}},\"required\":[\"version\"],\"type\":\"object\"}},\"required\":[\"jsonapi\",\"errors\"],\"type\""+
		":\"object\"}\n", qt.JSONEquals, errDoc)
}

func TestCollator(t *testing.T) {
	c := qt.New(t)
	collator := vervet.NewCollator()
	projects, err := vervet.LoadResourceVersions(testdata.Path("conflict-components/projects"))
	c.Assert(err, qt.IsNil)
	projectV, err := projects.At("2021-06-04~experimental")
	c.Assert(err, qt.IsNil)
	examples, err := vervet.LoadResourceVersions(testdata.Path("conflict-components/_examples"))
	c.Assert(err, qt.IsNil)
	examplesV, err := examples.At("2021-06-01~experimental")
	c.Assert(err, qt.IsNil)

	err = collator.Collate(projectV)
	c.Assert(err, qt.IsNil)
	err = collator.Collate(examplesV)
	c.Assert(err, qt.IsNil)

	result := collator.Result()
	c.Assert(
		result.Paths["/orgs/{orgId}/projects"].
			Get.Responses["200"].
			Value.
			Content["application/vnd.api+json"].
			Schema.Value.Properties["jsonapi"].Ref,
		qt.Equals,
		"#/components/schemas/JsonApi",
	)
	schemaRef := result.
		Paths["/examples/hello-world/{id}"].
		Get.
		Responses["200"].
		Value.
		Content["application/vnd.api+json"].
		Schema.
		Value.
		Properties["jsonapi"]
	c.Assert(schemaRef.Ref, qt.Equals, "")
	c.Assert("{\"additionalProperties\":false,\"example\":{\"version\":\"1.0\"},\"properties\":{\"version\":"+
		"{\"description\":\"Version of the JSON API specification this server supports.\",\"example\":\"1.0\","+
		"\"pattern\":\"^(0|[1-9]\\\\d*)\\\\.(0|[1-9]\\\\d*)$\",\"type\":\"string\"}},\"required\":[\"version\"],\"type\""+
		":\"object\"}\n", qt.JSONEquals, schemaRef.Value)
	c.Assert(result.Components.Schemas["JsonApi"], qt.IsNotNil)

	projectParameterRef := result.Paths["/orgs/{orgId}/projects"].Get.Parameters[0]
	c.Assert(projectParameterRef.Ref, qt.Equals, "#/components/parameters/Version")
	exampleParameterRef := result.Paths["/examples/hello-world/{id}"].Get.Parameters[0]
	c.Assert(exampleParameterRef.Ref, qt.Equals, "")
	//nolint:lll // acked
	c.Assert("{\"description\":\"The requested version of the endpoint to process the request\",\"example\""+
		":\"2021-06-04\",\"in\":\"query\",\"name\":\"version\",\"required\":true,\"schema\":{\"description\":"+
		"\"Requested API version\",\"pattern\":\"^(wip|work-in-progress|experimental|beta|((([0-9]{4})-([0-1][0-9]))"+
		"-((3[01])|(0[1-9])|([12][0-9]))(~(wip|work-in-progress|experimental|beta))?))$\",\"type\":\"string\"}}\n", qt.JSONEquals, exampleParameterRef.Value)

	projectConflictRef := result.Paths["/orgs/{orgId}/projects"].Get.Parameters[6]
	exampleConflictRef := result.Paths["/examples/hello-world/{id}"].Get.Parameters[3]
	c.Assert(projectConflictRef.Ref, qt.Not(qt.Equals), exampleConflictRef.Ref)

	projectResp400Ref := result.Paths["/orgs/{orgId}/projects"].Get.Responses["400"]
	c.Assert(projectResp400Ref.Ref, qt.Equals, "#/components/responses/400")
	exampleResp400Ref := result.Paths["/examples/hello-world/{id}"].Get.Responses["400"]
	c.Assert(exampleResp400Ref.Ref, qt.Equals, "")
	c.Assert("{\"content\":{\"application/vnd.api+json\":{\"schema\":{\"additionalProperties\":false,\"example\":{"+
		"\"errors\":[{\"detail\":\"Permission denied for this resource\",\"status\":\"403\"}],\"jsonapi\":{\"version\":"+
		"\"1.0\"}},\"properties\":{\"errors\":{\"example\":[{\"detail\":\"Permission denied for this resource\",\"status"+
		"\":\"403\"}],\"items\":{\"additionalProperties\":false,\"example\":{\"detail\":\"Not Found\",\"status\":\"404\"}"+
		",\"properties\":{\"detail\":{\"description\":\"A human-readable explanation specific to this occurrence of the "+
		"problem.\",\"example\":\"The request was missing these required fields: ...\",\"type\":\"string\"},\"id\":"+
		"{\"description\":\"A unique identifier for this particular occurrence of the problem.\",\"example\":"+
		"\"f16c31b5-6129-4571-add8-d589da9be524\",\"format\":\"uuid\",\"type\":\"string\"},\"meta\":"+
		"{\"additionalProperties\":true,\"example\":{\"key\":\"value\"},\"type\":\"object\"},\"source\":"+
		"{\"additionalProperties\":false,\"example\":{\"pointer\":\"/data/attributes\"},\"properties\":"+
		"{\"parameter\":{\"description\":\"A string indicating which URI query parameter caused the error."+
		"\",\"example\":\"param1\",\"type\":\"string\"},\"pointer\":{\"description\":\"A JSON Pointer [RFC6901] to the "+
		"associated entity in the request document.\",\"example\":\"/data/attributes\",\"type\":\"string\"}},\"type\":"+
		"\"object\"},\"status\":{\"description\":\"The HTTP status code applicable to this problem, expressed as a "+
		"string value.\",\"example\":\"400\",\"pattern\":\"^[45]\\\\d\\\\d$\",\"type\":\"string\"}},\"required\":"+
		"[\"status\",\"detail\"],\"type\":\"object\"},\"minItems\":1,\"type\":\"array\"},\"jsonapi\":"+
		"{\"additionalProperties\":false,\"example\":{\"version\":\"1.0\"},\"properties\":{\"version\":"+
		"{\"description\":\"Version of the JSON API specification this server supports.\",\"example\":\"1.0\","+
		"\"pattern\":\"^(0|[1-9]\\\\d*)\\\\.(0|[1-9]\\\\d*)$\",\"type\":\"string\"}},\"required\":[\"version\"],\"type\""+
		":\"object\"}},\"required\":[\"jsonapi\",\"errors\"],\"type\":\"object\"}}},\"description\":\"Bad Request: A "+
		"parameter provided as a part of the request was invalid.\",\"headers\":{\"deprecation\":{\"description\":\""+
		"A header containing the deprecation date of the underlying endpoint. For more information, please refer to "+
		"the deprecation header RFC:\\nhttps://tools.ietf.org/id/draft-dalal-deprecation-header-01.html\\n\",\"example"+
		"\":\"2021-07-01T00:00:00Z\",\"schema\":{\"format\":\"date-time\",\"type\":\"string\"}},\"snyk-request-id\":"+
		"{\"description\":\"A header containing a unique id used for tracking this request. If you are reporting an "+
		"issue to Snyk it's very helpful to provide this ID.\\n\",\"example\":\"4b58e274-ec62-4fab-917b-1d2c48d6bdef\""+
		",\"schema\":{\"format\":\"uuid\",\"type\":\"string\"}},\"snyk-version-lifecycle-stage\":{\"description\":"+
		"\"A header containing the version stage of the endpoint. This stage describes the guarantees snyk provides "+
		"surrounding stability of the endpoint.\\n\",\"schema\":{\"enum\":[\"wip\",\"experimental\",\"beta\",\"ga\","+
		"\"deprecated\",\"sunset\"],\"example\":\"ga\",\"type\":\"string\"}},\"snyk-version-requested\":{\"description\""+
		":\"A header containing the version of the endpoint requested by the caller.\",\"example\":\"2021-06-04\",\""+
		"schema\":{\"description\":\"Requested API version\",\"pattern\":\"^(wip|work-in-progress|experimental|beta|"+
		"((([0-9]{4})-([0-1][0-9]))-((3[01])|(0[1-9])|([12][0-9]))(~(wip|work-in-progress|experimental|beta))?))$\""+
		",\"type\":\"string\"}},\"snyk-version-served\":{\"description\":\"A header containing the version of the "+
		"endpoint that was served by the API.\",\"example\":\"2021-06-04\",\"schema\":{\"description\":\"Resolved API "+
		"version\",\"pattern\":\"^((([0-9]{4})-([0-1][0-9]))-((3[01])|(0[1-9])|([12][0-9]))(~"+
		"(wip|work-in-progress|experimental|beta))?)$\",\"type\":\"string\"}},\"sunset\":{\"description\":"+
		"\"A header containing the date of when the underlying endpoint will be removed. This header is only present if "+
		"the endpoint has been deprecated. Please refer to the RFC for more information:"+
		"\\nhttps://datatracker.ietf.org/doc/html/rfc8594\\n\",\"example\":\"2021-08-02T00:00:00Z\",\"schema\":"+
		"{\"format\":\"date-time\",\"type\":\"string\"}}}}\n", qt.JSONEquals, exampleResp400Ref.Value)
}

func TestCollateUseFirstRoute(t *testing.T) {
	c := qt.New(t)
	collator := vervet.NewCollator(vervet.UseFirstRoute(true))
	examples1, err := vervet.LoadResourceVersions(testdata.Path("conflict/_examples"))
	c.Assert(err, qt.IsNil)
	examples1v, err := examples1.At("2021-06-15~experimental")
	c.Assert(err, qt.IsNil)

	examples2, err := vervet.LoadResourceVersions(testdata.Path("conflict/_examples2"))
	c.Assert(err, qt.IsNil)
	examples2v, err := examples2.At("2021-06-15~experimental")
	c.Assert(err, qt.IsNil)

	err = collator.Collate(examples1v)
	c.Assert(err, qt.IsNil)
	err = collator.Collate(examples2v)
	c.Assert(err, qt.IsNil)

	result := collator.Result()

	// First path chosen, route matching rules ignore path variable
	c.Assert(result.Paths["/examples/hello-world/{id1}"], qt.Not(qt.IsNil))
	c.Assert(result.Paths["/examples/hello-world/{id2}"], qt.IsNil)

	// First chosen path has description expected
	c.Assert(result.Paths["/examples/hello-world/{id1}"].Get.Description, qt.Contains, " - from example 1")
}

func TestCollatePathConflict(t *testing.T) {
	c := qt.New(t)
	collator := vervet.NewCollator(vervet.UseFirstRoute(false))
	examples1, err := vervet.LoadResourceVersions(testdata.Path("conflict/_examples"))
	c.Assert(err, qt.IsNil)
	examples1v, err := examples1.At("2021-06-15~experimental")
	c.Assert(err, qt.IsNil)

	examples2, err := vervet.LoadResourceVersions(testdata.Path("conflict/_examples2"))
	c.Assert(err, qt.IsNil)
	examples2v, err := examples2.At("2021-06-15~experimental")
	c.Assert(err, qt.IsNil)

	err = collator.Collate(examples1v)
	c.Assert(err, qt.IsNil)
	err = collator.Collate(examples2v)
	c.Assert(err, qt.ErrorMatches, `.*conflict in #/paths /examples/hello-world/{id2}: declared in both.*`)
	c.Assert(err, qt.ErrorMatches, `.*conflict in #/paths /examples/hello-world: declared in both.*`)
}

func TestCollateMergingResources(t *testing.T) {
	c := qt.New(t)
	collator := vervet.NewCollator(vervet.UseFirstRoute(true))

	newService, err := vervet.LoadResourceVersions(testdata.Path("competing-specs/special_projects"))
	c.Assert(err, qt.IsNil)
	specV1, err := newService.At("2023-03-13~experimental")
	c.Assert(err, qt.IsNil)

	originalService, err := vervet.LoadResourceVersions(testdata.Path("competing-specs/projects"))
	c.Assert(err, qt.IsNil)
	specV2, err := originalService.At("2021-08-20~experimental")
	c.Assert(err, qt.IsNil)

	err = collator.Collate(specV2)
	c.Assert(err, qt.IsNil)
	err = collator.Collate(specV1)
	c.Assert(err, qt.IsNil)

	result := collator.Result()
	c.Assert(result.Paths["/orgs/{org_id}/projects/{project_id}"].Delete.Responses["204"], qt.IsNotNil)
}
