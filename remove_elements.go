package vervet

import (
	"reflect"
	"regexp"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/mitchellh/reflectwalk"
)

// ExcludePatterns defines patterns matching elements to be removed from an
// OpenAPI document.
type ExcludePatterns struct {
	ExtensionPatterns []string
	HeaderPatterns    []string
	Paths             []string
}

type excluder struct {
	doc *openapi3.T

	extensionPatterns []*regexp.Regexp
	headerPatterns    []*regexp.Regexp
	paths             []string
}

// RemoveElements removes those elements from an OpenAPI document matching the
// given exclude patterns.
func RemoveElements(doc *openapi3.T, excludes ExcludePatterns) error {
	ex := &excluder{
		doc:               doc,
		extensionPatterns: make([]*regexp.Regexp, len(excludes.ExtensionPatterns)),
		headerPatterns:    make([]*regexp.Regexp, len(excludes.HeaderPatterns)),
		paths:             excludes.Paths,
	}
	for i, pat := range excludes.ExtensionPatterns {
		re, err := regexp.Compile(pat)
		if err != nil {
			return err
		}
		ex.extensionPatterns[i] = re
	}
	for i, pat := range excludes.HeaderPatterns {
		re, err := regexp.Compile(pat)
		if err != nil {
			return err
		}
		ex.headerPatterns[i] = re
	}
	// Remove excluded paths
	excludedPaths := map[string]struct{}{}
	for path := range doc.Paths.Map() {
		if ex.isExcludedPath(path) {
			excludedPaths[path] = struct{}{}
		}
	}
	for path := range excludedPaths {
		// Dangerous, relies on the fact Map returns a reference to the
		// underlying map. There doesn't seem to be a safe way to delete a
		// PathItem from a Paths object.
		delete(doc.Paths.Map(), path)
	}
	// Remove excluded elements
	if err := ex.apply(); err != nil {
		return err
	}
	return nil
}

func (ex *excluder) apply() error {
	return reflectwalk.Walk(ex.doc, ex)
}

// Struct implements reflectwalk.StructWalker.
func (ex *excluder) Struct(v reflect.Value) error {
	if !v.CanInterface() {
		return nil
	}

	switch v.Interface().(type) {
	case openapi3.Operation:
		ex.applyOperation(v.Addr().Interface().(*openapi3.Operation))
	}

	return nil
}

// StructField implements reflectwalk.StructWalker.
func (ex *excluder) StructField(field reflect.StructField, v reflect.Value) error {
	if field.Name != "Extensions" || !v.CanInterface() {
		return nil
	}

	switch v.Interface().(type) {
	case map[string]interface{}:
		ex.applyExtensions(v.Addr().Interface().(*map[string]interface{}))
	}

	return nil
}

func (ex *excluder) applyExtensions(extensions *map[string]interface{}) {
	exts := make(map[string]interface{}, len(*extensions))
	for k, v := range *extensions {
		if !ex.isExcludedExtension(k) {
			exts[k] = v
		}
	}
	*extensions = exts
}

func (ex *excluder) applyOperation(op *openapi3.Operation) {
	var params []*openapi3.ParameterRef
	for _, p := range op.Parameters {
		if !ex.isExcludedHeaderParam(p) {
			params = append(params, p)
		}
	}
	op.Parameters = params

	for _, resp := range op.Responses.Map() {
		if resp.Value == nil {
			continue
		}
		headers := openapi3.Headers{}
		for headerName, header := range resp.Value.Headers {
			var matched bool
			for _, re := range ex.headerPatterns {
				if re.MatchString(headerName) {
					matched = true
					break
				}
			}
			if !matched {
				headers[headerName] = header
			}
		}
		resp.Value.Headers = headers
	}
}

func (ex *excluder) isExcludedExtension(name string) bool {
	for _, re := range ex.extensionPatterns {
		if re.MatchString(name) {
			return true
		}
	}
	return false
}

func (ex *excluder) isExcludedPath(path string) bool {
	for _, matchPath := range ex.paths {
		if matchPath == path {
			return true
		}
	}
	return false
}

func (ex *excluder) isExcludedHeaderParam(p *openapi3.ParameterRef) bool {
	if p.Value == nil {
		return false
	}
	if p.Value.In != "header" {
		return false
	}
	for _, re := range ex.headerPatterns {
		if re.MatchString(p.Value.Name) {
			return true
		}
	}
	return false
}
