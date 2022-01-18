package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/urfave/cli/v2"

	"github.com/snyk/vervet/v3/internal/scaffold"
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
	if err == scaffold.ErrAlreadyInitialized {
		// If the project files already exist, prompt the user to see if they want to overwrite them.
		vervetApp, err := appFromContext(ctx.Context)
		if err != nil {
			return err
		}
		prompt := vervetApp.Params.Prompt
		overwrite, err := prompt.Confirm("Scaffold already initialized; do you want to overwrite")
		if err != nil {
			return err
		}
		if overwrite {
			forceFn := scaffold.Force(true)
			forceFn(sc)
			// If an error happens with --force enabled, something new has gone wrong.
			if err = sc.Organize(); err != nil {
				return err
			}
		}
		return nil
	} else if err != nil {
		return err
	}
	err = sc.Init()
	if err != nil {
		return err
	}
	return nil
}
