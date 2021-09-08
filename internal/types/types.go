package types

import "context"

// A Linter checks that a set of files conform to some set of rules and
// standards.
type Linter interface {
	NewRules(ctx context.Context, files ...string) (Linter, error)
	Run(ctx context.Context, files ...string) error
}
