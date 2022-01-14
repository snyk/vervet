package output

import "embed"

// Embed compiled OpenAPI specs in Go projects.

//go:embed 2021-06-01~experimental/spec.json
//go:embed 2021-06-01~experimental/spec.yaml
//go:embed 2021-06-04~experimental/spec.json
//go:embed 2021-06-04~experimental/spec.yaml
//go:embed 2021-06-07~experimental/spec.json
//go:embed 2021-06-07~experimental/spec.yaml
//go:embed 2021-06-13~experimental/spec.json
//go:embed 2021-06-13~experimental/spec.yaml
//go:embed 2021-06-13~beta/spec.json
//go:embed 2021-06-13~beta/spec.yaml
//go:embed 2021-08-20~experimental/spec.json
//go:embed 2021-08-20~experimental/spec.yaml
//go:embed 2021-08-20~beta/spec.json
//go:embed 2021-08-20~beta/spec.yaml
// Versions contains OpenAPI specs for each distinct release version.
var Versions embed.FS
