package simplebuild

import (
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
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
func (docs DocSet) WriteOutputs(cfg config.Output, appendOutputFiles bool) error {
	paths := getOutputPaths(cfg)

	if !appendOutputFiles {
		for _, dir := range paths {
			err := os.RemoveAll(dir)
			if err != nil {
				return fmt.Errorf("clear output directory: %w", err)
			}
		}
	}

	err := docs.Write(paths[0], appendOutputFiles)
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
func (docs DocSet) Write(dir string, appendOutputFiles bool) error {
	err := os.MkdirAll(dir, 0777)
	if err != nil {
		return err
	}
	existingFiles, err := getExisingSpecFiles(dir)
	if err != nil {
		return fmt.Errorf("list existing files: %w", err)
	}

	versionSpecFiles := make([]string, 0, len(existingFiles)+len(docs)*2)
	versionSpecFiles = append(versionSpecFiles, existingFiles...)
	for _, doc := range docs {
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
		versionSpecFiles = append(versionSpecFiles, jsonEmbedPath)
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
		versionSpecFiles = append(versionSpecFiles, yamlEmbedPath)
		err = os.WriteFile(yamlSpecPath, yamlBuf, 0644)
		if err != nil {
			return fmt.Errorf("write yaml file: %w", err)
		}
		fmt.Println(yamlSpecPath)
	}
	return writeEmbedGo(dir, versionSpecFiles)
}

func getExisingSpecFiles(dir string) ([]string, error) {
	var outputFiles []string
	err := filepath.WalkDir(dir, func(filePath string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || d.Name() == "embed.go" {
			return nil
		}
		relativePath, err := filepath.Rel(dir, filePath)
		if err != nil {
			return err
		}
		outputFiles = append(outputFiles, relativePath)
		return nil
	})
	// Sort files for consistency
	sort.Strings(outputFiles)
	return outputFiles, err
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
