package cmd

import (
	"fmt"

	"github.com/urfave/cli/v2"

	"github.com/snyk/vervet/v8"
	"github.com/snyk/vervet/v8/config"
	"github.com/snyk/vervet/v8/internal/compiler"
	"github.com/snyk/vervet/v8/internal/simplebuild"
)

var defaultVersioningUrl = "https://api.snyk.io/rest/openapi"

var pivotDateCLIFlagName = "pivot-version"
var versioningUrlCLIFlagName = "versioning-url"

var buildFlags = []cli.Flag{
	&cli.StringFlag{
		Name:    "config",
		Aliases: []string{"c", "conf"},
		Usage:   "Project configuration file",
	},
	&cli.StringFlag{
		Name:    "include",
		Aliases: []string{"I"},
		Usage:   "OpenAPI specification to include in build output",
	},
	&cli.StringFlag{
		Name:    pivotDateCLIFlagName,
		Aliases: []string{"P"},
		Usage: fmt.Sprintf(
			"Pivot version after which new strategy versioning is used."+
				" Flag for testing only, recommend to use the default date(%s)", vervet.DefaultPivotDate.String()),
		Value: vervet.DefaultPivotDate.String(),
	},
	&cli.StringFlag{
		Name:    versioningUrlCLIFlagName,
		Aliases: []string{"U"},
		Usage:   fmt.Sprintf("URL to fetch versioning information. Default is %q", defaultVersioningUrl),
		Value:   defaultVersioningUrl,
	},
}

// BuildCommand is the `vervet build` subcommand.
var BuildCommand = cli.Command{
	Name:      "build",
	Usage:     "Build versioned resources into versioned OpenAPI specs",
	ArgsUsage: "[input resources root] [output api root]",
	Flags:     buildFlags,
	Action:    CombinedBuild,
}

// RetroBuild is the `vervet build` subcommand.
var RetroBuildCommand = cli.Command{
	Name:      "retrobuild",
	Usage:     "Build versioned resources into versioned OpenAPI specs",
	ArgsUsage: "[input resources root] [output api root]",
	Flags:     buildFlags,
	Action:    RetroBuild,
}

var SimpleBuildCommand = cli.Command{
	Name:      "simplebuild",
	Usage:     "Build versioned resources into versioned OpenAPI specs",
	ArgsUsage: "[input resources root]",
	Flags:     buildFlags,
	Action:    SimpleBuild,
}

// SimpleBuild compiles versioned resources into versioned API specs using the rolled up versioning strategy.
func SimpleBuild(ctx *cli.Context) error {
	project, err := projectFromContext(ctx)
	if err != nil {
		return err
	}
	pivotDate, err := parsePivotDate(ctx)
	if err != nil {
		return fmt.Errorf("failed to parse pivot date %q: %w", pivotDate, err)
	}

	versioningURL := ctx.String(versioningUrlCLIFlagName)

	err = simplebuild.Build(ctx.Context, project, pivotDate, versioningURL, false)
	return err
}

// CombinedBuild compiles versioned resources into versioned API specs
// invokes retorbuild and simplebuild based on the context.
func CombinedBuild(ctx *cli.Context) error {
	project, err := projectFromContext(ctx)
	if err != nil {
		return err
	}
	pivotDate, err := parsePivotDate(ctx)
	if err != nil {
		return fmt.Errorf("failed to parse pivot date %q: %w", pivotDate, err)
	}

	versioningURL := ctx.String(versioningUrlCLIFlagName)

	comp, err := compiler.New(ctx.Context, project)
	if err != nil {
		return err
	}
	err = comp.BuildAll(ctx.Context, pivotDate)
	if err != nil {
		return err
	}

	return simplebuild.Build(ctx.Context, project, pivotDate, versioningURL, true)
}

func parsePivotDate(ctx *cli.Context) (vervet.Version, error) {
	return vervet.ParseVersion(ctx.String(pivotDateCLIFlagName))
}

// RetroBuild compiles versioned resources into versioned API specs using the older versioning strategy.
// This is used for regenerating old versioned API specs only.
func RetroBuild(ctx *cli.Context) error {
	project, err := projectFromContext(ctx)
	if err != nil {
		return err
	}
	pivotDate, err := parsePivotDate(ctx)
	if err != nil {
		return fmt.Errorf("failed to parse pivot date %q: %w", pivotDate, err)
	}
	comp, err := compiler.New(ctx.Context, project)
	if err != nil {
		return err
	}
	return comp.BuildAll(ctx.Context, pivotDate)
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
		return config.FromFile(configPath)
	}

	api := &config.API{
		Resources: []*config.ResourceSet{{
			Path: ctx.Args().Get(0),
		}},
		Output: &config.Output{
			Paths: []string{ctx.Args().Get(1)},
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
	return project, nil
}
