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
	// Restore current working directory upon returning
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	defer func() {
		os.Chdir(cwd)
	}()

	// `cd` to the path containing the spec file, so that relative paths
	// resolve.
	specBase, specDir := filepath.Base(specFile), filepath.Dir(specFile)
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
	t, err := l.LoadFromFile(specBase)
	if err != nil {
		return nil, fmt.Errorf("failed to load %q: %w", specBase, err)
	}
	return t, nil
}

// ToSpecJSON renders an OpenAPI document object as JSON.
func ToSpecJSON(v interface{}) ([]byte, error) {
	return json.MarshalIndent(v, "", "  ")
}

// ToSpecYAML renders an OpenAPI document object as YAML.
func ToSpecYAML(v interface{}) ([]byte, error) {
	jsonBuf, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSON: %w", err)
	}
	return yaml.JSONToYAML(jsonBuf)
}
