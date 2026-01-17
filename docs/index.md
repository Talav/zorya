# Zorya

A Go HTTP API framework for building type-safe, RFC-compliant REST APIs with automatic request/response handling, content negotiation, validation, and comprehensive error handling.

## What is Zorya?

Zorya is a modern Go framework that brings type safety and declarative programming to HTTP APIs. Define your request and response types, add struct tags, and Zorya handles the rest - parsing, validation, serialization, and OpenAPI documentation.

## Key Features

- **Type-Safe Request/Response Handling** - Decode requests and encode responses using Go structs
- **Router Adapters** - Works with Chi, Fiber, and Go 1.22+ standard library
- **Content Negotiation** - Automatic content type negotiation (JSON, CBOR, and custom formats)
- **Request Validation** - Pluggable validation with go-playground/validator support
- **File Upload Support** - Multipart/form-data with binary file handling
- **Route Security** - Declarative authentication, role-based, permission-based, and resource-based authorization
- **RFC 9457 Error Handling** - Structured error responses with machine-readable codes
- **Conditional Requests** - Support for If-Match, If-None-Match, If-Modified-Since, If-Unmodified-Since
- **Streaming Responses** - Server-Sent Events (SSE) and chunked transfer support
- **Response Transformers** - Modify response bodies before serialization
- **Middleware Support** - API-level and route-level middleware chains
- **Route Groups** - Group routes with shared prefixes, middleware, and transformers
- **Request Limits** - Configurable body size limits and read timeouts
- **Default Parameter Values** - Automatic default value application using struct tags
- **OpenAPI 3.1 Generation** - Automatic OpenAPI documentation from code
- **Interactive Documentation** - Built-in API documentation UI with Stoplight Elements

## Quick Example

```go
package main

import (
    "context"
    "net/http"
    
    "github.com/go-chi/chi/v5"
    "github.com/talav/talav/pkg/component/zorya"
    "github.com/talav/talav/pkg/component/zorya/adapters"
)

type CreateUserInput struct {
    Body struct {
        Name  string `json:"name" validate:"required"`
        Email string `json:"email" validate:"required,email"`
    }
}

type CreateUserOutput struct {
    Status int `json:"-"`
    Body struct {
        ID    int    `json:"id"`
        Name  string `json:"name"`
        Email string `json:"email"`
    }
}

func main() {
    router := chi.NewMux()
    adapter := adapters.NewChi(router)
    api := zorya.NewAPI(adapter)
    
    zorya.Post(api, "/users", func(ctx context.Context, input *CreateUserInput) (*CreateUserOutput, error) {
        // Your business logic here
        return &CreateUserOutput{
            Status: http.StatusCreated,
            Body: struct {
                ID    int    `json:"id"`
                Name  string `json:"name"`
                Email string `json:"email"`
            }{
                ID:    1,
                Name:  input.Body.Name,
                Email: input.Body.Email,
            },
        }, nil
    })
    
    http.ListenAndServe(":8080", router)
}
```

## Getting Started

### Installation

```bash
go get github.com/talav/talav/pkg/component/zorya
```

### Next Steps

- **[Tutorial](tutorial/quick-start.md)** - Build your first API in 5 minutes
- **[Introduction](introduction/overview.md)** - Learn about Zorya's architecture and design
- **[Features](features/router-adapters.md)** - Explore all features in depth
- **[How-To Guides](how-to/custom-validators.md)** - Solve specific problems

## Why Zorya?

- **Type Safety**: Catch errors at compile time, not runtime
- **Less Boilerplate**: No manual request parsing or response encoding
- **Self-Documenting**: OpenAPI docs generated from code, never out of sync
- **RFC Compliant**: Built-in support for HTTP standards (RFC 9457 errors, conditional requests, etc.)
- **Framework Agnostic**: Works with your favorite router (Chi, Fiber, stdlib)
- **Production Ready**: Request limits, timeouts, security, streaming

## Documentation

- [Introduction](introduction/overview.md) - Architecture, concepts, why Zorya
- [Tutorial](tutorial/quick-start.md) - Step-by-step guide to building APIs
- [Features](features/router-adapters.md) - Complete feature documentation
- [How-To Guides](how-to/custom-validators.md) - Solutions to common problems
- [API Reference](reference/api.md) - Complete API documentation

## Installation

```bash
go get github.com/talav/talav/pkg/component/zorya
```

## License

See root LICENSE file.
