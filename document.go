package vervet

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"

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
	openapi3.DefineStringFormatCallback("url", func(v string) error {
		_, err := url.Parse(v)
		return err
	})
}

// Document is an OpenAPI 3 document object model.
type Document struct {
	*openapi3.T
	path string
	url  *url.URL
}

// NewDocumentFile loads an OpenAPI spec file from the given file path,
// returning a document object.
func NewDocumentFile(specFile string) (_ *Document, returnErr error) {
	// Restore current working directory upon returning
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	defer func() {
		err := os.Chdir(cwd)
		if err != nil {
			log.Println("warning: failed to restore working directory: %w", err)
			if returnErr == nil {
				returnErr = err
			}
		}
	}()

	specFile, err = filepath.Abs(specFile)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	// `cd` to the path containing the spec file, so that relative paths
	// resolve.
	specBase, specDir := filepath.Base(specFile), filepath.Dir(specFile)
	err = os.Chdir(specDir)
	if err != nil {
		return nil, fmt.Errorf("failed to chdir %q: %w", specDir, err)
	}

	specURL, err := url.Parse(specFile)
	if err != nil {
		return nil, err
	}

	var t openapi3.T
	contents, err := ioutil.ReadFile(specFile)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(contents, &t)
	if err != nil {
		return nil, err
	}
	resolver, err := newRefAliasResolver(&t)
	if err != nil {
		return nil, err
	}
	err = resolver.resolve()
	if err != nil {
		return nil, err
	}

	l := openapi3.NewLoader()
	l.IsExternalRefsAllowed = true
	err = l.ResolveRefsIn(&t, specURL)
	if err != nil {
		return nil, fmt.Errorf("failed to load %q: %w", specBase, err)
	}
	return &Document{
		T:    &t,
		path: specFile,
		url:  specURL,
	}, nil
}

// MarshalJSON implements json.Marshaler.
func (d *Document) MarshalJSON() ([]byte, error) {
	return d.T.MarshalJSON()
}

// RelativePath returns the relative path for resolving references from the
// file path location of the top-level document: the directory which contains
// the file from which the top-level document was loaded.
func (d *Document) RelativePath() string {
	return filepath.Dir(d.path)
}

// Location returns the URL from where the document was loaded.
func (d *Document) Location() *url.URL {
	return d.url
}

// ResolveRefs resolves all Ref types in the document, causing the Value field
// of each Ref to be loaded and populated from its referenced location.
func (d *Document) ResolveRefs() error {
	l := openapi3.NewLoader()
	l.IsExternalRefsAllowed = true
	return l.ResolveRefsIn(d.T, d.url)
}

// LoadReference loads a reference from refPath, relative to relPath, into
// target. The relative path of the reference is returned, so that references
// may be chain-loaded with successive calls.
func (d *Document) LoadReference(relPath, refPath string, target interface{}) (_ string, returnErr error) {
	refUrl, err := url.Parse(refPath)
	if err != nil {
		return "", err
	}
	if refUrl.Scheme == "" || refUrl.Scheme == "file" {
		refPath, err = filepath.Abs(filepath.Join(relPath, refUrl.Path))
		if err != nil {
			return "", err
		}
		refUrl.Path = refPath
	}

	// Parse and load the contents of the referenced document.
	l := openapi3.NewLoader()
	l.IsExternalRefsAllowed = true
	contents, err := openapi3.DefaultReadFromURI(l, refUrl)
	if err != nil {
		return "", fmt.Errorf("failed to read %q: %w", refUrl, err)
	}
	// If the reference is to an element in the referenced document, further resolve that.
	if refUrl.Fragment != "" {
		parts := strings.Split(refUrl.Fragment, "/")
		// TODO: support actual jsonpaths if/when needed. For now only
		// top-level properties are supported.
		if parts[0] != "" || len(parts) > 2 {
			return "", fmt.Errorf("URL %q not supported", refUrl)
		}
		elements := map[string]interface{}{}
		err := yaml.Unmarshal(contents, &elements)
		if err != nil {
			return "", err
		}
		elementDoc, ok := elements[parts[1]]
		if !ok {
			return "", fmt.Errorf("element %q not found in %q", parts[1], refUrl.Path)
		}
		contents, err = json.Marshal(elementDoc)
		if err != nil {
			return "", err
		}
	}

	// Unmarshal the resolved reference into target object.
	err = yaml.Unmarshal(contents, target)
	if err != nil {
		return "", err
	}

	return filepath.Abs(filepath.Dir(refUrl.Path))
}

// Version returns the version of the document.
func (d *Document) Version() (Version, error) {
	vs, err := ExtensionString(d.ExtensionProps, ExtSnykApiVersion)
	if err != nil {
		return Version{}, err
	}
	return ParseVersion(vs)
}

// Lifecycle returns the lifecycle of the document.
func (d *Document) Lifecycle() (Lifecycle, error) {
	ls, err := ExtensionString(d.ExtensionProps, ExtSnykApiLifecycle)
	if err != nil {
		if IsExtensionNotFound(err) {
			// If it's not marked as deprecated or sunset, assume it's released.
			return LifecycleReleased, err
		}
		return lifecycleUndefined, err
	}
	return ParseLifecycle(ls)
}
