# spectral

```go
import "github.com/snyk/vervet/v3/internal/linter/spectral"
```

## Index

- [type Spectral](<#type-spectral>)
  - [func New(ctx context.Context, cfg *config.SpectralLinter) (*Spectral, error)](<#func-new>)
  - [func (s *Spectral) Match(rcConfig *config.ResourceSet) ([]string, error)](<#func-spectral-match>)
  - [func (l *Spectral) Run(ctx context.Context, _ string, paths ...string) error](<#func-spectral-run>)
  - [func (s *Spectral) WithOverride(ctx context.Context, override *config.Linter) (linter.Linter, error)](<#func-spectral-withoverride>)


## type Spectral

Spectral runs spectral on collections of files with a set of rules\.

```go
type Spectral struct {
    // contains filtered or unexported fields
}
```

### func New

```go
func New(ctx context.Context, cfg *config.SpectralLinter) (*Spectral, error)
```

New returns a new Spectral instance\.

### func \(\*Spectral\) Match

```go
func (s *Spectral) Match(rcConfig *config.ResourceSet) ([]string, error)
```

Match implements linter\.Linter\.

### func \(\*Spectral\) Run

```go
func (l *Spectral) Run(ctx context.Context, _ string, paths ...string) error
```

Run runs spectral on the given paths\. Linting output is written to standard output by spectral\. Returns an error when lint fails configured rules\.

### func \(\*Spectral\) WithOverride

```go
func (s *Spectral) WithOverride(ctx context.Context, override *config.Linter) (linter.Linter, error)
```

WithOverride implements linter\.Linter\.
