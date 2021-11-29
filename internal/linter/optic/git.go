package optic

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
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
