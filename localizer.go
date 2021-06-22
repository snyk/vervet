package apiutil

import (
	"log"
	"path"
	"reflect"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/mitchellh/reflectwalk"
)

type Localizer struct {
	doc *openapi3.T

	curRefType    reflect.Value
	curRefField   reflect.Value
	curValueField reflect.Value
}

func NewLocalizer(doc *openapi3.T) *Localizer {
	return &Localizer{doc: doc}
}

func (l *Localizer) Localize() error {
	err := reflectwalk.Walk(l.doc, l)
	if err != nil {
		return err
	}
	// Some of the localized components may have non-localized references,
	// since they were just added to the document object tree in the prior
	// walk. Brute-forcing them into the fold...
	return reflectwalk.Walk(l.doc.Components, l)
}

/*
	for _, pi := range l.doc.Paths {
		// TODO: add all the ops
		for _, op := range []*openapi3.Operation{pi.Get} {
			for _, paramRef := range op.Parameters {
				l.localizeParam(paramRef)
			}
		}
		for _, paramRef := range pi.Parameters {
			l.localizeParam(paramRef)
		}
	}
	return nil
}
*/

func (l *Localizer) Struct(v reflect.Value) error {
	l.curRefType, l.curRefField, l.curValueField = v, v.FieldByName("Ref"), v.FieldByName("Value")
	return nil
}

