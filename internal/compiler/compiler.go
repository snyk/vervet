package compiler

import (
	"context"
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/ghodss/yaml"

	"github.com/snyk/vervet"
	"github.com/snyk/vervet/config"
	"github.com/snyk/vervet/internal/spectral"
)

// A Linter checks that a set of files conform to some set of rules and
// standards.
type Linter interface {
	Run(ctx context.Context, files ...string) error
}

// A Compiler checks and builds versioned API resource inputs into aggregated
// OpenAPI versioned outputs, as determined by an API project configuration.
type Compiler struct {
	apis    map[string]*api
	linters map[string]Linter

	newLinter func(ctx context.Context, lc *config.Linter) (Linter, error)
}

// CompilerOption applies a configuration option to a Compiler.
type CompilerOption func(*Compiler) error

// LinterFactory configures a Compiler to use a custom factory function for
// instantiating Linters.
func LinterFactory(f func(ctx context.Context, lc *config.Linter) (Linter, error)) CompilerOption {
	return func(c *Compiler) error {
		c.newLinter = f
		return nil
	}
}

func defaultLinterFactory(ctx context.Context, lc *config.Linter) (Linter, error) {
	if lc.Spectral == nil {
		return nil, fmt.Errorf("unsupported linter (linters.%s)", lc.Name)
	}
	// This can be a linter variant dispatch off non-nil if/when more linter
	// types are supported.
	return spectral.New(ctx, lc.Spectral.Rules)
}

type api struct {
	resources        []*resource
	overlayIncludes  []*vervet.Document
	overlayTemplates []*openapi3.T
	output           *output
}

type resource struct {
	linter       Linter
	matchedFiles []string
}

type output struct {
	path   string
	linter Linter
}

// New returns a new Compiler for a given project configuration.
func New(ctx context.Context, proj *config.Project, options ...CompilerOption) (*Compiler, error) {
	compiler := &Compiler{
		apis:      map[string]*api{},
		linters:   map[string]Linter{},
		newLinter: defaultLinterFactory,
	}
	for i := range options {
		err := options[i](compiler)
		if err != nil {
			return nil, err
		}
	}
	// set up linters
	for linterName, linterConfig := range proj.Linters {
		linter, err := compiler.newLinter(ctx, linterConfig)
		if err != nil {
			return nil, fmt.Errorf("%w (linters.%s)", err, linterName)
		}
		compiler.linters[linterName] = linter
	}
	// set up APIs
	for apiName, apiConfig := range proj.APIs {
		a := api{}

		// Build resources
		for rcIndex, rcConfig := range apiConfig.Resources {
			excludes := rcConfig.Excludes
			r := &resource{
				linter: compiler.linters[rcConfig.Linter],
			}
			err := doublestar.GlobWalk(os.DirFS(rcConfig.Path),
				vervet.SpecGlobPattern,
				func(path string, d fs.DirEntry) error {
					rcPath := filepath.Join(rcConfig.Path, path)
					for i := range excludes {
						if ok, err := doublestar.Match(excludes[i], rcPath); ok {
							return nil
						} else if err != nil {
							// Shouldn't happen; pattern is validated on config.Load
							panic(err)
						}
					}
					r.matchedFiles = append(r.matchedFiles, rcPath)
					return nil
				})
			if err != nil {
				return nil, fmt.Errorf("%w: (apis.%s.resources[%d].path)", err, apiName, rcIndex)
			}
			a.resources = append(a.resources, r)
		}

		// Build overlays
		for overlayIndex, overlayConfig := range apiConfig.Overlays {
			if overlayConfig.Include != "" {
				doc, err := vervet.NewDocumentFile(overlayConfig.Include)
				if err != nil {
					return nil, fmt.Errorf("failed to load overlay %q: %w (apis.%s.overlays[%d])",
						overlayConfig.Include, err, apiName, overlayIndex)
				}
				err = vervet.Localize(doc)
				if err != nil {
					return nil, fmt.Errorf("failed to localize references in %q: %w (apis.%s.overlays[%d]",
						overlayConfig.Include, err, apiName, overlayIndex)
				}
				a.overlayIncludes = append(a.overlayIncludes, doc)
			} else if overlayConfig.Template != "" {
				docString := os.ExpandEnv(overlayConfig.Template)
				l := openapi3.NewLoader()
				doc, err := l.LoadFromData([]byte(docString))
				if err != nil {
					return nil, fmt.Errorf("failed to load template: %w (apis.%s.overlays[%d].template)",
						err, apiName, overlayIndex)
				}
				a.overlayTemplates = append(a.overlayTemplates, doc)
			}
		}

		// Build output
		if apiConfig.Output != nil && apiConfig.Output.Path != "" {
			a.output = &output{
				path:   apiConfig.Output.Path,
				linter: compiler.linters[apiConfig.Output.Linter],
			}
		}

		compiler.apis[apiName] = &a
	}
	return compiler, nil
}

