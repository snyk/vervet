package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/ghodss/yaml"
	"github.com/urfave/cli/v2"

	"github.com/snyk/vervet"
)

func main() {
	app := &cli.App{
		Name:  "vervet",
		Usage: "API endpoint versioning tool",
		Commands: []*cli.Command{{
			Name:      "resolve",
			Usage:     "Aggregate, render and validate endpoint specs at a particular version",
			ArgsUsage: "[spec path]",
			Flags: []cli.Flag{
				&cli.StringFlag{Name: "at"},
			},
			Action: func(ctx *cli.Context) error {
				if ctx.Args().Len() < 1 {
					return fmt.Errorf("missing spec path")
				}
				specDir := ctx.Args().Get(0)
				specVersions, err := vervet.LoadSpecVersions(specDir)
				if err != nil {
					return err
				}
				specVersion, err := specVersions.At(ctx.String("at"))
				if err != nil {
					return err
				}

				yamlBuf, err := vervet.ToSpecYAML(specVersion)
				if err != nil {
					return fmt.Errorf("failed to convert JSON to YAML: %w", err)
				}
				fmt.Printf(string(yamlBuf))

				err = specVersion.Validate(ctx.Context)
				if err != nil {
					return fmt.Errorf("error: spec validation failed: %w", err)
				}
				return nil
			},
		}, {
			Name:      "localize",
			Usage:     "Localize references and validate a single OpenAPI spec file",
			ArgsUsage: "[spec.yaml file]",
			Action: func(ctx *cli.Context) error {
				if ctx.Args().Len() < 1 {
					return fmt.Errorf("missing spec.yaml file")
				}
				specFile := ctx.Args().Get(0)
				t, err := vervet.LoadSpecFile(specFile)
				if err != nil {
					return fmt.Errorf("failed to load spec from %q: %v", specFile, err)
				}

				// Localize all references, so we emit a completely self-contained OpenAPI document.
				err = vervet.NewLocalizer(t).Localize()
				if err != nil {
					return fmt.Errorf("failed to localize refs: %w", err)
				}

				yamlBuf, err := vervet.ToSpecYAML(t)
				if err != nil {
					return fmt.Errorf("failed to convert JSON to YAML: %w", err)
				}
				fmt.Printf(string(yamlBuf))

				err = t.Validate(ctx.Context)
				if err != nil {
					return fmt.Errorf("error: spec validation failed: %w", err)
				}
				return nil
			},
		}, {
			Name:      "versions",
			Usage:     "List all endpoint versions declared in a spec",
			ArgsUsage: "[spec path]",
			Action: func(ctx *cli.Context) error {
				if ctx.Args().Len() < 1 {
					return fmt.Errorf("missing spec path")
				}
				specDir := ctx.Args().Get(0)
				specVersions, err := vervet.LoadSpecVersions(specDir)
				if err != nil {
					return fmt.Errorf("failed to load spec from %q: %w", specDir, err)
				}
				jsonBuf, err := json.Marshal(specVersions.Versions())
				if err != nil {
					return fmt.Errorf("failed to marshal spec versions: %w", err)
				}
				yamlBuf, err := yaml.JSONToYAML(jsonBuf)
				if err != nil {
					return fmt.Errorf("failed to convert to YAML: %w", err)
				}
				fmt.Printf(string(yamlBuf))
				return nil
			},
		}},
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
