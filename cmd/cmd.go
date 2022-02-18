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
	"time"

	"github.com/manifoldco/promptui"
	"github.com/urfave/cli/v2"
)

// VervetVersion contains vervet's version, set by ldflags at build time:
// go build -ldflags "-X 'github.com/snyk/vervet/cmd.VervetVersion=$VERSION'"
var VervetVersion = "0.0.1"

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

func appFromContext(ctx context.Context) (*VervetApp, error) {
	v, ok := ctx.Value(vervetKey).(*VervetApp)
	if !ok {
		return nil, errors.New("could not retrieve vervet app from context")
	}
	return v, nil
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
			Version:   VervetVersion,
			Flags: []cli.Flag{
				&cli.BoolFlag{
					Name:  "debug",
					Usage: "Turn on debug logging to troubleshoot templates",
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
				Name: "scaffold",
				Subcommands: []*cli.Command{{
					Name:      "init",
					Usage:     "Initialize a new project from a scaffold",
					ArgsUsage: "[path to scaffold directory]",
					Flags: []cli.Flag{
						&cli.BoolFlag{
							Name:    "force",
							Aliases: []string{"f", "overwrite"},
							Usage:   "Overwrite existing files",
						},
					},
					Action: ScaffoldInit,
				}},
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
				Name: "version",
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
					Action:    VersionFiles,
				}, {
					Name:      "list",
					Usage:     "List resource versions in a vervet project",
					ArgsUsage: "[api [resource]]",
					Action:    VersionList,
				}, {
					Name:      "new",
					Usage:     "Create a new resource version",
					ArgsUsage: "<api> <resource>",
					Flags: []cli.Flag{
						&cli.BoolFlag{
							Name:    "force",
							Aliases: []string{"f", "overwrite"},
							Usage:   "Overwrite existing files",
						},
						&cli.StringFlag{
							Name:  "version",
							Usage: "Set version date (defaults to today UTC)",
							Value: time.Now().UTC().Format("2006-01-02"),
						},
						&cli.StringFlag{
							Name:  "stability",
							Usage: "Stability level of this version",
							Value: "wip",
						},
					},
					Action: VersionNew,
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
