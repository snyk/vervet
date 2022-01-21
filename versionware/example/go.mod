module github.com/snyk/vervet/v3/versionware/example

go 1.16

require (
	github.com/frankban/quicktest v1.13.0 // indirect
	github.com/getkin/kin-openapi v0.88.0
	github.com/go-chi/chi/v5 v5.0.7
	github.com/gorilla/mux v1.8.0
	github.com/prometheus/client_golang v1.11.0
	github.com/prometheus/procfs v0.6.0 // indirect
	github.com/slok/go-http-metrics v0.10.0
	github.com/snyk/vervet/v3 v3.0.0-00010101000000-000000000000
)

replace github.com/snyk/vervet/v3 => ../..
