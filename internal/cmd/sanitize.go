package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/urfave/cli/v2"

	"github.com/snyk/vervet/v8"
	"github.com/snyk/vervet/v8/internal/storage"
)

var SanitizeCommand = cli.Command{
	Name:      "sanitize",
	Usage:     "Manually load compiled specs from subfolders, strip sensitive fields, write sanitized output",
	ArgsUsage: " ",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "compiled-path",
			Required: true,
			Usage:    "Directory containing subfolders for each version, e.g. internal/rest/api/versions/",
		},
		&cli.StringFlag{
			Name:     "out",
			Usage:    "Output directory for sanitized versions",
			Value:    "sanitized",
			Required: false,
		},
		&cli.StringSliceFlag{
			Name:  "exclude-extension",
			Usage: "Regex patterns of extension names to remove (repeatable)",
		},
		&cli.StringSliceFlag{
			Name:  "exclude-header",
			Usage: "Regex patterns of header names to remove (repeatable)",
		},
		&cli.StringSliceFlag{
			Name:  "exclude-path",
			Usage: "Exact path(s) to remove from the final OpenAPI specs",
		},
	},
	Action: sanitizeAction,
}

func sanitizeAction(c *cli.Context) error {
	excludePatterns := vervet.ExcludePatterns{
		ExtensionPatterns: c.StringSlice("exclude-extension"),
		HeaderPatterns:    c.StringSlice("exclude-header"),
		Paths:             c.StringSlice("exclude-path"),
	}

	compiledPath := c.String("compiled-path")

	outDir := c.String("out")
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return fmt.Errorf("failed to create out directory %q: %w", outDir, err)
	}

	versionDirs, err := findVersionDirs(compiledPath)
	if err != nil {
		return fmt.Errorf("failed to find version subfolders in %q: %w", compiledPath, err)
	}

	if len(versionDirs) == 0 {
		fmt.Fprintf(c.App.Writer, "No version subfolders found in %q\n", compiledPath)
		return nil
	}

	coll, err := storage.NewCollator(
		storage.CollatorExcludePattern(excludePatterns),
	)

	if err != nil {
		return fmt.Errorf("failed to create collator: %w", err)
	}

	// for each version folder, read openapi.yaml/spec.yaml, parse version, add to Collator
	const serviceName = "api" // or any label you prefer
	for _, dir := range versionDirs {
		specFile := filepath.Join(dir, "spec.yaml")
		if _, statErr := os.Stat(specFile); statErr != nil {
			// No recognized spec found, skip this version folder
			continue
		}

		blob, err := os.ReadFile(specFile)
		if err != nil {
			return fmt.Errorf("failed to read spec file %q: %w", specFile, err)
		}

		versionName := filepath.Base(dir)
		v, err := vervet.ParseVersion(versionName)
		if err != nil {
			fmt.Fprintf(c.App.Writer, "Skipping folder %q: not a valid version: %v\n", dir, err)
			continue
		}

		digest := storage.NewDigest(blob)
		rev := storage.ContentRevision{
			Service: serviceName,
			Version: v,
			Digest:  digest,
			Blob:    blob,
		}
		coll.Add(serviceName, rev)
	}

	sanitized, err := coll.Collate()
	if err != nil {
		return fmt.Errorf("collate failed: %w", err)
	}

	for version, doc := range sanitized {
		versionOutDir := filepath.Join(outDir, version.String())
		if err := os.MkdirAll(versionOutDir, 0o755); err != nil {
			return fmt.Errorf("failed to create version output dir %q: %w", versionOutDir, err)
		}
		outPath := filepath.Join(versionOutDir, "openapi.yaml")

		jsonBytes, err := doc.MarshalJSON()
		if err != nil {
			return fmt.Errorf("failed to marshal JSON for version %s: %w", version, err)
		}

		// Write the final sanitized file
		if err := os.WriteFile(outPath, jsonBytes, 0o644); err != nil {
			return fmt.Errorf("failed to write sanitized spec %q: %w", outPath, err)
		}
	}

	fmt.Fprintf(c.App.Writer, "Wrote sanitized specs to %s\n", outDir)
	return nil
}

// findVersionDirs enumerates subdirectories of compiledPath and returns them sorted.
func findVersionDirs(compiledPath string) ([]string, error) {
	entries, err := os.ReadDir(compiledPath)
	if err != nil {
		return nil, err
	}
	var dirs []string
	for _, e := range entries {
		if e.IsDir() {
			dirs = append(dirs, filepath.Join(compiledPath, e.Name()))
		}
	}
	sort.Strings(dirs)
	return dirs, nil
}
