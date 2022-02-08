# compiler

```go
import "github.com/snyk/vervet/v3/internal/compiler"
```

## Index

- [func ResourceSpecFiles(rcConfig *config.ResourceSet) ([]string, error)](<#func-resourcespecfiles>)
- [type Compiler](<#type-compiler>)
  - [func New(ctx context.Context, proj *config.Project, options ...CompilerOption) (*Compiler, error)](<#func-new>)
  - [func (c *Compiler) Build(ctx context.Context, apiName string) error](<#func-compiler-build>)
  - [func (c *Compiler) BuildAll(ctx context.Context) error](<#func-compiler-buildall>)
  - [func (c *Compiler) LintOutput(ctx context.Context, apiName string) error](<#func-compiler-lintoutput>)
  - [func (c *Compiler) LintOutputAll(ctx context.Context) error](<#func-compiler-lintoutputall>)
  - [func (c *Compiler) LintResources(ctx context.Context, apiName string) error](<#func-compiler-lintresources>)
  - [func (c *Compiler) LintResourcesAll(ctx context.Context) error](<#func-compiler-lintresourcesall>)
- [type CompilerOption](<#type-compileroption>)
  - [func LinterFactory(f func(ctx context.Context, lc *config.Linter) (linter.Linter, error)) CompilerOption](<#func-linterfactory>)


## func ResourceSpecFiles

```go
func ResourceSpecFiles(rcConfig *config.ResourceSet) ([]string, error)
```

ResourceSpecFiles returns all matching spec files for a config\.Resource\.

## type Compiler

A Compiler checks and builds versioned API resource inputs into aggregated OpenAPI versioned outputs\, as determined by an API project configuration\.

```go
type Compiler struct {
    // contains filtered or unexported fields
}
```

### func New

```go
func New(ctx context.Context, proj *config.Project, options ...CompilerOption) (*Compiler, error)
```

New returns a new Compiler for a given project configuration\.

### func \(\*Compiler\) Build

```go
func (c *Compiler) Build(ctx context.Context, apiName string) error
```

Build builds an aggregate versioned OpenAPI spec for a specific API by name in the project\.

### func \(\*Compiler\) BuildAll

```go
func (c *Compiler) BuildAll(ctx context.Context) error
```

BuildAll builds all APIs in the project\.

### func \(\*Compiler\) LintOutput

```go
func (c *Compiler) LintOutput(ctx context.Context, apiName string) error
```

LintOutput applies configured linting rules to the build output\.

### func \(\*Compiler\) LintOutputAll

```go
func (c *Compiler) LintOutputAll(ctx context.Context) error
```

LintOutputAll lints output of all APIs in the project\.

### func \(\*Compiler\) LintResources

```go
func (c *Compiler) LintResources(ctx context.Context, apiName string) error
```

LintResources checks the inputs of an API's resources with the configured linter\.

### func \(\*Compiler\) LintResourcesAll

```go
func (c *Compiler) LintResourcesAll(ctx context.Context) error
```

LintResourcesAll lints resources in all APIs in the project\.

## type CompilerOption

CompilerOption applies a configuration option to a Compiler\.

```go
type CompilerOption func(*Compiler) error
```

### func LinterFactory

```go
func LinterFactory(f func(ctx context.Context, lc *config.Linter) (linter.Linter, error)) CompilerOption
```

LinterFactory configures a Compiler to use a custom factory function for instantiating Linters\.

