package files

import (
	"io/fs"
	"os"
	"path/filepath"

	"github.com/bmatcuk/doublestar/v4"

	"github.com/snyk/vervet/v4"
	"github.com/snyk/vervet/v4/config"
)

// FileSource defines a source of spec files to lint. This abstraction allows
// linters to operate seamlessly over version control systems and local files.
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

// NilSource is a FileSource that does not have any files in it.
type NilSource struct{}

// Name implements FileSource.
func (NilSource) Name() string { return "does not exist" }

// Match implements FileSource.
func (NilSource) Match(*config.ResourceSet) ([]string, error) { return nil, nil }

// Prefetch implements FileSource.
func (NilSource) Prefetch(root string) (string, error) {
	return "", nil
}

// Fetch implements FileSource.
func (NilSource) Fetch(path string) (string, error) {
	return "", nil
}

// Close implements FileSource.
func (NilSource) Close() error { return nil }

// LocalFSSource is a FileSource that resolves files from the local filesystem
// relative to the current working directory.
type LocalFSSource struct{}

// Name implements FileSource.
func (LocalFSSource) Name() string { return "local file" }

// Match implements FileSource.
func (LocalFSSource) Match(rcConfig *config.ResourceSet) ([]string, error) {
	var result []string
	err := doublestar.GlobWalk(os.DirFS(rcConfig.Path),
		vervet.SpecGlobPattern,
		func(path string, d fs.DirEntry) error {
			rcPath := filepath.Join(rcConfig.Path, path)
			for i := range rcConfig.Excludes {
				if ok, err := doublestar.Match(rcConfig.Excludes[i], rcPath); ok {
					return nil
				} else if err != nil {
					return err
				}
			}
			result = append(result, rcPath)
			return nil
		})
	return result, err
}

// Prefetch implements FileSource.
func (LocalFSSource) Prefetch(root string) (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return filepath.Join(cwd, root), nil
}

// Fetch implements FileSource.
func (LocalFSSource) Fetch(path string) (string, error) {
	if _, err := os.Stat(path); err == nil {
		return filepath.Abs(path)
	} else if os.IsNotExist(err) {
		return "", nil
	} else {
		return "", err
	}
}

// Close implements FileSource.
func (LocalFSSource) Close() error { return nil }
