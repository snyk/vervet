package compiler

import (
	"context"
	"fmt"
	"html/template"
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/ghodss/yaml"
	"go.uber.org/multierr"

	"github.com/snyk/vervet/v5"
	"github.com/snyk/vervet/v5/config"
	"github.com/snyk/vervet/v5/internal/files"
	"github.com/snyk/vervet/v5/internal/linter"
	"github.com/snyk/vervet/v5/internal/linter/optic"
	"github.com/snyk/vervet/v5/internal/linter/spectral"
)

// A Compiler checks and builds versioned API resource inputs into aggregated
// OpenAPI versioned outputs, as determined by an API project configuration.
type Compiler struct {
	apis    map[string]*api
	linters map[string]linter.Linter

	newLinter func(ctx context.Context, lc *config.Linter) (linter.Linter, error)
}

// CompilerOption applies a configuration option to a Compiler.
type CompilerOption func(*Compiler) error

// LinterFactory configures a Compiler to use a custom factory function for
// instantiating Linters.
func LinterFactory(f func(ctx context.Context, lc *config.Linter) (linter.Linter, error)) CompilerOption {
	return func(c *Compiler) error {
		c.newLinter = f
		return nil
	}
}

func defaultLinterFactory(ctx context.Context, lc *config.Linter) (linter.Linter, error) {
	if lc.Spectral != nil {
		return spectral.New(ctx, lc.Spectral)
	} else if lc.SweaterComb != nil {
		return optic.New(ctx, lc.SweaterComb)
	} else if lc.OpticCI != nil {
		return optic.New(ctx, lc.OpticCI)
	}
	return nil, fmt.Errorf("invalid linter (linters.%s)", lc.Name)
}

type api struct {
	resources       []*resourceSet
	overlayIncludes []*vervet.Document
	overlayInlines  []*openapi3.T
	output          *output
}

type resourceSet struct {
	path            string
	linter          linter.Linter
	linterOverrides map[string]map[string]config.Linter
	sourceFiles     []string
	lintFiles       []string
}

type output struct {
	paths  []string
	linter linter.Linter
}

// New returns a new Compiler for a given project configuration.
func New(ctx context.Context, proj *config.Project, lint bool, options ...CompilerOption) (*Compiler, error) {
	compiler := &Compiler{
		apis:      map[string]*api{},
		linters:   map[string]linter.Linter{},
		newLinter: defaultLinterFactory,
	}

	for i := range options {
		err := options[i](compiler)
		if err != nil {
			return nil, err
		}
	}

	if lint {
		// set up linters
		for linterName, linterConfig := range proj.Linters {
			linter, err := compiler.newLinter(ctx, linterConfig)
			if err != nil {
				return nil, fmt.Errorf("%w (linters.%s)", err, linterName)
			}
			compiler.linters[linterName] = linter
		}
	}

	// set up APIs
	for apiName, apiConfig := range proj.APIs {
		a := api{}

		// Build resources
		for rcIndex, rcConfig := range apiConfig.Resources {
			var err error
			r := &resourceSet{
				path:            rcConfig.Path,
				linter:          compiler.linters[rcConfig.Linter],
				linterOverrides: map[string]map[string]config.Linter{},
			}
			if lint && r.linter != nil {
				r.lintFiles, err = r.linter.Match(rcConfig)
				if err != nil {
					return nil, fmt.Errorf("%w: (apis.%s.resources[%d].path)", err, apiName, rcIndex)
				}
				// TODO: overrides are deprecated by Optic CI, remove soon
				linterOverrides := map[string]map[string]config.Linter{}
				for rcName, versionMap := range rcConfig.LinterOverrides {
					linterOverrides[rcName] = map[string]config.Linter{}
					for version, linter := range versionMap {
						linterOverrides[rcName][version] = *linter
					}
				}
				r.linterOverrides = linterOverrides
			}
			r.sourceFiles, err = ResourceSpecFiles(rcConfig)
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
			} else if overlayConfig.Inline != "" {
				docString := os.ExpandEnv(overlayConfig.Inline)
				l := openapi3.NewLoader()
				doc, err := l.LoadFromData([]byte(docString))
				if err != nil {
					return nil, fmt.Errorf("failed to load template: %w (apis.%s.overlays[%d].template)",
						err, apiName, overlayIndex)
				}
				a.overlayInlines = append(a.overlayInlines, doc)
			}
		}

		// Build output
		if apiConfig.Output != nil {
			paths := apiConfig.Output.Paths
			if len(paths) == 0 && apiConfig.Output.Path != "" {
				paths = []string{apiConfig.Output.Path}
			}
			if len(paths) > 0 {
				a.output = &output{
					paths:  paths,
					linter: compiler.linters[apiConfig.Output.Linter],
				}
			}
		}

		compiler.apis[apiName] = &a
	}
	return compiler, nil
}

