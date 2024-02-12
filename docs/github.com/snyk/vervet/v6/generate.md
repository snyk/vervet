# generate

```go
import "github.com/snyk/vervet/v6/generate"
```

## Index

- [func Generate(params GeneratorParams) error](<#func-generate>)
- [type GeneratorParams](<#type-generatorparams>)


## func [Generate](<https://github.com/snyk/vervet/blob/main/generate/generate.go#L29>)

```go
func Generate(params GeneratorParams) error
```

Generate executes code generators against OpenAPI specs\.

## type [GeneratorParams](<https://github.com/snyk/vervet/blob/main/generate/generate.go#L16-L26>)

GeneratorParams contains the metadata needed to execute code generators\.

```go
type GeneratorParams struct {
    ProjectDir     string
    ConfigFile     string
    Generators     []string
    GeneratorsFile string
    Force          bool
    Debug          bool
    DryRun         bool
    FS             fs.FS
    Functions      template.FuncMap
}
```

