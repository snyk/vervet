# config

```go
import "github.com/snyk/vervet/v6/config"
```

## Index

- [Constants](<#constants>)
- [func Save(w io.Writer, proj *Project) error](<#func-save>)
- [type API](<#type-api>)
- [type APIs](<#type-apis>)
- [type Generator](<#type-generator>)
- [type GeneratorScope](<#type-generatorscope>)
- [type Generators](<#type-generators>)
  - [func LoadGenerators(r io.Reader) (Generators, error)](<#func-loadgenerators>)
- [type Output](<#type-output>)
  - [func (o *Output) ResolvePaths() []string](<#func-output-resolvepaths>)
- [type Overlay](<#type-overlay>)
- [type Project](<#type-project>)
  - [func Load(r io.Reader) (*Project, error)](<#func-load>)
  - [func (p *Project) APINames() []string](<#func-project-apinames>)
- [type ResourceSet](<#type-resourceset>)


## Constants

```go
const (
    // GeneratorScopeDefault indicates the default scope should be used in
    // configuration.
    GeneratorScopeDefault = ""

    // GeneratorScopeVersion indicates the generator operates on a single
    // resource version.
    GeneratorScopeVersion = "version"

    // GeneratorScopeResource indicates the generator operates on all versions
    // in a resource. This is useful for generating version routers, for
    // example.
    GeneratorScopeResource = "resource"
)
```

## func [Save](<https://github.com/snyk/vervet/blob/main/config/project.go#L85>)

```go
func Save(w io.Writer, proj *Project) error
```

Save saves a Project configuration to YAML\.

## type [API](<https://github.com/snyk/vervet/blob/main/config/api.go#L15-L20>)

An API defines how and where to build versioned OpenAPI documents from a source collection of individual resource specifications and additional overlay content to merge\.

```go
type API struct {
    Name      string         `json:"-"`
    Resources []*ResourceSet `json:"resources"`
    Overlays  []*Overlay     `json:"overlays"`
    Output    *Output        `json:"output"`
}
```

## type [APIs](<https://github.com/snyk/vervet/blob/main/config/api.go#L10>)

APIs defines a named map of API instances\.

```go
type APIs map[string]*API
```

## type [Generator](<https://github.com/snyk/vervet/blob/main/config/generator.go#L11-L18>)

Generator describes how files are generated for a resource\.

```go
type Generator struct {
    Name      string         `json:"-"`
    Scope     GeneratorScope `json:"scope"`
    Filename  string         `json:"filename,omitempty"`
    Template  string         `json:"template"`
    Files     string         `json:"files,omitempty"`
    Functions string         `json:"functions,omitempty"`
}
```

## type [GeneratorScope](<https://github.com/snyk/vervet/blob/main/config/generator.go#L39>)

GeneratorScope determines the template context when running the generator\. Different scopes allow templates to operate over a single resource version\, or all versions in a resource\, for example\.

```go
type GeneratorScope string
```

## type [Generators](<https://github.com/snyk/vervet/blob/main/config/generator.go#L8>)

Generators defines a named map of Generator instances\.

```go
type Generators map[string]*Generator
```

### func [LoadGenerators](<https://github.com/snyk/vervet/blob/main/config/project.go#L71>)

```go
func LoadGenerators(r io.Reader) (Generators, error)
```

LoadGenerators loads Generators from their YAML representation\.

## type [Output](<https://github.com/snyk/vervet/blob/main/config/api.go#L71-L74>)

Output defines where the aggregate versioned OpenAPI specs should be created during compilation\.

```go
type Output struct {
    Path  string   `json:"path,omitempty"`
    Paths []string `json:"paths,omitempty"`
}
```

### func \(\*Output\) [ResolvePaths](<https://github.com/snyk/vervet/blob/main/config/api.go#L78>)

```go
func (o *Output) ResolvePaths() []string
```

EffectivePaths returns a slice of effective configured output paths\, whether a single or multiple output paths have been configured\.

## type [Overlay](<https://github.com/snyk/vervet/blob/main/config/api.go#L64-L67>)

An Overlay defines additional OpenAPI documents to merge into the aggregate OpenAPI spec when compiling an API\. These might include special endpoints that should be included in the aggregate API but are not versioned\, or top\-level descriptions of the API itself\.

```go
type Overlay struct {
    Include string `json:"include"`
    Inline  string `json:"inline"`
}
```

## type [Project](<https://github.com/snyk/vervet/blob/main/config/project.go#L12-L16>)

Project defines collection of APIs and the standards they adhere to\.

```go
type Project struct {
    Version    string     `json:"version"`
    Generators Generators `json:"generators,omitempty"`
    APIs       APIs       `json:"apis"`
}
```

### func [Load](<https://github.com/snyk/vervet/blob/main/config/project.go#L56>)

```go
func Load(r io.Reader) (*Project, error)
```

Load loads a Project configuration from its YAML representation\.

### func \(\*Project\) [APINames](<https://github.com/snyk/vervet/blob/main/config/project.go#L19>)

```go
func (p *Project) APINames() []string
```

APINames returns the API names in deterministic ascending order\.

## type [ResourceSet](<https://github.com/snyk/vervet/blob/main/config/api.go#L45-L49>)

A ResourceSet defines a set of versioned resources that adhere to the same standards\.

Versioned resources are expressed as individual OpenAPI documents in a directory structure:

\+\-resource

```
|
+-2021-08-01
| |
| +-spec.yaml
| +-<implementation code, etc. can go here>
|
+-2021-08-15
| |
| +-spec.yaml
| +-<implementation code, etc. can go here>
...
```

Each YYYY\-mm\-dd directory under a resource is a version\.  The spec\.yaml in each version is a complete OpenAPI document describing the resource at that version\.

```go
type ResourceSet struct {
    Description string   `json:"description"`
    Path        string   `json:"path"`
    Excludes    []string `json:"excludes"`
}
```

