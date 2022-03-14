package vervet

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/getkin/kin-openapi/openapi3"
)

// ComponentDeduplicator relocating name collisions in components across a
// collection of OpenAPI documents, so that they can be non-destructively
// merged.
type ComponentDeduplicator struct {
	contentDB   map[componentKey]string
	docDB       map[componentKey]*openapi3.T
	relocations map[componentKey]string
}

// NewComponentDeduplicator returns a new ComponentDeduplicator.
func NewComponentDeduplicator() *ComponentDeduplicator {
	return &ComponentDeduplicator{
		contentDB: map[componentKey]string{},
		docDB:     map[componentKey]*openapi3.T{},
	}
}

type componentKey struct {
	location string
	ref      string
}

type componentKeys []componentKey

func (ks componentKeys) Len() int { return len(ks) }
func (ks componentKeys) Less(i, j int) bool {
	if ks[i].location == ks[j].location {
		return ks[i].ref < ks[j].ref
	}
	return ks[i].location < ks[j].location
}
func (ks componentKeys) Swap(i, j int) { ks[i], ks[j] = ks[j], ks[i] }

func (dd *ComponentDeduplicator) Index(location string, doc *openapi3.T) error {
	for name, value := range doc.Components.Schemas {
		key := componentKey{location, "#/components/schemas/" + name}
		digest, err := componentHash(value)
		if err != nil {
			return err
		}
		dd.contentDB[key] = digest
		dd.docDB[key] = doc
	}
	for name, value := range doc.Components.Parameters {
		key := componentKey{location, "#/components/parameters/" + name}
		digest, err := componentHash(value)
		if err != nil {
			return err
		}
		dd.contentDB[key] = digest
		dd.docDB[key] = doc
	}
	for name, value := range doc.Components.Headers {
		key := componentKey{location, "#/components/headers/" + name}
		digest, err := componentHash(value)
		if err != nil {
			return err
		}
		dd.contentDB[key] = digest
		dd.docDB[key] = doc
	}
	for name, value := range doc.Components.RequestBodies {
		key := componentKey{location, "#/components/requestBodies/" + name}
		digest, err := componentHash(value)
		if err != nil {
			return err
		}
		dd.contentDB[key] = digest
		dd.docDB[key] = doc
	}
	for name, value := range doc.Components.Responses {
		key := componentKey{location, "#/components/responses/" + name}
		digest, err := componentHash(value)
		if err != nil {
			return err
		}
		dd.contentDB[key] = digest
		dd.docDB[key] = doc
	}
	for name, value := range doc.Components.SecuritySchemes {
		key := componentKey{location, "#/components/securitySchemes/" + name}
		digest, err := componentHash(value)
		if err != nil {
			return err
		}
		dd.contentDB[key] = digest
		dd.docDB[key] = doc
	}
	for name, value := range doc.Components.Examples {
		key := componentKey{location, "#/components/examples/" + name}
		digest, err := componentHash(value)
		if err != nil {
			return err
		}
		dd.contentDB[key] = digest
		dd.docDB[key] = doc
	}
	for name, value := range doc.Components.Links {
		key := componentKey{location, "#/components/links/" + name}
		digest, err := componentHash(value)
		if err != nil {
			return err
		}
		dd.contentDB[key] = digest
		dd.docDB[key] = doc
	}
	for name, value := range doc.Components.Callbacks {
		key := componentKey{location, "#/components/callbacks/" + name}
		digest, err := componentHash(value)
		if err != nil {
			return err
		}
		dd.contentDB[key] = digest
		dd.docDB[key] = doc
	}
	return nil
}

func componentHash(v interface{}) (string, error) {
	h := sha256.New()
	err := json.NewEncoder(h).Encode(&v)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func (dd *ComponentDeduplicator) Deduplicate() error {
	dd.calculateRelocations()
	fmt.Println(dd.relocations)
	return nil
}

func (dd *ComponentDeduplicator) calculateRelocations() {
	var keys componentKeys
	for k := range dd.contentDB {
		keys = append(keys, k)
	}
	sort.Sort(keys)

	dd.relocations = map[componentKey]string{}

	dedup := map[string]string{}
	relocatedDigests := map[string]string{}
	for _, key := range keys {
		digest := dd.contentDB[key]
		if relocatedRef, ok := relocatedDigests[digest]; ok {
			dd.relocations[key] = relocatedRef
		} else if priorDigest, ok := dedup[key.ref]; ok && digest != priorDigest {
			relocatedRef := key.ref + "$" + digest // TODO: shorten this
			relocatedDigests[digest] = relocatedRef
			dd.relocations[key] = relocatedRef
		} else {
			dedup[key.ref] = digest
		}
	}
}
