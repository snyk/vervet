package spectral

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/ghodss/yaml"
)

// Spectral runs spectral on collections of files with a set of rules.
type Spectral struct {
	spectralPath string
	rulesPath    string
}

// New returns a new Spectral instance configured with the given rules.
func New(ctx context.Context, rules []string) (*Spectral, error) {
	if len(rules) == 0 {
		return nil, fmt.Errorf("missing spectral rules")
	}
	spectralPath, ok := findSpectralAdjacent()
	if !ok {
		spectralPath, ok = findSpectralFromPath()
	}
	if !ok {
		return nil, fmt.Errorf("cannot find spectral linter: `npm install -g spectral-cli` and try again?")
	}

	var rulesPath string
	rulesFile, err := ioutil.TempFile("", "*.yaml")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp rules file: %w", err)
	}
	defer rulesFile.Close()
	rulesDoc := map[string]interface{}{
		"extends": rules,
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
		select {
		case <-ctx.Done():
			os.Remove(rulesPath)
		}
	}()
	return &Spectral{
		spectralPath: spectralPath,
		rulesPath:    rulesPath,
	}, nil
}

// Run runs spectral on the given paths. Linting output is written to standard
// output by spectral. Returns an error when lint fails configured rules.
func (l *Spectral) Run(ctx context.Context, paths ...string) error {
	cmd := exec.CommandContext(ctx, l.spectralPath, append([]string{"lint", "-r", l.rulesPath}, paths...)...)
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
