package optic

import (
	"os"
	"path/filepath"
)

// fileSource defines a source of files.
type fileSource interface {
	// Fetch retrieves the contents of the requested logical path as a local
	// file and returns the absolute path where it may be found. An empty
	// string, rather than an error, is returned if the file does not exist.
	Fetch(path string) (string, error)

	// Close releases any resources consumed in content retrieval. Any files
	// returned by Fetch will no longer be available after calling Close, and
	// any further calls to Fetch will error.
	Close() error
}

// nilSource is a fileSource that does not have any files in it.
type nilSource struct{}

// Fetch implements fileSource.
func (nilSource) Fetch(path string) (string, error) {
	return "", nil
}

// Close implements fileSource.
func (nilSource) Close() error { return nil }

// workingCopySource is a fileSource that resolves files relative to the
// current working directory.
type workingCopySource struct{}

// Fetch implements fileSource.
func (s workingCopySource) Fetch(path string) (string, error) {
	if _, err := os.Stat(path); err == nil {
		return filepath.Abs(path)
	} else if os.IsNotExist(err) {
		return "", nil
	} else {
		return "", err
	}
}

// Close implements fileSource.
func (workingCopySource) Close() error { return nil }
