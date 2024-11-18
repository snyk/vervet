package cmd

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"

	"github.com/urfave/cli/v2"

	"github.com/snyk/vervet/v8"
	"github.com/snyk/vervet/v8/config"
	"github.com/snyk/vervet/v8/internal/backstage"
)

// BackstageCommand is the `vervet backstage` subcommand.
var BackstageCommand = cli.Command{
	Name: "backstage",
	Subcommands: []*cli.Command{{
		Name:  "update-catalog",
		Usage: "Update Backstage catalog-info.yaml with Vervet API versions",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "config",
				Aliases: []string{"c", "conf"},
				Usage:   "Project configuration file",
			},
			&cli.StringFlag{
				Name:    pivotDateCLIFlagName,
				Aliases: []string{"P"},
				Usage: fmt.Sprintf(
					"Pivot version after which new strategy versioning is used."+
						" Flag for testing only, recommend to use the default date(%s)", vervet.DefaultPivotDate.String()),
				Value: vervet.DefaultPivotDate.String(),
			},
		},
		Action: UpdateCatalog,
	}, {
		Name:  "preview-catalog",
		Usage: "Preview changes to Backstage catalog-info.yaml",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "config",
				Aliases: []string{"c", "conf"},
				Usage:   "Project configuration file",
			},
			&cli.StringFlag{
				Name:    pivotDateCLIFlagName,
				Aliases: []string{"P"},
				Usage: fmt.Sprintf(
					"Pivot version after which new strategy versioning is used."+
						" Flag for testing only, recommend to use the default date(%s)", vervet.DefaultPivotDate.String()),
				Value: vervet.DefaultPivotDate.String(),
			},
		},
		Action: PreviewCatalog,
	}, {
		Name:  "check-catalog",
		Usage: "Check for uncommitted changes in Backstage catalog-info.yaml",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "config",
				Aliases: []string{"c", "conf"},
				Usage:   "Project configuration file",
			},
		},
		Action: CheckCatalog,
	}},
}

// UpdateCatalog updates the catalog-info.yaml from Vervet versions.
func UpdateCatalog(ctx *cli.Context) error {
	return processCatalog(ctx, nil)
}

// PreviewCatalog updates the catalog-info.yaml from Vervet versions.
func PreviewCatalog(ctx *cli.Context) error {
	return processCatalog(ctx, os.Stdout)
}

// CheckCatalog checks whether the catalog-info.yaml or tracked compiled
// versions it references have uncommitted changes. This is primarily useful in
// CI checks to make sure everything is checked into git for Backstage.
func CheckCatalog(ctx *cli.Context) error {
	projectDir, _, err := projectConfig(ctx)
	if err != nil {
		return err
	}

	if st, err := os.Stat(filepath.Join(projectDir, ".git")); err != nil || !st.IsDir() {
		// no git, no problem, just note
		log.Println(projectDir, "does not seem to be tracked in a git repository")
		return nil
	}

	catalogInfoPath := filepath.Join(projectDir, "catalog-info.yaml")
	fr, err := os.Open(catalogInfoPath)
	if err != nil {
		return err
	}
	defer fr.Close()

	err = checkUncommittedChanges(catalogInfoPath)
	if err != nil {
		return err
	}
	catalogInfo, err := backstage.LoadCatalogInfo(fr)
	if err != nil {
		return err
	}

	for _, vervetAPI := range catalogInfo.VervetAPIs {
		specPath := filepath.Join(projectDir, vervetAPI.Spec.Definition.Text)
		if err := checkUncommittedChanges(specPath); err != nil {
			return err
		}
	}
	return nil
}

func checkUncommittedChanges(path string) error {
	cmd := exec.Command("git", "status", "--porcelain", path)
	out, err := cmd.Output()
	if err != nil {
		log.Println("failed to execute git:", err)
	}
	if len(out) > 0 {
		return fmt.Errorf("%s has uncommitted changes", path)
	}
	return nil
}

func processCatalog(ctx *cli.Context, w io.Writer) error {
	projectDir, configFile, err := projectConfig(ctx)
	if err != nil {
		return err
	}
	pivotDate, err := parsePivotDate(ctx)
	if err != nil {
		return fmt.Errorf("failed to parse pivot date %q: %w", pivotDate, err)
	}

	f, err := os.Open(configFile)
	if err != nil {
		return err
	}
	defer f.Close()
	proj, err := config.Load(f)
	if err != nil {
		return err
	}

	catalogInfoPath := filepath.Join(projectDir, "catalog-info.yaml")
	fr, err := os.Open(catalogInfoPath)
	if err != nil {
		return err
	}
	defer fr.Close()
	catalogInfo, err := backstage.LoadCatalogInfo(fr)
	if err != nil {
		return err
	}

	matchPath := func(path string) bool { return true }
	if st, err := os.Stat(filepath.Join(projectDir, ".git")); err == nil && st.IsDir() {
		matchPath = func(path string) bool {
			cmd := exec.Command("git", "ls-files", path)
			out, err := cmd.Output()
			if err != nil {
				log.Println("failed to execute git to test output path:", err)
				return false
			}
			if len(out) == 0 {
				return false
			}
			return true
		}
	}

	// range over maps does not specify order and is not guaranteed to be the
	// same from one iteration to the next, stability is important when
	// generating catalog-info to produce reproducible results
	apiNames := []string{}
	for k := range proj.APIs {
		apiNames = append(apiNames, k)
	}
	sort.Strings(apiNames)
	for _, apiName := range apiNames {
		apiConf := proj.APIs[apiName]
		outputPaths := apiConf.Output.ResolvePaths()
		for _, outputPath := range outputPaths {
			outputPath = filepath.Join(projectDir, outputPath)
			if matchPath(outputPath) {
				if err := catalogInfo.LoadVervetAPIs(projectDir, outputPath, pivotDate.Date, apiName); err != nil {
					return err
				}
				break
			}
		}
	}

	if w == nil {
		fw, err := os.Create(catalogInfoPath)
		if err != nil {
			return err
		}
		defer fw.Close()
		w = fw
	}

	return catalogInfo.Save(w)
}
