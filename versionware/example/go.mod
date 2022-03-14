module github.com/snyk/vervet/v4/versionware/example

go 1.16

require (
	github.com/frankban/quicktest v1.13.0
	github.com/getkin/kin-openapi v0.92.0
	github.com/go-chi/chi/v5 v5.0.7
	github.com/gorilla/mux v1.8.0
	github.com/prometheus/client_golang v1.11.0
	github.com/slok/go-http-metrics v0.10.0
	github.com/snyk/vervet/v4 v4.0.0-00010101000000-000000000000
)

replace github.com/snyk/vervet/v4 => ../..
