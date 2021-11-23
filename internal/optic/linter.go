// Package optic supports linting OpenAPI specs with Optic CI and Sweater Comb.
package optic

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/snyk/vervet/config"
	"github.com/snyk/vervet/internal/types"
)

// Optic runs a Docker image containing Optic CI and built-in rules.
type Optic struct {
	image      string
	fromSource fileSource
	toSource   fileSource
	runner     commandRunner
}

type commandRunner interface {
	run(cmd *exec.Cmd) error
}

type execCommandRunner struct{}

func (*execCommandRunner) run(cmd *exec.Cmd) error {
	return cmd.Run()
}

// New returns a new Optic instance configured to run the given OCI image and
// file sources. File sources may be a Git "treeish" (commit hash or anything
// that resolves to one such as a branch or tag) where the current working
// directory is a cloned git repository. If `from` is empty string, comparison
// assumes all changes are new "from scratch" additions. If `to` is empty
// string, spec files are assumed to be relative to the current working
// directory.
//
// Temporary resources may be created by the linter, which are reclaimed when
// the context cancels.
func New(ctx context.Context, cfg *config.OpticCILinter) (*Optic, error) {
	image, from, to := cfg.Image, cfg.Original, cfg.Proposed
	var fromSource, toSource fileSource
	var err error
	if from == "" {
		fromSource = nilSource{}
	} else {
		fromSource, err = newGitRepoSource(".", from)
		if err != nil {
			return nil, err
		}
	}
	if to == "" {
		toSource = workingCopySource{}
	} else {
		toSource, err = newGitRepoSource(".", to)
		if err != nil {
			return nil, err
		}
	}
	go func() {
		<-ctx.Done()
		fromSource.Close()
		toSource.Close()
	}()
	return &Optic{
		image:      image,
		fromSource: fromSource,
		toSource:   toSource,
		runner:     &execCommandRunner{},
	}, nil
}

// WithOverride implements types.Linter.
func (l *Optic) WithOverride(ctx context.Context, override *config.Linter) (types.Linter, error) {
	if override.OpticCI == nil {
		return nil, fmt.Errorf("invalid linter override")
	}
	return New(ctx, override.OpticCI)
}

// Run runs Optic CI on the given paths. Linting output is written to standard
// output by Optic CI. Returns an error when lint fails configured rules.
func (o *Optic) Run(ctx context.Context, paths ...string) error {
	for i := range paths {
		err := o.runCompare(ctx, paths[i])
		if err != nil {
			return err
		}
	}
	return nil
}

func (o *Optic) runCompare(ctx context.Context, path string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	var args []string
	fromFile, err := o.fromSource.Fetch(path)
	if err != nil {
		return err
	}
	if fromFile != "" {
		args = append(args, "--from", fromFile)
	}
	toFile, err := o.toSource.Fetch(path)
	if err != nil {
		return err
	}
	if toFile != "" {
		args = append(args, "--to", toFile)
	}
	// TODO: provide context JSON object in --context
	// TODO: link to command line arguments for optic-ci when available.
	cmdline := append([]string{
		"run", "--rm",
		"-v", cwd + ":/target",
		o.image,
		"compare",
	}, args...)
	cmd := exec.CommandContext(ctx, "docker", cmdline...)
	return o.runner.run(cmd)
}
