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
}

// NewCollator returns a new Collator instance.
func NewCollator() *Collator {
	return &Collator{
		componentSources: map[string]string{},
		pathSources:      map[string]string{},
	}
}

// Result returns the merged result. If no versions have been merged, returns
// nil.
func (c *Collator) Result() *openapi3.T {
	return c.result
}

// Collate merges a resource version into the current result.
func (c *Collator) Collate(rv *ResourceVersion) error {
	if c.result == nil {
		c.result = &openapi3.T{}
	}
	err := c.mergeComponents(rv)
	if err != nil {
		return err
	}
	mergeInfo(c.result, rv.T, false)
	err = c.mergePaths(rv)
	if err != nil {
		return err
	}
	mergeSecurityRequirements(c.result, rv.T, false)
	mergeServers(c.result, rv.T, false)
	err = c.mergeTags(rv)
	if err != nil {
		return err
	}
	mergeOpenAPIVersion(c.result, rv.T, false)
	return nil
}

func (c *Collator) mergeTags(rv *ResourceVersion) error {
	m := map[string]*openapi3.Tag{}
	for _, t := range c.result.Tags {
		m[t.Name] = t
	}
	for _, t := range rv.T.Tags {
		if _, ok := m[t.Name]; ok {
			return fmt.Errorf("conflicting tag")
		} else {
			m[t.Name] = t
		}
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
		if current, ok := c.result.Components.Schemas[k]; ok && !cmpEqual(current, v) {
			errs = multierr.Append(errs, fmt.Errorf("conflict in %s: %s and %s differ", ref, rv.path, c.componentSources[ref]))
		} else {
			c.result.Components.Schemas[k] = v
			c.componentSources[ref] = rv.path
		}
	}
	for k, v := range rv.T.Components.Parameters {
		ref := "#/components/parameters/" + k
		if current, ok := c.result.Components.Parameters[k]; ok && !cmpEqual(current, v) {
			errs = multierr.Append(errs, fmt.Errorf("conflict in %s: %s and %s differ", ref, rv.path, c.componentSources[ref]))
		} else {
			c.result.Components.Parameters[k] = v
			c.componentSources[ref] = rv.path
		}
	}
	for k, v := range rv.T.Components.Headers {
		ref := "#/components/headers/" + k
		if current, ok := c.result.Components.Headers[k]; ok && !cmpEqual(current, v) {
			errs = multierr.Append(errs, fmt.Errorf("conflict in %s: %s and %s differ", ref, rv.path, c.componentSources[ref]))
		} else {
			c.result.Components.Headers[k] = v
			c.componentSources[ref] = rv.path
		}
	}
	for k, v := range rv.T.Components.RequestBodies {
		ref := "#/components/requestBodies/" + k
		if current, ok := c.result.Components.RequestBodies[k]; ok && !cmpEqual(current, v) {
			errs = multierr.Append(errs, fmt.Errorf("conflict in %s: %s and %s differ", ref, rv.path, c.componentSources[ref]))
		} else {
			c.result.Components.RequestBodies[k] = v
			c.componentSources[ref] = rv.path
		}
	}
	for k, v := range rv.T.Components.Responses {
		ref := "#/components/responses/" + k
		if current, ok := c.result.Components.Responses[k]; ok && !cmpEqual(current, v) {
			errs = multierr.Append(errs, fmt.Errorf("conflict in %s: %s and %s differ", ref, rv.path, c.componentSources[ref]))
		} else {
			c.result.Components.Responses[k] = v
			c.componentSources[ref] = rv.path
		}
	}
	for k, v := range rv.T.Components.SecuritySchemes {
		ref := "#/components/securitySchemas/" + k
		if current, ok := c.result.Components.SecuritySchemes[k]; ok && !cmpEqual(current, v) {
			errs = multierr.Append(errs, fmt.Errorf("conflict in %s: %s and %s differ", ref, rv.path, c.componentSources[ref]))
		} else {
			c.result.Components.SecuritySchemes[k] = v
			c.componentSources[ref] = rv.path
		}
	}
	for k, v := range rv.T.Components.Examples {
		ref := "#/components/examples/" + k
		if current, ok := c.result.Components.Examples[k]; ok && !cmpEqual(current, v) {
			errs = multierr.Append(errs, fmt.Errorf("conflict in %s: %s and %s differ", ref, rv.path, c.componentSources[ref]))
		} else {
			c.result.Components.Examples[k] = v
			c.componentSources[ref] = rv.path
		}
	}
	for k, v := range rv.T.Components.Links {
		ref := "#/components/links/" + k
		if current, ok := c.result.Components.Links[k]; ok && !cmpEqual(current, v) {
			errs = multierr.Append(errs, fmt.Errorf("conflict in %s: %s and %s differ", ref, rv.path, c.componentSources[ref]))
		} else {
			c.result.Components.Links[k] = v
			c.componentSources[ref] = rv.path
		}
	}
	for k, v := range rv.T.Components.Callbacks {
		ref := "#/components/callbacks/" + k
		if current, ok := c.result.Components.Callbacks[k]; ok && !cmpEqual(current, v) {
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

func cmpEqual(x, y interface{}) bool {
	return cmp.Equal(x, y, cmpComponents)
}

func (c *Collator) mergePaths(rv *ResourceVersion) error {
	if rv.T.Paths != nil && c.result.Paths == nil {
		c.result.Paths = make(openapi3.Paths)
	}
	for k, v := range rv.T.Paths {
		if _, ok := c.result.Paths[k]; ok {
			return fmt.Errorf("conflicting paths")
		}
		c.result.Paths[k] = v
	}
	return nil
}
