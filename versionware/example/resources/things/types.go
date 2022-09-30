// Package things defines the wire format types for the Things API.
package things

import (
	"time"

	"github.com/snyk/vervet/v5/versionware/example/store"
)

// Attributes represents the data contents of a Thing resource.
type Attributes struct {
	Name        string `json:"name"`
	Color       string `json:"color"`
	Strangeness int    `json:"strangeness"`
}

// Response represents a resource JSON response.
type Response struct {
	Id         string     `json:"id"`
	Created    time.Time  `json:"created"`
	Attributes Attributes `json:"attributes"`
}

// CollectionResponse represents a collection resource JSON response.
type CollectionResponse struct {
	Things []Response `json:"things"`
}

// FromAttributes converts Attributes from wire-format to a Thing data model.
func FromAttributes(attrs Attributes) store.Thing {
	return store.Thing{
		Name:        attrs.Name,
		Color:       attrs.Color,
		Strangeness: attrs.Strangeness,
	}
}

// ToResponse renders a Thing data model to a response format.
func ToResponse(id string, thing store.Thing) *Response {
	return &Response{
		Id:      id,
		Created: thing.Created,
		Attributes: Attributes{
			Name:        thing.Name,
			Color:       thing.Color,
			Strangeness: thing.Strangeness,
		},
	}
}

// ToResponse renders a collection of Thing models to a response.
func ToCollectionResponse(ids []string, things []store.Thing) *CollectionResponse {
	coll := &CollectionResponse{Things: make([]Response, len(ids))}
	for i := range ids {
		coll.Things[i] = Response{
			Id:      ids[i],
			Created: things[i].Created,
			Attributes: Attributes{
				Name:        things[i].Name,
				Color:       things[i].Color,
				Strangeness: things[i].Strangeness,
			},
		}
	}
	return coll
}
