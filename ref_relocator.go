package vervet

/*
func replaceRefs(doc *openapi3.T, targetRef, suffix string) error {
	r := &refReplacer{targetRef: targetRef, suffix: suffix}
	err := r.replace(doc)
	if err != nil {
		return err
	}
	return nil
}

// refReplacer replaces references in a self-contained OpenAPI document object.
type refReplacer struct {
	targetRef string
	suffix    string

	curRefType  reflect.Value
	curRefField reflect.Value
}

func (r *refReplacer) replace(doc *openapi3.T) error {
	return reflectwalk.Walk(doc, r)
}

// Struct implements reflectwalk.StructWalker
func (r *refReplacer) Struct(v reflect.Value) error {
	r.curRefType, r.curRefField = v, v.FieldByName("Ref")
	return nil
}

// StructField implements reflectwalk.StructWalker
func (r *refReplacer) StructField(sf reflect.StructField, v reflect.Value) error {
	if !r.curRefField.IsValid() {
		return nil
	}
	ref := r.curRefField.String()
	if ref == r.oldRef {
		r.curRefField.Set(reflect.ValueOf(r.newRef))
	}
	return nil
}
*/
