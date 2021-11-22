package config

import "fmt"

// Generators defines a named map of Generator instances.
type Generators map[string]*Generator

// Generator describes how files are generated for a resource.
type Generator struct {
	Name     string                    `json:"-"`
	Scope    GeneratorScope            `json:"scope"`
	Filename string                    `json:"filename,omitempty"`
	Template string                    `json:"template"`
	Files    string                    `json:"files,omitempty"`
	Data     map[string]*GeneratorData `json:"data,omitempty"`
}

func (g *Generator) validate() error {
	switch g.Scope {
	case GeneratorScopeVersion:
	//case GeneratorScopeResource:  // TODO: support resource scope
	default:
		return fmt.Errorf("invalid scope %q (generators.%s.scope)", g.Scope, g.Name)
	}
	if g.Template == "" {
		return fmt.Errorf("required field not specified (generators.%s.contents)", g.Name)
	}
	if g.Filename == "" && g.Files == "" {
		return fmt.Errorf("filename or files must be specified (generators.%s)", g.Name)
	}
	for k, v := range g.Data {
		if k == "" {
			return fmt.Errorf("empty key not allowed (generators.%s.data)", g.Name)
		}
		if v.Include == "" {
			return fmt.Errorf("required field not specified (generators.%s.data.%s.include)", g.Name, k)
		}
	}
	return nil
}

// GeneratorScope determines the template context when running the generator.
// Different scopes allow templates to operate over a single resource version,
// or all versions in a resource, for example.
type GeneratorScope string

const (
	// GeneratorScopeDefault indicates the default scope should be used in
	// configuration.
	GeneratorScopeDefault = ""

	// GeneratorScopeVersion indicates the generator operates on a single
	// resource version.
	GeneratorScopeVersion = "version"

	// GeneratorScopeResource indicates the generator operates on all versions
	// in a resource. This is useful for generating version routers, for
	// example.
	GeneratorScopeResource = "resource"
)

// GeneratorData describes an item that is added to a generator's template data
// context.
type GeneratorData struct {
	FieldName string `json:"-"`
	Include   string `json:"include"`
}

func (g Generators) init() error {
	for name, gen := range g {
		gen.Name = name
		if gen.Scope == GeneratorScopeDefault {
			gen.Scope = GeneratorScopeVersion
		}
		if err := gen.validate(); err != nil {
			return err
		}
	}
	return nil
}
