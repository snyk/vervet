package main

import (
	"fmt"
	"log"
	"os"

	"github.com/snyk/apiutil"
)

func main() {
	specFile := os.Args[1]
	t, err := apiutil.LoadSpecFile(specFile)
	if err != nil {
		log.Fatalf("failed to load spec from %q: %v", specFile, err)
	}

	// Localize all references, so we emit a completely self-contained OpenAPI document.
	err = apiutil.NewLocalizer(t).Localize()
	if err != nil {
		log.Fatalf("failed to localize refs: %v", err)
	}

	yamlBuf, err := apiutil.ToSpecYAML(t)
	if err != nil {
		log.Fatalf("failed to convert JSON to YAML: %v", err)
	}
	fmt.Printf(string(yamlBuf))
}
