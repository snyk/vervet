package cmd

import (
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/urfave/cli/v2"
)

// ScaffoldInit creates a new project configuration from a provided scaffold directory.
func ScaffoldInit(ctx *cli.Context) error {
	scaffoldDir := ctx.Args().Get(0)
	if scaffoldDir == "" {
		return fmt.Errorf("a scaffold name is required")
	}
	var err error
	scaffoldDir, err = filepath.Abs(scaffoldDir)
	if err != nil {
		return err
	}
	if _, err := os.Stat(scaffoldDir); err != nil && os.IsNotExist(err) {
		return err
	}
	force := ctx.Bool("force")
	debug := ctx.Bool("debug")
	err = copyFile(".vervet.yaml", filepath.Join(scaffoldDir, "vervet.yaml"), force)
	if err != nil {
		return err
	}
	err = filepath.WalkDir(scaffoldDir, func(path string, d fs.DirEntry, err error) error {
		name, err := filepath.Rel(scaffoldDir, path)
		if err != nil {
			return err
		}
		src := filepath.Join(scaffoldDir, name)
		if d.IsDir() {
			if strings.HasPrefix(d.Name(), ".") {
				if debug {
					log.Printf("%s: skipped", src)
				}
				return fs.SkipDir
			}
			return nil
		}
		err = os.MkdirAll(filepath.Join(".vervet", filepath.Dir(name)), 0777)
		if err != nil {
			return err
		}
		dst := filepath.Join(".vervet", name)
		if debug {
			log.Printf("%s: copying to %s, force=%v", src, dst, force)
		}
		err = copyFile(dst, src, force)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}

	// Run init script if defined
	initScript := filepath.Join(scaffoldDir, "init")
	if _, err := os.Stat(initScript); err == nil {
		cmd := exec.Command(initScript)
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		cmd.Dir = cwd
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err = cmd.Run()
		if err != nil {
			return fmt.Errorf("init script failed: %w", err)
		}
	}
	return nil
}

func copyFile(dst, src string, force bool) error {
	srcf, err := os.Open(src)
	if err != nil {
		return err
	}
	flags := os.O_CREATE | os.O_WRONLY | os.O_TRUNC
	if !force {
		flags = flags | os.O_EXCL
	}
	dstf, err := os.OpenFile(dst, flags, 0666)
	if err != nil {
		return err
	}
	_, err = io.Copy(dstf, srcf)
	if err != nil {
		return err
	}
	return nil
}
