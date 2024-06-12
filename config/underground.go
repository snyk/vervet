// Package config supports configuring the Vervet Underground service.
package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"

	"github.com/snyk/vervet/v7"
)

// StorageType describes backend implementations supported by Vervet Underground.
type StorageType string

const (
	StorageTypeDisk StorageType = "disk"
	StorageTypeS3   StorageType = "s3"
	StorageTypeGCS  StorageType = "gcs"
)

// ServerConfig defines the configuration options for the Vervet Underground service.
type ServerConfig struct {
	Host     string
	Services []ServiceConfig
	Storage  StorageConfig
	Merging  MergeConfig
}

// ServiceFilter provides a map of service names to quickly filter old services.
func (c *ServerConfig) ServiceFilter() map[string]bool {
	services := make(map[string]bool)
	for _, s := range c.Services {
		services[s.Name] = true
	}
	return services
}

func (c *ServerConfig) validate() error {
	serviceNames := map[string]struct{}{}
	for _, svc := range c.Services {
		if svc.Name == "" {
			return fmt.Errorf("missing service name")
		}
		if _, ok := serviceNames[svc.Name]; ok {
			return fmt.Errorf("duplicate service name %q", svc.Name)
		}
		serviceNames[svc.Name] = struct{}{}
	}
	return nil
}

// ServiceConfig defines configuration options on a service.
type ServiceConfig struct {
	Name string
	URL  string
}

// MergeConfig contains configuration options defining how to merge OpenAPI
// documents when collating aggregate OpenAPI specifications across all
// services.
type MergeConfig struct {
	ExcludePatterns vervet.ExcludePatterns
}

// StorageConfig defines the configuration options for storage.
// The value of Type determines which of S3 or GCS will be used.
type StorageConfig struct {
	Type           StorageType
	BucketName     string
	IamRoleEnabled bool
	S3             S3Config
	GCS            GcsConfig
	Disk           DiskConfig
}

// DiskConfig defines configuration options for local disk storage.
type DiskConfig struct {
	Path string
}

// S3Config defines configuration options for AWS S3 storage.
type S3Config struct {
	Region     string
	Endpoint   string
	AccessKey  string
	SecretKey  string
	SessionKey string
}

// GcsConfig defines configuration options for Google Cloud Storage (GCS).
type GcsConfig struct {
	Region    string
	Endpoint  string
	ProjectId string
	Filename  string
}

// setDefaults sets default values for the ServerConfig.
func setDefaults() {
	viper.SetDefault("host", "localhost")
	viper.SetDefault("storage.type", StorageTypeDisk)
}

// loadEnv sets up the config store to load values from environment variables,
// these will take precedent over values defined in config files.
func loadEnv() {
	viper.SetEnvPrefix("SNYK")
	// Set nested values, eg SNYK_STORAGE_S3_REGION=
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()
}

// LoadServerConfig returns a ServerConfig instance loaded from the given paths
// to a JSON config file.
func LoadServerConfig(configPaths ...string) (*ServerConfig, error) {
	setDefaults()
	loadEnv()

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

	err = config.validate()
	if err != nil {
		return nil, err
	}
	return &config, nil
}
