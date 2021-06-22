package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/ghodss/yaml"
	"github.com/urfave/cli/v2"

	"github.com/snyk/apiutil"
)

func main() {
	app := &cli.App{
		Name: "apiutil",
		Commands: []*cli.Command{{
			Name: "resolve",
			Action: func(ctx *cli.Context) error {
				specFile := ctx.Args().Get(0)
				t, err := apiutil.LoadSpecFile(specFile)
				if err != nil {
					return fmt.Errorf("failed to load spec from %q: %v", specFile, err)
				}

				// Localize all references, so we emit a completely self-contained OpenAPI document.
				err = apiutil.NewLocalizer(t).Localize()
				if err != nil {
					return fmt.Errorf("failed to localize refs: %w", err)
				}

				yamlBuf, err := apiutil.ToSpecYAML(t)
				if err != nil {
					return fmt.Errorf("failed to convert JSON to YAML: %w", err)
				}
				fmt.Printf(string(yamlBuf))

				err = t.Validate(context.TODO())
				if err != nil {
					return fmt.Errorf("error: spec validation failed: %w", err)
				}
				return nil
			},
		}, {
			Name: "endpoint",
			Subcommands: []*cli.Command{{
				Name: "resolve",
				Action: func(ctx *cli.Context) error {
					specDir := ctx.Args().Get(0)
					epVersions, err := apiutil.LoadEndpointVersions(specDir)
					if err != nil {
						return fmt.Errorf("failed to load end from %q: %w", specDir, err)
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
			}},
		}},
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
