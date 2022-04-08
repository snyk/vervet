package config_test

import (
	"os"
	"testing"

	qt "github.com/frankban/quicktest"

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
			"services": ["localhost"]
		}`))

		conf, err := config.Load(f.Name())
		c.Assert(err, qt.IsNil)

		expected := config.ServerConfig{
			Host:     "0.0.0.0",
			Services: []string{"localhost"},
			Storage: config.StorageConfig{
				Type: config.StorageTypeMemory,
			},
		}
		c.Assert(*conf, qt.DeepEquals, expected)
	})

	c.Run("s3 config", func(c *qt.C) {
		f := createTestFile(c, []byte(`{
			"host": "0.0.0.0",
			"services": ["localhost"],
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
			Services: []string{"localhost"},
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
			"services": ["localhost"],
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
			Services: []string{"localhost"},
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
			"services": ["http://user:password@localhost"],
			"storage": {
				"type": "memory"
			}
		}`))

		conf, err := config.Load(defaultConfig.Name(), secretConfig.Name())
		c.Assert(err, qt.IsNil)

		expected := config.ServerConfig{
			Host:     "0.0.0.0",
			Services: []string{"http://user:password@localhost"},
			Storage: config.StorageConfig{
				Type: config.StorageTypeMemory,
			},
		}
		c.Assert(*conf, qt.DeepEquals, expected)
	})
}
