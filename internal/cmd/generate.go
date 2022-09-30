package cmd

import (
	"path/filepath"

	"github.com/urfave/cli/v2"

	"github.com/snyk/vervet/v5/generate"
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

	var generators []string
	for i := 0; i < ctx.Args().Len(); i++ {
		generators = append(generators, ctx.Args().Get(i))
	}

	genFile := ctx.String("generators")
	genFile, err = filepath.Abs(genFile)
	if err != nil {
		return err
	}

	params := generate.GeneratorParams{
		ProjectDir:     projectDir,
		ConfigFile:     configFile,
		Generators:     generators,
		GeneratorsFile: genFile,
		Debug:          ctx.Bool("debug"),
		DryRun:         ctx.Bool("dry-run"),
	}

	return generate.Generate(params)
}
