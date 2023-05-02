package releases

import "embed"

// Embed compiled OpenAPI specs in Go projects.

//go:embed 2021-11-01~experimental/spec.json
//go:embed 2021-11-01~experimental/spec.yaml
//go:embed 2021-11-08~experimental/spec.json
//go:embed 2021-11-08~experimental/spec.yaml
//go:embed 2021-11-20~experimental/spec.json
//go:embed 2021-11-20~experimental/spec.yaml

// Versions contains OpenAPI specs for each distinct release version.
var Versions embed.FS