// LintResources checks the inputs of an API's resources with the configured linter.
func (c *Compiler) LintResources(ctx context.Context, apiName string) error {
	api, ok := c.apis[apiName]
	if !ok {
		return fmt.Errorf("api not found (apis.%s)", apiName)
	}
	for rcIndex, rc := range api.resources {
		if rc.linter == nil {
			continue
		}
		err := rc.linter.Run(ctx, rc.matchedFiles...)
		if err != nil {
			return fmt.Errorf("lint failed (apis.%s.resources[%d])", apiName, rcIndex)
		}
	}
	return nil
}

// LintResourcesAll lints resources in all APIs in the project.
func (c *Compiler) LintResourcesAll(ctx context.Context) error {
	return c.apisEach(ctx, c.LintResources)
}

func (c *Compiler) apisEach(ctx context.Context, f func(ctx context.Context, apiName string) error) error {
	for apiName := range c.apis {
		err := f(ctx, apiName)
		if err != nil {
			return err
		}
	}
	return nil
}

// Build builds an aggregate versioned OpenAPI spec for a specific API by name
// in the project.
func (c *Compiler) Build(ctx context.Context, apiName string) error {
	api, ok := c.apis[apiName]
	if !ok {
		return fmt.Errorf("api not found (apis.%s)", apiName)
	}
	if api.output == nil || api.output.path == "" {
		return nil
	}
	for rcIndex, rc := range api.resources {
		specVersions, err := vervet.LoadSpecVersionsFileset(rc.matchedFiles)
		if err != nil {
			return fmt.Errorf("failed to load spec versions: %w (apis.%s.resources[%d])",
				err, apiName, rcIndex)
		}
		buildErr := func(err error) error {
			return fmt.Errorf("%w (apis.%s.resources[%d])", err, apiName, rcIndex)
		}
		versions := specVersions.Versions()
		versionDates := vervet.VersionDateStrings(versions)
		stabilities := []string{"~experimental", "~beta", ""}
		for _, versionDate := range versionDates {
			for _, stabilitySuffix := range stabilities {
				version, err := vervet.ParseVersion(versionDate + stabilitySuffix)
				if err != nil {
					return buildErr(err)
				}
				versionDir := api.output.path + "/" + version.String()
				err = os.MkdirAll(versionDir, 0755)
				if err != nil {
					return buildErr(err)
				}
				spec, err := specVersions.At(version.String())
				if err == vervet.ErrNoMatchingVersion {
					continue
				} else if err != nil {
					return buildErr(err)
				}

				// Merge all overlays
				for _, doc := range api.overlayIncludes {
					vervet.MergeSpec(spec, doc.T)
				}
				for _, doc := range api.overlayTemplates {
					vervet.MergeSpec(spec, doc)
				}

				// Write the compiled spec to JSON and YAML
				jsonBuf, err := vervet.ToSpecJSON(spec)
				if err != nil {
					return buildErr(err)
				}
				err = ioutil.WriteFile(versionDir+"/spec.json", jsonBuf, 0644)
				if err != nil {
					return buildErr(err)
				}
				yamlBuf, err := yaml.JSONToYAML(jsonBuf)
				if err != nil {
					return buildErr(err)
				}
				yamlBuf, err = vervet.WithGeneratedComment(yamlBuf)
				err = ioutil.WriteFile(versionDir+"/spec.yaml", yamlBuf, 0644)
				if err != nil {
					return buildErr(err)
				}
			}
		}
	}
	return nil
}

// BuildAll builds all APIs in the project.
func (c *Compiler) BuildAll(ctx context.Context) error {
	return c.apisEach(ctx, c.Build)
}

// LintOutput applies configured linting rules to the build output.
func (c *Compiler) LintOutput(ctx context.Context, apiName string) error {
	api, ok := c.apis[apiName]
	if !ok {
		return fmt.Errorf("api not found (apis.%s)", apiName)
	}
	if api.output != nil && api.output.linter != nil {
		var outputFiles []string
		err := doublestar.GlobWalk(os.DirFS(api.output.path), "**/spec.{json,yaml}",
			func(path string, d fs.DirEntry) error {
				outputFiles = append(outputFiles, filepath.Join(api.output.path, path))
				return nil
			})
		if err != nil {
			return fmt.Errorf("failed to match output files for linting: %w (apis.%s.output)",
				err, apiName)
		}
		if len(outputFiles) == 0 {
			return fmt.Errorf("lint failed: no output files were produced")
		}
		err = api.output.linter.Run(ctx, outputFiles...)
		if err != nil {
			return fmt.Errorf("lint failed (apis.%s.output)", apiName)
		}
	}
	return nil
}

// LintOutputAll lints output of all APIs in the project.
func (c *Compiler) LintOutputAll(ctx context.Context) error {
	return c.apisEach(ctx, c.LintOutput)
}