// ResourceSpecFiles returns all matching spec files for a config.Resource.
func ResourceSpecFiles(rcConfig *config.ResourceSet) ([]string, error) {
	return files.LocalFSSource{}.Match(rcConfig)
}

// LintResources checks the inputs of an API's resources with the configured linter.
func (c *Compiler) LintResources(ctx context.Context, apiName string) error {
	api, ok := c.apis[apiName]
	if !ok {
		return fmt.Errorf("api not found (apis.%s)", apiName)
	}
	var errs error
	for rcIndex, rc := range api.resources {
		if rc.linter == nil {
			continue
		}
		if len(rc.linterOverrides) > 0 {
			err := c.lintWithOverrides(ctx, rc, apiName, rcIndex)
			if err != nil {
				errs = multierr.Append(errs, fmt.Errorf("%w (apis.%s.resources[%d])", err, apiName, rcIndex))
			}
		} else {
			err := rc.linter.Run(ctx, rc.path, rc.lintFiles...)
			if err != nil {
				errs = multierr.Append(errs, fmt.Errorf("%w (apis.%s.resources[%d])", err, apiName, rcIndex))
			}
		}
	}
	return errs
}

func (c *Compiler) lintWithOverrides(ctx context.Context, rc *resourceSet, apiName string, rcIndex int) error {
	var pending []string
	for _, matchedFile := range rc.lintFiles {
		versionDir := filepath.Dir(matchedFile)
		rcDir := filepath.Dir(versionDir)
		versionName := filepath.Base(versionDir)
		rcName := filepath.Base(rcDir)
		if linter, ok := rc.linterOverrides[rcName][versionName]; ok {
			linter, err := rc.linter.WithOverride(ctx, &linter)
			if err != nil {
				return fmt.Errorf("failed to apply overrides to linter: %w (apis.%s.resources[%d].linter-overrides.%s.%s)",
					err, apiName, rcIndex, rcName, versionName)
			}
			err = linter.Run(ctx, matchedFile)
			if err != nil {
				return fmt.Errorf("lint failed on %q: %w (apis.%s.resources[%d])", matchedFile, err, apiName, rcIndex)
			}
		} else {
			pending = append(pending, matchedFile)
		}
	}
	if len(pending) == 0 {
		return nil
	}
	err := rc.linter.Run(ctx, rc.path, pending...)
	if err != nil {
		return fmt.Errorf("lint failed (apis.%s.resources[%d])", apiName, rcIndex)
	}
	return nil
}

func (c *Compiler) apisEach(ctx context.Context, f func(ctx context.Context, apiName string) error) error {
	var errs error
	for apiName := range c.apis {
		err := f(ctx, apiName)
		if err != nil {
			errs = multierr.Append(errs, err)
		}
	}
	return errs
}

