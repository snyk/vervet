package release_2021_11_01

import (
	"encoding/json"
	"log"
	"net/http"
	"path"

	"github.com/snyk/vervet"

	"github.com/snyk/vervet/versionware/example/resources/things"
	"github.com/snyk/vervet/versionware/example/store"
)

// Version is the resource release version of handlers in this package.
var Version = vervet.MustParseVersion("2021-11-01~experimental")

// GetThing returns a request handler that uses the given data store.
// It creates a new thing.
func CreateThing(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var thingReq things.Attributes
		err := json.NewDecoder(r.Body).Decode(&thingReq)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		id, thing := s.InsertThing(things.FromAttributes(thingReq))
		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(things.ToResponse(id, thing))
		if err != nil {
			log.Println("failed to encode response", err)
		}
	}
}

// GetThing returns a request handler that uses the given data store.
// It gets a thing.
func GetThing(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Avoiding use of routing-specific URL parameter parsing here so that
		// the same code can be used with either the chi or gorilla demo. If
		// if we were settled on chi, we could use chi.URLParam instead.
		id := path.Base(r.URL.Path)
		thing, ok := s.SelectThing(id)
		if !ok {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode(things.ToResponse(id, thing))
		if err != nil {
			log.Println("failed to encode response", err)
		}
	}
}
