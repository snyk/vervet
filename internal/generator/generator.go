package generator

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/ghodss/yaml"

	"github.com/snyk/vervet"
	"github.com/snyk/vervet/config"
)

// Generator generates files for new resources from data models and templates.
type Generator struct {
	name     string
	filename *template.Template
	contents *template.Template
	files    *template.Template
	data     map[string]*template.Template

	debug bool
	force bool
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
		"replaceall": strings.ReplaceAll,
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

// NewMap instanstiates a map of all Generators defined in a
// Project.
func NewMap(proj *config.Project, options ...Option) (map[string]*Generator, error) {
	result := map[string]*Generator{}
	for name, genConf := range proj.Generators {
		g, err := New(genConf, options...)
		if err != nil {
			return nil, err
		}
		result[name] = g
	}
	return result, nil
}

// New returns a new Generator from config.
func New(conf *config.Generator, options ...Option) (*Generator, error) {
	g := &Generator{
		name: conf.Name,
		data: map[string]*template.Template{},
	}
	for i := range options {
		options[i](g)
	}
	if g.debug {
		log.Printf("generator %s: debug logging enabled", g.name)
	}

	contentsTemplate, err := ioutil.ReadFile(conf.Template)
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
	if len(conf.Data) > 0 {
		for fieldName, genData := range conf.Data {
			g.data[fieldName], err = template.New("include").Funcs(templateFuncs).Parse(genData.Include)
			if err != nil {
				return nil, fmt.Errorf("%w: (generators.%s.data.%s.include)", err, conf.Name, fieldName)
			}
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

// VersionScope identifies a distinct resource version that the generator is
// building for.
type VersionScope struct {
	API       string
	Resource  string
	Version   string
	Stability string
}

func (s *VersionScope) validate() error {
	_, err := vervet.ParseVersion(s.Version)
	if err != nil {
		return err
	}
	_, err = vervet.ParseStability(s.Stability)
	if err != nil {
		return err
	}
	return nil
}

type versionScope struct {
	*VersionScope
	Data map[string]interface{}
}

// Run executes the Generator. If generated artifacts already exist, a warning
// is logged but the file is not overwritten, unless force is true.
func (g *Generator) Run(scope *VersionScope) error {
	err := scope.validate()
	if err != nil {
		return err
	}

	// Derive data
	data := map[string]interface{}{}
	for fieldName, tmpl := range g.data {
		var buf bytes.Buffer
		err := tmpl.ExecuteTemplate(&buf, "include", scope)
		if err != nil {
			return fmt.Errorf("failed to resolve filename: %w (generators.%s.data.%s.include)", err, g.name, fieldName)
		}
		filename := strings.TrimSpace(buf.String())
		if g.debug {
			log.Printf("interpolated generators.%s.data.%s.include => %q", g.name, fieldName, filename)
		}
		contents, err := ioutil.ReadFile(filename)
		if err != nil {
			return fmt.Errorf("%w (generators.%s.data.%s.include)", err, g.name, fieldName)
		}
		fieldValue := map[string]interface{}{}
		switch filepath.Ext(filename) {
		case ".yaml":
			err = yaml.Unmarshal(contents, &fieldValue)
			if err != nil {
				return fmt.Errorf("failed to load %q: %w (generators.%s.data.%s.include)", filename, err, g.name, fieldName)
			}
		case ".json":
			err = json.Unmarshal(contents, &fieldValue)
			if err != nil {
				return fmt.Errorf("failed to load %q: %w (generators.%s.data.%s.include)", filename, err, g.name, fieldName)
			}
		default:
			return fmt.Errorf("don't know how to load %q: %w (generators.%s.data.%s.include)", filename, err, g.name, fieldName)
		}
		data[fieldName] = fieldValue
	}
	gsc := &versionScope{
		VersionScope: scope,
		Data:         data,
	}
	if g.files != nil {
		return g.runFiles(gsc)
	}
	return g.runFile(gsc)
}

func (g *Generator) runFile(scope *versionScope) error {
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

func (g *Generator) runFiles(scope *versionScope) error {
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
