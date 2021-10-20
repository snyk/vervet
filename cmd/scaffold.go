package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/urfave/cli/v2"

	"github.com/snyk/vervet/internal/scaffold"
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
	targetDir, err := os.Getwd()
	if err != nil {
		return err
	}
	sc, err := scaffold.New(targetDir, scaffoldDir, scaffold.Force(ctx.Bool("force")))
	if err != nil {
		return err
	}
	err = sc.Organize()
	if err != nil {
		return err
	}
	err = sc.Init()
	if err != nil {
		return err
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
