// Package testdata provides utility functions for locating files used in
// tests.
package testdata

import (
	"fmt"
	"path/filepath"
	"runtime"
)

// Path returns the absolute path, given a path relative to the testdata
// package directory. This function is intended for use in vervet tests.
func Path(path string) string {
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		panic(fmt.Errorf("cannot locate caller"))
	}
	result, err := filepath.Abs(filepath.Dir(thisFile) + "/" + path)
	if err != nil {
		panic(err)
	}
	return result
}
