package chi_test

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/go-chi/chi/v5"
	chiware "github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	metrics "github.com/slok/go-http-metrics/metrics/prometheus"
	promware "github.com/slok/go-http-metrics/middleware"
	promware_std "github.com/slok/go-http-metrics/middleware/std"

	"github.com/snyk/vervet/v5"
	"github.com/snyk/vervet/v5/versionware"
	. "github.com/snyk/vervet/v5/versionware/example"
	"github.com/snyk/vervet/v5/versionware/example/releases"
	release_2021_11_01 "github.com/snyk/vervet/v5/versionware/example/resources/things/2021-11-01"
	release_2021_11_08 "github.com/snyk/vervet/v5/versionware/example/resources/things/2021-11-08"
	release_2021_11_20 "github.com/snyk/vervet/v5/versionware/example/resources/things/2021-11-20"
	"github.com/snyk/vervet/v5/versionware/example/store"
)

func Example() {
	// Set up a test HTTP server
	var h http.Handler
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.ServeHTTP(w, r)
	}))
	defer srv.Close()

	// Top level router for test server
	root := chi.NewRouter()
	h = root

	// Middleware to wrap all requests
	root.Use(chiware.RequestID)
	root.Use(chiware.RealIP)
	root.Use(chiware.Logger)
	root.Use(chiware.Recoverer)
	root.Use(chiware.Timeout(30 * time.Second))
	root.Use(promware_std.HandlerProvider("", promware.New(promware.Config{
		Recorder: metrics.NewRecorder(metrics.Config{}),
	})))

	// Create a router for just the versioned API
	apiRouter := chi.NewRouter()

	// Load OpenAPI specs for all released API versions.
	specs, err := vervet.LoadVersions(releases.Versions)
	if err != nil {
		log.Fatal(err) //nolint:gocritic //acked
	}

	// Add request and response validation middleware to the API router
	validator, err := versionware.NewValidator(&versionware.ValidatorConfig{
		// We're going to mount our API at /api below...
		ServerURL: srv.URL + "/api",
		Options:   []openapi3filter.ValidatorOption{openapi3filter.Strict(true)},
	}, specs...)
	if err != nil {
		log.Fatal(err)
	}
	// Only validate the API requests (not other top-level stuff)
	apiRouter.Use(validator.Middleware)

	// A new storage backend
	s := store.New()

	// Router for the "things" resource
	// As the service grows, these could be pulled out into per-resource sub-packages...
	thingsRouter := chi.NewRouter()
	thingsRouter.Get("/{id}", versionware.NewHandler([]versionware.VersionHandler{{
		Version: release_2021_11_01.Version,
		Handler: http.HandlerFunc(release_2021_11_01.GetThing(s)),
	}}...).ServeHTTP)
	thingsRouter.Get("/", versionware.NewHandler([]versionware.VersionHandler{{
		Version: release_2021_11_08.Version,
		Handler: release_2021_11_08.ListThings(s),
	}}...).ServeHTTP)
	thingsRouter.Post("/", versionware.NewHandler([]versionware.VersionHandler{{
		Version: release_2021_11_01.Version,
		Handler: http.HandlerFunc(release_2021_11_01.CreateThing(s)),
	}}...).ServeHTTP)
	thingsRouter.Delete("/{id}", versionware.NewHandler([]versionware.VersionHandler{{
		Version: release_2021_11_20.Version,
		Handler: release_2021_11_20.DeleteThing(s),
	}}...).ServeHTTP)

	// Mount the "things" resource router at /things in the API
	apiRouter.Mount("/things", thingsRouter)

	// Mount the entire API at /api
	root.Mount("/api", apiRouter)

	// Observability stuff at the top-level, not part of the API
	root.Get("/metrics", promhttp.Handler().ServeHTTP)
	root.Get("/healthcheck", func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte("OK"))
		if err != nil {
			panic(err)
		}
	})

	// Do a health check
	PrintResp(srv.Client().Get(srv.URL + "/healthcheck"))

	// Create some things
	PrintResp(srv.Client().Post(
		srv.URL+"/api/things?version=2021-11-01~experimental", "application/json",
		bytes.NewBufferString(`{"name":"foo","color":"blue","strangeness":32}`)))
	PrintResp(srv.Client().Post(
		srv.URL+"/api/things?version=2021-11-01~experimental", "application/json",
		bytes.NewBufferString(`{"name":"shiny","color":"green","strangeness":99}`)))
	PrintResp(srv.Client().Post(
		srv.URL+"/api/things?version=2021-11-01~experimental", "application/json",
		bytes.NewBufferString(`{"name":"cochineal","color":"red","strangeness":5}`)))

	// 404: no matching version
	PrintResp(srv.Client().Post(
		srv.URL+"/api/things?version=2021-10-01~experimental", "application/json",
		bytes.NewBufferString(`{"name":"cochineal","color":"red","strangeness":5}`)))

	// 400: create an invalid thing
	PrintResp(srv.Client().Post(
		srv.URL+"/api/things?version=2021-11-01~experimental", "application/json",
		bytes.NewBufferString(`{"name":"eggplant","color":"purple","banality":17}`)))

	// 200: get a thing
	PrintResp(srv.Client().Get(srv.URL + "/api/things/1?version=2021-11-10~experimental"))

	// Output:
	// 200 OK
	// 200 {"id":"1","created":"2022-01-14T00:23:50Z","attributes":{"name":"foo","color":"blue","strangeness":32}}
	// 200 {"id":"2","created":"2022-01-14T00:23:50Z","attributes":{"name":"shiny","color":"green","strangeness":99}}
	// 200 {"id":"3","created":"2022-01-14T00:23:50Z","attributes":{"name":"cochineal","color":"red","strangeness":5}}
	// 404 Not Found
	// 400 bad request
	// 200 {"id":"1","created":"2022-01-14T00:23:50Z","attributes":{"name":"foo","color":"blue","strangeness":32}}
}
