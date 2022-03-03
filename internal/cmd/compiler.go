package cmd

import (
	"fmt"
	"os"

	"github.com/urfave/cli/v2"

	"github.com/snyk/vervet/v4/config"
	"github.com/snyk/vervet/v4/internal/compiler"
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
			Usage: "Enable linting during build",
			Value: true,
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
	return runCompiler(ctx, project, ctx.Bool("lint"), true)
}

// LintCommand is the `vervet lint` subcommand.
var LintCommand = cli.Command{
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
}

// Lint checks versioned resources against linting rules.
func Lint(ctx *cli.Context) error {
	project, err := projectFromContext(ctx)
	if err != nil {
		return err
	}
	return runCompiler(ctx, project, true, false)
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

func runCompiler(ctx *cli.Context, project *config.Project, lint, build bool) error {
	comp, err := compiler.New(ctx.Context, project)
	if err != nil {
		return err
	}
	if lint {
		err = comp.LintResourcesAll(ctx.Context)
		if err != nil {
			return err
		}
	}
	if build {
		err = comp.BuildAll(ctx.Context)
		if err != nil {
			return err
		}
	}
	if lint {
		err = comp.LintOutputAll(ctx.Context)
		if err != nil {
			return err
		}
	}
	return nil
}
