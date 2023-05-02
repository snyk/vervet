package vervet

import (
	"fmt"
	"net/url"
	"path/filepath"

	"github.com/getkin/kin-openapi/openapi3"
)

const (
	// ExtSnykIncludeHeaders is used to annotate a response with a list of
	// headers. While OpenAPI supports header references, it does not yet
	// support including a collection of common headers. This extension is used
	// by vervet to include headers from a referenced document when compiling
	// OpenAPI specs.
	ExtSnykIncludeHeaders = "x-snyk-include-headers"
)

// IncludeHeaders adds response headers included with the ExtSnykIncludeHeaders
// extension property.
func IncludeHeaders(doc *Document) error {
	w := &includeHeaders{doc: doc}
	if err := w.apply(); err != nil {
		return err
	}
	return doc.ResolveRefs()
}

type includeHeaders struct {
	doc *Document
}

func (w *includeHeaders) apply() error {
	for _, pathItem := range w.doc.Paths {
		if err := w.applyOperation(pathItem.Connect); err != nil {
			return err
		}
		if err := w.applyOperation(pathItem.Delete); err != nil {
			return err
		}
		if err := w.applyOperation(pathItem.Get); err != nil {
			return err
		}
		if err := w.applyOperation(pathItem.Head); err != nil {
			return err
		}
		if err := w.applyOperation(pathItem.Options); err != nil {
			return err
		}
		if err := w.applyOperation(pathItem.Patch); err != nil {
			return err
		}
		if err := w.applyOperation(pathItem.Post); err != nil {
			return err
		}
		if err := w.applyOperation(pathItem.Put); err != nil {
			return err
		}
	}
	return nil
}

func (w *includeHeaders) applyOperation(op *openapi3.Operation) error {
	if op == nil {
		return nil // nothing to do
	}
	for _, respRef := range op.Responses {
		resp := respRef.Value
		headersContents, ok := resp.Extensions[ExtSnykIncludeHeaders].(map[string]interface{})
		if !ok {
			continue
		}
		ref, ok := headersContents["$ref"].(string)
		if !ok {
			continue
		}
		val := openapi3.Headers{}
		relPath, err := w.doc.LoadReference(w.doc.RelativePath(), ref, &val)
		if err != nil {
			return fmt.Errorf("failed to load reference: %w", err)
		}

		if resp.Headers == nil {
			resp.Headers = openapi3.Headers{}
		}
		for headerKey, headerRef := range val {
			if _, ok := resp.Headers[headerKey]; ok {
				continue // Response's declared headers take precedence over includes.
			}
			if isRelativePath(headerRef.Ref) {
				headerRef.Ref = filepath.Join(relPath, headerRef.Ref)
			}
			resp.Headers[headerKey] = headerRef
		}
		// Remove the extension once it has been processed
		delete(resp.Extensions, ExtSnykIncludeHeaders)
	}
	return nil
}

func isRelativePath(s string) bool {
	if u, err := url.Parse(s); err == nil && u.Scheme != "" {
		return false
	}
	return !filepath.IsAbs(s)
}
