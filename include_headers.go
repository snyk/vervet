package vervet

import (
	"encoding/json"
	"fmt"
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
	err := w.apply()
	if err != nil {
		return err
	}
	return doc.ResolveRefs()
}

type includeHeaders struct {
	relPath string
	doc     *Document
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

type includeHeadersRef struct {
	Ref   string           `json:"$ref"`
	Value openapi3.Headers `json:"-"`
}

func (w *includeHeaders) applyOperation(op *openapi3.Operation) error {
	if op == nil {
		return nil // nothing to do
	}
	for _, respRef := range op.Responses {
		resp := respRef.Value
		headersRefJson := resp.ExtensionProps.Extensions[ExtSnykIncludeHeaders]
		if headersRefJson == nil {
			continue
		}
		inclRef := &includeHeadersRef{Value: openapi3.Headers{}}
		err := json.Unmarshal(headersRefJson.(json.RawMessage), &inclRef)
		if err != nil {
			return err
		}
		relPath, err := w.doc.LoadReference(w.doc.RelativePath(), inclRef.Ref, &inclRef.Value)
		if err != nil {
			return fmt.Errorf("failed to load reference: %w", err)
		}

		if resp.Headers == nil {
			resp.Headers = openapi3.Headers{}
		}
		for headerKey, headerRef := range inclRef.Value {
			if _, ok := resp.Headers[headerKey]; ok {
				continue // Response's declared headers take precedence over includes.
			}
			headerRef.Ref = filepath.Join(relPath, headerRef.Ref)
			resp.Headers[headerKey] = headerRef
		}
	}
	return nil
}
