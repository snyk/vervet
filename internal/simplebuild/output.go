package simplebuild

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/ghodss/yaml"

	"github.com/snyk/vervet/v7"
	"github.com/snyk/vervet/v7/config"
	"github.com/snyk/vervet/v7/internal/compiler"
	"github.com/snyk/vervet/v7/internal/files"
)

// Some services have a need to write specs to multiple destinations. This
// tends to happen in Typescript services in which we want to write specs to
// two places:
//   - src/** for committing into git and ingesting into Backstage
//   - dist/** for runtime module access to compiled specs.
//
// To maintain backwards compatibility we still allow a single path in the
// config file then normalise that here to an array.
func getOutputPaths(cfg config.Output) []string {
	paths := cfg.Paths
	if len(paths) == 0 && cfg.Path != "" {
		paths = []string{cfg.Path}
	}
	return paths
}

// WriteOutputs writes compiled specs to all directories specified by the given
// api config. Removes any existing builds if they are present.
func (docs DocSet) WriteOutputs(cfg config.Output) error {
	paths := getOutputPaths(cfg)

	for _, dir := range paths {
		err := os.RemoveAll(dir)
		if err != nil {
			return fmt.Errorf("clear output directory: %w", err)
		}
	}

	err := docs.Write(paths[0])
	if err != nil {
		return fmt.Errorf("write output files: %w", err)
	}

	for _, dir := range paths[1:] {
		err := files.CopyDir(dir, paths[0], true)
		if err != nil {
			return fmt.Errorf("copy outputs: %w", err)
		}
	}

	return nil
}

// Write writes compiled specs to a single directory in YAML and JSON formats.
// Unlike WriteOutputs this function assumes the destination directory does not
// already exist.
func (docs DocSet) Write(dir string) error {
	err := os.MkdirAll(dir, 0777)
	if err != nil {
		return err
	}

	versionSpecFiles := make([]string, len(docs)*2)
	for idx, doc := range docs {
		versionDir := path.Join(dir, doc.VersionDate.Format(time.DateOnly))
		err = os.MkdirAll(versionDir, 0755)
		if err != nil {
			return fmt.Errorf("make output directory: %w", err)
		}

		jsonBuf, err := vervet.ToSpecJSON(doc.Doc)
		if err != nil {
			return fmt.Errorf("serialise spec to json: %w", err)
		}
		jsonSpecPath := path.Join(versionDir, "spec.json")
		jsonEmbedPath, err := filepath.Rel(dir, jsonSpecPath)
		if err != nil {
			return fmt.Errorf("get relative output path: %w", err)
		}
		versionSpecFiles[idx*2] = jsonEmbedPath
		err = os.WriteFile(jsonSpecPath, jsonBuf, 0644)
		if err != nil {
			return fmt.Errorf("write json file: %w", err)
		}
		fmt.Println(jsonSpecPath)

		yamlBuf, err := yaml.JSONToYAML(jsonBuf)
		if err != nil {
			return fmt.Errorf("convert spec to yaml: %w", err)
		}
		yamlBuf, err = vervet.WithGeneratedComment(yamlBuf)
		if err != nil {
			return fmt.Errorf("prepend yaml comment: %w", err)
		}
		yamlSpecPath := path.Join(versionDir, "spec.yaml")
		yamlEmbedPath, err := filepath.Rel(dir, yamlSpecPath)
		if err != nil {
			return fmt.Errorf("get relative output path: %w", err)
		}
		versionSpecFiles[idx*2+1] = yamlEmbedPath
		err = os.WriteFile(yamlSpecPath, yamlBuf, 0644)
		if err != nil {
			return fmt.Errorf("write yaml file: %w", err)
		}
		fmt.Println(yamlSpecPath)
	}
	return writeEmbedGo(dir, versionSpecFiles)
}

// Go services embed the compiled specs in the binary to avoid loading them
// from the file system at runtime, this is done with the embed package.
func writeEmbedGo(dir string, versionSpecFiles []string) error {
	embedPath := filepath.Join(dir, "embed.go")
	f, err := os.Create(embedPath)
	if err != nil {
		return fmt.Errorf("create embed.go: %w", err)
	}
	defer f.Close()

	return compiler.EmbedGoTmpl.Execute(f, struct {
		Package          string
		VersionSpecFiles []string
	}{
		Package:          filepath.Base(dir),
		VersionSpecFiles: versionSpecFiles,
	})
}
