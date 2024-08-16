package versionware_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
	"time"

	qt "github.com/frankban/quicktest"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"

	"github.com/snyk/vervet/v8/versionware"
)

const (
	v20210820 = `
openapi: 3.0.0
x-snyk-api-version: '2021-08-20'
info:
  title: 'Validator'
  version: '0.0.0'
paths:
  /test/{id}:
    get:
      operationId: getTest
      description: get a test
      parameters:
        - in: path
          name: id
          schema:
            type: string
          required: true
        - in: query
          name: version
          schema:
            type: string
          required: true
      responses:
        '200':
          description: 'respond with test resource'
          content:
            application/json:
              schema: { $ref: '#/components/schemas/TestResource' }
        '400': { $ref: '#/components/responses/ErrorResponse' }
        '404': { $ref: '#/components/responses/ErrorResponse' }
        '500': { $ref: '#/components/responses/ErrorResponse' }
components:
  schemas:
    TestContents:
      type: object
      properties:
        name:
          type: string
        expected:
          type: number
        actual:
          type: number
      required: [name, expected, actual]
      additionalProperties: false
    TestResource:
      type: object
      properties:
        id:
          type: string
        contents:
          { $ref: '#/components/schemas/TestContents' }
      required: [id, contents]
      additionalProperties: false
    Error:
      type: object
      properties:
        code:
          type: string
        message:
          type: string
      required: [code, message]
      additionalProperties: false
  responses:
    ErrorResponse:
      description: 'an error occurred'
      content:
        application/json:
          schema: { $ref: '#/components/schemas/Error' }
`
	v20210916 = `
openapi: 3.0.0
x-snyk-api-version: '2021-09-16'
info:
  title: 'Validator'
  version: '0.0.0'
paths:
  /test:
    post:
      operationId: newTest
      description: create a new test
      parameters:
        - in: query
          name: version
          schema:
            type: string
          required: true
      requestBody:
        required: true
        content:
          application/json:
            schema: { $ref: '#/components/schemas/TestContents' }
      responses:
        '201':
          description: 'created test'
          content:
            application/json:
              schema: { $ref: '#/components/schemas/TestResource' }
        '400': { $ref: '#/components/responses/ErrorResponse' }
        '500': { $ref: '#/components/responses/ErrorResponse' }
  /test/{id}:
    get:
      operationId: getTest
      description: get a test
      parameters:
        - in: path
          name: id
          schema:
            type: string
          required: true
        - in: query
          name: version
          schema:
            type: string
          required: true
      responses:
        '200':
          description: 'respond with test resource'
          content:
            application/json:
              schema: { $ref: '#/components/schemas/TestResource' }
        '400': { $ref: '#/components/responses/ErrorResponse' }
        '404': { $ref: '#/components/responses/ErrorResponse' }
        '500': { $ref: '#/components/responses/ErrorResponse' }
components:
  schemas:
    TestContents:
      type: object
      properties:
        name:
          type: string
        expected:
          type: number
        actual:
          type: number
        noodles:
          type: boolean
      required: [name, expected, actual, noodles]
      additionalProperties: false
    TestResource:
      type: object
      properties:
        id:
          type: string
        contents:
          { $ref: '#/components/schemas/TestContents' }
      required: [id, contents]
      additionalProperties: false
    Error:
      type: object
      properties:
        code:
          type: string
        message:
          type: string
      required: [code, message]
      additionalProperties: false
  responses:
    ErrorResponse:
      description: 'an error occurred'
      content:
        application/json:
          schema: { $ref: '#/components/schemas/Error' }
`
)

type validatorTestHandler struct {
	contentType       string
	getBody, postBody string
	errBody           string
	errStatusCode     int
}

const v20210916_Body = `{"id": "42", "contents": {"name": "foo", "expected": 9, "actual": 10, "noodles": true}}`

func (h validatorTestHandler) withDefaults() validatorTestHandler {
	if h.contentType == "" {
		h.contentType = "application/json"
	}
	if h.getBody == "" {
		h.getBody = v20210916_Body
	}
	if h.postBody == "" {
		h.postBody = v20210916_Body
	}
	if h.errBody == "" {
		h.errBody = `{"code":"bad","message":"bad things"}`
	}
	return h
}

var testUrlRE = regexp.MustCompile(`^/test(/\d+)?$`)

