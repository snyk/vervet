# cmd

```go
import "github.com/snyk/vervet/v3/cmd"
```

Package cmd provides subcommands for the vervet CLI\.

## Index

- [Variables](<#variables>)
- [func Compile(ctx *cli.Context) error](<#func-compile>)
- [func Lint(ctx *cli.Context) error](<#func-lint>)
- [func Localize(ctx *cli.Context) error](<#func-localize>)
- [func Resolve(ctx *cli.Context) error](<#func-resolve>)
- [func ScaffoldInit(ctx *cli.Context) error](<#func-scaffoldinit>)
- [func VersionFiles(ctx *cli.Context) error](<#func-versionfiles>)
- [func VersionList(ctx *cli.Context) error](<#func-versionlist>)
- [func VersionNew(ctx *cli.Context) error](<#func-versionnew>)
- [type Prompt](<#type-prompt>)
  - [func (p Prompt) Confirm(label string) (bool, error)](<#func-prompt-confirm>)
  - [func (p Prompt) Entry(label string) (string, error)](<#func-prompt-entry>)
  - [func (p Prompt) Select(label string, items []string) (string, error)](<#func-prompt-select>)
- [type VervetApp](<#type-vervetapp>)
  - [func NewApp(vp VervetParams) *VervetApp](<#func-newapp>)
  - [func (v *VervetApp) Run(args []string) error](<#func-vervetapp-run>)
- [type VervetParams](<#type-vervetparams>)
- [type VervetPrompt](<#type-vervetprompt>)


## Variables

Vervet is the vervet application with the CLI application\.

```go
var Vervet = NewApp(VervetParams{
    Stdin:  os.Stdin,
    Stdout: os.Stdout,
    Stderr: os.Stderr,
    Prompt: Prompt{},
})
```

## func Compile

```go
func Compile(ctx *cli.Context) error
```

Compile compiles versioned resources into versioned API specs\.

## func Lint

```go
func Lint(ctx *cli.Context) error
```

Lint checks versioned resources against linting rules\.

## func Localize

```go
func Localize(ctx *cli.Context) error
```

Localize references and validate a single OpenAPI spec file

## func Resolve

```go
func Resolve(ctx *cli.Context) error
```

Resolve aggregates\, renders and validates resource specs at a particular version\.

## func ScaffoldInit

```go
func ScaffoldInit(ctx *cli.Context) error
```

ScaffoldInit creates a new project configuration from a provided scaffold directory\.

## func VersionFiles

```go
func VersionFiles(ctx *cli.Context) error
```

VersionFiles is a command that lists all versioned OpenAPI spec files of matching resources\. It takes optional arguments to filter the output: api resource

## func VersionList

```go
func VersionList(ctx *cli.Context) error
```

VersionList is a command that lists all the versions of matching resources\. It takes optional arguments to filter the output: api resource

## func VersionNew

```go
func VersionNew(ctx *cli.Context) error
```

VersionNew generates a new resource\.

## type Prompt

Prompt is the default interactive prompt for vervet\.

```go
type Prompt struct{}
```

### func \(Prompt\) Confirm

```go
func (p Prompt) Confirm(label string) (bool, error)
```

Confirm implements VervetPrompt\.Confirm

### func \(Prompt\) Entry

```go
func (p Prompt) Entry(label string) (string, error)
```

Entry implements VervetPrompt\.Entry

### func \(Prompt\) Select

```go
func (p Prompt) Select(label string, items []string) (string, error)
```

Select implements VervetPrompt\.Select

## type VervetApp

VervetApp contains the cli Application\.

```go
type VervetApp struct {
    App    *cli.App
    Params VervetParams
}
```

### func NewApp

```go
func NewApp(vp VervetParams) *VervetApp
```

NewApp returns a new VervetApp with the provided params\.

### func \(\*VervetApp\) Run

```go
func (v *VervetApp) Run(args []string) error
```

Run runs the cli\.App with the Vervet config params\.

## type VervetParams

VervetParams contains configuration parameters for the Vervet CLI application\.

```go
type VervetParams struct {
    Stdin  io.ReadCloser
    Stdout io.WriteCloser
    Stderr io.WriteCloser
    Prompt VervetPrompt
}
```

## type VervetPrompt

VervetPrompt defines the interface for interactive prompts in vervet\.

```go
type VervetPrompt interface {
    Confirm(label string) (bool, error)                  // Confirm y/n an action
    Entry(label string) (string, error)                  // Gather a freeform entry in response to a question
    Select(label string, items []string) (string, error) // Select from a limited number of entries
}
```

