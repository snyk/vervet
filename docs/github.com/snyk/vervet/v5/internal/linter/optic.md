# optic

```go
import "github.com/snyk/vervet/v5/internal/linter/optic"
```

Package optic supports linting OpenAPI specs with Optic CI and Sweater Comb\.

## Index

- [type Context](<#type-context>)
- [type Optic](<#type-optic>)
  - [func New(ctx context.Context, cfg *config.OpticCILinter) (*Optic, error)](<#func-new>)
  - [func (o *Optic) Match(rcConfig *config.ResourceSet) ([]string, error)](<#func-optic-match>)
  - [func (o *Optic) Run(ctx context.Context, root string, paths ...string) error](<#func-optic-run>)
  - [func (*Optic) WithOverride(ctx context.Context, override *config.Linter) (linter.Linter, error)](<#func-optic-withoverride>)
- [type Release](<#type-release>)
- [type ResourceVersionReleases](<#type-resourceversionreleases>)
- [type StabilityReleases](<#type-stabilityreleases>)
- [type Version](<#type-version>)
- [type VersionStabilityReleases](<#type-versionstabilityreleases>)


## type [Context](<https://github.com/snyk/vervet/blob/main/internal/linter/optic/context.go#L7-L19>)

Context provides Optic with external information needed in order to process API versioning lifecycle rules\. For example\, lifecycle rules need to know when a change is occurring\, and what other versions have deprecated the OpenAPI spec version being evaluated\.

```go
type Context struct {
    // ChangeDate is when the proposed change would occur.
    ChangeDate string `json:"changeDate"`

    // ChangeResource is the proposed change resource name.
    ChangeResource string `json:"changeResource"`

    // ChangeVersion is the proposed change version.
    ChangeVersion Version `json:"changeVersion"`

    // ResourceVersions describes other resource version releases.
    ResourceVersions ResourceVersionReleases `json:"resourceVersions,omitempty"`
}
```

## type [Optic](<https://github.com/snyk/vervet/blob/main/internal/linter/optic/linter.go#L32-L42>)

Optic runs a Docker image containing Optic CI and built\-in rules\.

```go
type Optic struct {
    // contains filtered or unexported fields
}
```

### func [New](<https://github.com/snyk/vervet/blob/main/internal/linter/optic/linter.go#L67>)

```go
func New(ctx context.Context, cfg *config.OpticCILinter) (*Optic, error)
```

New returns a new Optic instance configured to run the given OCI image and file sources\. File sources may be a Git "treeish" \(commit hash or anything that resolves to one such as a branch or tag\) where the current working directory is a cloned git repository\. If \`from\` is empty string\, comparison assumes all changes are new "from scratch" additions\. If \`to\` is empty string\, spec files are assumed to be relative to the current working directory\.

Temporary resources may be created by the linter\, which are reclaimed when the context cancels\.

### func \(\*Optic\) [Match](<https://github.com/snyk/vervet/blob/main/internal/linter/optic/linter.go#L117>)

```go
func (o *Optic) Match(rcConfig *config.ResourceSet) ([]string, error)
```

Match implements linter\.Linter\.

### func \(\*Optic\) [Run](<https://github.com/snyk/vervet/blob/main/internal/linter/optic/linter.go#L153>)

```go
func (o *Optic) Run(ctx context.Context, root string, paths ...string) error
```

Run runs Optic CI on the given paths\. Linting output is written to standard output by Optic CI\. Returns an error when lint fails configured rules\.

### func \(\*Optic\) [WithOverride](<https://github.com/snyk/vervet/blob/main/internal/linter/optic/linter.go#L144>)

```go
func (*Optic) WithOverride(ctx context.Context, override *config.Linter) (linter.Linter, error)
```

WithOverride implements linter\.Linter\.

## type [Release](<https://github.com/snyk/vervet/blob/main/internal/linter/optic/context.go#L38-L42>)

Release describes a single resource\-version\-stability release\.

```go
type Release struct {
    // DeprecatedBy indicates the other release version that deprecates this
    // release.
    DeprecatedBy Version `json:"deprecatedBy"`
}
```

## type [ResourceVersionReleases](<https://github.com/snyk/vervet/blob/main/internal/linter/optic/context.go#L29>)

ResourceVersionReleases describes resource version releases\.

```go
type ResourceVersionReleases map[string]VersionStabilityReleases
```

## type [StabilityReleases](<https://github.com/snyk/vervet/blob/main/internal/linter/optic/context.go#L35>)

StabilityReleases describes stability releases\.

```go
type StabilityReleases map[string]Release
```

## type [Version](<https://github.com/snyk/vervet/blob/main/internal/linter/optic/context.go#L23-L26>)

Version describes an API resource version\, a date and a stability\. Stability is assumed to be GA if not specified\.

```go
type Version struct {
    Date      string `json:"date"`
    Stability string `json:"stability,omitempty"`
}
```

## type [VersionStabilityReleases](<https://github.com/snyk/vervet/blob/main/internal/linter/optic/context.go#L32>)

VersionStabilityReleases describes version releases\.

```go
type VersionStabilityReleases map[string]StabilityReleases
```

