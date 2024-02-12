# scaffold

```go
import "github.com/snyk/vervet/v6/internal/scaffold"
```

## Index

- [Variables](<#variables>)
- [type Manifest](<#type-manifest>)
- [type Option](<#type-option>)
  - [func Force(force bool) Option](<#func-force>)
- [type Scaffold](<#type-scaffold>)
  - [func New(dst, src string, options ...Option) (*Scaffold, error)](<#func-new>)
  - [func (s *Scaffold) Init() error](<#func-scaffold-init>)
  - [func (s *Scaffold) Organize() error](<#func-scaffold-organize>)


## Variables

ErrAlreadyInitialized is used when scaffolding is being run on a project that is already setup\.

```go
var ErrAlreadyInitialized = fmt.Errorf("project files already exist")
```

## type [Manifest](<https://github.com/snyk/vervet/blob/main/internal/scaffold/scaffold.go#L27-L34>)

Manifest defines the scaffold manifest model\.

```go
type Manifest struct {
    Version string

    // Organize contains a mapping of files relative to Scaffold src, to be
    // copied into dst, relative to dst. Missing intermediate directories will
    // be created as needed.
    Organize map[string]string `json:"organize"`
}
```

## type [Option](<https://github.com/snyk/vervet/blob/main/internal/scaffold/scaffold.go#L38>)

Option defines a functional option that modifies a new Scaffold in the constructor\.

```go
type Option func(*Scaffold)
```

### func [Force](<https://github.com/snyk/vervet/blob/main/internal/scaffold/scaffold.go#L42>)

```go
func Force(force bool) Option
```

Force sets the force flag on a Scaffold\, which determines whether existing destination files will be overwritten\. Default is false\.

## type [Scaffold](<https://github.com/snyk/vervet/blob/main/internal/scaffold/scaffold.go#L18-L22>)

Scaffold defines a Vervet API project scaffold\.

```go
type Scaffold struct {
    // contains filtered or unexported fields
}
```

### func [New](<https://github.com/snyk/vervet/blob/main/internal/scaffold/scaffold.go#L51>)

```go
func New(dst, src string, options ...Option) (*Scaffold, error)
```

New returns a new Scaffold loaded from source directory \`src\` for operation on destination directory \`dst\`\. The Scaffold src must contain a \`manifest\.yaml\` which defines how dst will be provisioned\.

### func \(\*Scaffold\) [Init](<https://github.com/snyk/vervet/blob/main/internal/scaffold/scaffold.go#L114>)

```go
func (s *Scaffold) Init() error
```

Init runs a script called \`init\` in the scaffold source if present\, in the destination directory\.

### func \(\*Scaffold\) [Organize](<https://github.com/snyk/vervet/blob/main/internal/scaffold/scaffold.go#L87>)

```go
func (s *Scaffold) Organize() error
```

Organize provisions files from the scaffold source into its destination\.

