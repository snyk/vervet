package config

import (
	"encoding/json"
	"os"

	"vervet-underground/lib"
)

func Load(configPath string) (*lib.ServerConfig, error) {
	file, err := os.Open(configPath)
	var config lib.ServerConfig
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