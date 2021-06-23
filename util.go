package vervet

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/ghodss/yaml"
	"github.com/google/uuid"
)

func init() {
	// Necessary for `format: uuid` to validate.
	openapi3.DefineStringFormatCallback("uuid", func(v string) error {
		_, err := uuid.Parse(v)
		return err
	})
}

// LoadSpecFile loads an OpenAPI spec file from the given path,
// returning a document object.
func LoadSpecFile(specFile string) (*openapi3.T, error) {
	var err error
	// `cd` to the path containing the spec file, so that relative paths
	// resolve.
	specDir := filepath.Dir(specFile)
	specDir, err = filepath.Abs(specDir)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}
	err = os.Chdir(specDir)
	if err != nil {
		return nil, fmt.Errorf("failed to chdir %q: %w", specDir, err)
	}

	l := openapi3.NewLoader()
	l.IsExternalRefsAllowed = true
	t, err := l.LoadFromFile(specFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load %q: %w", specFile, err)
	}
	return t, nil
}

// ToSpecYAML renders an OpenAPI document object as YAML.
func ToSpecYAML(v interface{}) ([]byte, error) {
	jsonBuf, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSON: %w", err)
	}
	return yaml.JSONToYAML(jsonBuf)
}
