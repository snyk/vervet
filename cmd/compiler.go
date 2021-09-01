package cmd

import (
	"fmt"
	"os"

	"github.com/urfave/cli/v2"

	"github.com/snyk/vervet/config"
	"github.com/snyk/vervet/internal/compiler"
)

// Compile compiles versioned resources into versioned API specs.
func Compile(ctx *cli.Context) error {
	project, err := projectFromContext(ctx)
	if err != nil {
		return err
	}
	return runCompiler(ctx, project, ctx.Bool("lint"), true)
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
