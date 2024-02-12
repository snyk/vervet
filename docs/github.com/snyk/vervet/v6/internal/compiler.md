# compiler

```go
import "github.com/snyk/vervet/v6/internal/compiler"
```

## Index

- [func ResourceSpecFiles(rcConfig *config.ResourceSet) ([]string, error)](<#func-resourcespecfiles>)
- [type Compiler](<#type-compiler>)
  - [func New(ctx context.Context, proj *config.Project, options ...CompilerOption) (*Compiler, error)](<#func-new>)
  - [func (c *Compiler) Build(ctx context.Context, apiName string) error](<#func-compiler-build>)
  - [func (c *Compiler) BuildAll(ctx context.Context) error](<#func-compiler-buildall>)
- [type CompilerOption](<#type-compileroption>)


## func [ResourceSpecFiles](<https://github.com/snyk/vervet/blob/main/internal/compiler/compiler.go#L120>)

```go
func ResourceSpecFiles(rcConfig *config.ResourceSet) ([]string, error)
```

ResourceSpecFiles returns all matching spec files for a config\.Resource\.

## type [Compiler](<https://github.com/snyk/vervet/blob/main/internal/compiler/compiler.go#L22-L24>)

A Compiler checks and builds versioned API resource inputs into aggregated OpenAPI versioned outputs\, as determined by an API project configuration\.

```go
type Compiler struct {
    // contains filtered or unexported fields
}
```

### func [New](<https://github.com/snyk/vervet/blob/main/internal/compiler/compiler.go#L46>)

```go
func New(ctx context.Context, proj *config.Project, options ...CompilerOption) (*Compiler, error)
```

New returns a new Compiler for a given project configuration\.

### func \(\*Compiler\) [Build](<https://github.com/snyk/vervet/blob/main/internal/compiler/compiler.go#L137>)

```go
func (c *Compiler) Build(ctx context.Context, apiName string) error
```

Build builds an aggregate versioned OpenAPI spec for a specific API by name in the project\.

### func \(\*Compiler\) [BuildAll](<https://github.com/snyk/vervet/blob/main/internal/compiler/compiler.go#L288>)

```go
func (c *Compiler) BuildAll(ctx context.Context) error
```

BuildAll builds all APIs in the project\.

## type [CompilerOption](<https://github.com/snyk/vervet/blob/main/internal/compiler/compiler.go#L27>)

CompilerOption applies a configuration option to a Compiler\.

```go
type CompilerOption func(*Compiler) error
```

