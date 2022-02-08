# versionware

```go
import "github.com/snyk/vervet/v3/versionware"
```

Package versionware provides routing and middleware for building versioned HTTP services\.

## Index

- [Constants](<#constants>)
- [func DefaultVersionError(w http.ResponseWriter, r *http.Request, status int, err error)](<#func-defaultversionerror>)
- [type Handler](<#type-handler>)
  - [func NewHandler(vhs ...VersionHandler) *Handler](<#func-newhandler>)
  - [func (h *Handler) HandleErrors(errFunc VersionErrorHandler)](<#func-handler-handleerrors>)
  - [func (h *Handler) Resolve(requested vervet.Version) (*vervet.Version, http.Handler, error)](<#func-handler-resolve>)
  - [func (h *Handler) ServeHTTP(w http.ResponseWriter, req *http.Request)](<#func-handler-servehttp>)
- [type Validator](<#type-validator>)
  - [func NewValidator(config *ValidatorConfig, docs ...*openapi3.T) (*Validator, error)](<#func-newvalidator>)
  - [func (v *Validator) Middleware(h http.Handler) http.Handler](<#func-validator-middleware>)
- [type ValidatorConfig](<#type-validatorconfig>)
- [type VersionErrorHandler](<#type-versionerrorhandler>)
- [type VersionHandler](<#type-versionhandler>)


## Constants

```go
const (
    // HeaderSnykVersionRequested is a response header acknowledging the API
    // version that was requested.
    HeaderSnykVersionRequested = "snyk-version-requested"

    // HeaderSnykVersionServed is a response header indicating the actual API
    // version that was matched and served the response.
    HeaderSnykVersionServed = "snyk-version-served"
)
```

## func DefaultVersionError

```go
func DefaultVersionError(w http.ResponseWriter, r *http.Request, status int, err error)
```

DefaultVersionError provides a basic implementation of VersionErrorHandler that uses http\.Error\.

## type Handler

Handler is a multiplexing http\.Handler that dispatches requests based on the version query parameter according to vervet's API version matching rules\.

```go
type Handler struct {
    // contains filtered or unexported fields
}
```

<details><summary>Example</summary>
<p>

```go
{
	h := versionware.NewHandler([]versionware.VersionHandler{{
		Version: vervet.MustParseVersion("2021-10-01"),
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if _, err := w.Write([]byte("oct")); err != nil {
				panic(err)
			}
		}),
	}, {
		Version: vervet.MustParseVersion("2021-11-01"),
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if _, err := w.Write([]byte("nov")); err != nil {
				panic(err)
			}
		}),
	}, {
		Version: vervet.MustParseVersion("2021-09-01"),
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if _, err := w.Write([]byte("sept")); err != nil {
				panic(err)
			}
		}),
	}}...)

	s := httptest.NewServer(h)
	defer s.Close()

	resp, err := s.Client().Get(s.URL + "?version=2021-10-31")
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	fmt.Print(string(respBody))

}
```

#### Output

```
oct
```

</p>
</details>

### func NewHandler

```go
func NewHandler(vhs ...VersionHandler) *Handler
```

NewHandler returns a new Handler instance\, which handles versioned requests with the matching version handler\.

### func \(\*Handler\) HandleErrors

```go
func (h *Handler) HandleErrors(errFunc VersionErrorHandler)
```

HandleErrors changes the default error handler to the provided function\. It may be used to control the format of versioning error responses\.

### func \(\*Handler\) Resolve

```go
func (h *Handler) Resolve(requested vervet.Version) (*vervet.Version, http.Handler, error)
```

Resolve returns the resolved version and its associated http\.Handler for the requested version\.

### func \(\*Handler\) ServeHTTP

```go
func (h *Handler) ServeHTTP(w http.ResponseWriter, req *http.Request)
```

ServeHTTP implements http\.Handler with the handler matching the version query parameter on the request\. If no matching version is found\, responds 404\.

## type Validator

Validator provides versioned OpenAPI validation middleware for HTTP requests and responses\.

```go
type Validator struct {
    // contains filtered or unexported fields
}
```

### func NewValidator

```go
func NewValidator(config *ValidatorConfig, docs ...*openapi3.T) (*Validator, error)
```

NewValidator returns a new validation middleware\, which validates versioned requests according to the given OpenAPI spec versions\. For configuration defaults\, a nil config may be used\.

### func \(\*Validator\) Middleware

```go
func (v *Validator) Middleware(h http.Handler) http.Handler
```

Middleware returns an http\.Handler which wraps the given handler with request and response validation according to the requested API version\.

## type ValidatorConfig

ValidatorConfig defines how a new Validator may be configured\.

```go
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
```

## type VersionErrorHandler

VersionErrorHandler defines a function which handles versioning error responses in requests\.

```go
type VersionErrorHandler func(w http.ResponseWriter, r *http.Request, status int, err error)
```

## type VersionHandler

VersionHandler expresses a pairing of Version and http\.Handler\.

```go
type VersionHandler struct {
    Version vervet.Version
    Handler http.Handler
}
```

