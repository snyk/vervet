package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/urfave/cli/v2"

	"github.com/snyk/vervet/v4/config"
	"github.com/snyk/vervet/v4/internal/generator"
)

// GenerateCommand is the `vervet generate` subcommand.
var GenerateCommand = cli.Command{
	Name:      "generate",
	Usage:     "Generate artifacts from resource versioned OpenAPI specs",
	ArgsUsage: "<generator> [<generator2>...]",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "config",
			Aliases: []string{"c", "conf"},
			Usage:   "Project configuration file",
		},
		&cli.BoolFlag{
			Name:    "dry-run",
			Aliases: []string{"n"},
			Usage:   "Dry-run, listing files that would be generated",
		},
		&cli.StringFlag{
			Name:    "generators",
			Aliases: []string{"g", "gen", "generator"},
			Usage:   "Generators definition file",
		},
	},
	Action: Generate,
}

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
	generatorsHere := map[string]string{}
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
			generatorsHere[k] = filepath.Dir(genFile)
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
	if ctx.Bool("dry-run") {
		options = append(options, generator.DryRun(true))
	}
	projectHere := filepath.Dir(configFile)
	generators := map[string]*generator.Generator{}
	for k, genConf := range proj.Generators {
		genHere, ok := generatorsHere[k]
		if !ok {
			genHere = projectHere
		}
		genHere, err = filepath.Abs(genHere)
		if err != nil {
			return err
		}
		gen, err := generator.New(genConf, append(options, generator.Here(genHere))...)
		if err != nil {
			return err
		}
		generators[k] = gen
	}

	err = os.Chdir(projectDir)
	if err != nil {
		return err
	}

	resources, err := generator.MapResources(proj)
	if err != nil {
		return err
	}

	var allGeneratedFiles []string
	for _, gen := range generators {
		generatedFiles, err := gen.Execute(resources)
		if err != nil {
			return err
		}
		allGeneratedFiles = append(allGeneratedFiles, generatedFiles...)
	}

	for _, generatedFile := range allGeneratedFiles {
		fmt.Println(generatedFile)
	}

	return nil
}
