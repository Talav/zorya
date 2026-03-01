package zorya

import (
	"context"
	"crypto/sha256"
	"fmt"
	"sync"

	"github.com/talav/openapi"
)

// openapiState manages OpenAPI specification state.
// It stores operations directly and generates the spec lazily with caching.
type openapiState struct {
	openapiAPI *openapi.API
	operations []openapi.Operation

	// Cache for lazy generation
	specCache []byte
	specETag  string

	mu sync.RWMutex
}

// newOpenapiState creates a new OpenAPI state manager.
func newOpenapiState(a *api) *openapiState {
	// Build openapi.API with Zorya's configuration
	opts := buildOpenapiOptions(a)
	openapiAPI := openapi.NewAPI(opts...)

	return &openapiState{
		openapiAPI: openapiAPI,
		operations: make([]openapi.Operation, 0),
	}
}

// AddOperation adds an operation to the OpenAPI spec.
// This invalidates the cached spec.
func (s *openapiState) AddOperation(op openapi.Operation) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.operations = append(s.operations, op)

	// Invalidate cache
	s.specCache = nil
	s.specETag = ""
}

// GenerateSpec generates the OpenAPI specification.
// Results are cached until a new operation is added.
func (s *openapiState) GenerateSpec(ctx context.Context) ([]byte, string, error) {
	// Fast path: check cache with read lock
	s.mu.RLock()
	if s.specCache != nil {
		cache, etag := s.specCache, s.specETag
		s.mu.RUnlock()
		return cache, etag, nil
	}
	s.mu.RUnlock()

	// Slow path: generate with write lock
	s.mu.Lock()
	defer s.mu.Unlock()

	// Double-check after acquiring write lock
	if s.specCache != nil {
		return s.specCache, s.specETag, nil
	}

	// Generate spec using API method
	result, err := s.openapiAPI.Generate(ctx, s.operations...)
	if err != nil {
		return nil, "", err
	}

	// Cache the result
	s.specCache = result.JSON
	s.specETag = fmt.Sprintf(`"%x"`, sha256.Sum256(result.JSON))

	return s.specCache, s.specETag, nil
}

// buildOpenapiOptions creates openapi.Option slice from Zorya's configuration.
func buildOpenapiOptions(a *api) []openapi.Option {
	opts := []openapi.Option{
		openapi.WithVersion("3.1.2"),
		openapi.WithInfoTitle(a.openAPI.Info.Title),
		openapi.WithInfoVersion(a.openAPI.Info.Version),
	}

	if a.openAPI.Info.Description != "" {
		opts = append(opts, openapi.WithInfoDescription(a.openAPI.Info.Description))
	}

	// Add servers
	for _, server := range a.openAPI.Servers {
		opts = append(opts, openapi.WithServer(server.URL,
			openapi.WithServerDescription(server.Description)))
	}

	// Add security schemes
	if a.openAPI.Components != nil && a.openAPI.Components.SecuritySchemes != nil {
		for name, scheme := range a.openAPI.Components.SecuritySchemes {
			if scheme.Type == "http" && scheme.Scheme == "bearer" {
				opts = append(opts, openapi.WithBearerAuth(name, scheme.Description))
			} else if scheme.Type == "apiKey" {
				opts = append(opts, openapi.WithAPIKey(name, scheme.Name,
					openapi.ParameterLocation(scheme.In), scheme.Description))
			}
		}
	}

	return opts
}
