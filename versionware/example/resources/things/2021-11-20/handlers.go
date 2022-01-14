package release_2021_11_20

import (
	"net/http"
	"path"

	"github.com/snyk/vervet"
	"github.com/snyk/vervet/versionware/example/store"
)

// Version is the resource release version of handlers in this package.
var Version = vervet.MustParseVersion("2021-11-20~experimental")

// DeleteThing returns a request handler that uses the given data store.
// It deletes a thing.
func DeleteThing(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Avoiding use of routing-specific URL parameter parsing here so that
		// the same code can be used with either the chi or gorilla demo. If
		// if we were settled on chi, we could use chi.URLParam instead.
		id := path.Base(r.URL.Path)
		ok := s.DeleteThing(id)
		if !ok {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}
