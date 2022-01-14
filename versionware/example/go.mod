module github.com/snyk/vervet/versionware/example

go 1.16

require (
	github.com/getkin/kin-openapi v0.87.0
	github.com/go-chi/chi/v5 v5.0.7
	github.com/gorilla/mux v1.8.0
	github.com/prometheus/client_golang v1.11.0
	github.com/slok/go-http-metrics v0.10.0
	github.com/snyk/vervet v1.5.1
)

replace github.com/snyk/vervet => ../..
