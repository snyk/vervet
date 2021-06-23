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
		Name: "vervet",
		Commands: []*cli.Command{{
			Name: "resolve",
			Subcommands: []*cli.Command{{
				Name:      "file",
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
				Name:      "endpoint",
				ArgsUsage: "[endpoint path]",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "at"},
				},
				Action: func(ctx *cli.Context) error {
					if ctx.Args().Len() < 1 {
						return fmt.Errorf("missing endpoint path")
					}
					specDir := ctx.Args().Get(0)
					epVersions, err := vervet.LoadEndpointVersions(specDir)
					if err != nil {
						return fmt.Errorf("failed to load end from %q: %w", specDir, err)
					}
					epVersion, err := epVersions.VersionAt(ctx.String("at"))
					if err != nil {
						return err
					}

					yamlBuf, err := vervet.ToSpecYAML(epVersion)
					if err != nil {
						return fmt.Errorf("failed to convert JSON to YAML: %w", err)
					}
					fmt.Printf(string(yamlBuf))

					err = epVersion.Validate(ctx.Context)
					if err != nil {
						return fmt.Errorf("error: spec validation failed: %w", err)
					}
					return nil
				},
			}},
		}, {
			Name: "versions",
			Subcommands: []*cli.Command{{
				Name:      "show",
				ArgsUsage: "[endpoint directory]",
				Action: func(ctx *cli.Context) error {
					if ctx.Args().Len() < 1 {
						return fmt.Errorf("missing endpoint path")
					}
					specDir := ctx.Args().Get(0)
					epVersions, err := vervet.LoadEndpointVersions(specDir)
					if err != nil {
						return fmt.Errorf("failed to load endpoint from %q: %w", specDir, err)
					}
					jsonBuf, err := json.Marshal(epVersions.Versions())
					if err != nil {
						return fmt.Errorf("failed to marshal endpoint versions: %w", err)
					}
					yamlBuf, err := yaml.JSONToYAML(jsonBuf)
					if err != nil {
						return fmt.Errorf("failed to convert to YAML: %w", err)
					}
					fmt.Printf(string(yamlBuf))
					return nil
				},
			}, {
				Name:      "list",
				ArgsUsage: "[endpoint directory]",
				Action: func(ctx *cli.Context) error {
					if ctx.Args().Len() < 1 {
						return fmt.Errorf("missing endpoint path")
					}
					specDir := ctx.Args().Get(0)
					epVersions, err := vervet.LoadEndpointVersions(specDir)
					if err != nil {
						return fmt.Errorf("failed to load end from %q: %w", specDir, err)
					}
					for k := range epVersions.Versions() {
						fmt.Println(k)
					}
					return nil
				},
			}},
		}},
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
