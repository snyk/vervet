package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/urfave/cli/v2"

	"github.com/snyk/vervet/v5/config"
	"github.com/snyk/vervet/v5/internal/compiler"
)

// BuildCommand is the `vervet build` subcommand.
var BuildCommand = cli.Command{
	Name:      "build",
	Usage:     "Build versioned resources into versioned OpenAPI specs",
	ArgsUsage: "[input resources root] [output api root]",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "config",
			Aliases: []string{"c", "conf"},
			Usage:   "Project configuration file",
		},
		&cli.BoolFlag{
			Name:  "lint",
			Usage: "DEPRECATED; Enable linting during build",
			Value: false,
		},
		&cli.StringFlag{
			Name:    "include",
			Aliases: []string{"I"},
			Usage:   "OpenAPI specification to include in build output",
		},
	},
	Action: Build,
}

// Build compiles versioned resources into versioned API specs.
func Build(ctx *cli.Context) error {
	project, err := projectFromContext(ctx)
	if err != nil {
		return err
	}
	comp, err := compiler.New(ctx.Context, project)
	if err != nil {
		return err
	}
	err = comp.BuildAll(ctx.Context)
	if err != nil {
		return err
	}
	if ctx.Bool("lint") {
		return Lint(ctx)
	}
	return nil
}

// LintCommand is the `vervet lint` subcommand.
var LintCommand = cli.Command{
	Name:      "lint",
	Usage:     "DEPRECATED; Lint  versioned resources",
	ArgsUsage: "[input resources root] [output api root]",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "config",
			Aliases: []string{"c", "conf"},
			Usage:   "Project configuration file",
		},
	},
	Action: Lint,
}

// Lint checks versioned resources against linting rules.
func Lint(ctx *cli.Context) error {
	fmt.Fprintln(os.Stderr, `
Vervet is no longer needed to perform API linting.
Update your projects to use 'spectral lint' or 'sweater-comb lint' instead.
`[1:])
	fmt.Fprintln(os.Stderr, `
Attempting to run 'sweater-comb lint' from your current working directory...
`[1:])
	scpath, err := exec.LookPath("sweater-comb")
	if err != nil {
		scpath = "node_modules/.bin/sweater-comb"
	}
	if _, err := os.Stat(scpath); err != nil {
		fmt.Fprintln(os.Stderr, `
Failed to find a 'sweater-comb' executable script.
'npm install @snyk/sweater-comb' and try again?
`[1:])
		return err
	}
	cmd := exec.Command(scpath, "lint")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func projectFromContext(ctx *cli.Context) (*config.Project, error) {
	var project *config.Project
	if ctx.Args().Len() == 0 {
		var configPath string
		if s := ctx.String("config"); s != "" {
			configPath = s
		} else {
			configPath = ".vervet.yaml"
		}
		f, err := os.Open(configPath)
		if err != nil {
			return nil, fmt.Errorf("failed to open %q: %w", configPath, err)
		}
		defer f.Close()
		project, err = config.Load(f)
		if err != nil {
			return nil, err
		}
	} else {
		api := &config.API{
			Resources: []*config.ResourceSet{{
				Path: ctx.Args().Get(0),
			}},
			Output: &config.Output{
				Path: ctx.Args().Get(1),
			},
		}
		if includePath := ctx.String("include"); includePath != "" {
			api.Overlays = append(api.Overlays, &config.Overlay{
				Include: includePath,
			})
		}
		project = &config.Project{
			APIs: map[string]*config.API{
				"": api,
			},
		}
	}
	return project, nil
}
