package linter

import (
	"context"

	"github.com/snyk/vervet/config"
)

// A Linter checks that a set of files conform to some set of rules and
// standards.
type Linter interface {
	WithOverride(ctx context.Context, cfg *config.Linter) (Linter, error)
	Run(ctx context.Context, files ...string) error
}