func (h *validatorTestHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", h.contentType)
	if h.errStatusCode != 0 {
		w.WriteHeader(h.errStatusCode)
		_, err := w.Write([]byte(h.errBody))
		if err != nil {
			panic(err)
		}
		return
	}
	if !testUrlRE.MatchString(r.URL.Path) {
		w.WriteHeader(http.StatusNotFound)
		_, err := w.Write([]byte(h.errBody))
		if err != nil {
			panic(err)
		}
		return
	}
	switch r.Method {
	case "GET":
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(h.getBody))
		if err != nil {
			panic(err)
		}
	case "POST":
		w.WriteHeader(http.StatusCreated)
		_, err := w.Write([]byte(h.postBody))
		if err != nil {
			panic(err)
		}
	default:
		http.Error(w, h.errBody, http.StatusMethodNotAllowed)
	}
}

func TestValidator(t *testing.T) {
	c := qt.New(t)
	ctx := context.Background()
	docs := make([]*openapi3.T, 2)
	for i, specStr := range []string{v20210820, v20210916} {
		doc, err := openapi3.NewLoader().LoadFromData([]byte(specStr))
		c.Assert(err, qt.IsNil)
		err = doc.Validate(ctx)
		c.Assert(err, qt.IsNil)
		docs[i] = doc
	}

	type testRequest struct {
		method, path, body, contentType string
	}
	type testResponse struct {
		statusCode int
		body       string
	}
	tests := []struct {
		name     string
		handler  validatorTestHandler
		options  []openapi3filter.ValidatorOption
		request  testRequest
		response testResponse
		strict   bool
	}{{
		name:    "valid GET",
		handler: validatorTestHandler{}.withDefaults(),
		request: testRequest{
			method: "GET",
			path:   "/test/42?version=2021-09-17",
		},
		response: testResponse{
			200, v20210916_Body,
		},
		strict: true,
	}, {
		name:    "valid POST",
		handler: validatorTestHandler{}.withDefaults(),
		request: testRequest{
			method:      "POST",
			path:        "/test?version=2021-09-17",
			body:        `{"name": "foo", "expected": 9, "actual": 10, "noodles": true}`,
			contentType: "application/json",
		},
		response: testResponse{
			201, v20210916_Body,
		},
		strict: true,
	}, {
		name:    "not found; no GET operation for /test",
		handler: validatorTestHandler{}.withDefaults(),
		request: testRequest{
			method: "GET",
			path:   "/test?version=2021-09-17",
		},
		response: testResponse{
			404, "Not Found\n",
		},
		strict: true,
	}, {
		name:    "not found; no POST operation for /test/42",
		handler: validatorTestHandler{}.withDefaults(),
		request: testRequest{
			method: "POST",
			path:   "/test/42?version=2021-09-17",
		},
		response: testResponse{
			404, "Not Found\n",
		},
		strict: true,
	}, {
		name:    "invalid request; missing version",
		handler: validatorTestHandler{}.withDefaults(),
		request: testRequest{
			method: "GET",
			path:   "/test/42",
		},
		response: testResponse{
			400, "Bad Request\n",
		},
		strict: true,
	}, {
		name:    "invalid POST request; wrong property type",
		handler: validatorTestHandler{}.withDefaults(),
		request: testRequest{
			method:      "POST",
			path:        "/test?version=2021-09-17",
			body:        `{"name": "foo", "expected": "nine", "actual": "ten", "noodles": false}`,
			contentType: "application/json",
		},
		response: testResponse{
			400, "Bad Request\n",
		},
		strict: true,
	}, {
		name:    "invalid POST request; missing property",
		handler: validatorTestHandler{}.withDefaults(),
		request: testRequest{
			method:      "POST",
			path:        "/test?version=2021-09-17",
			body:        `{"name": "foo", "expected": 9}`,
			contentType: "application/json",
		},
		response: testResponse{
			400, "Bad Request\n",
		},
		strict: true,
	}, {
		name:    "invalid POST request; extra property",
		handler: validatorTestHandler{}.withDefaults(),
		request: testRequest{
			method:      "POST",
			path:        "/test?version=2021-09-17",
			body:        `{"name": "foo", "expected": 9, "actual": 10, "noodles": false, "ideal": 8}`,
			contentType: "application/json",
		},
		response: testResponse{
			400, "Bad Request\n",
		},
		strict: true,
	}, {
		name: "valid response; 404 error",
		handler: validatorTestHandler{
			contentType:   "application/json",
			errBody:       `{"code": "404", "message": "not found"}`,
			errStatusCode: 404,
		}.withDefaults(),
		request: testRequest{
			method: "GET",
			path:   "/test/42?version=2021-09-17",
		},
		response: testResponse{
			404, `{"code": "404", "message": "not found"}`,
		},
		strict: true,
	}, {
		name: "invalid response; invalid error",
		handler: validatorTestHandler{
			errBody:       `"not found"`,
			errStatusCode: 404,
		}.withDefaults(),
		request: testRequest{
			method: "GET",
			path:   "/test/42?version=2021-09-17",
		},
		response: testResponse{
			500, "Internal Server Error\n",
		},
		strict: true,
	}, {
		name: "invalid POST response; not strict",
		handler: validatorTestHandler{
			postBody: `{"id": "42", "contents": {"name": "foo", "expected": 9, "actual": 10}, "extra": true}`,
		}.withDefaults(),
		request: testRequest{
			method:      "POST",
			path:        "/test?version=2021-09-17",
			body:        `{"name": "foo", "expected": 9, "actual": 10, "noodles": true}`,
			contentType: "application/json",
		},
		response: testResponse{
			statusCode: 201,
			body:       `{"id": "42", "contents": {"name": "foo", "expected": 9, "actual": 10}, "extra": true}`,
		},
		strict: false,
	}, {
		name:    "invalid GET for API in the future",
		handler: validatorTestHandler{}.withDefaults(),
		request: testRequest{
			method: "GET",
			path:   "/test/42?version=2023-09-17",
		},
		response: testResponse{
			400, "Bad Request\n",
		},
		strict: true,
	}}
	for i, test := range tests {
		c.Run(fmt.Sprintf("%d %s", i, test.name), func(c *qt.C) {
			// Set up a test HTTP server
			var h http.Handler
			s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				h.ServeHTTP(w, r)
			}))
			defer s.Close()

			config := versionware.DefaultValidatorConfig
			config.ServerURL = s.URL
			config.Options = append(config.Options, append(test.options, openapi3filter.Strict(test.strict))...)
			v, err := versionware.NewValidator(&config, docs...)
			c.Assert(err, qt.IsNil)
			v.SetToday(func() time.Time {
				return time.Date(2022, time.January, 21, 0, 0, 0, 0, time.UTC)
			})
			h = v.Middleware(&test.handler)

			// Test: make a client request
			var requestBody io.Reader
			if test.request.body != "" {
				requestBody = bytes.NewBufferString(test.request.body)
			}
			req, err := http.NewRequest(test.request.method, s.URL+test.request.path, requestBody)
			c.Assert(err, qt.IsNil)

			if test.request.contentType != "" {
				req.Header.Set("Content-Type", test.request.contentType)
			}
			resp, err := s.Client().Do(req)
			c.Assert(err, qt.IsNil)
			defer resp.Body.Close()
			c.Assert(test.response.statusCode, qt.Equals, resp.StatusCode)

			body, err := io.ReadAll(resp.Body)
			c.Assert(err, qt.IsNil)
			c.Assert(test.response.body, qt.Equals, string(body))
		})
	}
}

