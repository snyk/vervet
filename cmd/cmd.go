// Package cmd provides subcommands for the vervet CLI.
package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ghodss/yaml"
	"github.com/urfave/cli/v2"

	"github.com/snyk/vervet"
	"github.com/snyk/vervet/config"
	"github.com/snyk/vervet/internal/compiler"
)

// App is the vervet CLI application.
var App = &cli.App{
	Name:  "vervet",
	Usage: "OpenAPI resource versioning tool",
	Commands: []*cli.Command{{
		Name:      "resolve",
		Usage:     "Aggregate, render and validate resource specs at a particular version",
		ArgsUsage: "[resource root]",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "at"},
		},
		Action: Resolve,
	}, {
		Name:      "compile",
		Usage:     "Compile versioned resources into versioned OpenAPI specs",
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
				Usage:   "OpenAPI specification to include in all compiled versions",
			},
		},
		Action: Compile,
	}, {
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
	}, {
		Name:      "localize",
		Usage:     "Localize references and validate a single OpenAPI spec file",
		ArgsUsage: "[spec.yaml file]",
		Action:    Localize,
	}, {
		Name:      "versions",
		Usage:     "List all resource versions declared in a spec",
		ArgsUsage: "[resource root]",
		Action:    Versions,
	}},
}

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

// Resolve aggregates, renders and validates resource specs at a particular
// version.
func Resolve(ctx *cli.Context) error {
	specDir, err := absPath(ctx.Args().Get(0))
	if err != nil {
		return err
	}
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
}

// Localize references and validate a single OpenAPI spec file
func Localize(ctx *cli.Context) error {
	if ctx.Args().Len() < 1 {
		return fmt.Errorf("missing spec.yaml file")
	}
	specFile, err := absPath(ctx.Args().Get(0))
	t, err := vervet.NewDocumentFile(specFile)
	if err != nil {
		return fmt.Errorf("failed to load spec from %q: %v", specFile, err)
	}

	// Localize all references, so we emit a completely self-contained OpenAPI document.
	err = vervet.Localize(t)
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
}

// Versions lists all resource versions declared in a spec.
func Versions(ctx *cli.Context) error {
	specDir, err := absPath(ctx.Args().Get(0))
	if err != nil {
		return err
	}
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
}

func absPath(path string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("path missing or empty")
	}
	return filepath.Abs(path)
}
