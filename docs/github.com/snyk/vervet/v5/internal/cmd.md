# cmd

```go
import "github.com/snyk/vervet/v5/internal/cmd"
```

Package cmd provides subcommands for the vervet CLI\.

## Index

- [Variables](<#variables>)
- [func Build(ctx *cli.Context) error](<#func-build>)
- [func CheckCatalog(ctx *cli.Context) error](<#func-checkcatalog>)
- [func Filter(ctx *cli.Context) error](<#func-filter>)
- [func Generate(ctx *cli.Context) error](<#func-generate>)
- [func Lint(ctx *cli.Context) error](<#func-lint>)
- [func Localize(ctx *cli.Context) error](<#func-localize>)
- [func PreviewCatalog(ctx *cli.Context) error](<#func-previewcatalog>)
- [func Resolve(ctx *cli.Context) error](<#func-resolve>)
- [func ResourceFiles(ctx *cli.Context) error](<#func-resourcefiles>)
- [func ResourceShow(ctx *cli.Context) error](<#func-resourceshow>)
- [func ScaffoldInit(ctx *cli.Context) error](<#func-scaffoldinit>)
- [func UpdateCatalog(ctx *cli.Context) error](<#func-updatecatalog>)
- [type Prompt](<#type-prompt>)
  - [func (p Prompt) Confirm(label string) (bool, error)](<#func-prompt-confirm>)
  - [func (p Prompt) Entry(label string) (string, error)](<#func-prompt-entry>)
  - [func (p Prompt) Select(label string, items []string) (string, error)](<#func-prompt-select>)
- [type VervetApp](<#type-vervetapp>)
  - [func NewApp(app *cli.App, vp VervetParams) *VervetApp](<#func-newapp>)
  - [func (v *VervetApp) Run(args []string) error](<#func-vervetapp-run>)
- [type VervetParams](<#type-vervetparams>)
- [type VervetPrompt](<#type-vervetprompt>)


## Variables

BackstageCommand is the \`vervet backstage\` subcommand\.

```go
var BackstageCommand = cli.Command{
    Name: "backstage",
    Subcommands: []*cli.Command{{
        Name:  "update-catalog",
        Usage: "Update Backstage catalog-info.yaml with Vervet API versions",
        Flags: []cli.Flag{
            &cli.StringFlag{
                Name:    "config",
                Aliases: []string{"c", "conf"},
                Usage:   "Project configuration file",
            },
        },
        Action: UpdateCatalog,
    }, {
        Name:  "preview-catalog",
        Usage: "Preview changes to Backstage catalog-info.yaml",
        Flags: []cli.Flag{
            &cli.StringFlag{
                Name:    "config",
                Aliases: []string{"c", "conf"},
                Usage:   "Project configuration file",
            },
        },
        Action: PreviewCatalog,
    }, {
        Name:  "check-catalog",
        Usage: "Check for uncommitted changes in Backstage catalog-info.yaml",
        Flags: []cli.Flag{
            &cli.StringFlag{
                Name:    "config",
                Aliases: []string{"c", "conf"},
                Usage:   "Project configuration file",
            },
        },
        Action: CheckCatalog,
    }},
}
```

BuildCommand is the \`vervet build\` subcommand\.

```go
var BuildCommand = cli.Command{
    Name:      "build",
    Usage:     "Build versioned resources into versioned OpenAPI specs",
    ArgsUsage: "[input resources root] [output api root]",
    Flags: []cli.Flag{
        &cli.StringFlag{
            Name:    "config",
            Aliases: []string{"c", "conf"},
            Usage:   "Project configuration file",
        },
        &cli.BoolFlag{
            Name:  "lint",
            Usage: "Enable linting during build",
            Value: true,
        },
        &cli.StringFlag{
            Name:    "include",
            Aliases: []string{"I"},
            Usage:   "OpenAPI specification to include in build output",
        },
    },
    Action: Build,
}
```

```go
var CLIApp = cli.App{
    Name:    "vervet",
    Usage:   "OpenAPI resource versioning tool",
    Version: "develop",
    Flags: []cli.Flag{
        &cli.BoolFlag{
            Name:  "debug",
            Usage: "Turn on debug logging",
        },
    },
    Commands: []*cli.Command{
        &BackstageCommand,
        &BuildCommand,
        &FilterCommand,
        &GenerateCommand,

        &LocalizeCommand,
        &ResourceCommand,
        &ResolveCommand,
    },
}
```

FilterCommand is the \`vervet filter\` subcommand

```go
var FilterCommand = cli.Command{
    Name:      "filter",
    Usage:     "Filter an OpenAPI document",
    ArgsUsage: "[spec.yaml file]",
    Flags: []cli.Flag{
        &cli.StringSliceFlag{Name: "include-paths", Aliases: []string{"I"}},
        &cli.StringSliceFlag{Name: "exclude-paths", Aliases: []string{"X"}},
    },
    Action: Filter,
}
```

GenerateCommand is the \`vervet generate\` subcommand\.

```go
var GenerateCommand = cli.Command{
    Name:      "generate",
    Usage:     "Generate artifacts from resource versioned OpenAPI specs",
    ArgsUsage: "<generator> [<generator2>...]",
    Flags: []cli.Flag{
        &cli.StringFlag{
            Name:    "config",
            Aliases: []string{"c", "conf"},
            Usage:   "Project configuration file",
        },
        &cli.BoolFlag{
            Name:    "dry-run",
            Aliases: []string{"n"},
            Usage:   "Dry-run, listing files that would be generated",
        },
        &cli.StringFlag{
            Name:    "generators",
            Aliases: []string{"g", "gen", "generator"},
            Usage:   "Generators definition file",
        },
    },
    Action: Generate,
}
```

LintCommand is the \`vervet lint\` subcommand\.

```go
var LintCommand = cli.Command{
    Name:      "lint",
    Usage:     "Lint  versioned resources",
    ArgsUsage: "[input resources root] [output api root]",
    Flags: []cli.Flag{
        &cli.StringFlag{
            Name:    "config",
            Aliases: []string{"c", "conf"},
            Usage:   "Project configuration file",
        },
    },
    Action: Lint,
}
```

LocalizeCommand is the \`vervet localize\` subcommand

```go
var LocalizeCommand = cli.Command{
    Name:      "localize",
    Aliases:   []string{"localise"},
    Usage:     "Localize references and validate a single OpenAPI spec file",
    ArgsUsage: "[spec.yaml file]",
    Action:    Localize,
}
```

ResolveCommand is the \`vervet resolve\` subcommand\.

```go
var ResolveCommand = cli.Command{
    Name:      "resolve",
    Usage:     "Aggregate, render and validate resource specs at a particular version",
    ArgsUsage: "[resource root]",
    Flags: []cli.Flag{
        &cli.StringFlag{Name: "at"},
    },
    Action: Resolve,
}
```

ResourceCommand is the \`vervet resource\` subcommand\.

```go
var ResourceCommand = cli.Command{
    Name:    "resource",
    Aliases: []string{"rc"},
    Flags: []cli.Flag{
        &cli.StringFlag{
            Name:    "config",
            Aliases: []string{"c", "conf"},
            Usage:   "Project configuration file",
        },
    },
    Subcommands: []*cli.Command{{
        Name:      "files",
        Usage:     "List OpenAPI files of versioned resources in a vervet project",
        ArgsUsage: "[api [resource]]",
        Action:    ResourceFiles,
    }, {
        Name:      "info",
        Usage:     "Information about versioned resources in a vervet project",
        ArgsUsage: "[api [resource]]",
        Action:    ResourceShow,
    }},
}
```

Scaffold is the \`vervet scaffold\` subcommand\.

```go
var Scaffold = cli.Command{
    Name: "scaffold",
    Subcommands: []*cli.Command{{
        Name:      "init",
        Usage:     "Initialize a new project from a scaffold",
        ArgsUsage: "[path to scaffold directory]",
        Flags: []cli.Flag{
            &cli.BoolFlag{
                Name:    "force",
                Aliases: []string{"f", "overwrite"},
                Usage:   "Overwrite existing files",
            },
        },
        Action: ScaffoldInit,
    }},
}
```

Vervet is the vervet application with the CLI application\.

```go
var Vervet = NewApp(&CLIApp, VervetParams{
    Stdin:  os.Stdin,
    Stdout: os.Stdout,
    Stderr: os.Stderr,
    Prompt: Prompt{},
})
```

## func [Build](<https://github.com/snyk/vervet/blob/main/internal/cmd/compiler.go#L39>)

```go
func Build(ctx *cli.Context) error
```

Build compiles versioned resources into versioned API specs\.

## func [CheckCatalog](<https://github.com/snyk/vervet/blob/main/internal/cmd/backstage.go#L69>)

```go
func CheckCatalog(ctx *cli.Context) error
```

CheckCatalog checks whether the catalog\-info\.yaml or tracked compiled versions it references have uncommitted changes\. This is primarily useful in CI checks to make sure everything is checked into git for Backstage\.

## func [Filter](<https://github.com/snyk/vervet/blob/main/internal/cmd/filter.go#L25>)

```go
func Filter(ctx *cli.Context) error
```

Filter an OpenAPI spec file\.

## func [Generate](<https://github.com/snyk/vervet/blob/main/internal/cmd/generate.go#L37>)

```go
func Generate(ctx *cli.Context) error
```

Generate executes code generators against OpenAPI specs\.

## func [Lint](<https://github.com/snyk/vervet/blob/main/internal/cmd/compiler.go#L63>)

```go
func Lint(ctx *cli.Context) error
```

Lint checks versioned resources against linting rules\.

## func [Localize](<https://github.com/snyk/vervet/blob/main/internal/cmd/localize.go#L21>)

```go
func Localize(ctx *cli.Context) error
```

Localize references and validate a single OpenAPI spec file

## func [PreviewCatalog](<https://github.com/snyk/vervet/blob/main/internal/cmd/backstage.go#L62>)

```go
func PreviewCatalog(ctx *cli.Context) error
```

PreviewCatalog updates the catalog\-info\.yaml from Vervet versions\.

## func [Resolve](<https://github.com/snyk/vervet/blob/main/internal/cmd/resolve.go#L24>)

```go
func Resolve(ctx *cli.Context) error
```

Resolve aggregates\, renders and validates resource specs at a particular version\.

## func [ResourceFiles](<https://github.com/snyk/vervet/blob/main/internal/cmd/resource.go#L118>)

```go
func ResourceFiles(ctx *cli.Context) error
```

ResourceFiles is a command that lists all versioned OpenAPI spec files of matching resources\. It takes optional arguments to filter the output: api resource

## func [ResourceShow](<https://github.com/snyk/vervet/blob/main/internal/cmd/resource.go#L43>)

```go
func ResourceShow(ctx *cli.Context) error
```

ResourceShow is a command that lists all the versions of matching resources\. It takes optional arguments to filter the output: api resource

## func [ScaffoldInit](<https://github.com/snyk/vervet/blob/main/internal/cmd/scaffold.go#L32>)

```go
func ScaffoldInit(ctx *cli.Context) error
```

ScaffoldInit creates a new project configuration from a provided scaffold directory\.

## func [UpdateCatalog](<https://github.com/snyk/vervet/blob/main/internal/cmd/backstage.go#L57>)

```go
func UpdateCatalog(ctx *cli.Context) error
```

UpdateCatalog updates the catalog\-info\.yaml from Vervet versions\.

## type [Prompt](<https://github.com/snyk/vervet/blob/main/internal/cmd/cmd.go#L90>)

Prompt is the default interactive prompt for vervet\.

```go
type Prompt struct{}
```

### func \(Prompt\) [Confirm](<https://github.com/snyk/vervet/blob/main/internal/cmd/cmd.go#L93>)

```go
func (p Prompt) Confirm(label string) (bool, error)
```

Confirm implements VervetPrompt\.Confirm

### func \(Prompt\) [Entry](<https://github.com/snyk/vervet/blob/main/internal/cmd/cmd.go#L113>)

```go
func (p Prompt) Entry(label string) (string, error)
```

Entry implements VervetPrompt\.Entry

### func \(Prompt\) [Select](<https://github.com/snyk/vervet/blob/main/internal/cmd/cmd.go#L131>)

```go
func (p Prompt) Select(label string, items []string) (string, error)
```

Select implements VervetPrompt\.Select

## type [VervetApp](<https://github.com/snyk/vervet/blob/main/internal/cmd/cmd.go#L28-L31>)

VervetApp contains the cli Application\.

```go
type VervetApp struct {
    App    *cli.App
    Params VervetParams
}
```

### func [NewApp](<https://github.com/snyk/vervet/blob/main/internal/cmd/cmd.go#L55>)

```go
func NewApp(app *cli.App, vp VervetParams) *VervetApp
```

NewApp returns a new VervetApp with the provided params\.

### func \(\*VervetApp\) [Run](<https://github.com/snyk/vervet/blob/main/internal/cmd/cmd.go#L49>)

```go
func (v *VervetApp) Run(args []string) error
```

Run runs the cli\.App with the Vervet config params\.

## type [VervetParams](<https://github.com/snyk/vervet/blob/main/internal/cmd/cmd.go#L20-L25>)

VervetParams contains configuration parameters for the Vervet CLI application\.

```go
type VervetParams struct {
    Stdin  io.ReadCloser
    Stdout io.WriteCloser
    Stderr io.WriteCloser
    Prompt VervetPrompt
}
```

## type [VervetPrompt](<https://github.com/snyk/vervet/blob/main/internal/cmd/cmd.go#L34-L38>)

VervetPrompt defines the interface for interactive prompts in vervet\.

```go
type VervetPrompt interface {
    Confirm(label string) (bool, error)                  // Confirm y/n an action
    Entry(label string) (string, error)                  // Gather a freeform entry in response to a question
    Select(label string, items []string) (string, error) // Select from a limited number of entries
}
```