func TestValidatorConfig(t *testing.T) {
	c := qt.New(t)

	// No specs provided
	_, err := versionware.NewValidator(&versionware.ValidatorConfig{ServerURL: "://"})
	c.Assert(err, qt.ErrorMatches, `no OpenAPI versions provided`)

	// Invalid server URL
	_, err = versionware.NewValidator(&versionware.ValidatorConfig{ServerURL: "://"}, &openapi3.T{})
	c.Assert(err, qt.ErrorMatches, `invalid ServerURL: parse "://": missing protocol scheme`)

	// Missing version in OpenAPI spec
	_, err = versionware.NewValidator(&versionware.ValidatorConfig{ServerURL: "http://example.com"}, &openapi3.T{})
	c.Assert(err, qt.ErrorMatches, `extension "x-snyk-api-version" not found`)

	docs := make([]*openapi3.T, 2)
	for i, specStr := range []string{v20210820, v20210916} {
		doc, err := openapi3.NewLoader().LoadFromData([]byte(specStr))
		c.Assert(err, qt.IsNil)
		err = doc.Validate(context.Background())
		c.Assert(err, qt.IsNil)
		docs[i] = doc
	}

	// Invalid server URL
	_, err = versionware.NewValidator(&versionware.ValidatorConfig{ServerURL: "localhost:8080"}, docs...)
	c.Assert(
		err,
		qt.ErrorMatches,
		`invalid ServerURL: unsupported scheme "localhost" \(did you forget to specify the scheme://\?\)`,
	)

	// Valid
	_, err = versionware.NewValidator(&versionware.ValidatorConfig{ServerURL: "http://localhost:8080"}, docs...)
	c.Assert(err, qt.IsNil)
	c.Assert(docs[0].Servers[0].URL, qt.Equals, "http://localhost:8080")
}
