package config

import (
	"fmt"
	"os"
)

// FromFile loads a vervet Project from a vervet.yaml file at given configPath,
// defaults to ".vervet.yaml".
func FromFile(configPath string) (*Project, error) {
	if configPath == "" {
		configPath = ".vervet.yaml"
	}
	f, err := os.Open(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open %q: %w", configPath, err)
	}
	defer f.Close()
	return Load(f)
}
