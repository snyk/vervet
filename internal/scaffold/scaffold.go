package scaffold

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/ghodss/yaml"

	"github.com/snyk/vervet/v8/internal/files"
)

// ErrAlreadyInitialized is used when scaffolding is being run on a project that is already setup.
var ErrAlreadyInitialized = fmt.Errorf("project files already exist")

// Scaffold defines a Vervet API project scaffold.
type Scaffold struct {
	dst, src string
	force    bool
	manifest *Manifest
}

const manifestV1 = "1"

// Manifest defines the scaffold manifest model.
type Manifest struct {
	Version string

	// Organize contains a mapping of files relative to Scaffold src, to be
	// copied into dst, relative to dst. Missing intermediate directories will
	// be created as needed.
	Organize map[string]string `json:"organize"`
}

// Option defines a functional option that modifies a new Scaffold in the
// constructor.
type Option func(*Scaffold)

// Force sets the force flag on a Scaffold, which determines whether existing
// destination files will be overwritten. Default is false.
func Force(force bool) Option {
	return func(s *Scaffold) {
		s.force = force
	}
}

// New returns a new Scaffold loaded from source directory `src` for operation
// on destination directory `dst`. The Scaffold src must contain a
// `manifest.yaml` which defines how dst will be provisioned.
func New(dst, src string, options ...Option) (*Scaffold, error) {
	if dst == "" || src == "" {
		return nil, fmt.Errorf("source and destination are required")
	}
	var err error
	dst, err = filepath.Abs(dst)
	if err != nil {
		return nil, err
	}
	src, err = filepath.Abs(src)
	if err != nil {
		return nil, err
	}

	manifestPath := filepath.Join(src, "manifest.yaml")
	contents, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, err
	}
	var manifest Manifest
	err = yaml.Unmarshal(contents, &manifest)
	if err != nil {
		return nil, err
	}
	err = manifest.validate(src)
	if err != nil {
		return nil, err
	}
	s := &Scaffold{src: src, dst: dst, manifest: &manifest}
	for i := range options {
		options[i](s)
	}
	return s, nil
}

// Organize provisions files from the scaffold source into its destination.
func (s *Scaffold) Organize() error {
	for dstItem, srcItem := range s.manifest.Organize {
		dstPath := filepath.Join(s.dst, dstItem)
		// If we're not force overwriting, check if files already exist.
		if !s.force {
			_, err := os.Stat(dstPath)
			if err == nil {
				// Project files already exist.
				return ErrAlreadyInitialized
			}
			if !os.IsNotExist(err) {
				// Something else went wrong; the file not existing is the desired
				// state.
				return err
			}
		}
		srcPath := filepath.Join(s.src, srcItem)
		err := files.CopyItem(dstPath, srcPath, s.force)
		if err != nil {
			return fmt.Errorf("failed to copy %q to %q: %w", srcPath, dstPath, err)
		}
	}
	return nil
}

// Init runs a script called `init` in the scaffold source if present,
// in the destination directory.
func (s *Scaffold) Init() error {
	initScript := filepath.Join(s.src, "init")
	if _, err := os.Stat(initScript); os.IsNotExist(err) {
		return nil // no init script
	} else if err != nil {
		return err
	}
	cmd := exec.Command(initScript)
	cmd.Dir = s.dst
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("init script failed: %w", err)
	}
	return nil
}

func (m *Manifest) validate(src string) error {
	if m.Version == "" {
		m.Version = manifestV1
	}
	if m.Version != manifestV1 {
		return fmt.Errorf("unsupported manifest version %q", m.Version)
	}

	if len(m.Organize) == 0 {
		return fmt.Errorf("empty manifest")
	}
	for _, srcItem := range m.Organize {
		srcPath := filepath.Join(src, srcItem)
		if _, err := os.Stat(srcPath); err != nil {
			return fmt.Errorf("cannot stat source item %q: %w", srcPath, err)
		}
	}
	return nil
}
