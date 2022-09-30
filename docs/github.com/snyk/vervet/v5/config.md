# config

```go
import "github.com/snyk/vervet/v5/config"
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
- [type Linter](<#type-linter>)
- [type Linters](<#type-linters>)
- [type OpticCILinter](<#type-opticcilinter>)
- [type Output](<#type-output>)
  - [func (o *Output) ResolvePaths() []string](<#func-output-resolvepaths>)
- [type Overlay](<#type-overlay>)
- [type Project](<#type-project>)
  - [func Load(r io.Reader) (*Project, error)](<#func-load>)
  - [func (p *Project) APINames() []string](<#func-project-apinames>)
- [type ResourceSet](<#type-resourceset>)
- [type SpectralLinter](<#type-spectrallinter>)


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

## func [Save](<https://github.com/snyk/vervet/blob/main/config/project.go#L94>)

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

### func [LoadGenerators](<https://github.com/snyk/vervet/blob/main/config/project.go#L80>)

```go
func LoadGenerators(r io.Reader) (Generators, error)
```

LoadGenerators loads Generators from their YAML representation\.

## type [Linter](<https://github.com/snyk/vervet/blob/main/config/linter.go#L15-L21>)

Linter describes a set of standards and rules that an API should satisfy\.

```go
type Linter struct {
    Name        string          `json:"-"`
    Description string          `json:"description,omitempty"`
    Spectral    *SpectralLinter `json:"spectral"`
    SweaterComb *OpticCILinter  `json:"sweater-comb"`
    OpticCI     *OpticCILinter  `json:"optic-ci"`
}
```

## type [Linters](<https://github.com/snyk/vervet/blob/main/config/linter.go#L12>)

Linters defines a named map of Linter instances\.

```go
type Linters map[string]*Linter
```

## type [OpticCILinter](<https://github.com/snyk/vervet/blob/main/config/linter.go#L66-L106>)

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

    // DEPRECATED: CIContext is no longer used and should be removed in the
    // next major release.
    CIContext string `json:"-"`

    // DEPRECATED: UploadResults is no longer used and should be removed in the
    // next major release. Uploading optic-ci comparison results to Optic
    // Cloud is determined by the presence of environment variables.
    UploadResults bool `json:"-"`

    // Exceptions are files that are excluded from CI checks. This is an escape
    // hatch of last resort, if a file needs to land and can't pass CI yet.
    // They are specified as a mapping from project relative path to sha256
    // sums of that spec file that is exempt. This makes the exception very
    // narrow -- only a specific version of a specific file is skipped, after
    // outside review and approval.
    Exceptions map[string][]string

    // ExtraArgs may be used to pass extra arguments to `optic-ci`.
    ExtraArgs []string `json:"extraArgs"`
}
```

## type [Output](<https://github.com/snyk/vervet/blob/main/config/api.go#L72-L76>)

Output defines where the aggregate versioned OpenAPI specs should be created during compilation\.

```go
type Output struct {
    Path   string   `json:"path,omitempty"`
    Paths  []string `json:"paths,omitempty"`
    Linter string   `json:"linter"`
}
```

### func \(\*Output\) [ResolvePaths](<https://github.com/snyk/vervet/blob/main/config/api.go#L80>)

```go
func (o *Output) ResolvePaths() []string
```

EffectivePaths returns a slice of effective configured output paths\, whether a single or multiple output paths have been configured\.

## type [Overlay](<https://github.com/snyk/vervet/blob/main/config/api.go#L65-L68>)

An Overlay defines additional OpenAPI documents to merge into the aggregate OpenAPI spec when compiling an API\. These might include special endpoints that should be included in the aggregate API but are not versioned\, or top\-level descriptions of the API itself\.

```go
type Overlay struct {
    Include string `json:"include"`
    Inline  string `json:"inline"`
}
```

## type [Project](<https://github.com/snyk/vervet/blob/main/config/project.go#L13-L18>)

Project defines collection of APIs and the standards they adhere to\.

```go
type Project struct {
    Version    string     `json:"version"`
    Linters    Linters    `json:"linters,omitempty"`
    Generators Generators `json:"generators,omitempty"`
    APIs       APIs       `json:"apis"`
}
```

### func [Load](<https://github.com/snyk/vervet/blob/main/config/project.go#L65>)

```go
func Load(r io.Reader) (*Project, error)
```

Load loads a Project configuration from its YAML representation\.

### func \(\*Project\) [APINames](<https://github.com/snyk/vervet/blob/main/config/project.go#L21>)

```go
func (p *Project) APINames() []string
```

APINames returns the API names in deterministic ascending order\.

## type [ResourceSet](<https://github.com/snyk/vervet/blob/main/config/api.go#L44-L50>)

A ResourceSet defines a set of versioned resources that adhere to the same standards\.

Versioned resources are expressed as individual OpenAPI documents in a directory structure:

\+\-resource | \+\-2021\-08\-01 | | | \+\-spec\.yaml | \+\-\<implementation code\, etc\. can go here\> | \+\-2021\-08\-15 | | | \+\-spec\.yaml | \+\-\<implementation code\, etc\. can go here\> \.\.\.

Each YYYY\-mm\-dd directory under a resource is a version\.  The spec\.yaml in each version is a complete OpenAPI document describing the resource at that version\.

```go
type ResourceSet struct {
    Description     string             `json:"description"`
    Linter          string             `json:"linter"`
    LinterOverrides map[string]Linters `json:"linter-overrides"`
    Path            string             `json:"path"`
    Excludes        []string           `json:"excludes"`
}
```

## type [SpectralLinter](<https://github.com/snyk/vervet/blob/main/config/linter.go#L45-L62>)

SpectralLinter identifies a Linter as a collection of Spectral rulesets\.

```go
type SpectralLinter struct {

    // Rules are a list of Spectral ruleset file locations
    Rules []string `json:"rules"`

    // Script identifies the path to the spectral script to use for linting.
    // If not defined linting will look for spectral-cli on $PATH.
    Script string `json:"script"`

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

