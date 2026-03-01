package zorya

import (
	"context"
	"errors"
	"fmt"
	"io"
	"maps"
	"net/http"
	"reflect"
	"strings"

	"github.com/talav/mapstructure"
	"github.com/talav/negotiation"
	"github.com/talav/openapi"
	"github.com/talav/schema"
)

// Adapter is an interface that allows the API to be used with different HTTP
// routers and frameworks. It is designed to work with the standard library
// `http.Request` and `http.ResponseWriter` types as well as types like
// `gin.Context` or `fiber.Ctx` that provide both request and response
// functionality in one place, by using the `zorya.Context` interface which
// abstracts away those router-specific differences.
type Adapter interface {
	// Handle registers a route with a standard http.HandlerFunc.
	Handle(route *BaseRoute, handler http.HandlerFunc)

	// ExtractRouterParams extracts router parameters from the request.
	ExtractRouterParams(r *http.Request, route *BaseRoute) map[string]string

	// ServeHTTP makes the adapter compatible with http.Handler interface.
	// This allows the adapter to be used directly with http.ListenAndServe,
	// mounted as a sub-handler in other routers, or used in testing scenarios
	// (e.g., with httptest.NewRecorder).
	ServeHTTP(http.ResponseWriter, *http.Request)
}

// Codec is an interface that allows to map request and router parameters into Go structures.
type Codec interface {
	// DecodeRequest decodes an HTTP request into the provided struct.
	DecodeRequest(request *http.Request, routerParams map[string]string, result any) error
}

// Validator validates input structs after request decoding.
// Each returned error should implement ErrorDetailer for RFC 9457 compliant responses.
type Validator interface {
	// Validate validates the input struct.
	// Returns nil if validation succeeds, or a slice of errors if validation fails.
	Validate(ctx context.Context, input any, metadata *schema.StructMetadata) []error
}

// Transformer is a function that transforms response bodies before serialization.
// Transformers are run in the order they are added.
// Parameters:
//   - r: The HTTP request (for context-aware transformations)
//   - status: The HTTP status code as an integer (e.g., 200, 404)
//   - result: The response body value to transform (must be a struct from the output struct's Body field)
//
// Returns the transformed value (which may be the same or a different type) and an error.
// Each transformer receives the output of the previous transformer in the chain.
// Note: Transformers are only called for struct body types. []byte and function bodies
// bypass transformers and are handled separately.
type Transformer func(r *http.Request, status int, result any) (any, error)

// API is the core interface for a Zorya API. It provides request/response handling,
// content negotiation, validation, and OpenAPI spec generation.
//
//nolint:interfacebloat // API is the core framework interface; 14 methods is reasonable for a complete API contract
type API interface {
	// Adapter returns the router adapter for this API, providing a generic
	// interface to get request information and write responses.
	Adapter() Adapter

	// Middlewares returns a slice of middleware handler functions that will be
	// run for all operations. Middleware are run in the order they are added.
	Middlewares() Middlewares

	// UseMiddleware adds one or more standard Go middleware functions to the API.
	// Middleware functions take an http.Handler and return an http.Handler.
	UseMiddleware(middlewares ...Middleware)

	// Codec returns the schema codec used for request decoding and response encoding.
	Codec() Codec

	// Metadata returns the schema metadata instance used by this API.
	Metadata() *schema.Metadata

	// Negotiate returns the best content type for the response based on the
	// Accept header. If no match is found, returns the default format.
	Negotiate(accept string) (string, error)

	// Marshal writes the value to the writer using the format for the given
	// content type. Supports plus-segment matching (e.g., application/vnd.api+json).
	// If marshaling fails, it falls back to plain text representation.
	Marshal(w io.Writer, contentType string, v any)

	// Validator returns the configured validator, or nil if validation is disabled.
	Validator() Validator

	// Transform runs all transformers on the response value.
	// Called automatically during response serialization.
	Transform(r *http.Request, status int, v any) (any, error)

	// UseTransformer adds one or more transformer functions that will be
	// run on all responses.
	UseTransformer(transformers ...Transformer)

	// OpenAPI returns the OpenAPI spec for this API. You may edit this spec
	// until the server starts.
	OpenAPI() *OpenAPI

	// addOperationToState registers an operation for OpenAPI generation.
	// Internal method used during route registration.
	addOperationToState(op openapi.Operation)
}

// Option configures an API.
type Option func(*api)

type api struct {
	adapter          Adapter
	middlewares      Middlewares
	codec            *schema.Codec
	metadata         *schema.Metadata
	formats          map[string]Format
	formatKeys       []string
	defaultFormat    string
	negotiator       *negotiation.Negotiator
	validator        Validator
	transformers     []Transformer
	config           *Config
	openAPI          *OpenAPI
	openapiState     *openapiState // Uses github.com/talav/openapi for schema generation
}

