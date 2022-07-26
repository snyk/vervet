// Package handler contains the HTTP handlers that serve Vervet Underground responses.
package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog/log"
	"github.com/snyk/vervet/v4"
	"github.com/snyk/vervet/v4/versionware"

	"vervet-underground/config"
	"vervet-underground/internal/scraper"
)

// Handler handles Vervet Underground HTTP requests.
type Handler struct {
	cfg    *config.ServerConfig
	sc     *scraper.Scraper
	router chi.Router
}

// New returns a new Handler.
func New(cfg *config.ServerConfig, sc *scraper.Scraper, routerOptions ...func(r chi.Router)) *Handler {
	h := &Handler{
		cfg:    cfg,
		sc:     sc,
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

// ServeHTTP implements http.Handler.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.router.ServeHTTP(w, r)
}

func (h *Handler) openapiVersions(w http.ResponseWriter, r *http.Request) {
	content, err := json.Marshal(h.sc.Versions().Strings())
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

	availableVersions := h.sc.Versions()
	resolvedVersion, err := availableVersions.Resolve(version)
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

	content, err := h.sc.Version(r.Context(), resolvedVersion.String())
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
