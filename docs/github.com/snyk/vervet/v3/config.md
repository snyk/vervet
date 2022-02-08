# config

```go
import "github.com/snyk/vervet/v3/config"
```

## Index

- [Constants](<#constants>)
- [func Save(w io.Writer, proj *Project) error](<#func-save>)
- [type API](<#type-api>)
- [type APIs](<#type-apis>)
- [type Generator](<#type-generator>)
- [type GeneratorData](<#type-generatordata>)
- [type GeneratorScope](<#type-generatorscope>)
- [type Generators](<#type-generators>)
- [type Linter](<#type-linter>)
- [type Linters](<#type-linters>)
- [type OpticCILinter](<#type-opticcilinter>)
- [type Output](<#type-output>)
- [type Overlay](<#type-overlay>)
- [type Project](<#type-project>)
  - [func Load(r io.Reader) (*Project, error)](<#func-load>)
  - [func (p *Project) APINames() []string](<#func-project-apinames>)
- [type ResourceSet](<#type-resourceset>)
- [type SpectralLinter](<#type-spectrallinter>)
- [type SweaterCombLinter](<#type-sweatercomblinter>)


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

## func Save

```go
func Save(w io.Writer, proj *Project) error
```

Save saves a Project configuration to YAML\.

## type API

An API defines how and where to build versioned OpenAPI documents from a source collection of individual resource specifications and additional overlay content to merge\.

```go
type API struct {
    Name      string         `json:"-"`
    Resources []*ResourceSet `json:"resources"`
    Overlays  []*Overlay     `json:"overlays"`
    Output    *Output        `json:"output"`
}
```

## type APIs

APIs defines a named map of API instances\.

```go
type APIs map[string]*API
```

## type Generator

Generator describes how files are generated for a resource\.

```go
type Generator struct {
    Name     string                    `json:"-"`
    Scope    GeneratorScope            `json:"scope"`
    Filename string                    `json:"filename,omitempty"`
    Template string                    `json:"template"`
    Files    string                    `json:"files,omitempty"`
    Data     map[string]*GeneratorData `json:"data,omitempty"`
}
```

## type GeneratorData

GeneratorData describes an item that is added to a generator's template data context\.

```go
type GeneratorData struct {
    FieldName string `json:"-"`
    Include   string `json:"include"`
}
```

## type GeneratorScope

GeneratorScope determines the template context when running the generator\. Different scopes allow templates to operate over a single resource version\, or all versions in a resource\, for example\.

```go
type GeneratorScope string
```

## type Generators

Generators defines a named map of Generator instances\.

```go
type Generators map[string]*Generator
```

## type Linter

Linter describes a set of standards and rules that an API should satisfy\.

```go
type Linter struct {
    Name        string             `json:"-"`
    Description string             `json:"description,omitempty"`
    Spectral    *SpectralLinter    `json:"spectral"`
    SweaterComb *SweaterCombLinter `json:"sweater-comb"`
    OpticCI     *OpticCILinter     `json:"optic-ci"`
}
```

## type Linters

Linters defines a named map of Linter instances\.

```go
type Linters map[string]*Linter
```

## type OpticCILinter

OpticCILinter identifies an Optic CI Linter\, which is distributed as a self\-contained docker image\.

```go
type OpticCILinter struct {
    // Image identifies the Optic CI docker image to use for linting.
    Image string

    // Script identifies the path to the Optic CI script to use for linting.
    // Mutually exclusive with Image; if Script is specified Docker will not be
    // used.
    Script string

    // Original is where to source the original version of an OpenAPI spec file
    // when comparing. If empty, all changes are assumed to be new additions.
    Original string `json:"original,omitempty"`

    // Proposed is where to source the proposed changed version of an OpenAPI
    // spec file when comparing. If empty, this is assumed to be the
    // local working copy.
    Proposed string `json:"proposed,omitempty"`

    // Debug turns on debug logging.
    Debug bool `json:"debug,omitempty"`
}
```

## type Output

Output defines where the aggregate versioned OpenAPI specs should be created during compilation\.

```go
type Output struct {
    Path   string `json:"path"`
    Linter string `json:"linter"`
}
```

## type Overlay

An Overlay defines additional OpenAPI documents to merge into the aggregate OpenAPI spec when compiling an API\. These might include special endpoints that should be included in the aggregate API but are not versioned\, or top\-level descriptions of the API itself\.

```go
type Overlay struct {
    Include string `json:"include"`
    Inline  string `json:"inline"`
}
```

## type Project

Project defines collection of APIs and the standards they adhere to\.

```go
type Project struct {
    Version    string     `json:"version"`
    Linters    Linters    `json:"linters,omitempty"`
    Generators Generators `json:"generators,omitempty"`
    APIs       APIs       `json:"apis"`
}
```

### func Load

```go
func Load(r io.Reader) (*Project, error)
```

Load loads a Project configuration from its YAML representation\.

### func \(\*Project\) APINames

```go
func (p *Project) APINames() []string
```

APINames returns the API names in deterministic ascending order\.

## type ResourceSet

A ResourceSet defines a set of versioned resources that adhere to the same standards\.

Versioned resources are expressed as individual OpenAPI documents in a directory structure:

\+\-resource | \+\-2021\-08\-01 | | | \+\-spec\.yaml | \+\-\<implementation code\, etc\. can go here\> | \+\-2021\-08\-15 | | | \+\-spec\.yaml | \+\-\<implementation code\, etc\. can go here\> \.\.\.

Each YYYY\-mm\-dd directory under a resource is a version\.  The spec\.yaml in each version is a complete OpenAPI document describing the resource at that version\.

```go
type ResourceSet struct {
    Description     string             `json:"description"`
    Linter          string             `json:"linter"`
    LinterOverrides map[string]Linters `json:"linter-overrides"`
    Generators      []string           `json:"generators"`
    Path            string             `json:"path"`
    Excludes        []string           `json:"excludes"`
}
```

## type SpectralLinter

SpectralLinter identifies a Linter as a collection of Spectral rulesets\.

```go
type SpectralLinter struct {

    // Rules are a list of Spectral ruleset file locations
    Rules []string `json:"rules"`

    // ExtraArgs may be used to pass extra arguments to `spectral lint`. If not
    // specified, the default arguments `--format text` are used when running
    // spectral. The `-r` flag must not be specified here, as this argument is
    // automatically added from the Rules setting above.
    //
    // See https://meta.stoplight.io/docs/spectral/ZG9jOjI1MTg1-spectral-cli
    // for the options supported.
    ExtraArgs []string `json:"extraArgs"`
}
```

## type SweaterCombLinter

SweaterCombLinter identifies a Sweater Comb Linter\, which is distributed as a self\-contained docker image\.

```go
type SweaterCombLinter struct {
    // Image identifies the Sweater Comb docker image to use for linting.
    Image string

    // Rules are a list of Spectral ruleset file locations
    // These may be absolute paths to Sweater Comb rules, such as /rules/apinext.yaml.
    // Or, they may be relative paths to files in this project.
    Rules []string `json:"rules"`

    // ExtraArgs may be used to pass extra arguments to `spectral lint`. The
    // Sweater Comb image includes Spectral. This has the same function as
    // SpectralLinter.ExtraArgs above.
    ExtraArgs []string `json:"extraArgs"`
}
```

