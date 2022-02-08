# files

```go
import "github.com/snyk/vervet/v3/internal/files"
```

## Index

- [type FileSource](<#type-filesource>)
- [type LocalFSSource](<#type-localfssource>)
  - [func (LocalFSSource) Close() error](<#func-localfssource-close>)
  - [func (LocalFSSource) Fetch(path string) (string, error)](<#func-localfssource-fetch>)
  - [func (LocalFSSource) Match(rcConfig *config.ResourceSet) ([]string, error)](<#func-localfssource-match>)
  - [func (LocalFSSource) Name() string](<#func-localfssource-name>)
  - [func (LocalFSSource) Prefetch(root string) (string, error)](<#func-localfssource-prefetch>)
- [type NilSource](<#type-nilsource>)
  - [func (NilSource) Close() error](<#func-nilsource-close>)
  - [func (NilSource) Fetch(path string) (string, error)](<#func-nilsource-fetch>)
  - [func (NilSource) Match(*config.ResourceSet) ([]string, error)](<#func-nilsource-match>)
  - [func (NilSource) Name() string](<#func-nilsource-name>)
  - [func (NilSource) Prefetch(root string) (string, error)](<#func-nilsource-prefetch>)


## type FileSource

FileSource defines a source of spec files to lint\. This abstraction allows linters to operate seamlessly over version control systems and local files\.

```go
type FileSource interface {
    // Name returns a string describing the file source.
    Name() string

    // Match returns a slice of logical paths to spec files that should be
    // linted from the given resource set configuration.
    Match(*config.ResourceSet) ([]string, error)

    // Prefetch retrieves an entire directory tree starting at the given root,
    // for remote sources which need to download and cache a local copy. For
    // such sources, a call to Fetch without a pre-fetched root will error.
    // The path to the local copy of the "root" is returned.
    //
    // For local sources, this method may be a no-op / passthrough.
    //
    // The root must contain all relative OpenAPI $ref references in all linted
    // specs, or the lint will fail.
    Prefetch(root string) (string, error)

    // Fetch retrieves the contents of the requested logical path as a local
    // file and returns the absolute path where it may be found. An empty
    // string, rather than an error, is returned if the file does not exist.
    Fetch(path string) (string, error)

    // Close releases any resources consumed in content retrieval. Any files
    // returned by Fetch will no longer be available after calling Close, and
    // any further calls to Fetch will error.
    Close() error
}
```

## type LocalFSSource

LocalFSSource is a FileSource that resolves files from the local filesystem relative to the current working directory\.

```go
type LocalFSSource struct{}
```

### func \(LocalFSSource\) Close

```go
func (LocalFSSource) Close() error
```

Close implements FileSource\.

### func \(LocalFSSource\) Fetch

```go
func (LocalFSSource) Fetch(path string) (string, error)
```

Fetch implements FileSource\.

### func \(LocalFSSource\) Match

```go
func (LocalFSSource) Match(rcConfig *config.ResourceSet) ([]string, error)
```

Match implements FileSource\.

### func \(LocalFSSource\) Name

```go
func (LocalFSSource) Name() string
```

Name implements FileSource\.

### func \(LocalFSSource\) Prefetch

```go
func (LocalFSSource) Prefetch(root string) (string, error)
```

Prefetch implements FileSource\.

## type NilSource

NilSource is a FileSource that does not have any files in it\.

```go
type NilSource struct{}
```

### func \(NilSource\) Close

```go
func (NilSource) Close() error
```

Close implements FileSource\.

### func \(NilSource\) Fetch

```go
func (NilSource) Fetch(path string) (string, error)
```

Fetch implements FileSource\.

### func \(NilSource\) Match

```go
func (NilSource) Match(*config.ResourceSet) ([]string, error)
```

Match implements FileSource\.

### func \(NilSource\) Name

```go
func (NilSource) Name() string
```

Name implements FileSource\.

### func \(NilSource\) Prefetch

```go
func (NilSource) Prefetch(root string) (string, error)
```

Prefetch implements FileSource\.

