package release_2021_11_08

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/snyk/vervet/v8"
	"github.com/snyk/vervet/v8/versionware/example/resources/things"
	"github.com/snyk/vervet/v8/versionware/example/store"
)

// Version is the resource release version of handlers in this package.
var Version = vervet.MustParseVersion("2021-11-08~experimental")

// ListThings returns a request handler that uses the given data store. It
// lists all the things.
func ListThings(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ids, storeThings := s.ListThings()
		w.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode(things.ToCollectionResponse(ids, storeThings))
		if err != nil {
			log.Println("failed to encode response", err)
		}
	}
}
