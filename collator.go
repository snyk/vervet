package vervet

import (
	"fmt"
	"reflect"
	"sort"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/mitchellh/reflectwalk"
	"go.uber.org/multierr"
)

// Collator merges resource versions into a single OpenAPI document.
type Collator struct {
	result           *openapi3.T
	componentSources map[string]string
	pathSources      map[string]string
	tagSources       map[string]string

	strictTags bool
}

// RefRemover removes the ref from the component
type RefRemover struct {
	target interface{}
}

// Inliner inlines the component
type Inliner struct {
	refs map[string]struct{}
}

func (rr *RefRemover) RemoveRef() error {
	return reflectwalk.Walk(rr.target, rr)
}

// Struct implements reflectwalk.StructWalker
func (rr *RefRemover) Struct(v reflect.Value) error {
	switch v.Interface().(type) {
	case openapi3.SchemaRef:
		valPointer := v.Addr().Interface().(*openapi3.SchemaRef)
		valPointer.Ref = ""
	case openapi3.ParameterRef:
		valPointer := v.Addr().Interface().(*openapi3.ParameterRef)
		valPointer.Ref = ""
	case openapi3.HeaderRef:
		valPointer := v.Addr().Interface().(*openapi3.HeaderRef)
		valPointer.Ref = ""
	case openapi3.RequestBodyRef:
		valPointer := v.Addr().Interface().(*openapi3.RequestBodyRef)
		valPointer.Ref = ""
	case openapi3.ResponseRef:
		valPointer := v.Addr().Interface().(*openapi3.ResponseRef)
		valPointer.Ref = ""
	case openapi3.SecuritySchemeRef:
		valPointer := v.Addr().Interface().(*openapi3.SecuritySchemeRef)
		valPointer.Ref = ""
	case openapi3.ExampleRef:
		valPointer := v.Addr().Interface().(*openapi3.ExampleRef)
		valPointer.Ref = ""
	case openapi3.LinkRef:
		valPointer := v.Addr().Interface().(*openapi3.LinkRef)
		valPointer.Ref = ""
	case openapi3.CallbackRef:
		valPointer := v.Addr().Interface().(*openapi3.CallbackRef)
		valPointer.Ref = ""
	}

	return nil
}

// StructField implements reflectwalk.StructWalker
func (rr *RefRemover) StructField(field reflect.StructField, v reflect.Value) error {
	return nil
}

func NewRefRemover(target interface{}) *RefRemover {
	return &RefRemover{target: target}
}

func NewInliner() *Inliner {
	return &Inliner{refs: map[string]struct{}{}}
}

func (in *Inliner) AddRef(ref string) {
	in.refs[ref] = struct{}{}
}

