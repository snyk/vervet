package main

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/ghodss/yaml"
)

func main() {
	schemaFile := os.Args[1]
	t, err := loadSchemaFile(schemaFile)
	if err != nil {
		log.Fatalf("failed to load schema from %q: %v", schemaFile, err)
	}

	err = printSchemaYAML(t)
	if err != nil {
		log.Fatalf("failed to print schema: %v", err)
	}
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

func loadSchemaFile(schemaFile string) (*openapi3.T, error) {
	var err error
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

	refUrlStr := "file://" + schemaDir
	refUrl, err := url.Parse(refUrlStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL %q: %w", refUrlStr, err)
	}
	err = l.ResolveRefsIn(t, refUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve refs in OpenAPI spec: %w", err)
	}
	return t, nil
}
