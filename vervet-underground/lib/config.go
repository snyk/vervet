package lib

import (
	"encoding/json"
	"os"
)

func Decode(configPath string, config *ServerConfig) error {
	file, err := os.Open(configPath)
	if err != nil {
		return err
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		return err
	}
	return nil
}