func (l *Localizer) StructField(sf reflect.StructField, v reflect.Value) error {
	if !l.curRefField.IsValid() || !l.curValueField.IsValid() {
		return nil
	}
	refPath := l.curRefField.String()
	if refPath == "" {
		return nil
	}
	// TODO: resolve unique names from external component refs?
	refBase := path.Base(refPath)
	if isLocalRef(refPath) {
		return nil
	}

	switch refObj := l.curRefType.Addr().Interface().(type) {
	case *openapi3.SchemaRef:
		refObj.Ref = "#/components/schemas/" + refBase
		if l.doc.Components.Schemas == nil {
			l.doc.Components.Schemas = map[string]*openapi3.SchemaRef{}
		}
		if l.doc.Components.Schemas[refBase] == nil {
			l.doc.Components.Schemas[refBase] = &openapi3.SchemaRef{Value: refObj.Value}
		}
	case *openapi3.ParameterRef:
		refObj.Ref = "#/components/parameters/" + refBase
		if l.doc.Components.Parameters == nil {
			l.doc.Components.Parameters = map[string]*openapi3.ParameterRef{}
		}
		if l.doc.Components.Parameters[refBase] == nil {
			l.doc.Components.Parameters[refBase] = &openapi3.ParameterRef{Value: refObj.Value}
		}
	case *openapi3.LinkRef:
		refObj.Ref = "#/components/links/" + refBase
		if l.doc.Components.Links == nil {
			l.doc.Components.Links = map[string]*openapi3.LinkRef{}
		}
		if l.doc.Components.Links[refBase] == nil {
			l.doc.Components.Links[refBase] = &openapi3.LinkRef{Value: refObj.Value}
		}
	case *openapi3.RequestBodyRef:
		refObj.Ref = "#/components/requests/" + refBase
		if l.doc.Components.RequestBodies == nil {
			l.doc.Components.RequestBodies = map[string]*openapi3.RequestBodyRef{}
		}
		if l.doc.Components.RequestBodies[refBase] == nil {
			l.doc.Components.RequestBodies[refBase] = &openapi3.RequestBodyRef{Value: refObj.Value}
		}
	case *openapi3.ResponseRef:
		refObj.Ref = "#/components/responses/" + refBase
		if l.doc.Components.Responses == nil {
			l.doc.Components.Responses = map[string]*openapi3.ResponseRef{}
		}
		if l.doc.Components.Responses[refBase] == nil {
			l.doc.Components.Responses[refBase] = &openapi3.ResponseRef{Value: refObj.Value}
		}
	case *openapi3.HeaderRef:
		refObj.Ref = "#/components/headers/" + refBase
		if l.doc.Components.Headers == nil {
			l.doc.Components.Headers = map[string]*openapi3.HeaderRef{}
		}
		if l.doc.Components.Headers[refBase] == nil {
			l.doc.Components.Headers[refBase] = &openapi3.HeaderRef{Value: refObj.Value}
		}
	default:
		log.Printf("got %T", refObj)
	}
	/*
		if sf.Name == "Ref" && sf.Type.Kind() == reflect.String {
			refBase := path.Base(v.String())
			log.Printf("refBase %v value %v", refBase, v.Interface())
			var refPath string
			refTypeName := l.currentStruct.Type().Name()
			refPtr := l.currentStruct.Addr().Interface()
			//refValPtr := l.refValueField.Interface()
			switch refTypeName {
			case "SchemaRef":
				refPath = "#/components/schemas/" + refBase
				ref := refPtr.(*openapi3.SchemaRef)
				if l.root.Components.Schemas == nil {
					l.root.Components.Schemas = map[string]*openapi3.SchemaRef{}
				}
				if l.root.Components.Schemas[refBase] == nil {
					//l.walkInto(refValPtr)
					l.root.Components.Schemas[refBase] = ref
				}
			case "ParameterRef":
				refPath = "#/components/parameters/" + refBase
				ref := refPtr.(*openapi3.ParameterRef)
				if l.root.Components.Parameters == nil {
					l.root.Components.Parameters = map[string]*openapi3.ParameterRef{}
				}
				if l.root.Components.Parameters[refBase] == nil {
					//l.walkInto(refValPtr)
					l.root.Components.Parameters[refBase] = ref
				}
				/*
					case "HeaderRef":
						refPath = "#/components/headers/" + refBase
						ref := refPtr.(*openapi3.HeaderRef)
						ref.Ref = ""
						if l.root.Components.Headers == nil {
							l.root.Components.Headers = map[string]*openapi3.HeaderRef{}
						}
						if l.root.Components.Headers[refBase] == nil {
							l.walkInto(refValPtr)
							l.root.Components.Headers[refBase] = ref
						}
					case "RequestBodyRef":
						refPath = "#/components/requestBodies/" + refBase
						ref := refPtr.(*openapi3.RequestBodyRef)
						ref.Ref = ""
						if l.root.Components.RequestBodies == nil {
							l.root.Components.RequestBodies = map[string]*openapi3.RequestBodyRef{}
						}
						if l.root.Components.RequestBodies[refBase] == nil {
							l.walkInto(refValPtr)
							l.root.Components.RequestBodies[refBase] = ref
						}
					case "ResponseRef":
						refPath = "#/components/responses/" + refBase
						ref := refPtr.(*openapi3.ResponseRef)
						ref.Ref = ""
						if l.root.Components.Responses == nil {
							l.root.Components.Responses = map[string]*openapi3.ResponseRef{}
						}
						if l.root.Components.Responses[refBase] == nil {
							l.walkInto(refValPtr)
							l.root.Components.Responses[refBase] = ref
						}
					case "SecuritySchemeRef":
						refPath = "#/components/securitySchemes/" + refBase
						ref := refPtr.(*openapi3.SecuritySchemeRef)
						ref.Ref = ""
						if l.root.Components.SecuritySchemes == nil {
							l.root.Components.SecuritySchemes = map[string]*openapi3.SecuritySchemeRef{}
						}
						if l.root.Components.SecuritySchemes[refBase] == nil {
							l.walkInto(refValPtr)
							l.root.Components.SecuritySchemes[refBase] = ref
						}
					case "ExampleRef":
						refPath = "#/components/examples/" + refBase
						ref := refPtr.(*openapi3.ExampleRef)
						ref.Ref = ""
						if l.root.Components.Examples == nil {
							l.root.Components.Examples = map[string]*openapi3.ExampleRef{}
						}
						if l.root.Components.Examples[refBase] == nil {
							l.walkInto(refValPtr)
							l.root.Components.Examples[refBase] = ref
						}
					case "LinkRef":
						refPath = "#/components/links/" + refBase
						ref := refPtr.(*openapi3.LinkRef)
						ref.Ref = ""
						if l.root.Components.Links == nil {
							l.root.Components.Links = map[string]*openapi3.LinkRef{}
						}
						if l.root.Components.Links[refBase] == nil {
							l.walkInto(refValPtr)
							l.root.Components.Links[refBase] = ref
						}
					case "CallbackRef":
						refPath = "#/components/callbacks/" + refBase
						ref := refPtr.(*openapi3.CallbackRef)
						ref.Ref = ""
						if l.root.Components.Callbacks == nil {
							l.root.Components.Callbacks = map[string]*openapi3.CallbackRef{}
						}
						if l.root.Components.Callbacks[refBase] == nil {
							l.walkInto(refValPtr)
							l.root.Components.Callbacks[refBase] = ref
						}
	*/
	return nil
}

func (l *Localizer) localizeParam(p *openapi3.ParameterRef) {
	if p.Ref != "" && !isLocalRef(p.Ref) {
		// TODO: make refBase unique, might have distinct types from different
		// paths, for example.
		refBase := path.Base(p.Ref)
		p.Ref = "#/components/parameters/" + refBase
		if l.doc.Components.Parameters == nil {
			l.doc.Components.Parameters = map[string]*openapi3.ParameterRef{}
		}
		l.doc.Components.Parameters[refBase] = &openapi3.ParameterRef{Value: p.Value}
	}
}

/*
func (l *Localizer) localizeSchema(p *openapi3.ParameterRef) {
	if p.Ref != "" && !isLocalRef(p.Ref) {
		// TODO: make refBase unique, might have distinct types from different
		// paths, for example.
		refBase := path.Base(p.Ref)
		p.Ref = "#/components/schemas/" + refBase
		if l.doc.Components.Schemas == nil {
			l.doc.Components.Schemas = map[string]*openapi3.SchemaRef{}
		}
		l.doc.Components.Schemas[refBase] = &openapi3.SchemaRef{Value: p.Value}
	}
}
*/

func isLocalRef(s string) bool {
	return strings.HasPrefix(s, "#/components/")
}