func (a *api) Adapter() Adapter {
	return a.adapter
}

func (a *api) Middlewares() Middlewares {
	return a.middlewares
}

// UseMiddleware adds one or more standard Go middleware functions to the API.
func (a *api) UseMiddleware(middlewares ...Middleware) {
	a.middlewares = append(a.middlewares, middlewares...)
}

func (a *api) Codec() Codec {
	return a.codec
}

func (a *api) Metadata() *schema.Metadata {
	return a.metadata
}

func (a *api) Validator() Validator {
	return a.validator
}

func (a *api) OpenAPI() *OpenAPI {
	return a.openAPI
}

func (a *api) addOperationToState(op openapi.Operation) {
	a.openapiState.AddOperation(op)
}

// buildOpenapiOperation converts Zorya operation metadata to openapi.Operation.
// This is called during route registration to build the operation immediately.
func buildOpenapiOperation(method, path string, inputType, outputType reflect.Type, route *BaseRoute) openapi.Operation {
	// Build operation options
	var opts []openapi.OperationDocOption

	// Add operation metadata
	if route.Operation != nil {
		if route.Operation.Summary != "" {
			opts = append(opts, openapi.WithSummary(route.Operation.Summary))
		}
		if route.Operation.Description != "" {
			opts = append(opts, openapi.WithDescription(route.Operation.Description))
		}
		if route.Operation.OperationID != "" {
			opts = append(opts, openapi.WithOperationID(route.Operation.OperationID))
		}
		if len(route.Operation.Tags) > 0 {
			opts = append(opts, openapi.WithTags(route.Operation.Tags...))
		}
		if route.Operation.Deprecated {
			opts = append(opts, openapi.WithDeprecated())
		}
	}

	// Add request schema if input type is provided
	if inputType != nil && inputType.Kind() == reflect.Struct {
		inputInstance := reflect.New(inputType).Elem().Interface()
		opts = append(opts, openapi.WithRequest(inputInstance))
	}

	// Add response schema if output type is provided
	if outputType != nil && outputType.Kind() == reflect.Struct {
		outputInstance := reflect.New(outputType).Elem().Interface()
		status := route.DefaultStatus
		if status == 0 {
			status = 200
		}
		opts = append(opts, openapi.WithResponse(status, outputInstance))
	}

	// Add error responses
	errorInstance := ErrorModel{}
	if len(route.Errors) > 0 {
		for _, code := range route.Errors {
			opts = append(opts, openapi.WithResponse(code, errorInstance))
		}
	}

	// Add automatic error responses (422 if has input, 500 always)
	if inputType != nil {
		opts = append(opts, openapi.WithResponse(422, errorInstance))
	}
	opts = append(opts, openapi.WithResponse(500, errorInstance))

	// Add security requirements
	if route.Security != nil {
		scopes := make([]string, 0, len(route.Security.Roles)+len(route.Security.Permissions))
		scopes = append(scopes, route.Security.Roles...)
		scopes = append(scopes, route.Security.Permissions...)
		opts = append(opts, openapi.WithSecurity("bearerAuth", scopes...))
	}

	// Create operation with appropriate HTTP method
	var op openapi.Operation
	switch method {
	case http.MethodGet:
		op = openapi.GET(path, opts...)
	case http.MethodPost:
		op = openapi.POST(path, opts...)
	case http.MethodPut:
		op = openapi.PUT(path, opts...)
	case http.MethodPatch:
		op = openapi.PATCH(path, opts...)
	case http.MethodDelete:
		op = openapi.DELETE(path, opts...)
	case http.MethodHead:
		op = openapi.HEAD(path, opts...)
	case http.MethodOptions:
		op = openapi.OPTIONS(path, opts...)
	default:
		op = openapi.GET(path, opts...) // Fallback
	}

	return op
}

// Transform runs all transformers on the response value in the order they were added.
func (a *api) Transform(r *http.Request, status int, v any) (any, error) {
	for _, t := range a.transformers {
		var err error
		v, err = t(r, status, v)
		if err != nil {
			return nil, err
		}
	}

	return v, nil
}

// UseTransformer adds one or more transformer functions that will be run on all responses.
func (a *api) UseTransformer(transformers ...Transformer) {
	a.transformers = append(a.transformers, transformers...)
}

