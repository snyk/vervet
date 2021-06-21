package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"reflect"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/ghodss/yaml"
	"github.com/mitchellh/reflectwalk"
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
	err = removeRefs(t)
	if err != nil {
		log.Fatalf("failed to remove refs: %v", err)
	}

	err = printSchemaYAML(t)
	if err != nil {
		log.Fatalf("failed to print schema: %v", err)
	}
}

type refRemover struct{}

func (*refRemover) Struct(reflect.Value) error {
	return nil
}
func (*refRemover) StructField(sf reflect.StructField, v reflect.Value) error {
	// Note this unconditionally clears the $ref URI, inlining all referenced
	// components. We might be more selective about this by adding a bit of
	// state to the walker.
	// TODO: might also check for a `json:"$ref"` tag on the field.
	if sf.Name == "Ref" && sf.Type.Kind() == reflect.String {
		v.Set(reflect.ValueOf(""))
	}
	return nil
}

func removeRefs(t *openapi3.T) error {
	return reflectwalk.Walk(t, &refRemover{})
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
