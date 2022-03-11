package generator

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"text/template"

	"github.com/ghodss/yaml"

	"github.com/snyk/vervet/v4"
	"github.com/snyk/vervet/v4/config"
)

// Generator generates files for new resources from data models and templates.
type Generator struct {
	name      string
	filename  *template.Template
	contents  *template.Template
	files     *template.Template
	functions template.FuncMap
	scope     config.GeneratorScope

	debug  bool
	dryRun bool
	force  bool
	here   string
}

// NewMap instanstiates a map of Generators from configuration.
func NewMap(generatorsConf config.Generators, options ...Option) (map[string]*Generator, error) {
	result := map[string]*Generator{}
	for name, genConf := range generatorsConf {
		g, err := New(genConf, options...)
		if err != nil {
			return nil, err
		}
		result[name] = g
	}
	return result, nil
}

// New returns a new Generator from configuration.
func New(conf *config.Generator, options ...Option) (*Generator, error) {
	g := &Generator{
		name:      conf.Name,
		scope:     conf.Scope,
		functions: template.FuncMap{},
	}
	for i := range options {
		options[i](g)
	}
	if g.debug {
		log.Printf("generator %s: debug logging enabled", g.name)
	}

	// If .Here isn't specified, we'll assume cwd.
	var err error
	if g.here == "" {
		g.here, err = os.Getwd()
		if err != nil {
			return nil, err
		}
	}

	// Resolve the template 'functions'... with a template. Only .Here is
	// supported, not full scope. Just enough to locate files relative to the
	// config.
	if conf.Functions != "" {
		functionsFilename, err := g.resolveFilename(conf.Functions)
		if err != nil {
			return nil, fmt.Errorf("%w: (generators.%s.functions)", err, conf.Name)
		}
		err = g.loadFunctions(functionsFilename)
		if err != nil {
			return nil, fmt.Errorf("%w: (generators.%s.functions)", err, conf.Name)
		}
	}

	// Resolve the template filename... with a template. Only .Here is
	// supported, not full scope. Just enough to locate files relative to the
	// config.
	templateFilename, err := g.resolveFilename(conf.Template)
	if err != nil {
		return nil, fmt.Errorf("%w: (generators.%s.template)", err, conf.Name)
	}

	// Parse & wire up other templates: contents, filename or files. These do
	// support full scope.
	contentsTemplate, err := ioutil.ReadFile(templateFilename)
	if err != nil {
		return nil, fmt.Errorf("%w: (generators.%s.contents)", err, conf.Name)
	}
	g.contents, err = withIncludeFunc(template.New("contents").
		Funcs(g.functions).
		Funcs(builtinFuncs)).
		Parse(string(contentsTemplate))
	if err != nil {
		return nil, fmt.Errorf("%w: (generators.%s.contents)", err, conf.Name)
	}
	if conf.Filename != "" {
		g.filename, err = template.New("filename").
			Funcs(g.functions).
			Funcs(builtinFuncs).
			Parse(conf.Filename)
		if err != nil {
			return nil, fmt.Errorf("%w: (generators.%s.filename)", err, conf.Name)
		}
	}
	if conf.Files != "" {
		g.files, err = withIncludeFunc(g.contents.New("files").Funcs(g.functions)).Parse(conf.Files)
		if err != nil {
			return nil, fmt.Errorf("%w: (generators.%s.files)", err, conf.Name)
		}
	}
	return g, nil
}

