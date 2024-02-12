package config_test

import (
	"os"
	"testing"

	qt "github.com/frankban/quicktest"
	"github.com/snyk/vervet/v6"

	"vervet-underground/config"
)

func createTestFile(c *qt.C, data []byte) *os.File {
	f, err := os.CreateTemp(c.TempDir(), "tmpfile-*.json")
	c.Assert(err, qt.IsNil)

	defer f.Close()
	_, err = f.Write(data)
	c.Assert(err, qt.IsNil)
	return f
}

func TestLoad(t *testing.T) {
	c := qt.New(t)

	c.Run("empty with defaults", func(c *qt.C) {
		f := createTestFile(c, []byte(`{}`))

		conf, err := config.Load(f.Name())
		c.Assert(err, qt.IsNil)

		expected := config.ServerConfig{
			Host:     "localhost",
			Services: nil,
			Storage: config.StorageConfig{
				Type: config.StorageTypeMemory,
			},
		}
		c.Assert(*conf, qt.DeepEquals, expected)
	})

	c.Run("specify config", func(c *qt.C) {
		f := createTestFile(c, []byte(`{
			"host": "0.0.0.0",
			"services": [{"url":"localhost","name":"localhost"}]
		}`))

		conf, err := config.Load(f.Name())
		c.Assert(err, qt.IsNil)

		expected := config.ServerConfig{
			Host:     "0.0.0.0",
			Services: []config.ServiceConfig{{URL: "localhost", Name: "localhost"}},
			Storage: config.StorageConfig{
				Type: config.StorageTypeMemory,
			},
		}
		c.Assert(*conf, qt.DeepEquals, expected)
	})

	c.Run("specify excludes", func(c *qt.C) {
		f := createTestFile(c, []byte(`{
			"merging": {
				"excludePatterns": {
					"extensionPatterns": "^x-snyk-.*",
					"headerPatterns": ".*-internal$"
				}
			}
		}`))

		conf, err := config.Load(f.Name())
		c.Assert(err, qt.IsNil)

		expected := config.ServerConfig{
			Host:     "localhost",
			Services: nil,
			Merging: config.MergeConfig{
				ExcludePatterns: vervet.ExcludePatterns{
					ExtensionPatterns: []string{"^x-snyk-.*"},
					HeaderPatterns:    []string{".*-internal$"},
				},
			},
			Storage: config.StorageConfig{
				Type: config.StorageTypeMemory,
			},
		}
		c.Assert(*conf, qt.DeepEquals, expected)
	})

	c.Run("s3 config", func(c *qt.C) {
		f := createTestFile(c, []byte(`{
			"host": "0.0.0.0",
			"services": [{"url":"localhost","name":"localhost"}],
			"storage": {
				"type": "s3",
				"s3": {
					"region": "us-east-2",
					"endpoint": "http://test",
					"accessKey": "access-key",
					"secretKey": "secret-key",
					"sessionKey": "session-key"
				}
			}
		}`))

		conf, err := config.Load(f.Name())
		c.Assert(err, qt.IsNil)

		expected := config.ServerConfig{
			Host:     "0.0.0.0",
			Services: []config.ServiceConfig{{URL: "localhost", Name: "localhost"}},
			Storage: config.StorageConfig{
				Type: config.StorageTypeS3,
				S3: config.S3Config{
					Region:     "us-east-2",
					Endpoint:   "http://test",
					AccessKey:  "access-key",
					SecretKey:  "secret-key",
					SessionKey: "session-key",
				},
			},
		}
		c.Assert(*conf, qt.DeepEquals, expected)
	})

	c.Run("gcs config", func(c *qt.C) {
		f := createTestFile(c, []byte(`{
			"host": "0.0.0.0",
			"services": [{"url":"localhost","name":"localhost"}],
			"storage": {
				"type": "gcs",
				"gcs": {
				  "region": "US-EAST1",
				  "endpoint": "http://fake-gcs:4443",
				  "projectId": "test",
				  "filename": "test"
				}
			}
		}`))

		conf, err := config.Load(f.Name())
		c.Assert(err, qt.IsNil)

		expected := config.ServerConfig{
			Host:     "0.0.0.0",
			Services: []config.ServiceConfig{{URL: "localhost", Name: "localhost"}},
			Storage: config.StorageConfig{
				Type: config.StorageTypeGCS,
				GCS: config.GcsConfig{
					Region:    "US-EAST1",
					Endpoint:  "http://fake-gcs:4443",
					Filename:  "test",
					ProjectId: "test",
				},
			},
		}
		c.Assert(*conf, qt.DeepEquals, expected)
	})

	c.Run("multiple configs", func(c *qt.C) {
		defaultConfig := createTestFile(c, []byte(`{
			"host": "0.0.0.0",
			"storage": {
				"type": "overwrite"
			}
		}`))
		secretConfig := createTestFile(c, []byte(`{
			"services": [{"url":"http://user:password@localhost","name":"localhost"}],
			"storage": {
				"type": "memory"
			}
		}`))

		conf, err := config.Load(defaultConfig.Name(), secretConfig.Name())
		c.Assert(err, qt.IsNil)

		expected := config.ServerConfig{
			Host:     "0.0.0.0",
			Services: []config.ServiceConfig{{URL: "http://user:password@localhost", Name: "localhost"}},
			Storage: config.StorageConfig{
				Type: config.StorageTypeMemory,
			},
		}
		c.Assert(*conf, qt.DeepEquals, expected)
	})

	c.Run("invalid service config - no name", func(c *qt.C) {
		cfg := createTestFile(c, []byte(`{
			"host": "0.0.0.0",
			"services": [{"url":"http://user:password@localhost"}],
			"storage": {
				"type": "memory"
			}
		}`))
		_, err := config.Load(cfg.Name())
		c.Assert(err, qt.ErrorMatches, `missing service name`)
	})

	c.Run("invalid service config - duplicate name", func(c *qt.C) {
		cfg := createTestFile(c, []byte(`{
			"host": "0.0.0.0",
			"services": [{"url":"http://service-a","name":"service-a"},{"url":"http://service-a","name":"service-a"}],
			"storage": {
				"type": "memory"
			}
		}`))
		_, err := config.Load(cfg.Name())
		c.Assert(err, qt.ErrorMatches, `duplicate service name "service-a"`)
	})
}
