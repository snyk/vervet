package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/ghodss/yaml"

	"github.com/snyk/apiutil"
)

func main() {
	schemaFile := os.Args[1]
	t, err := loadSchemaFile(schemaFile)
	if err != nil {
		log.Fatalf("failed to load schema from %q: %v", schemaFile, err)
	}

	// kin-openapi has a preference for displaying refs rather than inlining types,
	// even though the references are followed and parsed.
	//
	// Clearing the Ref field causes the referenced component to be inlined
	// when the OpenAPI document root is marshaled.
	err = apiutil.NewLocalizer(t).Localize()
	if err != nil {
		log.Fatalf("failed to localize refs: %v", err)
	}

	err = printSchemaYAML(t)
	if err != nil {
		log.Fatalf("failed to print schema: %v", err)
	}
}

func loadSchemaFile(schemaFile string) (*openapi3.T, error) {
	var err error
	// `cd` to the path containing the schema file, so that relative paths
	// resolve.
	schemaDir := filepath.Dir(schemaFile)
	schemaDir, err = filepath.Abs(schemaDir)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}
	err = os.Chdir(schemaDir)
	if err != nil {
		return nil, fmt.Errorf("failed to chdir %q: %w", schemaDir, err)
	}

	l := openapi3.NewLoader()
	l.IsExternalRefsAllowed = true
	t, err := l.LoadFromFile(schemaFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load %q: %w", schemaFile, err)
	}
	return t, nil
}

func printSchemaYAML(t *openapi3.T) error {
	jsonBuf, err := t.MarshalJSON()
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	yamlBuf, err := yaml.JSONToYAML(jsonBuf)
	if err != nil {
		return fmt.Errorf("failed to convert JSON to YAML: %w", err)
	}
	_, err = fmt.Printf(string(yamlBuf))
	return err
}