// Build builds an aggregate versioned OpenAPI spec for a specific API by name
// in the project.
func (c *Compiler) Build(ctx context.Context, apiName string) error {
	api, ok := c.apis[apiName]
	if !ok {
		return fmt.Errorf("api not found (apis.%s)", apiName)
	}
	if api.output == nil || len(api.output.paths) == 0 {
		return nil
	}
	for _, path := range api.output.paths {
		err := os.RemoveAll(path)
		if err != nil {
			return fmt.Errorf("failed to clear output directory: %w", err)
		}
	}
	err := os.MkdirAll(api.output.paths[0], 0777)
	if err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}
	log.Printf("compiling API %s to output versions", apiName)
	var versionSpecFiles []string

	specMap := make(map[vervet.Version]*openapi3.T)

	for rcIndex, rc := range api.resources {
		specVersions, err := vervet.LoadSpecVersionsFileset(rc.sourceFiles)
		if err != nil {
			return fmt.Errorf("failed to load spec versions: %+v (apis.%s.resources[%d])",
				err, apiName, rcIndex)
		}
		buildErr := func(err error) error {
			return fmt.Errorf("%w (apis.%s.resources[%d])", err, apiName, rcIndex)
		}
		versions := specVersions.Versions()
		for _, version := range versions {
			spec, err := specVersions.At(version)
			if err == vervet.ErrNoMatchingVersion {
				continue
			} else if err != nil {
				return buildErr(err)
			}
			if specMap[version] == nil {
				specMap[version] = &openapi3.T{}
			}
			vervet.Merge(specMap[version], spec, false)
		}
	}
	for version, spec := range specMap {
		buildErr := func(err error) error {
			return fmt.Errorf("%w (apis.%s.resources)", err, apiName)
		}

		// Create the directories, but only if a spec file exists for it.
		versionDir := api.output.paths[0] + "/" + version.String()

		err = os.MkdirAll(versionDir, 0755)
		if err != nil {
			return buildErr(err)
		}

		// Merge all overlays
		for _, doc := range api.overlayIncludes {
			vervet.Merge(spec, doc.T, true)
		}
		for _, doc := range api.overlayInlines {
			vervet.Merge(spec, doc, true)
		}

		// Write the compiled spec to JSON and YAML
		jsonBuf, err := vervet.ToSpecJSON(spec)
		if err != nil {
			return buildErr(err)
		}
		jsonSpecPath := versionDir + "/spec.json"
		jsonEmbedPath, err := filepath.Rel(api.output.paths[0], jsonSpecPath)
		if err != nil {
			return buildErr(err)
		}
		versionSpecFiles = append(versionSpecFiles, jsonEmbedPath)
		err = ioutil.WriteFile(jsonSpecPath, jsonBuf, 0644)
		if err != nil {
			return buildErr(err)
		}
		log.Println(jsonSpecPath)
		yamlBuf, err := yaml.JSONToYAML(jsonBuf)
		if err != nil {
			return buildErr(err)
		}
		yamlBuf, err = vervet.WithGeneratedComment(yamlBuf)
		if err != nil {
			return buildErr(err)
		}
		yamlSpecPath := versionDir + "/spec.yaml"
		yamlEmbedPath, err := filepath.Rel(api.output.paths[0], yamlSpecPath)
		if err != nil {
			return buildErr(err)
		}
		versionSpecFiles = append(versionSpecFiles, yamlEmbedPath)
		err = ioutil.WriteFile(yamlSpecPath, yamlBuf, 0644)
		if err != nil {
			return buildErr(err)
		}
		log.Println(yamlSpecPath)
	}
	err = c.writeEmbedGo(filepath.Base(api.output.paths[0]), api, versionSpecFiles)
	if err != nil {
		return fmt.Errorf("failed to create embed.go: %w", err)
	}
	// Copy output to multiple paths if specified
	src := api.output.paths[0]
	for _, dst := range api.output.paths[1:] {
		if err := files.CopyDir(dst, src, true); err != nil {
			return fmt.Errorf("failed to copy %q to %q: %w", src, dst, err)
		}
	}
	return nil
}

func (c *Compiler) writeEmbedGo(pkgName string, a *api, versionSpecFiles []string) error {
	embedPath := filepath.Join(a.output.paths[0], "embed.go")
	f, err := os.Create(embedPath)
	if err != nil {
		return err
	}
	defer f.Close()
	err = embedGoTmpl.Execute(f, struct {
		Package          string
		API              *api
		VersionSpecFiles []string
	}{
		Package:          pkgName,
		API:              a,
		VersionSpecFiles: versionSpecFiles,
	})
	if err != nil {
		return err
	}
	for _, dst := range a.output.paths[1:] {
		if err := files.CopyFile(filepath.Join(dst, "embed.go"), embedPath, true); err != nil {
			return err
		}
	}
	return nil
}

var embedGoTmpl = template.Must(template.New("embed.go").Parse(`
package {{ .Package }}

import "embed"

// Embed compiled OpenAPI specs in Go projects.

{{ range .VersionSpecFiles -}}
//go:embed {{ . }}
{{ end -}}
// Versions contains OpenAPI specs for each distinct release version.
var Versions embed.FS
`[1:]))

// BuildAll builds all APIs in the project.
func (c *Compiler) BuildAll(ctx context.Context) error {
	if err := c.apisEach(ctx, c.LintResources); err != nil {
		return err
	}
	if err := c.apisEach(ctx, c.Build); err != nil {
		return err
	}
	return c.apisEach(ctx, c.LintOutput)
}

// LintOutput applies configured linting rules to the build output.
func (c *Compiler) LintOutput(ctx context.Context, apiName string) error {
	api, ok := c.apis[apiName]
	if !ok {
		return fmt.Errorf("api not found (apis.%s)", apiName)
	}
	if api.output != nil && len(api.output.paths) > 0 && api.output.linter != nil {
		var outputFiles []string
		err := doublestar.GlobWalk(os.DirFS(api.output.paths[0]), "**/spec.{json,yaml}",
			func(path string, d fs.DirEntry) error {
				outputFiles = append(outputFiles, filepath.Join(api.output.paths[0], path))
				return nil
			})
		if err != nil {
			return fmt.Errorf("failed to match output files for linting: %w (apis.%s.output)",
				err, apiName)
		}
		if len(outputFiles) == 0 {
			return fmt.Errorf("lint failed: no output files were produced")
		}
		err = api.output.linter.Run(ctx, api.output.paths[0], outputFiles...)
		if err != nil {
			return fmt.Errorf("lint failed (apis.%s.output)", apiName)
		}
	}
	return nil
}
