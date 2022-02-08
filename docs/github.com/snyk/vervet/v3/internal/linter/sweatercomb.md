# sweatercomb

```go
import "github.com/snyk/vervet/v3/internal/linter/sweatercomb"
```

## Index

- [type SweaterComb](<#type-sweatercomb>)
  - [func New(ctx context.Context, cfg *config.SweaterCombLinter) (*SweaterComb, error)](<#func-new>)
  - [func (s *SweaterComb) Match(rcConfig *config.ResourceSet) ([]string, error)](<#func-sweatercomb-match>)
  - [func (s *SweaterComb) Run(ctx context.Context, _ string, paths ...string) error](<#func-sweatercomb-run>)
  - [func (s *SweaterComb) WithOverride(ctx context.Context, override *config.Linter) (linter.Linter, error)](<#func-sweatercomb-withoverride>)


## type SweaterComb

SweaterComb runs a Docker image containing Spectral and some built\-in rules\, along with additional user\-specified rules\.

```go
type SweaterComb struct {
    // contains filtered or unexported fields
}
```

### func New

```go
func New(ctx context.Context, cfg *config.SweaterCombLinter) (*SweaterComb, error)
```

New returns a new SweaterComb instance configured with the given rules\.

### func \(\*SweaterComb\) Match

```go
func (s *SweaterComb) Match(rcConfig *config.ResourceSet) ([]string, error)
```

Match implements linter\.Linter\.

### func \(\*SweaterComb\) Run

```go
func (s *SweaterComb) Run(ctx context.Context, _ string, paths ...string) error
```

Run runs spectral on the given paths\. Linting output is written to standard output by spectral\. Returns an error when lint fails configured rules\.

### func \(\*SweaterComb\) WithOverride

```go
func (s *SweaterComb) WithOverride(ctx context.Context, override *config.Linter) (linter.Linter, error)
```

WithOverride implements linter\.Linter\.

