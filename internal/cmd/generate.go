package cmd

import (
	"os"

	"github.com/urfave/cli/v2"

	"github.com/snyk/vervet/v3/config"
	"github.com/snyk/vervet/v3/internal/generator"
)

// Generate executes code generators against OpenAPI specs.
func Generate(ctx *cli.Context) error {
	projectDir, configFile, err := projectConfig(ctx)
	if err != nil {
		return err
	}
	f, err := os.Open(configFile)
	if err != nil {
		return err
	}
	defer f.Close()
	proj, err := config.Load(f)
	if err != nil {
		return err
	}

	selectedGenerators := map[string]struct{}{}
	for i := 0; i < ctx.Args().Len(); i++ {
		selectedGenerators[ctx.Args().Get(i)] = struct{}{}
	}

	// Option to load generators and overlay onto project config
	if genFile := ctx.String("generators"); genFile != "" {
		f, err := os.Open(genFile)
		if err != nil {
			return err
		}
		defer f.Close()
		generators, err := config.LoadGenerators(f)
		if err != nil {
			return err
		}
		for k, v := range generators {
			proj.Generators[k] = v
		}
	}
	// If a list of specific generators were specified, only instantiate those.
	if len(selectedGenerators) > 0 {
		for k := range proj.Generators {
			if _, ok := selectedGenerators[k]; !ok {
				delete(proj.Generators, k)
			}
		}
	}

	options := []generator.Option{generator.Force(true)}
	if ctx.Bool("debug") {
		options = append(options, generator.Debug(true))
	}

	generators, err := generator.NewMap(proj.Generators, options...)
	if err != nil {
		return err
	}

	err = os.Chdir(projectDir)
	if err != nil {
		return err
	}

	// TODO: everything below here can probably move to internal/generator/...
	resources, err := generator.MapResources(proj)
	if err != nil {
		return err
	}

	for _, gen := range generators {
		err := gen.Execute(resources)
		if err != nil {
			return err
		}
	}
	return nil
}
