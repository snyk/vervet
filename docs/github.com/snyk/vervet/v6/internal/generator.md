# generator

```go
import "github.com/snyk/vervet/v6/internal/generator"
```

## Index

- [func MapPathOperations(p *openapi3.PathItem) map[string]*openapi3.Operation](<#func-mappathoperations>)
- [func NewMap(generatorsConf config.Generators, options ...Option) (map[string]*Generator, error)](<#func-newmap>)
- [type Generator](<#type-generator>)
  - [func New(conf *config.Generator, options ...Option) (*Generator, error)](<#func-new>)
  - [func (g *Generator) Execute(resources ResourceMap) ([]string, error)](<#func-generator-execute>)
  - [func (g *Generator) Scope() config.GeneratorScope](<#func-generator-scope>)
- [type OperationMap](<#type-operationmap>)
  - [func MapResourceOperations(resourceVersions *vervet.ResourceVersions) (OperationMap, error)](<#func-mapresourceoperations>)
- [type OperationVersion](<#type-operationversion>)
- [type Option](<#type-option>)
  - [func Debug(debug bool) Option](<#func-debug>)
  - [func DryRun(dryRun bool) Option](<#func-dryrun>)
  - [func Filesystem(fileSystem fs.FS) Option](<#func-filesystem>)
  - [func Force(force bool) Option](<#func-force>)
  - [func Functions(funcs template.FuncMap) Option](<#func-functions>)
  - [func Here(here string) Option](<#func-here>)
- [type ResourceKey](<#type-resourcekey>)
- [type ResourceMap](<#type-resourcemap>)
  - [func MapResources(proj *config.Project) (ResourceMap, error)](<#func-mapresources>)
- [type ResourceScope](<#type-resourcescope>)
  - [func (s *ResourceScope) Resource() string](<#func-resourcescope-resource>)
- [type VersionScope](<#type-versionscope>)
  - [func (s *VersionScope) Resource() string](<#func-versionscope-resource>)
  - [func (s *VersionScope) Version() *vervet.Version](<#func-versionscope-version>)


## func [MapPathOperations](<https://github.com/snyk/vervet/blob/main/internal/generator/resources.go#L64>)

```go
func MapPathOperations(p *openapi3.PathItem) map[string]*openapi3.Operation
```

MapPathOperations returns a mapping from HTTP method to \*openapi3\.Operation for a given \*openapi3\.PathItem\.

## func [NewMap](<https://github.com/snyk/vervet/blob/main/internal/generator/generator.go#L37>)

```go
func NewMap(generatorsConf config.Generators, options ...Option) (map[string]*Generator, error)
```

NewMap instanstiates a map of Generators from configuration\.

## type [Generator](<https://github.com/snyk/vervet/blob/main/internal/generator/generator.go#L21-L34>)

Generator generates files for new resources from data models and templates\.

```go
type Generator struct {
    // contains filtered or unexported fields
}
```

### func [New](<https://github.com/snyk/vervet/blob/main/internal/generator/generator.go#L50>)

```go
func New(conf *config.Generator, options ...Option) (*Generator, error)
```

New returns a new Generator from configuration\.

### func \(\*Generator\) [Execute](<https://github.com/snyk/vervet/blob/main/internal/generator/generator.go#L210>)

```go
func (g *Generator) Execute(resources ResourceMap) ([]string, error)
```

Execute runs the generator on the given resources\.

### func \(\*Generator\) [Scope](<https://github.com/snyk/vervet/blob/main/internal/generator/generator.go#L314>)

```go
func (g *Generator) Scope() config.GeneratorScope
```

Scope returns the configured scope type of the generator\.

## type [OperationMap](<https://github.com/snyk/vervet/blob/main/internal/generator/resources.go#L25>)

OperationMap defines a mapping from operation name to all versions of that operation within a resource\.

```go
type OperationMap map[string][]OperationVersion
```

### func [MapResourceOperations](<https://github.com/snyk/vervet/blob/main/internal/generator/resources.go#L38>)

```go
func MapResourceOperations(resourceVersions *vervet.ResourceVersions) (OperationMap, error)
```

MapResourceOperations returns a mapping from operation ID to all versions of that operation\.

## type [OperationVersion](<https://github.com/snyk/vervet/blob/main/internal/generator/resources.go#L29-L34>)

OperationVersion represents a version of an operation within a collection of resource versions\.

```go
type OperationVersion struct {
    *vervet.ResourceVersion
    Path      string
    Method    string
    Operation *openapi3.Operation
}
```

## type [Option](<https://github.com/snyk/vervet/blob/main/internal/generator/generator.go#L162>)

Option configures a Generator\.

```go
type Option func(g *Generator)
```

### func [Debug](<https://github.com/snyk/vervet/blob/main/internal/generator/generator.go#L172>)

```go
func Debug(debug bool) Option
```

Debug turns on template debug logging\.

### func [DryRun](<https://github.com/snyk/vervet/blob/main/internal/generator/generator.go#L180>)

```go
func DryRun(dryRun bool) Option
```

DryRun executes templates and lists the files that would be generated without actually generating them\.

### func [Filesystem](<https://github.com/snyk/vervet/blob/main/internal/generator/generator.go#L195>)

```go
func Filesystem(fileSystem fs.FS) Option
```

Filesystem sets the filesytem that the generator checks for templates\.

### func [Force](<https://github.com/snyk/vervet/blob/main/internal/generator/generator.go#L165>)

```go
func Force(force bool) Option
```

Force configures the Generator to overwrite generated artifacts\.

### func [Functions](<https://github.com/snyk/vervet/blob/main/internal/generator/generator.go#L201>)

```go
func Functions(funcs template.FuncMap) Option
```

### func [Here](<https://github.com/snyk/vervet/blob/main/internal/generator/generator.go#L188>)

```go
func Here(here string) Option
```

Here sets the \.Here scope property\. This is typically relative to the location of the generators config file\.

## type [ResourceKey](<https://github.com/snyk/vervet/blob/main/internal/generator/resources.go#L14-L18>)

ResourceKey uniquely identifies an API resource\.

```go
type ResourceKey struct {
    API      string
    Resource string
    Path     string
}
```

## type [ResourceMap](<https://github.com/snyk/vervet/blob/main/internal/generator/resources.go#L21>)

ResourceMap defines a mapping from API resource identity to its versions\.

```go
type ResourceMap map[ResourceKey]*vervet.ResourceVersions
```

### func [MapResources](<https://github.com/snyk/vervet/blob/main/internal/generator/resources.go#L98>)

```go
func MapResources(proj *config.Project) (ResourceMap, error)
```

MapResources returns a mapping of all resources managed within a Vervet project\.

## type [ResourceScope](<https://github.com/snyk/vervet/blob/main/internal/generator/generator.go#L271-L282>)

ResourceScope identifies a resource that the generator is building for\.

```go
type ResourceScope struct {
    // ResourceVersions contains all the versions of this resource.
    *vervet.ResourceVersions
    // API is name of the API containing this resource.
    API string
    // Path is the path to the resource directory.
    Path string
    // Here is the directory containing the executing template.
    Here string
    // Env is a map of template values read from the os environment.
    Env map[string]string
}
```

### func \(\*ResourceScope\) [Resource](<https://github.com/snyk/vervet/blob/main/internal/generator/generator.go#L285>)

```go
func (s *ResourceScope) Resource() string
```

Resource returns the name of the resource in scope\.

## type [VersionScope](<https://github.com/snyk/vervet/blob/main/internal/generator/generator.go#L291-L301>)

VersionScope identifies a distinct version of a resource that the generator is building for\.

```go
type VersionScope struct {
    *vervet.ResourceVersion
    // API is name of the API containing this resource.
    API string
    // Path is the path to the resource directory.
    Path string
    // Here is the directory containing the generator template.
    Here string
    // Env is a map of template values read from the os environment.
    Env map[string]string
}
```

### func \(\*VersionScope\) [Resource](<https://github.com/snyk/vervet/blob/main/internal/generator/generator.go#L304>)

```go
func (s *VersionScope) Resource() string
```

Resource returns the name of the resource in scope\.

### func \(\*VersionScope\) [Version](<https://github.com/snyk/vervet/blob/main/internal/generator/generator.go#L309>)

```go
func (s *VersionScope) Version() *vervet.Version
```

Version returns the version of the resource in scope\.

