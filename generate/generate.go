package generate

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/snyk/vervet/v4/config"
	"github.com/snyk/vervet/v4/internal/generator"
)

// GeneratorParams contains the metadata needed to execute code generators.
type GeneratorParams struct {
	ProjectDir     string
	ConfigFile     string
	Generators     []string
	GeneratorsFile string
	Force          bool
	Debug          bool
	DryRun         bool
}

// Generate executes code generators against OpenAPI specs.
func Generate(params GeneratorParams) error {
	f, err := os.Open(params.ConfigFile)
	if err != nil {
		return err
	}
	defer f.Close()
	proj, err := config.Load(f)
	if err != nil {
		return err
	}

	selectedGenerators := map[string]struct{}{}
	for _, generator := range params.Generators {
		selectedGenerators[generator] = struct{}{}
	}

	// Option to load generators and overlay onto project config
	generatorsHere := map[string]string{}
	if genFile := params.GeneratorsFile; genFile != "" {
		f, err := os.Open(genFile)
		if err != nil {
			return err
		}
		defer f.Close()
		generators, err := config.LoadGenerators(f)
		if err != nil {
			return err
		}
		for k, v := range generators {
			proj.Generators[k] = v
			generatorsHere[k] = filepath.Dir(genFile)
		}
	}
	// If a list of specific generators were specified, only instantiate those.
	if len(selectedGenerators) > 0 {
		for k := range proj.Generators {
			if _, ok := selectedGenerators[k]; !ok {
				delete(proj.Generators, k)
			}
		}
	}

	options := []generator.Option{generator.Force(true)}
	if params.Debug {
		options = append(options, generator.Debug(true))
	}
	if params.DryRun {
		options = append(options, generator.DryRun(true))
	}
	projectHere := filepath.Dir(params.ConfigFile)
	generators := map[string]*generator.Generator{}
	for k, genConf := range proj.Generators {
		genHere, ok := generatorsHere[k]
		if !ok {
			genHere = projectHere
		}
		genHere, err = filepath.Abs(genHere)
		if err != nil {
			return err
		}
		gen, err := generator.New(genConf, append(options, generator.Here(genHere))...)
		if err != nil {
			return err
		}
		generators[k] = gen
	}

	err = os.Chdir(params.ProjectDir)
	if err != nil {
		return err
	}

	resources, err := generator.MapResources(proj)
	if err != nil {
		return err
	}

	var allGeneratedFiles []string
	for _, gen := range generators {
		generatedFiles, err := gen.Execute(resources)
		if err != nil {
			return err
		}
		allGeneratedFiles = append(allGeneratedFiles, generatedFiles...)
	}

	for _, generatedFile := range allGeneratedFiles {
		fmt.Println(generatedFile)
	}

	return nil
}
