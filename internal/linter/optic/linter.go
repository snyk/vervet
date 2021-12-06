// Package optic supports linting OpenAPI specs with Optic CI and Sweater Comb.
package optic

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"time"

	"github.com/ghodss/yaml"
	"go.uber.org/multierr"

	"github.com/snyk/vervet"
	"github.com/snyk/vervet/config"
	"github.com/snyk/vervet/internal/files"
	"github.com/snyk/vervet/internal/linter"
)

// Optic runs a Docker image containing Optic CI and built-in rules.
type Optic struct {
	image      string
	fromSource files.FileSource
	toSource   files.FileSource
	runner     commandRunner
	timeNow    func() time.Time
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
	var fromSource, toSource files.FileSource
	var err error

	if from == "" {
		fromSource = files.NilSource{}
	} else {
		fromSource, err = newGitRepoSource(".", from)
		if err != nil {
			return nil, err
		}
	}

	if to == "" {
		toSource = files.LocalFSSource{}
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
		timeNow:    time.Now,
	}, nil
}

// Match implements linter.Linter.
func (o *Optic) Match(rcConfig *config.ResourceSet) ([]string, error) {
	fromFiles, err := o.fromSource.Match(rcConfig)
	if err != nil {
		return nil, err
	}
	toFiles, err := o.toSource.Match(rcConfig)
	if err != nil {
		return nil, err
	}
	// Unique set of files
	// TODO: normalization needed? or if not needed, tested to prove it?
	filesMap := map[string]struct{}{}
	for i := range fromFiles {
		filesMap[fromFiles[i]] = struct{}{}
	}
	for i := range toFiles {
		filesMap[toFiles[i]] = struct{}{}
	}
	var result []string
	for k := range filesMap {
		result = append(result, k)
	}
	sort.Strings(result)
	log.Println(result)
	return result, nil
}

// WithOverride implements linter.Linter.
func (*Optic) WithOverride(ctx context.Context, override *config.Linter) (linter.Linter, error) {
	if override.OpticCI == nil {
		return nil, fmt.Errorf("invalid linter override")
	}
	return New(ctx, override.OpticCI)
}

// Run runs Optic CI on the given paths. Linting output is written to standard
// output by Optic CI. Returns an error when lint fails configured rules.
func (o *Optic) Run(ctx context.Context, paths ...string) error {
	var errs error
	var comparisons []comparison
	var dockerArgs []string
	for i := range paths {
		comparison, volumeArgs, err := o.newComparison(paths[i])
		if err != nil {
			errs = multierr.Append(errs, err)
		} else {
			comparisons = append(comparisons, comparison)
			dockerArgs = append(dockerArgs, volumeArgs...)
		}
	}
	err := o.bulkCompare(ctx, comparisons, dockerArgs)
	errs = multierr.Append(errs, err)
	return errs
}

type comparison struct {
	From    string  `json:"from"`
	To      string  `json:"to"`
	Context Context `json:"context"`
}

type bulkCompareInput struct {
	Comparisons []comparison `json:"comparisons"`
}

func (o *Optic) newComparison(path string) (comparison, []string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return comparison{}, nil, err
	}
	var volumeArgs []string

	// TODO: This assumes the file being linted is a resource version spec
	// file, and not a compiled one. We don't yet have rules that support
	// diffing _compiled_ specs; that will require a different context and rule
	// set for Vervet Underground integration.
	opticCtx, err := o.contextFromPath(path)
	if err != nil {
		return comparison{}, nil, fmt.Errorf("failed to get context from path %q: %w", path, err)
	}

	cmp := comparison{
		Context: *opticCtx,
	}

	fromFile, err := o.fromSource.Fetch(path)
	if err != nil {
		return comparison{}, nil, err
	}
	if fromFile != "" {
		cmp.From = "/from/" + path
		volumeArgs = append(volumeArgs,
			"-v", cwd+":/from",
			"-v", fromFile+":/from/"+path,
		)
	}

	toFile, err := o.toSource.Fetch(path)
	if err != nil {
		return comparison{}, nil, err
	}
	if toFile != "" {
		cmp.To = "/to/" + path
		volumeArgs = append(volumeArgs,
			"-v", cwd+":/to",
			"-v", toFile+":/to/"+path,
		)
	}

	return cmp, volumeArgs, nil
}

var opticFromOutputRE = regexp.MustCompile(`/from/`)
var opticToOutputRE = regexp.MustCompile(`/to/`)

func (o *Optic) bulkCompare(ctx context.Context, comparisons []comparison, dockerArgs []string) error {
	input := &bulkCompareInput{
		Comparisons: comparisons,
	}
	inputFile, err := ioutil.TempFile("", "*-input.json")
	if err != nil {
		return err
	}
	defer inputFile.Close()
	err = json.NewEncoder(inputFile).Encode(&input)
	if err != nil {
		return err
	}
	if err := inputFile.Sync(); err != nil {
		return err
	}

	// TODO: link to command line arguments for optic-ci when available.
	cmdline := append([]string{"run", "--rm", "-v", inputFile.Name() + ":/input.json"}, dockerArgs...)
	cmdline = append(cmdline, o.image, "bulk-compare", "--input", "/input.json")
	log.Println(cmdline)
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
			return
		case <-ctx.Done():
			return
		case <-time.After(cmdTimeout):
			log.Printf("warning: timeout waiting for output to flush")
			return
		}
	}()
	go func() {
		defer pipeReader.Close()
		sc := bufio.NewScanner(pipeReader)
		for sc.Scan() {
			line := sc.Text()
			line = opticFromOutputRE.ReplaceAllString(line, "("+o.fromSource.Name()+"):")
			line = opticToOutputRE.ReplaceAllString(line, "("+o.toSource.Name()+"):")
			fmt.Println(line)
		}
		if err := sc.Err(); err != nil {
			fmt.Fprintf(os.Stderr, "error reading stdout: %v", err)
		}
		close(ch)
	}()
	cmd.Stdin = os.Stdin
	cmd.Stdout = pipeWriter
	cmd.Stderr = os.Stderr
	err = o.runner.run(cmd)
	if err != nil {
		return fmt.Errorf("lint failed: %w", err)
	}
	return nil
}

func (o *Optic) contextFromPath(path string) (*Context, error) {
	dateDir := filepath.Dir(path)
	resourceDir := filepath.Dir(dateDir)
	date, resource := filepath.Base(dateDir), filepath.Base(resourceDir)
	if _, err := time.Parse("2006-01-02", date); err != nil {
		return nil, err
	}
	stability, err := o.loadStability(path)
	if err != nil {
		return nil, err
	}
	if _, err := vervet.ParseStability(stability); err != nil {
		return nil, err
	}
	return &Context{
		ChangeDate:     o.timeNow().UTC().Format("2006-01-02"),
		ChangeResource: resource,
		ChangeVersion: Version{
			Date:      date,
			Stability: stability,
		},
	}, nil
}

func (o *Optic) loadStability(path string) (string, error) {
	var (
		doc struct {
			Stability string `json:"x-snyk-api-stability"`
		}
		contentsFile string
		err          error
	)
	contentsFile, err = o.fromSource.Fetch(path)
	if err != nil {
		return "", err
	}
	if contentsFile == "" {
		contentsFile, err = o.toSource.Fetch(path)
		if err != nil {
			return "", err
		}
	}
	contents, err := ioutil.ReadFile(contentsFile)
	if err != nil {
		return "", err
	}
	err = yaml.Unmarshal(contents, &doc)
	if err != nil {
		return "", err
	}
	return doc.Stability, nil
}

const cmdTimeout = time.Second * 30
