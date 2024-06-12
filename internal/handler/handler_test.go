package handler_test

import (
	"context"
	"io"
	"net/http/httptest"
	"testing"
	"time"

	qt "github.com/frankban/quicktest"

	"github.com/snyk/vervet/v6"
	"github.com/snyk/vervet/v6/config"
	"github.com/snyk/vervet/v6/internal/handler"
)

func TestHealth(t *testing.T) {
	c := qt.New(t)
	cfg, h := setup()

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	h.ServeHTTP(w, req)
	c.Assert(w.Code, qt.Equals, 200)
	contents, err := io.ReadAll(w.Result().Body)
	c.Assert(err, qt.IsNil)
	c.Assert(contents, qt.JSONEquals, map[string]interface{}{
		"msg":      "success",
		"services": cfg.Services,
	})
}

func TestOpenapi(t *testing.T) {
	c := qt.New(t)
	_, h := setup()

	for _, path := range []string{"/openapi", "/openapi/"} {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", path, nil)
		h.ServeHTTP(w, req)
		c.Assert(w.Code, qt.Equals, 200)
		contents, err := io.ReadAll(w.Result().Body)
		c.Assert(err, qt.IsNil)
		c.Assert(contents, qt.JSONEquals, []string{
			"2021-06-04~experimental",
			"2021-10-20~experimental",
			"2021-10-20~beta",
			"2022-01-16~experimental",
			"2022-01-16~beta",
			"2022-01-16",
		})
	}
}

func TestMetrics(t *testing.T) {
	c := qt.New(t)
	_, h := setup()

	w := httptest.NewRecorder()
	// NOTE: Metrics are counted globally, so in order for this metrics test to
	// reliably pass, this particular version should not be requested in any of
	// the other tests within this package. Otherwise the count will be thrown
	// off (and will race among tests bumping the same tag).
	req := httptest.NewRequest("GET", "/openapi/2021-10-20~beta", nil)
	h.ServeHTTP(w, req)
	c.Assert(w.Code, qt.Equals, 200)

	w = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/metrics", nil)
	h.ServeHTTP(w, req)
	c.Assert(w.Code, qt.Equals, 200)
	contents, err := io.ReadAll(w.Result().Body)
	c.Assert(err, qt.IsNil)
	// Metrics captured the /openapi request above
	c.Assert(
		string(contents),
		qt.Contains,
		`vu_http_response_size_bytes_count{code="200",handler="/openapi/2021-10-20~beta",method="GET",service=""} 1`,
	)
}

func TestOpenapiVersion(t *testing.T) {
	c := qt.New(t)
	_, h := setup()

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/openapi/2022-01-16~beta", nil)
	h.ServeHTTP(w, req)
	c.Assert(w.Code, qt.Equals, 200)
	contents, err := io.ReadAll(w.Result().Body)
	c.Assert(err, qt.IsNil)
	c.Assert(contents, qt.DeepEquals, []byte("got 2022-01-16~beta"))
}

func TestOpenapiVersionNotFound(t *testing.T) {
	c := qt.New(t)
	_, h := setup()

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/openapi/2021-01-16~beta", nil)
	h.ServeHTTP(w, req)
	c.Assert(w.Code, qt.Equals, 404)
	contents, err := io.ReadAll(w.Result().Body)
	c.Assert(err, qt.IsNil)
	c.Assert(contents, qt.DeepEquals, []byte("Version not found\n"))
}

func setup() (*config.ServerConfig, *handler.Handler) {
	cfg := &config.ServerConfig{
		Services: []config.ServiceConfig{{
			Name: "petfood", URL: "http://petfood.svc.cluster.local",
		}, {
			Name: "animals", URL: "http://animals.svc.cluster.local",
		}},
	}
	st := &mockStorage{}
	h := handler.New(cfg, st, handler.UseDefaultMiddleware)
	return cfg, h
}

type mockStorage struct{}

func (s *mockStorage) NotifyVersions(ctx context.Context, name string, versions []string, scrapeTime time.Time) error {
	return nil
}

func (s *mockStorage) CollateVersions(ctx context.Context, serviceFilter map[string]bool) error {
	return nil
}

func (s *mockStorage) HasVersion(ctx context.Context, name string, version string, digest string) (bool, error) {
	return true, nil
}

func (s *mockStorage) NotifyVersion(
	ctx context.Context,
	name string,
	version string,
	contents []byte,
	scrapeTime time.Time,
) error {
	return nil
}

func (s *mockStorage) VersionIndex(ctx context.Context) (vervet.VersionIndex, error) {
	return vervet.NewVersionIndex(vervet.VersionSlice{
		vervet.MustParseVersion("2021-06-04~experimental"),
		vervet.MustParseVersion("2021-10-20~experimental"),
		vervet.MustParseVersion("2021-10-20~beta"),
		vervet.MustParseVersion("2022-01-16~experimental"),
		vervet.MustParseVersion("2022-01-16~beta"),
		vervet.MustParseVersion("2022-01-16~ga"),
	}), nil
}

func (s *mockStorage) Version(ctx context.Context, version string) ([]byte, error) {
	return []byte("got " + version), nil
}
