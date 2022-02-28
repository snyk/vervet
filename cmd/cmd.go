// Package cmd provides subcommands for the vervet CLI.
package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/manifoldco/promptui"
	"github.com/urfave/cli/v2"
)

//go:generate ../scripts/genversion.bash

// VervetParams contains configuration parameters for the Vervet CLI application.
type VervetParams struct {
	Stdin  io.ReadCloser
	Stdout io.WriteCloser
	Stderr io.WriteCloser
	Prompt VervetPrompt
}

// VervetApp contains the cli Application.
type VervetApp struct {
	App    *cli.App
	Params VervetParams
}

// VervetPrompt defines the interface for interactive prompts in vervet.
type VervetPrompt interface {
	Confirm(label string) (bool, error)                  // Confirm y/n an action
	Entry(label string) (string, error)                  // Gather a freeform entry in response to a question
	Select(label string, items []string) (string, error) // Select from a limited number of entries
}

type runKey string

var vervetKey = runKey("vervet")

func contextWithApp(ctx context.Context, v *VervetApp) context.Context {
	return context.WithValue(ctx, vervetKey, v)
}

// Run runs the cli.App with the Vervet config params.
func (v *VervetApp) Run(args []string) error {
	ctx := contextWithApp(context.Background(), v)
	return v.App.RunContext(ctx, args)
}

// NewApp returns a new VervetApp with the provided params.
func NewApp(vp VervetParams) *VervetApp {
	return &VervetApp{
		App: &cli.App{
			Name:      "vervet",
			Usage:     "OpenAPI resource versioning tool",
			Reader:    vp.Stdin,
			Writer:    vp.Stdout,
			ErrWriter: vp.Stderr,
			Version:   "develop", // Set in init created with go generate.
			Flags: []cli.Flag{
				&cli.BoolFlag{
					Name:  "debug",
					Usage: "Turn on debug logging",
				},
			},
			Commands: []*cli.Command{{
				Name:      "resolve",
				Usage:     "Aggregate, render and validate resource specs at a particular version",
				ArgsUsage: "[resource root]",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "at"},
				},
				Action: Resolve,
			}, {
				Name:      "generate",
				Usage:     "Generate artifacts from resource versioned OpenAPI specs",
				ArgsUsage: "<generator> [<generator2>...]",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "config",
						Aliases: []string{"c", "conf"},
						Usage:   "Project configuration file",
					},
					&cli.StringFlag{
						Name:    "generators",
						Aliases: []string{"g", "gen", "generator"},
						Usage:   "Generators definition file",
					},
				},
				Action: Generate,
			}, {
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
				Name:    "resource",
				Aliases: []string{"rc"},
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "config",
						Aliases: []string{"c", "conf"},
						Usage:   "Project configuration file",
					},
				},
				Subcommands: []*cli.Command{{
					Name:      "files",
					Usage:     "List resource spec files in a vervet project",
					ArgsUsage: "[api [resource]]",
					Action:    ResourceFiles,
				}, {
					Name:      "operations",
					Aliases:   []string{"op", "ops"},
					Usage:     "List versioned resource operations in a vervet project",
					ArgsUsage: "[api [resource]]",
					Action:    ResourceOperations,
				}},
			}},
		},
		Params: vp,
	}
}

// Prompt is the default interactive prompt for vervet.
type Prompt struct{}

// Confirm implements VervetPrompt.Confirm
func (p Prompt) Confirm(label string) (bool, error) {
	prompt := promptui.Prompt{
		Label:   fmt.Sprintf("%v (y/N)?", label),
		Default: "N",
		Validate: func(input string) error {
			input = strings.ToLower(input)
			if input != "n" && input != "y" {
				return errors.New("you must pick y or n")
			}
			return nil
		},
	}
	result, err := prompt.Run()
	if err != nil {
		return false, err
	}
	return (result == "y"), nil
}

// Entry implements VervetPrompt.Entry
func (p Prompt) Entry(label string) (string, error) {
	prompt := promptui.Prompt{
		Label: label,
		Validate: func(result string) error {
			if result == "" {
				return errors.New("you must provide a non-empty response")
			}
			return nil
		},
	}
	result, err := prompt.Run()
	if err != nil {
		return "", err
	}
	return result, nil
}

// Select implements VervetPrompt.Select
func (p Prompt) Select(label string, items []string) (string, error) {
	prompt := promptui.Select{
		Label: label,
		Items: items,
	}
	_, result, err := prompt.Run()
	if err != nil {
		return "", err
	}
	return result, nil
}

// Vervet is the vervet application with the CLI application.
var Vervet = NewApp(VervetParams{
	Stdin:  os.Stdin,
	Stdout: os.Stdout,
	Stderr: os.Stderr,
	Prompt: Prompt{},
})

func absPath(path string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("path missing or empty")
	}
	return filepath.Abs(path)
}

func projectConfig(ctx *cli.Context) (string, string, error) {
	var projectDir, configFile string
	var err error
	if cf := ctx.String("config"); cf != "" {
		configFile, err = filepath.Abs(cf)
		if err != nil {
			return "", "", err
		}
		projectDir = filepath.Dir(configFile)
	} else {
		configFile = filepath.Join(projectDir, ".vervet.yaml")
		projectDir, err = os.Getwd()
		if err != nil {
			return "", "", err
		}
	}
	return projectDir, configFile, nil
}
