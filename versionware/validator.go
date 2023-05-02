package versionware

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers/gorillamux"

	"github.com/snyk/vervet/v5"
)

// Validator provides versioned OpenAPI validation middleware for HTTP requests
// and responses.
type Validator struct {
	versions   vervet.VersionIndex
	validators map[vervet.Version]*openapi3filter.Validator
	errFunc    VersionErrorHandler
	today      func() time.Time
}

// ValidatorConfig defines how a new Validator may be configured.
type ValidatorConfig struct {
	// ServerURL overrides the server URLs in the given OpenAPI specs to match
	// the URL of requests reaching the backend service. If unset, requests
	// must match the servers defined in OpenAPI specs.
	ServerURL string

	// VersionError is called on any error that occurs when trying to resolve the
	// API version.
	VersionError VersionErrorHandler

	// Options further configure the request and response validation. See
	// https://pkg.go.dev/github.com/getkin/kin-openapi/openapi3filter#ValidatorOption
	// for available options.
	Options []openapi3filter.ValidatorOption
}

var defaultValidatorConfig = ValidatorConfig{
	VersionError: DefaultVersionError,
	Options: []openapi3filter.ValidatorOption{
		openapi3filter.OnErr(func(w http.ResponseWriter, status int, code openapi3filter.ErrCode, _ error) {
			statusText := http.StatusText(http.StatusInternalServerError)
			switch code {
			case openapi3filter.ErrCodeCannotFindRoute:
				statusText = "Not Found"
			case openapi3filter.ErrCodeRequestInvalid:
				statusText = "Bad Request"
			}
			http.Error(w, statusText, status)
		}),
	},
}

func today() time.Time {
	now := time.Now().UTC()
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
}

// NewValidator returns a new validation middleware, which validates versioned
// requests according to the given OpenAPI spec versions. For configuration
// defaults, a nil config may be used.
func NewValidator(config *ValidatorConfig, docs ...*openapi3.T) (*Validator, error) {
	if len(docs) == 0 {
		return nil, fmt.Errorf("no OpenAPI versions provided")
	}
	if config == nil {
		config = &defaultValidatorConfig
	}
	if config.ServerURL != "" {
		serverURL, err := url.Parse(config.ServerURL)
		if err != nil {
			return nil, fmt.Errorf("invalid ServerURL: %w", err)
		}
		switch serverURL.Scheme {
		case "http", "https":
		case "":
			return nil, errors.New("invalid ServerURL: missing scheme")
		default:
			return nil, fmt.Errorf(
				"invalid ServerURL: unsupported scheme %q (did you forget to specify the scheme://?)",
				serverURL.Scheme,
			)
		}
	}
	if config.VersionError == nil {
		config.VersionError = DefaultVersionError
	}
	v := &Validator{
		validators: map[vervet.Version]*openapi3filter.Validator{},
		errFunc:    config.VersionError,
		today:      today,
	}
	serviceVersions := make(vervet.VersionSlice, len(docs))
	for i := range docs {
		if config.ServerURL != "" {
			docs[i].Servers = []*openapi3.Server{{URL: config.ServerURL}}
		}
		versionStr, err := vervet.ExtensionString(docs[i].Extensions, vervet.ExtSnykApiVersion)
		if err != nil {
			return nil, err
		}
		version, err := vervet.ParseVersion(versionStr)
		if err != nil {
			return nil, err
		}
		serviceVersions[i] = version
		router, err := gorillamux.NewRouter(docs[i])
		if err != nil {
			return nil, err
		}
		v.validators[version] = openapi3filter.NewValidator(router, config.Options...)
	}
	v.versions = vervet.NewVersionIndex(serviceVersions)
	return v, nil
}

// Middleware returns an http.Handler which wraps the given handler with
// request and response validation according to the requested API version.
func (v *Validator) Middleware(h http.Handler) http.Handler {
	handlers := map[vervet.Version]http.Handler{}
	for version, validator := range v.validators {
		handlers[version] = validator.Middleware(h)
	}
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		versionParam := req.URL.Query().Get("version")
		if versionParam == "" {
			v.errFunc(w, req, http.StatusBadRequest, errors.New("missing required query parameter 'version'"))
			return
		}
		requested, err := vervet.ParseVersion(versionParam)
		if err != nil {
			v.errFunc(w, req, http.StatusBadRequest, err)
			return
		}
		if t := v.today(); requested.Date.After(t) {
			v.errFunc(w, req, http.StatusBadRequest,
				fmt.Errorf("requested version newer than present date %s", t))
			return
		}
		resolvedVersion, err := v.versions.Resolve(requested)
		if err != nil {
			v.errFunc(w, req, http.StatusNotFound, err)
			return
		}
		h, ok := handlers[resolvedVersion]
		if !ok {
			// Crash noisily, as this indicates a serious bug. Should not
			// happen if we've initialized our version maps correctly.
			panic(fmt.Sprintf("missing validator for version %q", resolvedVersion))
		}
		h.ServeHTTP(w, req)
	})
}
