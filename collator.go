package vervet

import (
	"fmt"
	"sort"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"go.uber.org/multierr"
)

// Collator merges resource versions into a single OpenAPI document.
type Collator struct {
	result           *openapi3.T
	componentSources map[string]string
	pathSources      map[string]string
	tagSources       map[string]string
}

// NewCollator returns a new Collator instance.
func NewCollator() *Collator {
	return &Collator{
		componentSources: map[string]string{},
		pathSources:      map[string]string{},
		tagSources:       map[string]string{},
	}
}

// Result returns the merged result. If no versions have been merged, returns
// nil.
func (c *Collator) Result() *openapi3.T {
	return c.result
}

// Collate merges a resource version into the current result.
func (c *Collator) Collate(rv *ResourceVersion) error {
	var errs error
	if c.result == nil {
		c.result = &openapi3.T{}
	}

	err := rv.cleanRefs()
	if err != nil {
		return err
	}

	err = c.mergeComponents(rv)
	if err != nil {
		errs = multierr.Append(errs, err)
	}
	mergeInfo(c.result, rv.T, false)
	err = c.mergePaths(rv)
	if err != nil {
		errs = multierr.Append(errs, err)
	}
	mergeSecurityRequirements(c.result, rv.T, false)
	mergeServers(c.result, rv.T, false)
	err = c.mergeTags(rv)
	if err != nil {
		errs = multierr.Append(errs, err)
	}
	mergeOpenAPIVersion(c.result, rv.T, false)
	return errs
}

func (c *Collator) mergeTags(rv *ResourceVersion) error {
	m := map[string]*openapi3.Tag{}
	for _, t := range c.result.Tags {
		m[t.Name] = t
	}
	var errs error
	for _, t := range rv.T.Tags {
		if current, ok := m[t.Name]; ok && !tagsEqual(current, t) {
			errs = multierr.Append(errs, fmt.Errorf("conflict in #/tags %s: %s and %s differ", t.Name, rv.path, c.tagSources[t.Name]))
		} else {
			m[t.Name] = t
			c.tagSources[t.Name] = rv.path
		}
	}
	if errs != nil {
		return errs
	}
	c.result.Tags = openapi3.Tags{}
	var tagNames []string
	for tagName := range m {
		tagNames = append(tagNames, tagName)
	}
	sort.Strings(tagNames)
	for _, tagName := range tagNames {
		c.result.Tags = append(c.result.Tags, m[tagName])
	}
	return nil
}

