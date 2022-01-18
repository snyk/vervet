package linter

import (
	"context"

	"github.com/snyk/vervet/v3/config"
)

// A Linter checks that a set of spec files conform to some set of rules and
// standards.
type Linter interface {
	// Match returns a slice of logical paths to spec files that should be
	// linted from the given resource set configuration.
	Match(*config.ResourceSet) ([]string, error)

	// WithOverride returns a new instance of a Linter with the given configuration.
	WithOverride(ctx context.Context, cfg *config.Linter) (Linter, error)

	// Run executes the linter checks on the given spec files.
	Run(ctx context.Context, root string, files ...string) error
}
