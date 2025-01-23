package cmd_test

import (
	"os"
	"path/filepath"
	"testing"

	qt "github.com/frankban/quicktest"
	"github.com/getkin/kin-openapi/openapi3"
	"gopkg.in/yaml.v3"

	"github.com/snyk/vervet/v8"
	"github.com/snyk/vervet/v8/internal/cmd"
)

func TestSanitize(t *testing.T) {
	c := qt.New(t)

	// Create temp directories for source and output
	srcDir := c.TempDir()
	outDir := c.TempDir()

	// Create test OpenAPI specs with sensitive data
	tests := []struct {
		version string
		spec    *openapi3.T
	}{
		{
			version: "2024-01-15",
			spec: &openapi3.T{
				OpenAPI: "3.0.0",
				Info: &openapi3.Info{
					Title:   "Test API",
					Version: "1.0.0",
				},
				Paths: func() *openapi3.Paths {
					paths := &openapi3.Paths{}
					paths.Set("/public", &openapi3.PathItem{
						Get: &openapi3.Operation{
							OperationID: "getPublic",
							Extensions: map[string]interface{}{
								"x-internal-op": "should-be-removed",
								"x-public-op":   "should-remain",
							},
							Parameters: []*openapi3.ParameterRef{
								{
									Value: &openapi3.Parameter{
										Name:        "X-Internal-Header",
										In:          "header",
										Description: "Should be removed",
									},
								},
								{
									Value: &openapi3.Parameter{
										Name:        "X-Public-Header",
										In:          "header",
										Description: "Should remain",
									},
								},
							},
						},
					})
					paths.Set("/internal", &openapi3.PathItem{
						Get: &openapi3.Operation{
							OperationID: "getInternal",
						},
					})
					return paths
				}(),
			},
		},
	}

	// Write test specs to source directory
	for _, test := range tests {
		versionDir := filepath.Join(srcDir, test.version)
		err := os.MkdirAll(versionDir, 0755)
		c.Assert(err, qt.IsNil)

		specBytes, err := yaml.Marshal(test.spec)
		c.Assert(err, qt.IsNil)

		err = os.WriteFile(filepath.Join(versionDir, "spec.yaml"), specBytes, 0644)
		c.Assert(err, qt.IsNil)
	}

	// Run sanitize command
	err := cmd.Vervet.Run([]string{
		"vervet",
		"sanitize",
		"--compiled-path", srcDir,
		"--out", outDir,
		"--exclude-extension", "^x-internal-.*$",
		"--exclude-header", "^X-Internal-.*$",
		"--exclude-path", "/internal",
	})
	c.Assert(err, qt.IsNil)

	// Verify output
	for _, test := range tests {
		c.Run("sanitized version "+test.version, func(c *qt.C) {
			outPath := filepath.Join(outDir, test.version, "openapi.yaml")
			doc, err := vervet.NewDocumentFile(outPath)
			c.Assert(err, qt.IsNil)

			// Check that internal path was removed
			internalPath := doc.Paths.Find("/internal")
			c.Assert(internalPath, qt.IsNil, qt.Commentf("Internal path should be removed"))

			// Check that public path and its proper operations remain
			publicPath := doc.Paths.Find("/public")
			c.Assert(publicPath, qt.Not(qt.IsNil), qt.Commentf("Public path should remain"))
			c.Assert(publicPath.Get, qt.Not(qt.IsNil))

			// Check operation extensions
			_, hasInternalOpExt := publicPath.Get.Extensions["x-internal-op"]
			c.Assert(hasInternalOpExt, qt.IsFalse, qt.Commentf("Internal operation extension should be removed"))

			_, hasPublicOpExt := publicPath.Get.Extensions["x-public-op"]
			c.Assert(hasPublicOpExt, qt.IsTrue, qt.Commentf("Public operation extension should remain"))

			// Check headers in parameters
			var foundInternalHeader, foundPublicHeader bool
			for _, param := range publicPath.Get.Parameters {
				if param.Value.In == "header" {
					if param.Value.Name == "X-Internal-Header" {
						foundInternalHeader = true
					}
					if param.Value.Name == "X-Public-Header" {
						foundPublicHeader = true
					}
				}
			}
			c.Assert(foundInternalHeader, qt.IsFalse, qt.Commentf("Internal header parameter should be removed"))
			c.Assert(foundPublicHeader, qt.IsTrue, qt.Commentf("Public header parameter should remain"))
		})
	}
}
