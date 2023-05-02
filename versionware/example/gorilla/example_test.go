package gorilla_test

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"

	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/gorilla/mux"
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

	// Top level router
	root := mux.NewRouter()
	// Wrap all requests with the prometheus middleware (API and non-API)
	promMiddleware := promware_std.HandlerProvider("", promware.New(promware.Config{
		Recorder: metrics.NewRecorder(metrics.Config{}),
	}))
	h = promMiddleware(root)

	// Create a subrouter for the versioned API
	apiRouter := root.PathPrefix("/api").Subrouter()

	// Load OpenAPI specs for all released API versions.
	specs, err := vervet.LoadVersions(releases.Versions)
	if err != nil {
		log.Fatal(err) //nolint:gocritic //acked
	}

	// Add request and response validation middleware to the API subrouter
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
	thingsRouter := apiRouter.PathPrefix("/things").Subrouter()
	thingsRouter.Handle("/{id}", versionware.NewHandler([]versionware.VersionHandler{{
		Version: release_2021_11_01.Version,
		Handler: http.HandlerFunc(release_2021_11_01.GetThing(s)),
	}}...)).Methods("GET")
	thingsRouter.Handle("", versionware.NewHandler([]versionware.VersionHandler{{
		Version: release_2021_11_01.Version,
		Handler: http.HandlerFunc(release_2021_11_01.CreateThing(s)),
	}}...)).Methods("POST")
	thingsRouter.Handle("", versionware.NewHandler([]versionware.VersionHandler{{
		Version: release_2021_11_08.Version,
		Handler: release_2021_11_08.ListThings(s),
	}}...)).Methods("GET")
	thingsRouter.Handle("/{id}", versionware.NewHandler([]versionware.VersionHandler{{
		Version: release_2021_11_20.Version,
		Handler: release_2021_11_20.DeleteThing(s),
	}}...)).Methods("DELETE")

	// Observability stuff at the top-level, not part of the API
	root.Handle("/metrics", promhttp.Handler())
	root.HandleFunc("/healthcheck", func(w http.ResponseWriter, r *http.Request) {
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
