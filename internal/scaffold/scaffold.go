package scaffold

import (
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/ghodss/yaml"
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
	contents, err := ioutil.ReadFile(manifestPath)
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
		err := s.copyItem(dstPath, srcPath)
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
	err := cmd.Run()
	if err != nil {
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

func (s *Scaffold) copyItem(dstPath, srcPath string) error {
	if st, err := os.Stat(srcPath); err == nil && st.IsDir() {
		return s.copyDir(dstPath, srcPath)
	} else if err == nil {
		return s.copyFile(dstPath, srcPath)
	} else {
		return err
	}
}

func (s *Scaffold) copyDir(dstPath, srcPath string) error {
	return filepath.WalkDir(srcPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		name, err := filepath.Rel(srcPath, path)
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		return s.copyFile(filepath.Join(dstPath, name), path)
	})
}

func (s *Scaffold) copyFile(dstPath, srcPath string) error {
	srcf, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	flags := os.O_CREATE | os.O_WRONLY | os.O_TRUNC
	if !s.force {
		flags = flags | os.O_EXCL
	}

	dstDir := filepath.Dir(dstPath)
	if dstDir != "." {
		err = os.MkdirAll(dstDir, 0777)
		if err != nil {
			return err
		}
	}

	dstf, err := os.OpenFile(dstPath, flags, 0666)
	if err != nil {
		return err
	}

	_, err = io.Copy(dstf, srcf)
	if err != nil {
		return err
	}
	return nil
}
