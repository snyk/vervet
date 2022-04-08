// Package config supports configuring the Vervet Underground service.
package config

import (
	"encoding/json"
	"os"
)

// ServerConfig defines the configuration options for the Vervet Underground service.
type ServerConfig struct {
	Host     string            `json:"host"`
	Services []string          `json:"services"`
	Storage  map[string]string `json:"storage"`
}

// Load returns a ServerConfig instance loaded from the given path to a JSON
// config file.
func Load(configPath string) (*ServerConfig, error) {
	file, err := os.Open(configPath)
	var config ServerConfig
	if err != nil {
		return nil, err
	}
	defer file.Close()
	err = json.NewDecoder(file).Decode(&config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}
