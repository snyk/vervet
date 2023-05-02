package config

import (
	"fmt"
	"io"
	"sort"

	"github.com/ghodss/yaml"
)

// Project defines collection of APIs and the standards they adhere to.
type Project struct {
	Version string `json:"version"`
	// NOTE: Linters are deprecated and may be removed in v5.
	Linters    Linters    `json:"linters,omitempty"`
	Generators Generators `json:"generators,omitempty"`
	APIs       APIs       `json:"apis"`
}

// APINames returns the API names in deterministic ascending order.
func (p *Project) APINames() []string {
	result := []string{}
	for k := range p.APIs {
		result = append(result, k)
	}
	sort.Strings(result)
	return result
}

func (p *Project) init() {
	if p.Linters == nil {
		p.Linters = Linters{}
	}
	if p.Generators == nil {
		p.Generators = Generators{}
	}
	if p.APIs == nil {
		p.APIs = APIs{}
	}
}

func (p *Project) validate() error {
	if p.Version == "" {
		p.Version = "1"
	}
	if p.Version != "1" {
		return fmt.Errorf("unsupported version %q", p.Version)
	}
	err := p.Linters.init()
	if err != nil {
		return err
	}
	err = p.Generators.init()
	if err != nil {
		return err
	}
	err = p.APIs.init(p)
	if err != nil {
		return err
	}
	return nil
}

// Load loads a Project configuration from its YAML representation.
func Load(r io.Reader) (*Project, error) {
	var p Project
	buf, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read project configuration: %w", err)
	}
	err = yaml.Unmarshal(buf, &p)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal project configuration: %w", err)
	}
	p.init()
	return &p, p.validate()
}

// LoadGenerators loads Generators from their YAML representation.
func LoadGenerators(r io.Reader) (Generators, error) {
	var g Generators
	buf, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read generators: %w", err)
	}
	err = yaml.Unmarshal(buf, &g)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal generators: %w", err)
	}
	return g, g.init()
}

// Save saves a Project configuration to YAML.
func Save(w io.Writer, proj *Project) error {
	buf, err := yaml.Marshal(proj)
	if err != nil {
		return err
	}
	_, err = w.Write(buf)
	return err
}
