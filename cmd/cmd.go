// Package cmd provides subcommands for the vervet CLI.
package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/urfave/cli/v2"
)

// App is the vervet CLI application.
var App = &cli.App{
	Name:  "vervet",
	Usage: "OpenAPI resource versioning tool",
	Commands: []*cli.Command{{
		Name:      "resolve",
		Usage:     "Aggregate, render and validate resource specs at a particular version",
		ArgsUsage: "[resource root]",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "at"},
		},
		Action: Resolve,
	}, {
		Name:      "compile",
		Usage:     "Compile versioned resources into versioned OpenAPI specs",
		ArgsUsage: "[input resources root] [output api root]",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "config",
				Aliases: []string{"c", "conf"},
				Usage:   "Project configuration file",
			},
			&cli.BoolFlag{
				Name:  "lint",
				Usage: "Enable linting during build",
				Value: true,
			},
			&cli.StringFlag{
				Name:    "include",
				Aliases: []string{"I"},
				Usage:   "OpenAPI specification to include in all compiled versions",
			},
		},
		Action: Compile,
	}, {
		Name:      "lint",
		Usage:     "Lint  versioned resources",
		ArgsUsage: "[input resources root] [output api root]",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "config",
				Aliases: []string{"c", "conf"},
				Usage:   "Project configuration file",
			},
		},
		Action: Lint,
	}, {
		Name:      "localize",
		Usage:     "Localize references and validate a single OpenAPI spec file",
		ArgsUsage: "[spec.yaml file]",
		Action:    Localize,
	}, {
		Name: "version",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "config",
				Aliases: []string{"c", "conf"},
				Usage:   "Project configuration file",
			},
		},
		Subcommands: []*cli.Command{{
			Name:      "files",
			Usage:     "List resource spec files in a vervet project",
			ArgsUsage: "[api [resource]]",
			Action:    VersionFiles,
		}, {
			Name:      "list",
			Usage:     "List resource versions in a vervet project",
			ArgsUsage: "[api [resource]]",
			Action:    VersionList,
		}, {
			Name:      "new",
			Usage:     "Create a new resource version",
			ArgsUsage: "<api> <resource>",
			Action:    VersionNew,
		}},
	}},
}

func absPath(path string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("path missing or empty")
	}
	return filepath.Abs(path)
}

func projectConfig(ctx *cli.Context) (string, string, error) {
	var projectDir, configFile string
	var err error
	if cf := ctx.String("config"); cf != "" {
		configFile, err = filepath.Abs(cf)
		if err != nil {
			return "", "", err
		}
		projectDir = filepath.Dir(configFile)
	} else {
		configFile = filepath.Join(projectDir, ".vervet.yaml")
		projectDir, err = os.Getwd()
		if err != nil {
			return "", "", err
		}
	}
	return projectDir, configFile, nil
}