func (c *Collator) mergeComponents(rv *ResourceVersion) error {
	initDestinationComponents(c.result, rv.T)
	var errs error
	for k, v := range rv.T.Components.Schemas {
		ref := "#/components/schemas/" + k
		if current, ok := c.result.Components.Schemas[k]; ok && !componentsEqual(current, v) {
			errs = multierr.Append(errs, fmt.Errorf("conflict in %s: %s and %s differ", ref, rv.path, c.componentSources[ref]))
		} else {
			c.result.Components.Schemas[k] = v
			c.componentSources[ref] = rv.path
		}
	}
	for k, v := range rv.T.Components.Parameters {
		ref := "#/components/parameters/" + k
		if current, ok := c.result.Components.Parameters[k]; ok && !componentsEqual(current, v) {
			errs = multierr.Append(errs, fmt.Errorf("conflict in %s: %s and %s differ", ref, rv.path, c.componentSources[ref]))
		} else {
			c.result.Components.Parameters[k] = v
			c.componentSources[ref] = rv.path
		}
	}
	for k, v := range rv.T.Components.Headers {
		ref := "#/components/headers/" + k
		if current, ok := c.result.Components.Headers[k]; ok && !componentsEqual(current, v) {
			errs = multierr.Append(errs, fmt.Errorf("conflict in %s: %s and %s differ", ref, rv.path, c.componentSources[ref]))
		} else {
			c.result.Components.Headers[k] = v
			c.componentSources[ref] = rv.path
		}
	}
	for k, v := range rv.T.Components.RequestBodies {
		ref := "#/components/requestBodies/" + k
		if current, ok := c.result.Components.RequestBodies[k]; ok && !componentsEqual(current, v) {
			errs = multierr.Append(errs, fmt.Errorf("conflict in %s: %s and %s differ", ref, rv.path, c.componentSources[ref]))
		} else {
			c.result.Components.RequestBodies[k] = v
			c.componentSources[ref] = rv.path
		}
	}
	for k, v := range rv.T.Components.Responses {
		ref := "#/components/responses/" + k
		if current, ok := c.result.Components.Responses[k]; ok && !componentsEqual(current, v) {
			errs = multierr.Append(errs, fmt.Errorf("conflict in %s: %s and %s differ", ref, rv.path, c.componentSources[ref]))
		} else {
			c.result.Components.Responses[k] = v
			c.componentSources[ref] = rv.path
		}
	}
	for k, v := range rv.T.Components.SecuritySchemes {
		ref := "#/components/securitySchemas/" + k
		if current, ok := c.result.Components.SecuritySchemes[k]; ok && !componentsEqual(current, v) {
			errs = multierr.Append(errs, fmt.Errorf("conflict in %s: %s and %s differ", ref, rv.path, c.componentSources[ref]))
		} else {
			c.result.Components.SecuritySchemes[k] = v
			c.componentSources[ref] = rv.path
		}
	}
	for k, v := range rv.T.Components.Examples {
		ref := "#/components/examples/" + k
		if current, ok := c.result.Components.Examples[k]; ok && !componentsEqual(current, v) {
			errs = multierr.Append(errs, fmt.Errorf("conflict in %s: %s and %s differ", ref, rv.path, c.componentSources[ref]))
		} else {
			c.result.Components.Examples[k] = v
			c.componentSources[ref] = rv.path
		}
	}
	for k, v := range rv.T.Components.Links {
		ref := "#/components/links/" + k
		if current, ok := c.result.Components.Links[k]; ok && !componentsEqual(current, v) {
			errs = multierr.Append(errs, fmt.Errorf("conflict in %s: %s and %s differ", ref, rv.path, c.componentSources[ref]))
		} else {
			c.result.Components.Links[k] = v
			c.componentSources[ref] = rv.path
		}
	}
	for k, v := range rv.T.Components.Callbacks {
		ref := "#/components/callbacks/" + k
		if current, ok := c.result.Components.Callbacks[k]; ok && !componentsEqual(current, v) {
			errs = multierr.Append(errs, fmt.Errorf("conflict in %s: %s and %s differ", ref, rv.path, c.componentSources[ref]))
		} else {
			c.result.Components.Callbacks[k] = v
			c.componentSources[ref] = rv.path
		}
	}
	return errs
}

var cmpComponents = cmp.Options{
	// openapi3.Schema has some unexported fields which are ignored for the
	// purposes of content comparison.
	cmpopts.IgnoreUnexported(openapi3.Schema{}),
	// Refs themselves can mutate during relocation, so they are excluded from
	// content comparison.
	cmp.FilterPath(func(p cmp.Path) bool {
		switch p.Last().String() {
		case ".Ref", ".Description", ".Example", ".Summary":
			return true
		}
		return false
	}, cmp.Ignore()),
}

func componentsEqual(x, y interface{}) bool {
	return cmp.Equal(x, y, cmpComponents)
}

func tagsEqual(x, y interface{}) bool {
	return cmp.Equal(x, y)
}

func (c *Collator) mergePaths(rv *ResourceVersion) error {
	if rv.T.Paths != nil && c.result.Paths == nil {
		c.result.Paths = make(openapi3.Paths)
	}
	var errs error
	for k, v := range rv.T.Paths {
		if _, ok := c.result.Paths[k]; ok {
			errs = multierr.Append(errs, fmt.Errorf("conflict in #/paths %s: declared in both %s and %s", k, rv.path, c.pathSources[k]))
		} else {
			c.result.Paths[k] = v
			c.pathSources[k] = rv.path
		}
	}
	return errs
}