// Negotiate returns the best content type based on the Accept header.
func (a *api) Negotiate(accept string) (string, error) {
	if accept == "" {
		return a.defaultFormat, nil
	}

	header, err := a.negotiator.Negotiate(accept, a.formatKeys, false)
	if errors.Is(err, negotiation.ErrNoMatch) {
		// Fallback to default format when no match
		return a.defaultFormat, nil
	}

	if err != nil {
		return "", fmt.Errorf("negotiation failed: %w", err)
	}

	return header.Type, nil
}

// Marshal writes the value using the format for the given content type.
// If marshaling fails, it falls back to plain text representation.
func (a *api) Marshal(w io.Writer, ct string, v any) {
	f, ok := a.formats[ct]
	if !ok {
		// Try extracting suffix from plus-segment (e.g., application/vnd.api+json -> json).
		if idx := strings.LastIndex(ct, "+"); idx != -1 {
			f, ok = a.formats[ct[idx+1:]]
		}
	}

	if !ok {
		// Unknown content type - fallback to plain text
		_, _ = fmt.Fprintf(w, "%v", v)

		return
	}

	if err := f.Marshal(w, v); err != nil {
		// Marshaling failed - fallback to plain text
		_, _ = fmt.Fprintf(w, "%v", v)
	}
}

// NewAPI creates a new API instance with the given adapter and options.
// The adapter is required; all other configuration is optional.
//
// Example:
//
//	api := zorya.NewAPI(adapter)
//	api := zorya.NewAPI(adapter, zorya.WithValidator(validator))
//	api := zorya.NewAPI(adapter, zorya.WithFormat("application/xml", xmlFormat))
//	api := zorya.NewAPI(adapter, zorya.WithFormats(customFormats))
//	api := zorya.NewAPI(adapter, zorya.WithFormatsReplace(formats)) // Replace all formats
func NewAPI(adapter Adapter, opts ...Option) API {
	a := &api{
		adapter:       adapter,
		middlewares:   Middlewares{},
		defaultFormat: "application/json",
		negotiator:    negotiation.NewMediaNegotiator(),
		transformers:  []Transformer{},
	}

	// Apply options
	for _, opt := range opts {
		opt(a)
	}

	// Set defaults for anything not configured
	// Create metadata first, then codec uses it
	if a.metadata == nil {
		a.metadata = NewMetadata()
	}
	if a.codec == nil {
		a.codec = schema.NewCodec(a.metadata, mapstructure.NewDefaultUnmarshaler(), schema.NewDefaultDecoder())
	}

	if a.formats == nil {
		a.formats = DefaultFormats()
	}

	initializeOpenAPI(a)

	if a.config == nil {
		a.config = DefaultConfig()
	}

	// Build format keys from formats
	if len(a.formatKeys) == 0 {
		a.formatKeys = make([]string, 0, len(a.formats))
		for k := range a.formats {
			// Only include full content types, not suffixes
			if strings.Contains(k, "/") {
				a.formatKeys = append(a.formatKeys, k)
			}
		}
	}

	// Initialize the openapi state that uses github.com/talav/openapi library
	a.openapiState = newOpenapiState(a)

	registerOpenAPIEndpoint(a)
	registerDocsEndpoint(a)

	return a
}

// initializeOpenAPI initializes the OpenAPI spec and its Components if needed.
func initializeOpenAPI(a *api) {
	if a.openAPI == nil {
		a.openAPI = DefaultOpenAPI("API", "1.0.0")
	}

	// Initialize OpenAPI Components if needed
	if a.openAPI.Components == nil {
		a.openAPI.Components = &Components{
			Schemas: make(map[string]*Schema),
		}
	} else if a.openAPI.Components.Schemas == nil {
		a.openAPI.Components.Schemas = make(map[string]*Schema)
	}
}

