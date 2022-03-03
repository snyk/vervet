package cmd

import (
	"fmt"

	"github.com/urfave/cli/v2"

	"github.com/snyk/vervet/v4"
)

// LocalizeCommand is the `vervet localize` subcommand
var LocalizeCommand = cli.Command{
	Name:      "localize",
	Aliases:   []string{"localise"},
	Usage:     "Localize references and validate a single OpenAPI spec file",
	ArgsUsage: "[spec.yaml file]",
	Action:    Localize,
}

// Localize references and validate a single OpenAPI spec file
func Localize(ctx *cli.Context) error {
	if ctx.Args().Len() < 1 {
		return fmt.Errorf("missing spec.yaml file")
	}
	specFile, err := absPath(ctx.Args().Get(0))
	if err != nil {
		return fmt.Errorf("failed to resolve %q", ctx.Args().Get(0))
	}
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
	fmt.Println(string(yamlBuf))

	err = t.Validate(ctx.Context)
	if err != nil {
		return fmt.Errorf("error: spec validation failed: %w", err)
	}
	return nil
}
