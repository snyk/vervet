package cmd

import (
	"fmt"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/urfave/cli/v2"

	"github.com/snyk/vervet/v6"
)

// FilterCommand is the `vervet filter` subcommand.
var FilterCommand = cli.Command{
	Name:      "filter",
	Usage:     "Filter an OpenAPI document",
	ArgsUsage: "[spec.yaml file]",
	Flags: []cli.Flag{
		&cli.StringSliceFlag{Name: "include-paths", Aliases: []string{"I"}},
		&cli.StringSliceFlag{Name: "exclude-paths", Aliases: []string{"X"}},
	},
	Action: Filter,
}

// Filter an OpenAPI spec file.
func Filter(ctx *cli.Context) error {
	if ctx.Args().Len() < 1 {
		return fmt.Errorf("missing spec.yaml file")
	}
	specFile, err := absPath(ctx.Args().Get(0))
	if err != nil {
		return fmt.Errorf("failed to resolve %q", ctx.Args().Get(0))
	}
	doc, err := vervet.NewDocumentFile(specFile)
	if err != nil {
		return fmt.Errorf("failed to load spec from %q: %v", specFile, err)
	}

	// Localize all references, so we emit a completely self-contained OpenAPI document.
	err = vervet.Localize(ctx.Context, doc)
	if err != nil {
		return fmt.Errorf("failed to localize refs: %w", err)
	}

	if excludePaths := ctx.StringSlice("exclude-paths"); len(excludePaths) > 0 {
		for _, excludePath := range excludePaths {
			delete(doc.Paths, excludePath)
		}
	}
	if includePaths := ctx.StringSlice("include-paths"); len(includePaths) > 0 {
		newPaths := openapi3.Paths{}
		for _, includePath := range includePaths {
			if pathInfo, ok := doc.Paths[includePath]; ok {
				newPaths[includePath] = pathInfo
			}
		}
		doc.Paths = newPaths
	}

	err = removeOrphanedComponents(doc.T)
	if err != nil {
		return fmt.Errorf("failed to remove orphaned components: %W", err)
	}

	yamlBuf, err := vervet.ToSpecYAML(doc)
	if err != nil {
		return fmt.Errorf("failed to convert JSON to YAML: %w", err)
	}
	fmt.Println(string(yamlBuf))

	err = doc.Validate(ctx.Context)
	if err != nil {
		return fmt.Errorf("error: spec validation failed: %w", err)
	}
	return nil
}

// TODO: refactor to reduce cyclomatic complexity.
func removeOrphanedComponents(t *openapi3.T) error { //nolint:gocyclo // acked
	ix, err := vervet.NewRefIndex(t)
	if err != nil {
		return err
	}
	if t.Components.Schemas != nil {
		var remove []string
		for key := range t.Components.Schemas {
			if !ix.HasRef("#/components/schemas/" + key) {
				remove = append(remove, key)
			}
		}
		for i := range remove {
			delete(t.Components.Schemas, remove[i])
		}
	}
	if t.Components.Parameters != nil {
		var remove []string
		for key := range t.Components.Parameters {
			if !ix.HasRef("#/components/parameters/" + key) {
				remove = append(remove, key)
			}
		}
		for i := range remove {
			delete(t.Components.Parameters, remove[i])
		}
	}
	if t.Components.Headers != nil {
		var remove []string
		for key := range t.Components.Headers {
			if !ix.HasRef("#/components/headers/" + key) {
				remove = append(remove, key)
			}
		}
		for i := range remove {
			delete(t.Components.Headers, remove[i])
		}
	}
	if t.Components.RequestBodies != nil {
		var remove []string
		for key := range t.Components.RequestBodies {
			if !ix.HasRef("#/components/requestbodies/" + key) {
				remove = append(remove, key)
			}
		}
		for i := range remove {
			delete(t.Components.RequestBodies, remove[i])
		}
	}
	if t.Components.Responses != nil {
		var remove []string
		for key := range t.Components.Responses {
			if !ix.HasRef("#/components/responses/" + key) {
				remove = append(remove, key)
			}
		}
		for i := range remove {
			delete(t.Components.Responses, remove[i])
		}
	}
	if t.Components.SecuritySchemes != nil {
		var remove []string
		for key := range t.Components.SecuritySchemes {
			if !ix.HasRef("#/components/securityschemes/" + key) {
				remove = append(remove, key)
			}
		}
		for i := range remove {
			delete(t.Components.SecuritySchemes, remove[i])
		}
	}
	if t.Components.Examples != nil {
		var remove []string
		for key := range t.Components.Examples {
			if !ix.HasRef("#/components/examples/" + key) {
				remove = append(remove, key)
			}
		}
		for i := range remove {
			delete(t.Components.Examples, remove[i])
		}
	}
	if t.Components.Links != nil {
		var remove []string
		for key := range t.Components.Links {
			if !ix.HasRef("#/components/links/" + key) {
				remove = append(remove, key)
			}
		}
		for i := range remove {
			delete(t.Components.Links, remove[i])
		}
	}
	if t.Components.Callbacks != nil {
		var remove []string
		for key := range t.Components.Callbacks {
			if !ix.HasRef("#/components/callbacks/" + key) {
				remove = append(remove, key)
			}
		}
		for i := range remove {
			delete(t.Components.Callbacks, remove[i])
		}
	}
	return nil
}
