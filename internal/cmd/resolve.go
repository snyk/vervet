package cmd

import (
	"fmt"

	"github.com/urfave/cli/v2"

	"github.com/snyk/vervet/v8"
)

// ResolveCommand is the `vervet resolve` subcommand.
var ResolveCommand = cli.Command{
	Name:      "resolve",
	Usage:     "Aggregate, render and validate resource specs at a particular version",
	ArgsUsage: "[resource root]",
	Flags: []cli.Flag{
		&cli.StringFlag{Name: "at"},
	},
	Action: Resolve,
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
	version, err := vervet.ParseVersion(ctx.String("at"))
	if err != nil {
		return err
	}
	specVersion, err := specVersions.At(version)
	if err != nil {
		return err
	}

	yamlBuf, err := vervet.ToSpecYAML(specVersion)
	if err != nil {
		return fmt.Errorf("failed to convert JSON to YAML: %w", err)
	}
	fmt.Println(string(yamlBuf))

	err = specVersion.Validate(ctx.Context)
	if err != nil {
		return fmt.Errorf("error: spec validation failed: %w", err)
	}
	return nil
}
