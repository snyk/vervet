// Package optic supports linting OpenAPI specs with Optic CI and Sweater Comb.
package optic

import (
	"context"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

// fileSource defines a source of files.
type fileSource interface {
	// Fetch retrieves the contents of the requested logical path as a local
	// file and returns the location where it may be found. An empty string,
	// rather than an error, is returned if the file does not exist.
	Fetch(path string) (string, error)

	// Close releases any resources consumed in content retrieval. Any files
	// returned by Fetch will no longer be available after calling Close, and
	// any further calls to Fetch will error.
	Close() error
}

// nilSource is a fileSource that does not have any files in it.
type nilSource struct{}

// Fetch implements fileSource.
func (nilSource) Fetch(path string) (string, error) {
	return "", nil
}

// Close implements fileSource.
func (nilSource) Close() error { return nil }

// workingCopySource is a fileSource that resolves files relative to the
// current working directory.
type workingCopySource struct{}

// Fetch implements fileSource.
func (workingCopySource) Fetch(path string) (string, error) {
	if _, err := os.Stat(path); err == nil {
		return path, nil
	} else if os.IsNotExist(err) {
		return "", nil
	} else {
		return "", err
	}
}

// Close implements fileSource.
func (workingCopySource) Close() error { return nil }

// gitRepoSource is a fileSource that resolves files out of a specific git
// commit.
type gitRepoSource struct {
	repo    *git.Repository
	commit  *object.Commit
	tag     string
	tempDir string
}

// newGitRepoSource returns a new gitRepoSource for the given git repository
// path and commit, which can be a branch, tag, commit hash or other "treeish".
func newGitRepoSource(path string, treeish string) (*gitRepoSource, error) {
	repo, err := git.PlainOpen(path)
	if err != nil {
		return nil, err
	}
	commitHash, err := repo.ResolveRevision(plumbing.Revision(treeish))
	if err != nil {
		return nil, err
	}
	commit, err := repo.CommitObject(*commitHash)
	if err != nil {
		return nil, err
	}
	tempDir, err := ioutil.TempDir("", "")
	if err != nil {
		return nil, err
	}
	return &gitRepoSource{repo: repo, commit: commit, tempDir: tempDir}, nil
}

// Fetch implements fileSource.
func (s *gitRepoSource) Fetch(path string) (string, error) {
	f, err := s.commit.File(path)
	if err != nil {
		if err == object.ErrFileNotFound {
			return "", nil
		}
		return "", err
	}
	r, err := f.Reader()
	if err != nil {
		return "", err
	}
	defer r.Close()
	fname := filepath.Join(s.tempDir, f.ID().String())
	tempf, err := os.Create(fname)
	if err != nil {
		return "", err
	}
	defer tempf.Close()
	_, err = io.Copy(tempf, r)
	if err != nil {
		return "", err
	}
	return fname, nil
}

// Close implements fileSource.
func (s *gitRepoSource) Close() (retErr error) {
	err := os.RemoveAll(s.tempDir)
	if err != nil {
		return err
	}
	return nil
}

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
func New(ctx context.Context, image string, from, to string) (*Optic, error) {
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

// Run runs Optic CI on the given paths. Linting output is written to standard
// output by Optic CI. Returns an error when lint fails configured rules.
func (l *Optic) Run(ctx context.Context, paths ...string) error {
	for i := range paths {
		err := l.runCompare(ctx, paths[i])
		if err != nil {
			return err
		}
	}
	return nil
}

func (l *Optic) runCompare(ctx context.Context, path string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	var args []string
	fromFile, err := l.fromSource.Fetch(path)
	if err != nil {
		return err
	}
	if fromFile != "" {
		args = append(args, "--from", fromFile)
	}
	toFile, err := l.toSource.Fetch(path)
	if err != nil {
		return err
	}
	if toFile != "" {
		args = append(args, "--to", toFile)
	}
	// TODO: provide rule context
	cmdline := append([]string{
		"run", "--rm",
		"-v", cwd + ":/target",
		l.image,
		"compare",
	}, args...)
	cmd := exec.CommandContext(ctx, "docker", cmdline...)
	return l.runner.run(cmd)
}
