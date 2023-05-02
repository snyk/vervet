package spectral

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/ghodss/yaml"

	"github.com/snyk/vervet/v5/config"
	"github.com/snyk/vervet/v5/internal/files"
	"github.com/snyk/vervet/v5/internal/linter"
)

// Spectral runs spectral on collections of files with a set of rules.
type Spectral struct {
	rules     []string
	extraArgs []string

	spectralPath string
	rulesPath    string
}

// New returns a new Spectral instance.
func New(ctx context.Context, cfg *config.SpectralLinter) (*Spectral, error) {
	rules, extraArgs := cfg.Rules, cfg.ExtraArgs
	if len(rules) == 0 {
		return nil, fmt.Errorf("missing spectral rules")
	}
	spectralPath := cfg.Script
	ok := (cfg.Script != "")
	if !ok {
		spectralPath, ok = findSpectralAdjacent()
	}
	if !ok {
		spectralPath, ok = findSpectralFromPath()
	}
	if !ok {
		return nil, fmt.Errorf("cannot find spectral linter: `npm install -g spectral-cli` and try again?")
	}

	var rulesPath string
	rulesFile, err := os.CreateTemp("", "*.yaml")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp rules file: %w", err)
	}
	defer rulesFile.Close()
	resolvedRules := make([]string, len(rules))
	for i := range rules {
		resolvedRules[i], err = filepath.Abs(rules[i])
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
	rulesPath = rulesFile.Name()
	go func() {
		<-ctx.Done()
		os.Remove(rulesPath)
	}()
	return &Spectral{
		rules:        resolvedRules,
		spectralPath: spectralPath,
		rulesPath:    rulesPath,
		extraArgs:    extraArgs,
	}, nil
}

// Match implements linter.Linter.
func (s *Spectral) Match(rcConfig *config.ResourceSet) ([]string, error) {
	return files.LocalFSSource{}.Match(rcConfig)
}

// WithOverride implements linter.Linter.
func (s *Spectral) WithOverride(ctx context.Context, override *config.Linter) (linter.Linter, error) {
	if override.Spectral == nil {
		return nil, fmt.Errorf("invalid linter override")
	}
	merged := *override.Spectral
	merged.Rules = append(s.rules, merged.Rules...)
	return New(ctx, &merged)
}

// Run runs spectral on the given paths. Linting output is written to standard
// output by spectral. Returns an error when lint fails configured rules.
func (l *Spectral) Run(ctx context.Context, _ string, paths ...string) error {
	cmd := exec.CommandContext(
		ctx,
		l.spectralPath,
		append(append([]string{"lint", "-r", l.rulesPath}, l.extraArgs...), paths...)...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func findSpectralAdjacent() (string, bool) {
	if len(os.Args) < 1 {
		// hmmm
		return "", false
	}
	binDir := filepath.Dir(os.Args[0])
	binFile := filepath.Join(binDir, "spectral")
	st, err := os.Stat(binFile)
	return binFile, err == nil && !st.IsDir() && st.Mode()&0111 != 0
}

func findSpectralFromPath() (string, bool) {
	path, err := exec.LookPath("spectral")
	return path, err == nil
}
