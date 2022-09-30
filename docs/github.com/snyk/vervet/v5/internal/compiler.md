# compiler

```go
import "github.com/snyk/vervet/v5/internal/compiler"
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


## func [ResourceSpecFiles](<https://github.com/snyk/vervet/blob/main/internal/compiler/compiler.go#L179>)

```go
func ResourceSpecFiles(rcConfig *config.ResourceSet) ([]string, error)
```

ResourceSpecFiles returns all matching spec files for a config\.Resource\.

## type [Compiler](<https://github.com/snyk/vervet/blob/main/internal/compiler/compiler.go#L28-L33>)

A Compiler checks and builds versioned API resource inputs into aggregated OpenAPI versioned outputs\, as determined by an API project configuration\.

```go
type Compiler struct {
    // contains filtered or unexported fields
}
```

### func [New](<https://github.com/snyk/vervet/blob/main/internal/compiler/compiler.go#L79>)

```go
func New(ctx context.Context, proj *config.Project, options ...CompilerOption) (*Compiler, error)
```

New returns a new Compiler for a given project configuration\.

### func \(\*Compiler\) [Build](<https://github.com/snyk/vervet/blob/main/internal/compiler/compiler.go#L258>)

```go
func (c *Compiler) Build(ctx context.Context, apiName string) error
```

Build builds an aggregate versioned OpenAPI spec for a specific API by name in the project\.

### func \(\*Compiler\) [BuildAll](<https://github.com/snyk/vervet/blob/main/internal/compiler/compiler.go#L407>)

```go
func (c *Compiler) BuildAll(ctx context.Context) error
```

BuildAll builds all APIs in the project\.

### func \(\*Compiler\) [LintOutput](<https://github.com/snyk/vervet/blob/main/internal/compiler/compiler.go#L412>)

```go
func (c *Compiler) LintOutput(ctx context.Context, apiName string) error
```

LintOutput applies configured linting rules to the build output\.

### func \(\*Compiler\) [LintOutputAll](<https://github.com/snyk/vervet/blob/main/internal/compiler/compiler.go#L440>)

```go
func (c *Compiler) LintOutputAll(ctx context.Context) error
```

LintOutputAll lints output of all APIs in the project\.

### func \(\*Compiler\) [LintResources](<https://github.com/snyk/vervet/blob/main/internal/compiler/compiler.go#L184>)

```go
func (c *Compiler) LintResources(ctx context.Context, apiName string) error
```

LintResources checks the inputs of an API's resources with the configured linter\.

### func \(\*Compiler\) [LintResourcesAll](<https://github.com/snyk/vervet/blob/main/internal/compiler/compiler.go#L241>)

```go
func (c *Compiler) LintResourcesAll(ctx context.Context) error
```

LintResourcesAll lints resources in all APIs in the project\.

## type [CompilerOption](<https://github.com/snyk/vervet/blob/main/internal/compiler/compiler.go#L36>)

CompilerOption applies a configuration option to a Compiler\.

```go
type CompilerOption func(*Compiler) error
```

### func [LinterFactory](<https://github.com/snyk/vervet/blob/main/internal/compiler/compiler.go#L40>)

```go
func LinterFactory(f func(ctx context.Context, lc *config.Linter) (linter.Linter, error)) CompilerOption
```

LinterFactory configures a Compiler to use a custom factory function for instantiating Linters\.

