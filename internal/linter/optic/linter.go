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
	"strings"
	"time"

	"github.com/ghodss/yaml"
	"go.uber.org/multierr"

	"github.com/snyk/vervet/v3"
	"github.com/snyk/vervet/v3/config"
	"github.com/snyk/vervet/v3/internal/files"
	"github.com/snyk/vervet/v3/internal/linter"
)

// Optic runs a Docker image containing Optic CI and built-in rules.
type Optic struct {
	image      string
	script     string
	fromSource files.FileSource
	toSource   files.FileSource
	runner     commandRunner
	timeNow    func() time.Time
	debug      bool
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
	image, script, from, to := cfg.Image, cfg.Script, cfg.Original, cfg.Proposed
	var fromSource, toSource files.FileSource
	var err error

	if !isDocker(script) {
		image = ""
	}

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
		script:     script,
		fromSource: fromSource,
		toSource:   toSource,
		runner:     &execCommandRunner{},
		timeNow:    time.Now,
		debug:      cfg.Debug,
	}, nil
}

func isDocker(script string) bool {
	return script == ""
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
func (o *Optic) Run(ctx context.Context, root string, paths ...string) error {
	var errs error
	var comparisons []comparison
	localFrom, err := o.fromSource.Prefetch(root)
	if err != nil {
		return err
	}
	localTo, err := o.toSource.Prefetch(root)
	if err != nil {
		return err
	}
	var dockerArgs []string
	var fromFilter, toFilter func(string) string
	if localFrom != "" {
		dockerArgs = append(dockerArgs, "-v", localFrom+":/from/"+root)
		if o.isDocker() {
			fromFilter = func(s string) string {
				return strings.Replace(s, localFrom, "/from/"+root, 1)
			}
		}
	}
	if localTo != "" {
		dockerArgs = append(dockerArgs, "-v", localTo+":/to/"+root)
		if o.isDocker() {
			toFilter = func(s string) string {
				return strings.Replace(s, localTo, "/to/"+root, 1)
			}
		}
	}
	for i := range paths {
		comparison, volumeArgs, err := o.newComparison(paths[i], fromFilter, toFilter)
		if err != nil {
			errs = multierr.Append(errs, err)
		} else {
			comparisons = append(comparisons, comparison)
			dockerArgs = append(dockerArgs, volumeArgs...)
		}
	}
	if o.isDocker() {
		err = o.bulkCompareDocker(ctx, comparisons, dockerArgs)
	} else {
		err = o.bulkCompareScript(ctx, comparisons)
	}
	errs = multierr.Append(errs, err)
	return errs
}

func (o *Optic) isDocker() bool {
	return isDocker(o.script)
}

type comparison struct {
	From    string  `json:"from,omitempty"`
	To      string  `json:"to,omitempty"`
	Context Context `json:"context,omitempty"`
}

type bulkCompareInput struct {
	Comparisons []comparison `json:"comparisons,omitempty"`
}

func (o *Optic) newComparison(path string, fromFilter, toFilter func(string) string) (comparison, []string, error) {
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
	cmp.From = fromFile
	if fromFilter != nil {
		cmp.From = fromFilter(cmp.From)
	}

	toFile, err := o.toSource.Fetch(path)
	if err != nil {
		return comparison{}, nil, err
	}
	cmp.To = toFile
	if toFilter != nil {
		cmp.To = toFilter(cmp.To)
	}

	return cmp, volumeArgs, nil
}

func (o *Optic) bulkCompareScript(ctx context.Context, comparisons []comparison) error {
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
	if o.debug {
		log.Println("input.json:")
		err = json.NewEncoder(os.Stderr).Encode(&input)
		if err != nil {
			return err
		}
	}
	if err := inputFile.Sync(); err != nil {
		return err
	}

	if o.debug {
		log.Print("bulk-compare input:")
		if err := json.NewEncoder(os.Stdout).Encode(&input); err != nil {
			log.Println("failed to encode input to stdout!")
		}
		log.Println()
	}

	cmd := exec.CommandContext(ctx, o.script, "bulk-compare", "--input", inputFile.Name())

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
			// TODO: this wanton breakage of FileSource encapsulation indicates
			// we probably need an abstraction if/when we support other
			// sources. VU might be such a future source...
			if fromGit, ok := o.fromSource.(*gitRepoSource); ok {
				for root, tempDir := range fromGit.roots {
					line = strings.ReplaceAll(line, tempDir, "("+fromGit.Name()+"):"+root)
				}
			}
			if toGit, ok := o.toSource.(*gitRepoSource); ok {
				for root, tempDir := range toGit.roots {
					line = strings.ReplaceAll(line, tempDir, "("+toGit.Name()+"):"+root)
				}
			}
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

var fromDockerOutputRE = regexp.MustCompile(`/from/`)
var toDockerOutputRE = regexp.MustCompile(`/to/`)

func (o *Optic) bulkCompareDocker(ctx context.Context, comparisons []comparison, dockerArgs []string) error {
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

	if o.debug {
		log.Print("bulk-compare input:")
		if err := json.NewEncoder(os.Stdout).Encode(&input); err != nil {
			log.Println("failed to encode input to stdout!")
		}
		log.Println()
	}

	// TODO: link to command line arguments for optic-ci when available.
	cmdline := append([]string{"run", "--rm", "-v", inputFile.Name() + ":/input.json"}, dockerArgs...)
	cmdline = append(cmdline, o.image, "bulk-compare", "--input", "/input.json")
	if o.debug {
		log.Printf("running: docker %s", strings.Join(cmdline, " "))
	}
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
			line = fromDockerOutputRE.ReplaceAllString(line, "("+o.fromSource.Name()+"):")
			line = toDockerOutputRE.ReplaceAllString(line, "("+o.toSource.Name()+"):")
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