// Struct implements reflectwalk.StructWalker
func (in *Inliner) Struct(v reflect.Value) error {
	switch val := v.Interface().(type) {
	case openapi3.SchemaRef:
		if _, ok := in.refs[val.Ref]; ok {
			valPointer := v.Addr().Interface().(*openapi3.SchemaRef)
			refRemover := NewRefRemover(valPointer)
			err := refRemover.RemoveRef()
			if err != nil {
				return err
			}
		}
	case openapi3.ParameterRef:
		if _, ok := in.refs[val.Ref]; ok {
			valPointer := v.Addr().Interface().(*openapi3.ParameterRef)
			refRemover := NewRefRemover(valPointer)
			err := refRemover.RemoveRef()
			if err != nil {
				return err
			}
		}
	case openapi3.HeaderRef:
		if _, ok := in.refs[val.Ref]; ok {
			valPointer := v.Addr().Interface().(*openapi3.HeaderRef)
			refRemover := NewRefRemover(valPointer)
			err := refRemover.RemoveRef()
			if err != nil {
				return err
			}
		}
	case openapi3.RequestBodyRef:
		if _, ok := in.refs[val.Ref]; ok {
			valPointer := v.Addr().Interface().(*openapi3.RequestBodyRef)
			refRemover := NewRefRemover(valPointer)
			err := refRemover.RemoveRef()
			if err != nil {
				return err
			}
		}
	case openapi3.ResponseRef:
		if _, ok := in.refs[val.Ref]; ok {
			valPointer := v.Addr().Interface().(*openapi3.ResponseRef)
			refRemover := NewRefRemover(valPointer)
			err := refRemover.RemoveRef()
			if err != nil {
				return err
			}
		}
	case openapi3.SecuritySchemeRef:
		if _, ok := in.refs[val.Ref]; ok {
			valPointer := v.Addr().Interface().(*openapi3.SecuritySchemeRef)
			refRemover := NewRefRemover(valPointer)
			err := refRemover.RemoveRef()
			if err != nil {
				return err
			}
		}
	case openapi3.ExampleRef:
		if _, ok := in.refs[val.Ref]; ok {
			valPointer := v.Addr().Interface().(*openapi3.ExampleRef)
			refRemover := NewRefRemover(valPointer)
			err := refRemover.RemoveRef()
			if err != nil {
				return err
			}
		}
	case openapi3.LinkRef:
		if _, ok := in.refs[val.Ref]; ok {
			valPointer := v.Addr().Interface().(*openapi3.LinkRef)
			refRemover := NewRefRemover(valPointer)
			err := refRemover.RemoveRef()
			if err != nil {
				return err
			}
		}
	case openapi3.CallbackRef:
		if _, ok := in.refs[val.Ref]; ok {
			valPointer := v.Addr().Interface().(*openapi3.CallbackRef)
			refRemover := NewRefRemover(valPointer)
			err := refRemover.RemoveRef()
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// StructField implements reflectwalk.StructWalker
func (in *Inliner) StructField(field reflect.StructField, v reflect.Value) error {
	return nil
}

func (in *Inliner) Inliner(doc *openapi3.T) error {
	return reflectwalk.Walk(doc, in)
}

// NewCollator returns a new Collator instance.
func NewCollator(options ...CollatorOption) *Collator {
	coll := &Collator{
		componentSources: map[string]string{},
		pathSources:      map[string]string{},
		tagSources:       map[string]string{},
		strictTags:       true,
	}
	for i := range options {
		options[i](coll)
	}
	return coll
}

// CollatorOption defines an option when creating a Collator.
type CollatorOption func(*Collator)

// StrictTags defines whether a collator should enforce a strict conflict check
// when merging tags.
func StrictTags(strict bool) CollatorOption {
	return func(coll *Collator) {
		coll.strictTags = strict
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

	mergeExtensions(c.result, rv.T, false)
	mergeInfo(c.result, rv.T, false)
	mergeOpenAPIVersion(c.result, rv.T, false)
	mergeSecurityRequirements(c.result, rv.T, false)
	mergeServers(c.result, rv.T, false)

	if err = c.mergeComponents(rv); err != nil {
		errs = multierr.Append(errs, err)
	}

	if err = c.mergePaths(rv); err != nil {
		errs = multierr.Append(errs, err)
	}

	if err = c.mergeTags(rv); err != nil {
		errs = multierr.Append(errs, err)
	}

	return errs
}

func (c *Collator) mergeTags(rv *ResourceVersion) error {
	m := map[string]*openapi3.Tag{}
	for _, t := range c.result.Tags {
		m[t.Name] = t
	}
	var errs error
	for _, t := range rv.T.Tags {
		if current, ok := m[t.Name]; ok && !tagsEqual(current, t) && c.strictTags {
			// If there is a conflict and we're collating with strict tags, indicate an error.
			errs = multierr.Append(errs, fmt.Errorf("conflict in #/tags %s: %s and %s differ", t.Name, rv.path, c.tagSources[t.Name]))
		} else {
			// Otherwise last tag with this key wins.
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
	inliner := NewInliner()
	for k, v := range rv.T.Components.Schemas {
		ref := "#/components/schemas/" + k
		if current, ok := c.result.Components.Schemas[k]; ok && !componentsEqual(current, v) {
			inliner.AddRef(ref)
		} else {
			c.result.Components.Schemas[k] = v
			c.componentSources[ref] = rv.path
		}
	}
	for k, v := range rv.T.Components.Parameters {
		ref := "#/components/parameters/" + k
		if current, ok := c.result.Components.Parameters[k]; ok && !componentsEqual(current, v) {
			inliner.AddRef(ref)
		} else {
			c.result.Components.Parameters[k] = v
			c.componentSources[ref] = rv.path
		}
	}
	for k, v := range rv.T.Components.Headers {
		ref := "#/components/headers/" + k
		if current, ok := c.result.Components.Headers[k]; ok && !componentsEqual(current, v) {
			inliner.AddRef(ref)
		} else {
			c.result.Components.Headers[k] = v
			c.componentSources[ref] = rv.path
		}
	}
	for k, v := range rv.T.Components.RequestBodies {
		ref := "#/components/requestBodies/" + k
		if current, ok := c.result.Components.RequestBodies[k]; ok && !componentsEqual(current, v) {
			inliner.AddRef(ref)
		} else {
			c.result.Components.RequestBodies[k] = v
			c.componentSources[ref] = rv.path
		}
	}
	for k, v := range rv.T.Components.Responses {
		ref := "#/components/responses/" + k
		if current, ok := c.result.Components.Responses[k]; ok && !componentsEqual(current, v) {
			inliner.AddRef(ref)
		} else {
			c.result.Components.Responses[k] = v
			c.componentSources[ref] = rv.path
		}
	}
	for k, v := range rv.T.Components.SecuritySchemes {
		ref := "#/components/securitySchemas/" + k
		if current, ok := c.result.Components.SecuritySchemes[k]; ok && !componentsEqual(current, v) {
			inliner.AddRef(ref)
		} else {
			c.result.Components.SecuritySchemes[k] = v
			c.componentSources[ref] = rv.path
		}
	}
	for k, v := range rv.T.Components.Examples {
		ref := "#/components/examples/" + k
		if current, ok := c.result.Components.Examples[k]; ok && !componentsEqual(current, v) {
			inliner.AddRef(ref)
		} else {
			c.result.Components.Examples[k] = v
			c.componentSources[ref] = rv.path
		}
	}
	for k, v := range rv.T.Components.Links {
		ref := "#/components/links/" + k
		if current, ok := c.result.Components.Links[k]; ok && !componentsEqual(current, v) {
			inliner.AddRef(ref)
		} else {
			c.result.Components.Links[k] = v
			c.componentSources[ref] = rv.path
		}
	}
	for k, v := range rv.T.Components.Callbacks {
		ref := "#/components/callbacks/" + k
		if current, ok := c.result.Components.Callbacks[k]; ok && !componentsEqual(current, v) {
			inliner.AddRef(ref)
		} else {
			c.result.Components.Callbacks[k] = v
			c.componentSources[ref] = rv.path
		}
	}
	if errs == nil {
		err := inliner.Inliner(rv.T)
		if err != nil {
			errs = multierr.Append(errs, err)
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