// registerOpenAPIEndpoint registers the OpenAPI spec endpoint if configured.
func registerOpenAPIEndpoint(a *api) {
	if a.config.OpenAPIPath == "" {
		return
	}

	a.adapter.Handle(&BaseRoute{
		Method: http.MethodGet,
		Path:   a.config.OpenAPIPath,
	}, func(w http.ResponseWriter, r *http.Request) {
		// Generate spec with caching and ETag support
		specJSON, etag, err := a.openapiState.GenerateSpec(r.Context())
		if err != nil {
			WriteErr(a, r, w, http.StatusInternalServerError, "failed to generate OpenAPI spec", err)
			return
		}

		// Set ETag header
		w.Header().Set("ETag", etag)

		// Check If-None-Match for 304 Not Modified
		if match := r.Header.Get("If-None-Match"); match == etag {
			w.WriteHeader(http.StatusNotModified)
			return
		}

		w.Header().Set("Content-Type", "application/vnd.oai.openapi+json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(specJSON)
	})
}

// WithValidator sets a validator for request validation.
func WithValidator(validator Validator) Option {
	return func(a *api) {
		a.validator = validator
	}
}

// WithFormat adds a single format for content negotiation.
// Multiple calls to WithFormat can be chained to add multiple formats.
// Formats are merged with default formats, with later formats taking precedence.
// Default formats are automatically included, so you don't need to add them manually.
func WithFormat(contentType string, format Format) Option {
	return func(a *api) {
		// Ensure defaults are loaded
		if a.formats == nil {
			a.formats = DefaultFormats()
		}
		a.formats[contentType] = format
	}
}

// WithFormats sets custom formats for content negotiation.
// Custom formats are merged with default formats, with custom formats taking precedence.
// Default formats are automatically included, so you don't need to add them manually.
func WithFormats(formats map[string]Format) Option {
	return func(a *api) {
		// Start with defaults and merge custom formats
		if a.formats == nil {
			a.formats = DefaultFormats()
		}
		maps.Copy(a.formats, formats)
	}
}

// WithFormatsReplace replaces all formats (does not merge with defaults).
// Use this when you want complete control over supported formats.
// Default formats are NOT included unless you explicitly add them.
func WithFormatsReplace(formats map[string]Format) Option {
	return func(a *api) {
		// Replace all formats - don't merge with defaults
		a.formats = make(map[string]Format, len(formats))
		maps.Copy(a.formats, formats)
	}
}

// WithMetadata sets a custom metadata instance for schema operations.
func WithMetadata(metadata *schema.Metadata) Option {
	return func(a *api) {
		a.metadata = metadata
	}
}

// WithCodec sets a custom codec for request/response encoding/decoding.
// Note: If you use WithCodec, you should also use WithMetadata to ensure
// the metadata instance matches the codec's metadata.
func WithCodec(codec *schema.Codec) Option {
	return func(a *api) {
		a.codec = codec
	}
}

// WithDefaultFormat sets the default content type when Accept header is missing or no match is found.
func WithDefaultFormat(format string) Option {
	return func(a *api) {
		a.defaultFormat = format
	}
}

// WithConfig sets the API configuration (OpenAPI path, docs path, default format, etc.).
func WithConfig(config *Config) Option {
	return func(a *api) {
		a.config = config
	}
}

// WithOpenAPI sets a custom OpenAPI spec for the API. If not set, a default spec is created.
func WithOpenAPI(openAPI *OpenAPI) Option {
	return func(a *api) {
		a.openAPI = openAPI
	}
}

// Register an operation handler for an API. The handler must be a function that
// takes a context and a pointer to the input struct and returns a pointer to the
// output struct and an error. The input struct must be a struct with fields
// for the request path/query/header/cookie parameters and/or body. The output
// struct must be a struct with fields for the output headers and body of the
// operation, if any.

func Register[I, O any](api API, route BaseRoute, handler func(context.Context, *I) (*O, error)) error {
	inputType := reflect.TypeFor[I]()
	if inputType.Kind() != reflect.Struct {
		return fmt.Errorf("input type %s must be a struct", inputType)
	}
	outputType := reflect.TypeFor[O]()
	if outputType.Kind() != reflect.Struct {
		return fmt.Errorf("output type %s must be a struct", outputType)
	}

	// Build and register OpenAPI operation immediately during route registration
	op := buildOpenapiOperation(route.Method, route.Path, inputType, outputType, &route)
	api.addOperationToState(op)

	// Create and register HTTP handler (routing logic remains unchanged)
	httpHandler := createRequestHandler(api, &route, handler)

	// Build middleware chain:
	// 1. Router params extraction
	// 2. Security metadata middleware (if Secure() was used)
	// 3. API-level middlewares
	// 4. Route-specific middlewares
	allMiddlewares := Middlewares{newRouterParamsMiddleware(api.Adapter(), &route)}
	if securityMiddleware := newSecurityMetadataMiddleware(route.Security); securityMiddleware != nil {
		allMiddlewares = append(allMiddlewares, securityMiddleware)
	}
	allMiddlewares = append(allMiddlewares, api.Middlewares()...)
	allMiddlewares = append(allMiddlewares, route.Middlewares...)
	finalHandler := allMiddlewares.Apply(http.HandlerFunc(httpHandler))

	api.Adapter().Handle(&route, finalHandler.ServeHTTP)

	return nil
}

// createRequestHandler creates the HTTP handler for processing requests.
func createRequestHandler[I, O any](api API, route *BaseRoute, handler func(context.Context, *I) (*O, error)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// Router params are extracted by RouterParamsMiddleware and stored in context
		routerParams := GetRouterParams(r)

		// Setup request limits
		setupRequestLimits(r, w, *route)

		// Decode and validate request
		input := new(I)
		if err := decodeAndValidateRequest(api, r, routerParams, input); err != nil {
			WriteErr(api, r, w, 0, "", err)

			return
		}

		// Execute handler
		output, err := handler(r.Context(), input)
		if err != nil {
			WriteErr(api, r, w, 0, "", err)

			return
		}

		// Transform and write response
		defaultStatus := route.DefaultStatus
		if defaultStatus == 0 {
			defaultStatus = http.StatusOK
		}
		if err := transformAndWriteResponse(api, r, w, output, defaultStatus); err != nil {
			return // Error already written
		}
	}
}

