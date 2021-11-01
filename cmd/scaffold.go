package cmd

import (
	"fmt"
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
