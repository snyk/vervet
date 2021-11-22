package sweatercomb

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/snyk/vervet/config"
)

func TestLinter(t *testing.T) {
	c := qt.New(t)
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	// Sanity check constructor
	l, err := New(ctx, &config.SweaterCombLinter{
		Image:     "some-image",
		Rules:     []string{"/sweater-comb/rules/rule1", "rule2"},
		ExtraArgs: []string{"--some-flag"},
	})
	c.Assert(err, qt.IsNil)
	c.Assert(l.image, qt.Equals, "some-image")
	c.Assert(l.rules, qt.DeepEquals, []string{"/sweater-comb/rules/rule1", "/sweater-comb/target/rule2"})
	c.Assert(l.extraArgs, qt.DeepEquals, []string{"--some-flag"})

	// Verify temp ruleset that joins all the rulesets.
	rulesetFile := filepath.Join(l.rulesDir, "ruleset.yaml")
	rulesetContents, err := ioutil.ReadFile(rulesetFile)
	c.Assert(err, qt.IsNil)
	c.Assert(string(rulesetContents), qt.Equals, `
extends:
- /sweater-comb/rules/rule1
- /sweater-comb/target/rule2
`[1:])

	// Capture stdout so we can test the output substitution that replaces
	// container paths with the current working directory in spectral's output.
	cwd, err := os.Getwd()
	c.Assert(err, qt.IsNil)
	tempDir := c.TempDir()
	tempFile, err := os.Create(tempDir + "/stdout")
	c.Assert(err, qt.IsNil)
	c.Patch(&os.Stdout, tempFile)
	defer tempFile.Close()

	// Verify mock runner ran what we'd expect
	runner := &mockRunner{}
	l.runner = runner
	err = l.Run(ctx, "my-api/**/*.yaml")
	c.Assert(err, qt.IsNil)
	c.Assert(runner.runs, qt.DeepEquals, [][]string{{
		"docker", "run", "--rm",
		"-v", l.rulesDir + ":/vervet",
		"-v", cwd + ":/sweater-comb/target",
		"some-image",
		"lint",
		"-r", "/vervet/ruleset.yaml",
		"--some-flag",
		"my-api/**/*.yaml",
	}})

	// Verify captured output was substituted. Mainly a convenience that makes
	// output host-relevant and cmd-clickable if possible.
	c.Assert(tempFile.Sync(), qt.IsNil)
	capturedOutput, err := ioutil.ReadFile(tempFile.Name())
	c.Assert(err, qt.IsNil)
	c.Assert(string(capturedOutput), qt.Equals, cwd+" is the path to things in your project\n")

	// Command failed.
	runner = &mockRunner{err: fmt.Errorf("nope")}
	l.runner = runner
	err = l.Run(ctx, "my-api/**/*.yaml")
	c.Assert(err, qt.ErrorMatches, "nope")
}

type mockRunner struct {
	runs [][]string
	err  error
}

func (r *mockRunner) run(cmd *exec.Cmd) error {
	fmt.Fprintln(cmd.Stdout, "/sweater-comb/target is the path to things in your project")
	r.runs = append(r.runs, cmd.Args)
	return r.err
}