// decodeAndValidateRequest decodes and validates the request input.
func decodeAndValidateRequest[I any](api API, r *http.Request, routerParams map[string]string, input *I) error {
	if err := api.Codec().DecodeRequest(r, routerParams, input); err != nil {
		return err
	}

	if errs := validateRequest(api, r, input); len(errs) > 0 {
		return NewError(http.StatusUnprocessableEntity, "validation failed", errs...)
	}

	return nil
}

// transformAndWriteResponse transforms the output and writes the response.
func transformAndWriteResponse[O any](api API, r *http.Request, w http.ResponseWriter, output *O, defaultStatus int) error {
	statusCode := defaultStatus
	transformed, err := api.Transform(r, statusCode, output)
	if err != nil {
		WriteErr(api, r, w, http.StatusInternalServerError, "transformer error", err)

		return err
	}

	transformedOutput, ok := transformed.(*O)
	if !ok {
		err := fmt.Errorf("transformer returned unexpected type")
		WriteErr(api, r, w, http.StatusInternalServerError, "transformer error", err)

		return err
	}

	if err := writeResponse(api, r, w, transformedOutput, statusCode); err != nil {
		WriteErr(api, r, w, http.StatusInternalServerError, "failed to write response", err)

		return err
	}

	return nil
}

// Get registers a GET route handler.
// Panics on errors since route registration happens during startup
// and errors represent programming/configuration mistakes.
func Get[I, O any](api API, path string, handler func(context.Context, *I) (*O, error), options ...func(*BaseRoute)) {
	convenience(api, http.MethodGet, path, handler, options...)
}

// Post registers a POST route handler.
// Panics on errors since route registration happens during startup
// and errors represent programming/configuration mistakes.
func Post[I, O any](api API, path string, handler func(context.Context, *I) (*O, error), options ...func(*BaseRoute)) {
	convenience(api, http.MethodPost, path, handler, options...)
}

// Put registers a PUT route handler.
// Panics on errors since route registration happens during startup
// and errors represent programming/configuration mistakes.
func Put[I, O any](api API, path string, handler func(context.Context, *I) (*O, error), options ...func(*BaseRoute)) {
	convenience(api, http.MethodPut, path, handler, options...)
}

// Delete registers a DELETE route handler.
// Panics on errors since route registration happens during startup
// and errors represent programming/configuration mistakes.
func Delete[I, O any](api API, path string, handler func(context.Context, *I) (*O, error), options ...func(*BaseRoute)) {
	convenience(api, http.MethodDelete, path, handler, options...)
}

// Patch registers a PATCH route handler.
// Panics on errors since route registration happens during startup
// and errors represent programming/configuration mistakes.
func Patch[I, O any](api API, path string, handler func(context.Context, *I) (*O, error), options ...func(*BaseRoute)) {
	convenience(api, http.MethodPatch, path, handler, options...)
}

// Head registers a HEAD route handler.
// Panics on errors since route registration happens during startup
// and errors represent programming/configuration mistakes.
func Head[I, O any](api API, path string, handler func(context.Context, *I) (*O, error), options ...func(*BaseRoute)) {
	convenience(api, http.MethodHead, path, handler, options...)
}

// convenience is a helper function used by Get, Post, Put, Delete, Patch, and Head.
// Panics on errors since route registration happens during startup and errors
// represent programming/configuration mistakes that should fail fast.
func convenience[I, O any](api API, method, path string, handler func(context.Context, *I) (*O, error), options ...func(o *BaseRoute)) {
	// generate operation id, generate summary, generate base route, execute all options
	route := BaseRoute{
		Method: method,
		Path:   path,
	}
	for _, o := range options {
		o(&route)
	}

	if err := Register(api, route, handler); err != nil {
		panic(fmt.Errorf("failed to register route %s %s: %w", method, path, err))
	}
}
