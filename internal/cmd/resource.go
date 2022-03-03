package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/olekukonko/tablewriter"
	"github.com/urfave/cli/v2"

	"github.com/snyk/vervet/v4"
	"github.com/snyk/vervet/v4/config"
	"github.com/snyk/vervet/v4/internal/compiler"
)

// ResourceCommand is the `vervet resource` subcommand.
var ResourceCommand = cli.Command{
	Name:    "resource",
	Aliases: []string{"rc"},
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "config",
			Aliases: []string{"c", "conf"},
			Usage:   "Project configuration file",
		},
	},
	Subcommands: []*cli.Command{{
		Name:      "files",
		Usage:     "List OpenAPI files of versioned resources in a vervet project",
		ArgsUsage: "[api [resource]]",
		Action:    ResourceFiles,
	}, {
		Name:      "info",
		Usage:     "Information about versioned resources in a vervet project",
		ArgsUsage: "[api [resource]]",
		Action:    ResourceShow,
	}},
}

// ResourceShow is a command that lists all the versions of matching resources.
// It takes optional arguments to filter the output: api resource
func ResourceShow(ctx *cli.Context) error {
	projectDir, configFile, err := projectConfig(ctx)
	if err != nil {
		return err
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
	err = os.Chdir(projectDir)
	if err != nil {
		return err
	}
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"API", "Resource", "Version", "Path", "Method", "Operation"})
	for _, apiName := range proj.APINames() {
		if apiArg := ctx.Args().Get(0); apiArg != "" && apiArg != apiName {
			continue
		}
		api := proj.APIs[apiName]
		for _, rcConfig := range api.Resources {
			specFiles, err := compiler.ResourceSpecFiles(rcConfig)
			if err != nil {
				return err
			}
			resources, err := vervet.LoadResourceVersionsFileset(specFiles)
			if err != nil {
				return err
			}
			for _, version := range resources.Versions() {
				rc, err := resources.At(version.String())
				if err != nil {
					return err
				}
				if rcArg := ctx.Args().Get(1); rcArg != "" && rcArg != rc.Name {
					continue
				}
				var pathNames []string
				for k := range rc.Paths {
					pathNames = append(pathNames, k)
				}
				sort.Strings(pathNames)
				for _, pathName := range pathNames {
					pathSpec := rc.Paths[pathName]
					if pathSpec.Get != nil {
						table.Append([]string{apiName, rc.Name, version.String(), pathName, "GET", pathSpec.Get.OperationID})
					}
					if pathSpec.Post != nil {
						table.Append([]string{apiName, rc.Name, version.String(), pathName, "POST", pathSpec.Post.OperationID})
					}
					if pathSpec.Put != nil {
						table.Append([]string{apiName, rc.Name, version.String(), pathName, "PUT", pathSpec.Put.OperationID})
					}
					if pathSpec.Patch != nil {
						table.Append([]string{apiName, rc.Name, version.String(), pathName, "PATCH", pathSpec.Patch.OperationID})
					}
					if pathSpec.Delete != nil {
						table.Append([]string{apiName, rc.Name, version.String(), pathName, "DELETE", pathSpec.Delete.OperationID})
					}
				}
			}
		}
	}
	table.Render()
	return nil
}

// ResourceFiles is a command that lists all versioned OpenAPI spec files of
// matching resources.
// It takes optional arguments to filter the output: api resource
func ResourceFiles(ctx *cli.Context) error {
	projectDir, configFile, err := projectConfig(ctx)
	if err != nil {
		return err
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
	err = os.Chdir(projectDir)
	if err != nil {
		return err
	}
	for _, apiName := range proj.APINames() {
		if apiArg := ctx.Args().Get(0); apiArg != "" && apiArg != apiName {
			continue
		}
		api := proj.APIs[apiName]
		for _, rcConfig := range api.Resources {
			specFiles, err := compiler.ResourceSpecFiles(rcConfig)
			if err != nil {
				return err
			}
			sort.Strings(specFiles)
			for i := range specFiles {
				rcName := filepath.Base(filepath.Dir(filepath.Dir(specFiles[i])))
				if rcArg := ctx.Args().Get(1); rcArg != "" && rcArg != rcName {
					continue
				}
				fmt.Println(specFiles[i])
			}
		}
	}
	return nil
}
