package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/olekukonko/tablewriter"
	"github.com/urfave/cli/v2"

	"github.com/snyk/vervet"
	"github.com/snyk/vervet/config"
	"github.com/snyk/vervet/internal/compiler"
	"github.com/snyk/vervet/internal/generator"
)

// VersionList is a command that lists all the versions of matching resources.
// It takes optional arguments to filter the output: api resource
func VersionList(ctx *cli.Context) error {
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
			specVersions, err := vervet.LoadSpecVersionsFileset(specFiles)
			if err != nil {
				return err
			}
			for _, rc := range specVersions.Resources() {
				if rcArg := ctx.Args().Get(1); rcArg != "" && rcArg != rc.Name() {
					continue
				}
				for _, version := range rc.Versions() {
					doc, err := rc.At(version.String())
					if err != nil {
						return err
					}
					var pathNames []string
					for k := range doc.Paths {
						pathNames = append(pathNames, k)
					}
					sort.Strings(pathNames)
					for _, pathName := range pathNames {
						pathSpec := doc.Paths[pathName]
						if pathSpec.Get != nil {
							table.Append([]string{apiName, rc.Name(), version.String(), pathName, "GET", pathSpec.Get.OperationID})
						}
						if pathSpec.Post != nil {
							table.Append([]string{apiName, rc.Name(), version.String(), pathName, "POST", pathSpec.Post.OperationID})
						}
						if pathSpec.Put != nil {
							table.Append([]string{apiName, rc.Name(), version.String(), pathName, "PUT", pathSpec.Put.OperationID})
						}
						if pathSpec.Patch != nil {
							table.Append([]string{apiName, rc.Name(), version.String(), pathName, "PATCH", pathSpec.Patch.OperationID})
						}
						if pathSpec.Delete != nil {
							table.Append([]string{apiName, rc.Name(), version.String(), pathName, "DELETE", pathSpec.Delete.OperationID})
						}
					}
				}
			}
		}
	}
	table.Render()
	return nil
}

// VersionFiles is a command that lists all versioned OpenAPI spec files of
// matching resources.
// It takes optional arguments to filter the output: api resource
func VersionFiles(ctx *cli.Context) error {
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

type specVersionKey struct {
	API      string
	Resource string
	Version  vervet.Version
}

// VersionNew generates a new resource.
func VersionNew(ctx *cli.Context) error {
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
	var options []generator.Option
	if ctx.Bool("force") {
		options = append(options, generator.Force(true))
	}
	if ctx.Bool("debug") {
		options = append(options, generator.Debug(true))
	}
	generators, err := generator.NewMap(proj, options...)
	if err != nil {
		return err
	}
	err = os.Chdir(projectDir)
	if err != nil {
		return err
	}
	apiName, resourceName := ctx.Args().Get(0), ctx.Args().Get(1)
	if apiName == "" || resourceName == "" {
		return fmt.Errorf("api and resource are required")
	}
	api, ok := proj.APIs[apiName]
	if !ok && len(proj.APIs) > 0 {
		var apiNames []string
		for k := range proj.APIs {
			apiNames = append(apiNames, k)
		}
		sort.Strings(apiNames)
		return fmt.Errorf(`API %q not found. Choose an existing one (%s) or
`+"`%s api new %s <resource path>`"+` to start a new API`,
			apiName, strings.Join(apiNames, ", "), os.Args[0], apiName)
	}
	if len(api.Resources) == 0 {
		return fmt.Errorf(`API %q does not seem to have a resource set defined.
Please add a `+"`resources:`"+` section to
%q and try again`, apiName, configFile)
	}

	versionTime, err := time.Parse("2006-01-02", ctx.String("version"))
	if err != nil {
		return err
	}
	version := versionTime.Format("2006-01-02")
	resourceDir := api.Resources[0].Path
	versionDir := filepath.Join(resourceDir, resourceName, version)
	err = os.MkdirAll(versionDir, 0777)
	if err != nil {
		return fmt.Errorf("failed to create version path %q: %w", versionDir, err)
	}

	for _, genName := range api.Resources[0].Generators {
		gen := generators[genName]
		context := &generator.VersionScope{
			API:       apiName,
			Resource:  resourceName,
			Version:   version,
			Stability: ctx.String("stability"),
		}
		err := gen.Run(context)
		if err != nil {
			return fmt.Errorf("%w (generators.%s)", err, genName)
		}
	}
	return nil
}
