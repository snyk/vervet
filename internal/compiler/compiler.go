package compiler

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/ghodss/yaml"
	"go.uber.org/multierr"

	"github.com/snyk/vervet/v7"
	"github.com/snyk/vervet/v7/config"
	"github.com/snyk/vervet/v7/internal/files"
)

// A Compiler checks and builds versioned API resource inputs into aggregated
// OpenAPI versioned outputs, as determined by an API project configuration.
type Compiler struct {
	apis map[string]*api
}

// CompilerOption applies a configuration option to a Compiler.
type CompilerOption func(*Compiler) error

type api struct {
	resources       []*resourceSet
	overlayIncludes []*vervet.Document
	overlayInlines  []*openapi3.T
	output          *output
}

type resourceSet struct {
	path        string
	sourceFiles []string
}

type output struct {
	paths []string
}

// New returns a new Compiler for a given project configuration.
func New(ctx context.Context, proj *config.Project, options ...CompilerOption) (*Compiler, error) {
	compiler := &Compiler{
		apis: map[string]*api{},
	}

	for i := range options {
		err := options[i](compiler)
		if err != nil {
			return nil, err
		}
	}

	// set up APIs
	for apiName, apiConfig := range proj.APIs {
		a := api{}

		// Build resources
		for rcIndex, rcConfig := range apiConfig.Resources {
			var err error
			r := &resourceSet{
				path: rcConfig.Path,
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
				err = vervet.Localize(ctx, doc)
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
					paths: paths,
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
	for rcIndex, rc := range api.resources {
		specVersions, err := vervet.LoadSpecVersionsFileset(rc.sourceFiles) //nolint:contextcheck //acked
		if err != nil {
			return fmt.Errorf("failed to load spec versions: %+v (apis.%s.resources[%d])",
				err, apiName, rcIndex)
		}
		buildErr := func(err error) error {
			return fmt.Errorf("%w (apis.%s.resources[%d])", err, apiName, rcIndex)
		}
		versions := specVersions.Versions()
		for _, version := range versions {
			if version.LifecycleAt(time.Now()) == vervet.LifecycleUnreleased {
				return buildErr(fmt.Errorf(
					"API spec with version %s is in the future. This is not supported as it may cause breakage",
					version,
				))
			}

			spec, err := specVersions.At(version)
			if err == vervet.ErrNoMatchingVersion {
				continue
			} else if err != nil {
				return buildErr(err)
			}

			// Create the directories, but only if a spec file exists for it.
			versionDir := api.output.paths[0] + "/" + version.String()

			if spec != nil {
				err = os.MkdirAll(versionDir, 0755)
				if err != nil {
					return buildErr(err)
				}
			}

			// Merge all overlays
			for _, doc := range api.overlayIncludes {
				err = vervet.Merge(spec, doc.T, true)
				if err != nil {
					return buildErr(err)
				}
			}
			for _, doc := range api.overlayInlines {
				err = vervet.Merge(spec, doc, true)
				if err != nil {
					return buildErr(err)
				}
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
			err = os.WriteFile(jsonSpecPath, jsonBuf, 0644)
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
			err = os.WriteFile(yamlSpecPath, yamlBuf, 0644)
			if err != nil {
				return buildErr(err)
			}
			log.Println(yamlSpecPath)
		}
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
	err = EmbedGoTmpl.Execute(f, struct {
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

var EmbedGoTmpl = template.Must(template.New("embed.go").Parse(`
// Code generated by Vervet. DO NOT EDIT.

package {{ .Package }}

import "embed"

// Embed compiled OpenAPI specs in Go projects.

{{ range .VersionSpecFiles -}}
//go:embed {{ . }}
{{ end }}
// Versions contains OpenAPI specs for each distinct release version.
var Versions embed.FS
`[1:]))

// BuildAll builds all APIs in the project.
func (c *Compiler) BuildAll(ctx context.Context) error {
	return c.apisEach(ctx, c.Build)
}
