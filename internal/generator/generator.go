package generator

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/ghodss/yaml"

	"github.com/snyk/vervet/v4"
	"github.com/snyk/vervet/v4/config"
)

// Generator generates files for new resources from data models and templates.
type Generator struct {
	name     string
	filename *template.Template
	contents *template.Template
	files    *template.Template
	scope    config.GeneratorScope

	debug bool
	force bool
	here  string
}

var (
	templateFuncs = template.FuncMap{
		"map": func(keyValues ...interface{}) (map[string]interface{}, error) {
			if len(keyValues)%2 != 0 {
				return nil, fmt.Errorf("invalid number of arguments to map")
			}
			m := make(map[string]interface{}, len(keyValues)/2)
			for i := 0; i < len(keyValues); i += 2 {
				k, ok := keyValues[i].(string)
				if !ok {
					return nil, fmt.Errorf("map keys must be strings")
				}
				m[k] = keyValues[i+1]
			}
			return m, nil
		},
		"indent": func(indent int, s string) string {
			return strings.ReplaceAll(s, "\n", "\n"+strings.Repeat(" ", indent))
		},
		"uncapitalize": func(s string) string {
			if len(s) > 1 {
				return strings.ToLower(s[0:1]) + s[1:]
			}
			return s
		},
		"capitalize": func(s string) string {
			if len(s) > 1 {
				return strings.ToUpper(s[0:1]) + s[1:]
			}
			return s
		},
		"replaceall":         strings.ReplaceAll,
		"pathOperations":     MapPathOperations,
		"resourceOperations": MapResourceOperations,
	}
)

func withIncludeFunc(t *template.Template) *template.Template {
	return t.Funcs(template.FuncMap{
		"include": func(name string, data interface{}) (string, error) {
			buf := bytes.NewBuffer(nil)
			if err := t.ExecuteTemplate(buf, name, data); err != nil {
				return "", err
			}
			return buf.String(), nil
		},
	})
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
		name:  conf.Name,
		scope: conf.Scope,
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

	// Resolve the template filename... with a template. Only .Here is
	// supported, not full scope. Just enough to locate files relative to the
	// config.
	templateTemplate, err := template.New("template").Parse(string(conf.Template))
	if err != nil {
		return nil, fmt.Errorf("%w: (generators.%s.template)", err, conf.Name)
	}
	var templateFilenameBuf bytes.Buffer
	err = templateTemplate.ExecuteTemplate(&templateFilenameBuf, "template", map[string]string{
		"Here": g.here,
	})
	if err != nil {
		return nil, fmt.Errorf("%w: (generators.%s.template)", err, conf.Name)
	}

	// Parse & wire up other templates: contents, filename or files. These do
	// support full scope.
	contentsTemplate, err := ioutil.ReadFile(templateFilenameBuf.String())
	if err != nil {
		return nil, fmt.Errorf("%w: (generators.%s.contents)", err, conf.Name)
	}
	g.contents, err = template.New("contents").Funcs(templateFuncs).Parse(string(contentsTemplate))
	if err != nil {
		return nil, fmt.Errorf("%w: (generators.%s.contents)", err, conf.Name)
	}
	if conf.Filename != "" {
		g.filename, err = template.New("filename").Funcs(templateFuncs).Parse(conf.Filename)
		if err != nil {
			return nil, fmt.Errorf("%w: (generators.%s.filename)", err, conf.Name)
		}
	}
	if conf.Files != "" {
		g.files, err = withIncludeFunc(g.contents.New("files")).Parse(conf.Files)
		if err != nil {
			return nil, fmt.Errorf("%w: (generators.%s.files)", err, conf.Name)
		}
	}
	return g, nil
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

// Here sets the .Here scope property. This is typically relative to the
// location of the generators config file.
func Here(here string) Option {
	return func(g *Generator) {
		g.here = here
	}
}

// Execute runs the generator on the given resources.
func (g *Generator) Execute(resources ResourceMap) error {
	switch g.Scope() {
	case config.GeneratorScopeDefault, config.GeneratorScopeVersion:
		for rcKey, rcVersions := range resources {
			for _, version := range rcVersions.Versions() {
				rc, err := rcVersions.At(version.String())
				if err != nil {
					return err
				}
				scope := &VersionScope{
					API:             rcKey.API,
					Path:            filepath.Join(rcKey.Path, version.DateString()),
					ResourceVersion: rc,
					Here:            g.here,
				}
				err = g.execute(scope)
				if err != nil {
					return err
				}
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
			err := g.execute(scope)
			if err != nil {
				return err
			}
		}
	default:
		return fmt.Errorf("unsupported generator scope %q", g.Scope())
	}
	return nil
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
func (g *Generator) execute(scope interface{}) error {
	if g.files != nil {
		return g.runFiles(scope)
	}
	return g.runFile(scope)
}

func (g *Generator) runFile(scope interface{}) error {
	var filenameBuf bytes.Buffer
	err := g.filename.ExecuteTemplate(&filenameBuf, "filename", scope)
	if err != nil {
		return fmt.Errorf("failed to resolve filename: %w (generators.%s.filename)", err, g.name)
	}
	filename := filenameBuf.String()
	if g.debug {
		log.Printf("interpolated generators.%s.filename => %q", g.name, filename)
	}
	if _, err := os.Stat(filename); err == nil && !g.force {
		log.Printf("not overwriting existing file %q", filename)
		return nil
	}
	parentDir := filepath.Dir(filename)
	err = os.MkdirAll(parentDir, 0777)
	if err != nil {
		return fmt.Errorf("failed to create %q: %w: (generators.%s.filename)", parentDir, err, g.name)
	}
	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create %q: %w: (generators.%s.filename)", filename, err, g.name)
	}
	defer f.Close()
	err = g.contents.ExecuteTemplate(f, "contents", scope)
	if err != nil {
		return fmt.Errorf("template failed: %w (generators.%s.filename)", err, g.name)
	}
	return nil
}

func (g *Generator) runFiles(scope interface{}) error {
	var filesBuf bytes.Buffer
	err := g.files.ExecuteTemplate(&filesBuf, "files", scope)
	if err != nil {
		return fmt.Errorf("%w: (generators.%s.files)", err, g.name)
	}
	if g.debug {
		log.Printf("interpolated generators.%s.files => %q", g.name, filesBuf.String())
	}
	files := map[string]string{}
	err = yaml.Unmarshal(filesBuf.Bytes(), &files)
	if err != nil {
		// TODO: dump output for debugging?
		return fmt.Errorf("failed to load output as yaml: %w: (generators.%s.files)", err, g.name)
	}
	for filename, contents := range files {
		dir := filepath.Dir(filename)
		err := os.MkdirAll(dir, 0777)
		if err != nil {
			return fmt.Errorf("failed to create directory %q: %w (generators.%s.files)", dir, err, g.name)
		}
		if _, err := os.Stat(filename); err == nil && !g.force {
			log.Printf("not overwriting existing file %q", filename)
			continue
		}
		err = ioutil.WriteFile(filename, []byte(contents), 0777)
		if err != nil {
			return fmt.Errorf("failed to write file %q: %w (generators.%s.files)", filename, err, g.name)
		}
	}
	return nil
}
