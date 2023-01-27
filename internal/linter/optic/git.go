package optic

import (
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"go.uber.org/multierr"

	"github.com/snyk/vervet/v5/config"
)

// gitRepoSource is a fileSource that resolves files out of a specific git
// commit.
type gitRepoSource struct {
	repo   *git.Repository
	commit *object.Commit
	roots  map[string]string
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
	return &gitRepoSource{repo: repo, commit: commit, roots: map[string]string{}}, nil
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

// Prefetch implements FileSource
func (g *gitRepoSource) Prefetch(root string) (string, error) {
	tree, err := g.commit.Tree()
	if err != nil {
		return "", err
	}
	tree, err = tree.Tree(root)
	if err != nil {
		return "", err
	}
	tempDir, err := os.MkdirTemp("", "")
	if err != nil {
		return "", err
	}
	err = func() error {
		// Wrap this in a closure to simplify walker cleanup
		w := object.NewTreeWalker(tree, true, map[plumbing.Hash]bool{})
		defer w.Close()
		for {
			ok, err := func() (bool, error) {
				// Wrap this in a closure to release fds early & often
				name, entry, err := w.Next()
				if err == io.EOF {
					return false, nil
				} else if err != nil {
					return false, err
				}
				if !entry.Mode.IsFile() {
					return true, nil
				}
				blob, err := object.GetBlob(g.repo.Storer, entry.Hash)
				if err != nil {
					return false, err
				}
				err = os.MkdirAll(filepath.Join(tempDir, filepath.Dir(name)), 0777)
				if err != nil {
					return false, err
				}
				tempFile, err := os.Create(filepath.Join(tempDir, name))
				if err != nil {
					return false, err
				}
				defer tempFile.Close()
				blobContents, err := blob.Reader()
				if err != nil {
					return false, err
				}
				_, err = io.Copy(tempFile, blobContents)
				if err != nil {
					return false, err
				}
				return true, nil
			}()
			if err != nil {
				return err
			}
			if !ok {
				return nil
			}
		}
	}()
	if err != nil {
		// Clean up temp dir if we failed to populate it
		errs := multierr.Append(nil, err)
		err := os.RemoveAll(tempDir)
		if err != nil {
			errs = multierr.Append(errs, err)
		}
		return "", errs
	}
	g.roots[root] = tempDir
	return tempDir, nil
}

// Fetch implements FileSource.
func (g *gitRepoSource) Fetch(path string) (string, error) {
	var matchRoot string
	// Linear search for this is probably good enough. Could use a trie if it
	// gets out of hand.
	for root := range g.roots {
		if strings.HasPrefix(path, root) {
			matchRoot = root
		}
	}
	if matchRoot == "" {
		return "", nil
	}
	matchPath := strings.Replace(path, matchRoot, g.roots[matchRoot], 1)
	if _, err := os.Stat(matchPath); os.IsNotExist(err) {
		return "", nil
	}
	return matchPath, nil
}

// Close implements fileSource.
func (g *gitRepoSource) Close() error {
	var errs error
	for _, tempDir := range g.roots {
		errs = multierr.Append(errs, os.RemoveAll(tempDir))
	}
	return errs
}
