# generator

```go
import "github.com/snyk/vervet/v3/internal/generator"
```

## Index

- [func NewMap(proj *config.Project, options ...Option) (map[string]*Generator, error)](<#func-newmap>)
- [type Generator](<#type-generator>)
  - [func New(conf *config.Generator, options ...Option) (*Generator, error)](<#func-new>)
  - [func (g *Generator) Run(scope *VersionScope) error](<#func-generator-run>)
- [type Option](<#type-option>)
  - [func Debug(debug bool) Option](<#func-debug>)
  - [func Force(force bool) Option](<#func-force>)
- [type VersionScope](<#type-versionscope>)


## func NewMap

```go
func NewMap(proj *config.Project, options ...Option) (map[string]*Generator, error)
```

NewMap instanstiates a map of all Generators defined in a Project\.

## type Generator

Generator generates files for new resources from data models and templates\.

```go
type Generator struct {
    // contains filtered or unexported fields
}
```

### func New

```go
func New(conf *config.Generator, options ...Option) (*Generator, error)
```

New returns a new Generator from config\.

### func \(\*Generator\) Run

```go
func (g *Generator) Run(scope *VersionScope) error
```

Run executes the Generator\. If generated artifacts already exist\, a warning is logged but the file is not overwritten\, unless force is true\.

## type Option

Option configures a Generator\.

```go
type Option func(g *Generator)
```

### func Debug

```go
func Debug(debug bool) Option
```

Debug turns on template debug logging\.

### func Force

```go
func Force(force bool) Option
```

Force configures the Generator to overwrite generated artifacts\.

## type VersionScope

VersionScope identifies a distinct resource version that the generator is building for\.

```go
type VersionScope struct {
    API       string
    Resource  string
    Version   string
    Stability string
}
```