func (g *Generator) resolveFilename(filenameTemplate string) (string, error) {
	t, err := template.New("").Funcs(g.functions).Parse(string(filenameTemplate))
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	err = t.ExecuteTemplate(&buf, "", map[string]string{
		"Here": g.here,
	})
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

// Option configures a Generator.
type Option func(g *Generator)

// Force configures the Generator to overwrite generated artifacts.
func Force(force bool) Option {
	return func(g *Generator) {
		g.force = true
	}
}

// Debug turns on template debug logging.
func Debug(debug bool) Option {
	return func(g *Generator) {
		g.debug = true
	}
}

// DryRun executes templates and lists the files that would be generated
// without actually generating them.
func DryRun(dryRun bool) Option {
	return func(g *Generator) {
		g.dryRun = dryRun
	}
}

// Here sets the .Here scope property. This is typically relative to the
// location of the generators config file.
func Here(here string) Option {
	return func(g *Generator) {
		g.here = here
	}
}

// Execute runs the generator on the given resources.
func (g *Generator) Execute(resources ResourceMap) ([]string, error) {
	var allFiles []string
	switch g.Scope() {
	case config.GeneratorScopeDefault, config.GeneratorScopeVersion:
		for rcKey, rcVersions := range resources {
			for _, version := range rcVersions.Versions() {
				rc, err := rcVersions.At(version.String())
				if err != nil {
					return nil, err
				}
				scope := &VersionScope{
					API:             rcKey.API,
					Path:            filepath.Join(rcKey.Path, version.DateString()),
					ResourceVersion: rc,
					Here:            g.here,
				}
				generatedFiles, err := g.execute(scope)
				if err != nil {
					return nil, err
				}
				allFiles = append(allFiles, generatedFiles...)
			}
		}
	case config.GeneratorScopeResource:
		for rcKey, rcVersions := range resources {
			scope := &ResourceScope{
				API:              rcKey.API,
				Path:             rcKey.Path,
				ResourceVersions: rcVersions,
				Here:             g.here,
			}
			generatedFiles, err := g.execute(scope)
			if err != nil {
				return nil, err
			}
			allFiles = append(allFiles, generatedFiles...)
		}
	default:
		return nil, fmt.Errorf("unsupported generator scope %q", g.Scope())
	}
	return allFiles, nil
}

// ResourceScope identifies a resource that the generator is building for.
type ResourceScope struct {
	// ResourceVersions contains all the versions of this resource.
	*vervet.ResourceVersions
	// API is name of the API containing this resource.
	API string
	// Path is the path to the resource directory.
	Path string
	// Here is the directory containing the executing template.
	Here string
}

// Resource returns the name of the resource in scope.
func (s *ResourceScope) Resource() string {
	return s.ResourceVersions.Name()
}

// VersionScope identifies a distinct version of a resource that the generator
// is building for.
type VersionScope struct {
	*vervet.ResourceVersion
	// API is name of the API containing this resource.
	API string
	// Path is the path to the resource directory.
	Path string
	// Here is the directory containing the generator template.
	Here string
}

// Resource returns the name of the resource in scope.
func (s *VersionScope) Resource() string {
	return s.ResourceVersion.Name
}

// Version returns the version of the resource in scope.
func (s *VersionScope) Version() *vervet.Version {
	return &s.ResourceVersion.Version
}

// Scope returns the configured scope type of the generator.
func (g *Generator) Scope() config.GeneratorScope {
	return g.scope
}

// execute the Generator. If generated artifacts already exist, a warning
// is logged but the file is not overwritten, unless force is true.
//
// TODO: in Go 1.18, declare scope as an interface{ VersionScope | ResourceScope }
func (g *Generator) execute(scope interface{}) ([]string, error) {
	if g.files != nil {
		return g.runFiles(scope)
	}
	return g.runFile(scope)
}

func (g *Generator) runFile(scope interface{}) ([]string, error) {
	var filenameBuf bytes.Buffer
	err := g.filename.ExecuteTemplate(&filenameBuf, "filename", scope)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve filename: %w (generators.%s.filename)", err, g.name)
	}
	filename := filenameBuf.String()
	if g.debug {
		log.Printf("interpolated generators.%s.filename => %q", g.name, filename)
	}
	if _, err := os.Stat(filename); err == nil && !g.force {
		log.Printf("not overwriting existing file %q", filename)
		return nil, nil
	}
	parentDir := filepath.Dir(filename)
	err = os.MkdirAll(parentDir, 0777)
	if err != nil {
		return nil, fmt.Errorf("failed to create %q: %w: (generators.%s.filename)", parentDir, err, g.name)
	}
	var out io.Writer
	if g.dryRun {
		out = io.Discard
	} else {
		f, err := os.Create(filename)
		if err != nil {
			return nil, fmt.Errorf("failed to create %q: %w: (generators.%s.filename)", filename, err, g.name)
		}
		defer f.Close()
		out = f
	}
	err = g.contents.ExecuteTemplate(out, "contents", scope)
	if err != nil {
		return nil, fmt.Errorf("template failed: %w (generators.%s.filename)", err, g.name)
	}
	return []string{filename}, nil
}

func (g *Generator) runFiles(scope interface{}) ([]string, error) {
	var filesBuf bytes.Buffer
	err := g.files.ExecuteTemplate(&filesBuf, "files", scope)
	if err != nil {
		return nil, fmt.Errorf("%w: (generators.%s.files)", err, g.name)
	}
	if g.debug {
		log.Printf("interpolated generators.%s.files => %q", g.name, filesBuf.String())
	}
	files := map[string]string{}
	err = yaml.Unmarshal(filesBuf.Bytes(), &files)
	if err != nil {
		// TODO: dump output for debugging?
		return nil, fmt.Errorf("failed to load output as yaml: %w: (generators.%s.files)", err, g.name)
	}
	var generatedFiles []string
	for filename, contents := range files {
		generatedFiles = append(generatedFiles, filename)
		dir := filepath.Dir(filename)
		err := os.MkdirAll(dir, 0777)
		if err != nil {
			return nil, fmt.Errorf("failed to create directory %q: %w (generators.%s.files)", dir, err, g.name)
		}
		if _, err := os.Stat(filename); err == nil && !g.force {
			log.Printf("not overwriting existing file %q", filename)
			continue
		}
		if g.dryRun {
			continue
		}
		err = ioutil.WriteFile(filename, []byte(contents), 0777)
		if err != nil {
			return nil, fmt.Errorf("failed to write file %q: %w (generators.%s.files)", filename, err, g.name)
		}
	}
	return generatedFiles, nil
}
