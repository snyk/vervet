package versionware

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	qt "github.com/frankban/quicktest"
	"github.com/getkin/kin-openapi/openapi3"
)

func TestResolveStability(t *testing.T) {
	c := qt.New(t)
	v2022_07_10_beta := parseOpenAPI(c, `
openapi: 3.0.0
x-snyk-api-version: 2022-07-10~beta
info:
  title: 'test'
paths:
  /foo:
    get:
      operationId: getFoo
      parameters:
        - in: query
          name: version
          schema:
            type: string
          required: true
        - in: query
          name: color
          schema:
            type: string
          required: true
      responses:
        '200':
          content:
            application/json:
              schema:
                type: object
`)
	v2022_07_20_ga := parseOpenAPI(c, `
openapi: 3.0.0
x-snyk-api-version: 2022-07-20~ga
info:
  title: 'test'
paths:
  /foo:
    get:
      operationId: getFoo
      parameters:
        - in: query
          name: version
          schema:
            type: string
          required: true
        - in: query
          name: flavor
          schema:
            type: string
          required: true
      responses:
        '200':
          content:
            application/json:
              schema:
                type: object
`)

	v, err := NewValidator(nil, v2022_07_10_beta, v2022_07_20_ga)
	c.Assert(err, qt.IsNil)
	h := v.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte("{}"))
		c.Assert(err, qt.IsNil)
	}))

	tests := []struct {
		desc   string
		url    string
		status int
	}{{
		desc:   "valid request, version resolves to 2022-07-10~beta",
		url:    "/foo?color=red&version=2022-07-19~beta",
		status: 200,
	}, {
		desc:   "invalid request, version resolves to 2022-07-10~beta",
		url:    "/foo?flavor=blue&version=2022-07-19~beta",
		status: 400,
	}, {
		desc:   "valid request, version resolves to 2022-07-20~ga",
		url:    "/foo?flavor=flav&version=2022-07-20",
		status: 200,
	}, {
		desc:   "valid request, version resolves to 2022-07-20~ga",
		url:    "/foo?flavor=flav&version=2022-07-21",
		status: 200,
	}, {
		desc:   "invalid request, version resolves to 2022-07-10~beta",
		url:    "/foo?flavor=flav&version=2022-07-21~beta",
		status: 400,
	}, {
		desc:   "valid request, version resolves to 2022-07-10~beta",
		url:    "/foo?color=octarine&version=2022-07-21~beta",
		status: 200,
	}, {
		desc:   "invalid request, version resolves to 2022-07-20~ga",
		url:    "/foo?color=red&version=2022-07-21~ga",
		status: 400,
	}, {
		desc:   "valid request, version resolves to 2022-07-10~beta",
		url:    "/foo?color=red&version=2022-07-20~beta",
		status: 200,
	}, {
		desc:   "valid request, version resolves to expected validator",
		url:    "/foo?flavor=flav&version=2022-07-21",
		status: 200,
	}, {
		desc:   "version does not resolve, no such version",
		url:    "/foo?version=2022-01-01",
		status: 404,
	}}

	for i, test := range tests {
		c.Run(fmt.Sprintf("test %d", i), func(c *qt.C) {
			req := httptest.NewRequest("GET", test.url, nil)
			w := httptest.NewRecorder()
			h.ServeHTTP(w, req)
			resp := w.Result()
			c.Check(resp.StatusCode, qt.Equals, test.status, qt.Commentf("%s %s", test.desc, test.url))
		})
	}
}

func parseOpenAPI(c *qt.C, docstr string) *openapi3.T {
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData([]byte(docstr))
	c.Assert(err, qt.IsNil)
	return doc
}
