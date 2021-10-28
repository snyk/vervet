
package output

import "embed"

// Embed compiled OpenAPI specs in Go projects.

//go:embed 2021-06-01~experimental/spec.json
//go:embed 2021-06-01~experimental/spec.yaml
//go:embed 2021-06-01~beta/spec.json
//go:embed 2021-06-01~beta/spec.yaml
//go:embed 2021-06-01/spec.json
//go:embed 2021-06-01/spec.yaml
//go:embed 2021-06-04~experimental/spec.json
//go:embed 2021-06-04~experimental/spec.yaml
//go:embed 2021-06-04~beta/spec.json
//go:embed 2021-06-04~beta/spec.yaml
//go:embed 2021-06-04/spec.json
//go:embed 2021-06-04/spec.yaml
//go:embed 2021-06-07~experimental/spec.json
//go:embed 2021-06-07~experimental/spec.yaml
//go:embed 2021-06-07~beta/spec.json
//go:embed 2021-06-07~beta/spec.yaml
//go:embed 2021-06-07/spec.json
//go:embed 2021-06-07/spec.yaml
//go:embed 2021-06-13~experimental/spec.json
//go:embed 2021-06-13~experimental/spec.yaml
//go:embed 2021-06-13~beta/spec.json
//go:embed 2021-06-13~beta/spec.yaml
//go:embed 2021-06-13/spec.json
//go:embed 2021-06-13/spec.yaml
var Versions embed.FS
