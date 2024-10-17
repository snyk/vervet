// Package versionware provides routing and middleware for building versioned
// HTTP services.
package versionware

import (
	"fmt"
	"net/http"

	"github.com/snyk/vervet/v8"
)

const (
	// HeaderSnykVersionRequested is a response header acknowledging the API
	// version that was requested.
	HeaderSnykVersionRequested = "snyk-version-requested"

	// HeaderSnykVersionServed is a response header indicating the actual API
	// version that was matched and served the response.
	HeaderSnykVersionServed = "snyk-version-served"
)

// Handler is a multiplexing http.Handler that dispatches requests based on the
// version query parameter according to vervet's API version matching rules.
type Handler struct {
	handlers map[vervet.Version]http.Handler
	index    vervet.VersionIndex
	errFunc  VersionErrorHandler
}

// VersionErrorHandler defines a function which handles versioning error
// responses in requests.
type VersionErrorHandler func(w http.ResponseWriter, r *http.Request, status int, err error)

// VersionHandler expresses a pairing of Version and http.Handler.
type VersionHandler struct {
	Version vervet.Version
	Handler http.Handler
}

// NewHandler returns a new Handler instance, which handles versioned requests
// with the matching version handler.
func NewHandler(vhs ...VersionHandler) *Handler {
	h := &Handler{
		handlers: map[vervet.Version]http.Handler{},
		errFunc:  DefaultVersionError,
	}
	versions := make([]vervet.Version, len(vhs))
	for i := range vhs {
		v := vhs[i].Version
		versions[i] = v
		h.handlers[v] = vhs[i].Handler
	}
	h.index = vervet.NewVersionIndex(versions)
	return h
}

// DefaultVersionError provides a basic implementation of VersionErrorHandler
// that uses http.Error.
func DefaultVersionError(w http.ResponseWriter, r *http.Request, status int, err error) {
	http.Error(w, http.StatusText(status), status)
}

// HandleErrors changes the default error handler to the provided function. It
// may be used to control the format of versioning error responses.
func (h *Handler) HandleErrors(errFunc VersionErrorHandler) {
	h.errFunc = errFunc
}

// Resolve returns the resolved version and its associated http.Handler for the
// requested version.
func (h *Handler) Resolve(requested vervet.Version) (*vervet.Version, http.Handler, error) {
	resolvedVersion, err := h.index.ResolveForBuild(requested)
	if err != nil {
		return nil, nil, err
	}
	return &resolvedVersion, h.handlers[resolvedVersion], nil
}

// ServeHTTP implements http.Handler with the handler matching the version
// query parameter on the request. If no matching version is found, responds
// 404.
func (h *Handler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	versionParam := req.URL.Query().Get("version")
	if versionParam == "" {
		h.errFunc(w, req, http.StatusBadRequest, fmt.Errorf("missing required query parameter 'version'"))
		return
	}
	requested, err := vervet.ParseVersion(versionParam)
	if err != nil {
		h.errFunc(w, req, http.StatusBadRequest, err)
		return
	}
	versionToResolve := requested
	// Check if the cerberus has resolved a specific version to be served.
	snykVersionInHeader := req.Header.Get(HeaderSnykVersionServed)
	if snykVersionInHeader != "" {
		versionToResolve, err = vervet.ParseVersion(snykVersionInHeader)
		if err != nil {
			h.errFunc(w, req, http.StatusBadRequest, fmt.Errorf(
				"cant parse header %s with version %s:%w", HeaderSnykVersionServed, snykVersionInHeader, err))
			return
		}
	}
	resolved, handler, err := h.Resolve(versionToResolve)
	if err != nil {
		h.errFunc(w, req, http.StatusNotFound, err)
		return
	}
	w.Header().Set(HeaderSnykVersionRequested, requested.String())
	w.Header().Set(HeaderSnykVersionServed, resolved.String())
	handler.ServeHTTP(w, req)
}
