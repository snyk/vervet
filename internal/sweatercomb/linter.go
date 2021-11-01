package sweatercomb

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"time"

	"github.com/ghodss/yaml"

	"github.com/snyk/vervet/internal/types"
)

// SweaterComb runs a Docker image containing Spectral and some built-in rules,
// along with additional user-specified rules.
type SweaterComb struct {
	image     string
	rules     []string
	extraArgs []string

	rulesDir string

	runner commandRunner
}

type commandRunner interface {
	run(cmd *exec.Cmd) error
}

type execCommandRunner struct{}

func (*execCommandRunner) run(cmd *exec.Cmd) error {
	return cmd.Run()
}

// New returns a new SweaterComb instance configured with the given rules.
func New(ctx context.Context, image string, rules []string, extraArgs []string) (*SweaterComb, error) {
	if len(rules) == 0 {
		return nil, fmt.Errorf("missing spectral rules")
	}

	rulesDir, err := ioutil.TempDir("", "*-scrules")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp rules directory: %w", err)
	}
	rulesFile, err := os.Create(filepath.Join(rulesDir, "ruleset.yaml"))
	if err != nil {
		return nil, fmt.Errorf("failed to create temp rules file: %w", err)
	}
	defer rulesFile.Close()
	resolvedRules := make([]string, len(rules))
	for i := range rules {
		rule := filepath.Clean(rules[i])
		if !filepath.IsAbs(rule) {
			rule = "/sweater-comb/target/" + rule
		}
		resolvedRules[i] = rule
		if err != nil {
			return nil, err
		}
	}
	rulesDoc := map[string]interface{}{
		"extends": resolvedRules,
	}
	rulesBuf, err := yaml.Marshal(&rulesDoc)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal temp rules file: %w", err)
	}
	_, err = rulesFile.Write(rulesBuf)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal temp rules file: %w", err)
	}
	go func() {
		<-ctx.Done()
		os.RemoveAll(rulesDir)
	}()
	return &SweaterComb{
		image:     image,
		rules:     resolvedRules,
		rulesDir:  rulesDir,
		extraArgs: extraArgs,
		runner:    &execCommandRunner{},
	}, nil
}

// NewRules returns a new Linter instance with additional rules appended.
func (l *SweaterComb) NewRules(ctx context.Context, rules ...string) (types.Linter, error) {
	return New(ctx, l.image, append(l.rules, rules...), l.extraArgs)
}

var sweaterCombOutputRE = regexp.MustCompile(`/sweater-comb/target`)

// Run runs spectral on the given paths. Linting output is written to standard
// output by spectral. Returns an error when lint fails configured rules.
func (l *SweaterComb) Run(ctx context.Context, paths ...string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	mountedPaths := make([]string, len(paths))
	for i := range paths {
		mountedPaths[i] = filepath.Join("./", paths[i])
	}
	cmdline := append(append([]string{
		"run", "--rm",
		"-v", l.rulesDir + ":/vervet", "-v", cwd + ":/sweater-comb/target",
		l.image,
		"lint",
		"-r", "/vervet/ruleset.yaml",
	}, l.extraArgs...), paths...)
	cmd := exec.CommandContext(ctx, "docker", cmdline...)

	pipeReader, pipeWriter := io.Pipe()
	ch := make(chan struct{})
	defer func() {
		err := pipeWriter.Close()
		if err != nil {
			log.Printf("warning: failed to close output: %v", err)
		}
		select {
		case <-ch:
		case <-ctx.Done():
		case <-time.After(cmdTimeout):
			log.Printf("warning: timeout waiting for output to flush")
		}
	}()
	go func() {
		defer pipeReader.Close()
		sc := bufio.NewScanner(pipeReader)
		for sc.Scan() {
			fmt.Println(sweaterCombOutputRE.ReplaceAllLiteralString(sc.Text(), cwd))
		}
		if err := sc.Err(); err != nil {
			fmt.Fprintf(os.Stderr, "error reading stdout: %v", err)
		}
		close(ch)
	}()
	cmd.Stdin = os.Stdin
	cmd.Stdout = pipeWriter
	cmd.Stderr = os.Stderr
	return l.runner.run(cmd)
}

const cmdTimeout = time.Second * 10
