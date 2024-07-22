package cmd

import (
	"fmt"
	"os"

	"github.com/urfave/cli/v2"

	"github.com/snyk/vervet/v7/config"
	"github.com/snyk/vervet/v7/internal/compiler"
	"github.com/snyk/vervet/v7/internal/simplebuild"
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
		&cli.StringFlag{
			Name:    "include",
			Aliases: []string{"I"},
			Usage:   "OpenAPI specification to include in build output",
		},
	},
	Action: Build,
}

var SimpleBuildCommand = cli.Command{
	Name:      "simplebuild",
	Usage:     "Build versioned resources into versioned OpenAPI specs",
	ArgsUsage: "[input resources root]",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "config",
			Aliases: []string{"c", "conf"},
			Usage:   "Project configuration file",
		},
		&cli.StringFlag{
			Name:    "include",
			Aliases: []string{"I"},
			Usage:   "OpenAPI specification to include in build output",
		},
	},
	Action: SimpleBuild,
}

func SimpleBuild(ctx *cli.Context) error {
	project, err := projectFromContext(ctx)
	if err != nil {
		return err
	}
	err = simplebuild.Build(ctx.Context, project)
	return err
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
	return nil
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
