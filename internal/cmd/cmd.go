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

// MANAGED BY scripts/genversion.bash DO NOT EDIT.
const cmdVersion = "v8.8.0"

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
func NewApp(app *cli.App, vp VervetParams) *VervetApp {
	app.Reader = vp.Stdin
	app.Writer = vp.Stdout
	app.ErrWriter = vp.Stderr
	return &VervetApp{
		App:    app,
		Params: vp,
	}
}

var CLIApp = cli.App{
	Name:    "vervet",
	Usage:   "OpenAPI resource versioning tool",
	Version: cmdVersion,
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:  "debug",
			Usage: "Turn on debug logging",
		},
	},
	Commands: []*cli.Command{
		&BackstageCommand,
		&BuildCommand,
		&RetroBuildCommand,
		&SimpleBuildCommand,
		&FilterCommand,
		&GenerateCommand,
		&LocalizeCommand,
		&ResourceCommand,
		&ResolveCommand,
	},
}

// Prompt is the default interactive prompt for vervet.
type Prompt struct{}

// Confirm implements VervetPrompt.Confirm.
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

// Entry implements VervetPrompt.Entry.
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

// Select implements VervetPrompt.Select.
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
var Vervet = NewApp(&CLIApp, VervetParams{
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
