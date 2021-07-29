// Package cmd provides subcommands for the vervet CLI.
package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/ghodss/yaml"
	"github.com/snyk/vervet"
	"github.com/urfave/cli/v2"
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
				Name:    "include",
				Aliases: []string{"I"},
				Usage:   "OpenAPI specification to include in all compiled versions",
			},
		},
		Action: Compile,
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
	if ctx.Args().Len() < 1 {
		return fmt.Errorf("missing resources root")
	}

	var includeSpec *openapi3.T
	var err error
	if includePath := ctx.String("include"); includePath != "" {
		includeSpec, err = vervet.LoadSpecFile(includePath)
		if err != nil {
			return fmt.Errorf("failed to load included spec %q: %w", includePath, err)
		}
		err = vervet.Localize(includeSpec)
		if err != nil {
			return fmt.Errorf("failed to localize included spec %q: %w", includePath, err)
		}
		// This marshal/unmarshal is needed to avoid local filesystem
		// references from re-appearing in the merge below.
		// TODO: Find out why, improve vervet.Localize.
		buf, err := vervet.ToSpecJSON(includeSpec)
		if err != nil {
			return err
		}
		includeSpec, err = openapi3.NewLoader().LoadFromData(buf)
		if err != nil {
			return err
		}
	}

	inputDir, err := absPath(ctx.Args().Get(0))
	if err != nil {
		return err
	}
	specVersions, err := vervet.LoadSpecVersions(inputDir)
	if err != nil {
		return err
	}
	versions := specVersions.Versions()
	outputDir, err := absPath(ctx.Args().Get(1))
	if err != nil {
		return err
	}
	versionDates := vervet.VersionDateStrings(versions)
	stabilities := []string{"~experimental", "~beta", ""}
	for _, versionDate := range versionDates {
		for _, stabilitySuffix := range stabilities {
			version, err := vervet.ParseVersion(versionDate + stabilitySuffix)
			if err != nil {
				return err
			}
			versionDir := outputDir + "/" + version.String()
			err = os.MkdirAll(versionDir, 0755)
			if err != nil {
				return err
			}
			spec, err := specVersions.At(version.String())
			if err == vervet.ErrNoMatchingVersion {
				continue
			} else if err != nil {
				return err
			}
			if includeSpec != nil {
				vervet.MergeSpec(spec, includeSpec)
			}
			jsonBuf, err := vervet.ToSpecJSON(spec)
			if err != nil {
				return fmt.Errorf("failed to convert to JSON: %w", err)
			}
			err = ioutil.WriteFile(versionDir+"/spec.json", jsonBuf, 0644)
			if err != nil {
				return err
			}
			yamlBuf, err := yaml.JSONToYAML(jsonBuf)
			if err != nil {
				return fmt.Errorf("failed to convert to YAML: %w", err)
			}
			yamlBuf, err = vervet.WithGeneratedComment(yamlBuf)
			err = ioutil.WriteFile(versionDir+"/spec.yaml", yamlBuf, 0644)
			if err != nil {
				return err
			}
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
	t, err := vervet.LoadSpecFile(specFile)
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
