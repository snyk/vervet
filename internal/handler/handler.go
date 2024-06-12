// Package handler contains the HTTP handlers that serve Vervet Underground responses.
package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog/log"
	metrics "github.com/slok/go-http-metrics/metrics/prometheus"
	prommiddleware "github.com/slok/go-http-metrics/middleware"
	prommiddlewarestd "github.com/slok/go-http-metrics/middleware/std"

	"github.com/snyk/vervet/v6"
	"github.com/snyk/vervet/v6/config"
	"github.com/snyk/vervet/v6/internal/storage"
	"github.com/snyk/vervet/v6/versionware"
)

// Handler handles Vervet Underground HTTP requests.
type Handler struct {
	cfg    *config.ServerConfig
	store  storage.ReadOnlyStorage
	router chi.Router
}

// New returns a new Handler.
func New(cfg *config.ServerConfig, store storage.ReadOnlyStorage, routerOptions ...func(r chi.Router)) *Handler {
	h := &Handler{
		cfg:    cfg,
		store:  store,
		router: chi.NewRouter(),
	}
	for i := range routerOptions {
		routerOptions[i](h.router)
	}
	h.router.Get("/openapi/{version}", h.openapiVersion)
	h.router.Get("/openapi", h.openapiVersions)
	h.router.Get("/metrics", promhttp.Handler().ServeHTTP)
	h.router.Get("/", h.health)
	return h
}

var promMiddlewareConfig = prommiddleware.Config{
	Recorder: metrics.NewRecorder(metrics.Config{
		Prefix: "vu",
	}),
}

// UseDefaultMiddleware configures a chi.Router to use the default middleware
// in the Vervet Underground service.
func UseDefaultMiddleware(r chi.Router) {
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.StripSlashes)

	promMiddleware := prommiddleware.New(promMiddlewareConfig)
	r.Use(prommiddlewarestd.HandlerProvider("", promMiddleware))
}

// ServeHTTP implements http.Handler.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.router.ServeHTTP(w, r)
}

func (h *Handler) openapiVersions(w http.ResponseWriter, r *http.Request) {
	versionIndex, err := h.store.VersionIndex(r.Context())
	if err != nil {
		logError(err)
		http.Error(w, "Cannot get versions", http.StatusInternalServerError)
		return
	}
	content, err := json.Marshal(versionIndex.Versions().Strings())
	if err != nil {
		logError(err)
		http.Error(w, "Failure to process request", http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(content)
	if err != nil {
		logError(err)
		http.Error(w, "Failure to write response", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) openapiVersion(w http.ResponseWriter, r *http.Request) {
	versionString := chi.URLParam(r, "version")
	w.Header().Set(versionware.HeaderSnykVersionRequested, versionString)

	version, err := vervet.ParseVersion(versionString)
	if err != nil {
		// Assume current date if only stability provided
		if stability, err := vervet.ParseStability(versionString); err == nil {
			version = vervet.Version{
				Date:      time.Now().UTC().Truncate(time.Hour * 24),
				Stability: stability,
			}
		} else {
			logError(err)
			http.Error(w, "Invalid version", http.StatusBadRequest)
			return
		}
	}

	ctx := r.Context()
	versionIndex, err := h.store.VersionIndex(ctx)
	if err != nil {
		logError(err)
		http.Error(w, "Cannot get versions", http.StatusInternalServerError)
		return
	}
	resolvedVersion, err := versionIndex.Resolve(version)
	if errors.Is(err, vervet.ErrNoMatchingVersion) {
		http.Error(w, "Version not found", http.StatusNotFound)
		return
	} else if err != nil {
		logError(err)
		http.Error(w, "Failure to resolve version", http.StatusInternalServerError)
		return
	}
	resolvedVersion.Stability = version.Stability
	w.Header().Set(versionware.HeaderSnykVersionServed, resolvedVersion.String())

	content, err := h.store.Version(ctx, resolvedVersion.String())
	if err != nil {
		logError(err)
		http.Error(w, "Failure to retrieve version", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(content)
	if err != nil {
		logError(err)
		http.Error(w, "Failure to write response", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)
	if err := encoder.Encode(map[string]interface{}{"msg": "success", "services": h.cfg.Services}); err != nil {
		http.Error(w, "Failure to write response", http.StatusInternalServerError)
		return
	}
}

func logError(err error) {
	log.
		Error().
		Stack().
		Err(err).
		Str("cause", fmt.Sprintf("%+v", errors.Cause(err))).
		Msg("UnhandledException")
}
