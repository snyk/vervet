# linter

```go
import "github.com/snyk/vervet/v5/internal/linter"
```

## Index

- [type Linter](<#type-linter>)


## type [Linter](<https://github.com/snyk/vervet/blob/main/internal/linter/linter.go#L11-L21>)

A Linter checks that a set of spec files conform to some set of rules and standards\.

```go
type Linter interface {
    // Match returns a slice of logical paths to spec files that should be
    // linted from the given resource set configuration.
    Match(*config.ResourceSet) ([]string, error)

    // WithOverride returns a new instance of a Linter with the given configuration.
    WithOverride(ctx context.Context, cfg *config.Linter) (Linter, error)

    // Run executes the linter checks on the given spec files.
    Run(ctx context.Context, root string, files ...string) error
}
```

