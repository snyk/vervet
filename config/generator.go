package config

import "fmt"

// Generators defines a named map of Generator instances.
type Generators map[string]*Generator

// Generator describes how files are generated for a resource.
type Generator struct {
	Name      string         `json:"-"`
	Scope     GeneratorScope `json:"scope"`
	Filename  string         `json:"filename,omitempty"`
	Template  string         `json:"template"`
	Files     string         `json:"files,omitempty"`
	Functions string         `json:"functions,omitempty"`
}

func (g *Generator) validate() error {
	switch g.Scope {
	case GeneratorScopeVersion:
	case GeneratorScopeResource:
	default:
		return fmt.Errorf("invalid scope %q (generators.%s.scope)", g.Scope, g.Name)
	}
	if g.Template == "" {
		return fmt.Errorf("required field not specified (generators.%s.contents)", g.Name)
	}
	if g.Filename == "" && g.Files == "" {
		return fmt.Errorf("filename or files must be specified (generators.%s)", g.Name)
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
