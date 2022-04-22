// Package config supports configuring the Vervet Underground service.
package config

import (
	"github.com/spf13/viper"
)

// StorageType describes backend implementations supported by Vervet Underground.
type StorageType string

const (
	StorageTypeMemory StorageType = "memory"
	StorageTypeS3     StorageType = "s3"
	StorageTypeGCS    StorageType = "gcs"
)

// ServerConfig defines the configuration options for the Vervet Underground service.
type ServerConfig struct {
	Host     string
	Services []string
	Storage  StorageConfig
}

type StorageConfig struct {
	Type StorageType
	S3   S3Config
	GCS  GcsConfig
}

type S3Config struct {
	Region     string
	Endpoint   string
	AccessKey  string
	SecretKey  string
	SessionKey string
}

type GcsConfig struct {
	Region    string
	Endpoint  string
	ProjectId string
	Filename  string
}

// setDefaults sets default values for the ServerConfig.
func setDefaults() {
	viper.SetDefault("host", "localhost")
	viper.SetDefault("storage.type", StorageTypeMemory)
}

// Load returns a ServerConfig instance loaded from the given paths to a JSON
// config file.
func Load(configPaths ...string) (*ServerConfig, error) {
	setDefaults()

	for i, c := range configPaths {
		if i == 0 {
			viper.SetConfigFile(c)
			if err := viper.ReadInConfig(); err != nil {
				return nil, err
			}
		} else {
			viper.SetConfigFile(c)
			if err := viper.MergeInConfig(); err != nil {
				return nil, err
			}
		}
	}

	var config ServerConfig
	err := viper.Unmarshal(&config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}
