package optic

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"

	"github.com/snyk/vervet/config"
)

// gitRepoSource is a fileSource that resolves files out of a specific git
// commit.
type gitRepoSource struct {
	repo    *git.Repository
	commit  *object.Commit
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

// Name implements FileSource.
func (s *gitRepoSource) Name() string {
	return "commit " + s.commit.Hash.String()
}

// Match implements FileSource.
func (s *gitRepoSource) Match(rcConfig *config.ResourceSet) ([]string, error) {
	tree, err := s.repo.TreeObject(s.commit.TreeHash)
	if err != nil {
		return nil, err
	}
	var matches []string
	matchPattern := rcConfig.Path + "/**/spec.yaml"
	err = tree.Files().ForEach(func(f *object.File) error {
		// Check if this file matches
		if ok, err := doublestar.Match(matchPattern, f.Name); err != nil {
			return err
		} else if !ok {
			return nil
		}
		// Check exclude patterns
		for i := range rcConfig.Excludes {
			if ok, err := doublestar.Match(rcConfig.Excludes[i], f.Name); err != nil {
				return err
			} else if ok {
				return nil
			}
		}
		matches = append(matches, f.Name)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return matches, nil
}

// Fetch implements fileSource.
func (g *gitRepoSource) Fetch(path string) (string, error) {
	f, err := g.commit.File(path)
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
	fname := filepath.Join(g.tempDir, f.ID().String())
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
func (g *gitRepoSource) Close() (retErr error) {
	err := os.RemoveAll(g.tempDir)
	if err != nil {
		return err
	}
	return nil
}
