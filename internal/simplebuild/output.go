package simplebuild

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	"time"

	"github.com/ghodss/yaml"

	"github.com/snyk/vervet/v8"
	"github.com/snyk/vervet/v8/config"
	"github.com/snyk/vervet/v8/internal/compiler"
	"github.com/snyk/vervet/v8/internal/files"
)

type DocWriter struct {
	cfg              config.Output
	paths            []string
	versionSpecFiles []string
}

// NewWriter initialises any output paths, removing existing files and
// directories if they are present.
func NewWriter(cfg config.Output, appendOutputFiles bool) (*DocWriter, error) {
	paths := cfg.Paths
	toClear := paths
	if appendOutputFiles {
		// We treat the first path as the source of truth and copy the whole
		// directory to the other paths in Finalize.
		toClear = toClear[1:]
	}

	for _, dir := range toClear {
		err := os.RemoveAll(dir)
		if err != nil {
			return nil, fmt.Errorf("clear output directory: %w", err)
		}
	}
	err := os.MkdirAll(paths[0], 0777)
	if err != nil {
		return nil, fmt.Errorf("make output directory: %w", err)
	}

	versionSpecFiles, err := getExisingSpecFiles(paths[0])
	if err != nil {
		return nil, fmt.Errorf("list existing files: %w", err)
	}

	return &DocWriter{
		cfg:              cfg,
		paths:            paths,
		versionSpecFiles: versionSpecFiles,
	}, nil
}

// Write writes compiled specs to a single directory in YAML and JSON formats.
// Call Finalize after to populate other directories.
func (out *DocWriter) Write(ctx context.Context, doc VersionedDoc) error {
	err := doc.Doc.Validate(ctx)
	if err != nil {
		return fmt.Errorf("invalid compiled document: %w", err)
	}

	// We write to the first directory then copy the entire directory
	// afterwards
	dir := out.paths[0]

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
	out.versionSpecFiles = append(out.versionSpecFiles, jsonEmbedPath)
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
	out.versionSpecFiles = append(out.versionSpecFiles, yamlEmbedPath)
	err = os.WriteFile(yamlSpecPath, yamlBuf, 0644)
	if err != nil {
		return fmt.Errorf("write yaml file: %w", err)
	}
	fmt.Println(yamlSpecPath)
	return nil
}

func (out *DocWriter) Finalize() error {
	err := writeEmbedGo(out.paths[0], out.versionSpecFiles)
	if err != nil {
		return err
	}
	for _, dir := range out.paths[1:] {
		err := files.CopyDir(dir, out.paths[0], true)
		if err != nil {
			return fmt.Errorf("copy outputs: %w", err)
		}
	}
	return nil
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
