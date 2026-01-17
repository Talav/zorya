# Overview

Zorya is a modern Go framework for building type-safe HTTP REST APIs. It leverages Go's type system and struct tags to provide automatic request parsing, validation, response encoding, and OpenAPI documentation generation.

## What is Zorya?

Zorya is designed around the principle that **your code should be the single source of truth**. Define your API contract once using Go structs and struct tags, and Zorya automatically:

- Parses HTTP requests into typed structs
- Validates input using go-playground/validator
- Encodes responses based on content negotiation
- Generates OpenAPI 3.1 documentation
- Provides interactive API documentation UI

## Core Concepts

### Type-Safe Handlers

Handlers are just Go functions that take a typed input struct and return a typed output struct:

```go
func createUser(ctx context.Context, input *CreateUserInput) (*CreateUserOutput, error) {
    // Your business logic here
}
```

No manual request parsing. No manual response encoding. Just pure business logic.

### Struct Tags as API Contract

Struct tags define your API contract:

```go
type CreateUserInput struct {
    // Path parameters
    OrgID string `schema:"orgId,location=path"`
    
    // Query parameters
    Page int `schema:"page,location=query" default:"1"`
    
    // Headers
    APIKey string `schema:"X-API-Key,location=header"`
    
    // Request body
    Body struct {
        Name  string `json:"name" validate:"required,min=3"`
        Email string `json:"email" validate:"required,email"`
    } `body:"structured"`
}
```

### Automatic OpenAPI Generation

OpenAPI documentation is generated from your struct tags and stays in sync with your code:

```go
// This code generates complete OpenAPI documentation:
zorya.Post(api, "/orgs/{orgId}/users", createUser,
    zorya.Secure(
        zorya.Roles("admin"),
    ),
)
```

## Design Philosophy

### 1. Code as Documentation

Your code **is** your API documentation. Struct tags define the contract, and OpenAPI docs are automatically generated. Documentation can never be out of sync because it's derived from the code.

### 2. Framework Agnostic

Zorya works with multiple routers through adapters:

- Chi
- Fiber  
- Go 1.22+ standard library

Your business logic stays the same regardless of which router you use.

### 3. Type Safety

Catch errors at compile time, not runtime. If your code compiles, you know:

- All required fields are present
- Types are correct
- Validation rules are defined

### 4. RFC Compliant

Built-in support for HTTP standards:

- RFC 9457 (Problem Details for HTTP APIs)
- Conditional requests (If-Match, If-None-Match, ETags)
- Content negotiation
- Standard status codes

### 5. Separation of Concerns

Zorya focuses on HTTP concerns (routing, parsing, encoding). Your business logic stays pure and testable:

```go
// Pure business logic - no HTTP dependencies
func (s *UserService) CreateUser(ctx context.Context, name, email string) (*User, error) {
    // Domain logic
}

// HTTP adapter - thin wrapper
func createUserHandler(ctx context.Context, input *CreateUserInput) (*CreateUserOutput, error) {
    user, err := userService.CreateUser(ctx, input.Body.Name, input.Body.Email)
    return &CreateUserOutput{Body: user}, err
}
```

## Architecture

```
┌──────────────────────────────────────────────────┐
│  HTTP Request                                     │
└────────────────┬─────────────────────────────────┘
                 │
                 ▼
┌──────────────────────────────────────────────────┐
│  Router Adapter (Chi/Fiber/Stdlib)               │
│  - Extracts path parameters                      │
│  - Converts to Zorya Context                     │
└────────────────┬─────────────────────────────────┘
                 │
                 ▼
┌──────────────────────────────────────────────────┐
│  API Middleware Chain                            │
│  - Security enforcement                          │
│  - Logging                                       │
│  - Custom middleware                             │
└────────────────┬─────────────────────────────────┘
                 │
                 ▼
┌──────────────────────────────────────────────────┐
│  Route Middleware Chain                          │
│  - Route-specific middleware                     │
└────────────────┬─────────────────────────────────┘
                 │
                 ▼
┌──────────────────────────────────────────────────┐
│  Request Decoder                                 │
│  - Parses path/query/header/cookie params        │
│  - Decodes body (JSON/CBOR/multipart)            │
│  - Applies default values                        │
│  - Creates typed input struct                    │
└────────────────┬─────────────────────────────────┘
                 │
                 ▼
┌──────────────────────────────────────────────────┐
│  Validator                                       │
│  - Validates using struct tags                   │
│  - Returns structured errors                     │
└────────────────┬─────────────────────────────────┘
                 │
                 ▼
┌──────────────────────────────────────────────────┐
│  Handler Function                                │
│  - Your business logic                           │
│  - Returns typed output struct or error          │
└────────────────┬─────────────────────────────────┘
                 │
                 ▼
┌──────────────────────────────────────────────────┐
│  Response Transformers                           │
│  - Modify response bodies                        │
│  - Group and API-level transformers              │
└────────────────┬─────────────────────────────────┘
                 │
                 ▼
┌──────────────────────────────────────────────────┐
│  Response Encoder                                │
│  - Content negotiation                           │
│  - Encodes body (JSON/CBOR/custom)               │
│  - Sets status code and headers                  │
└────────────────┬─────────────────────────────────┘
                 │
                 ▼
┌──────────────────────────────────────────────────┐
│  HTTP Response                                    │
└──────────────────────────────────────────────────┘
```

## Key Components

### API

The main API instance that manages routes, middleware, and configuration.

```go
api := zorya.NewAPI(adapter,
    zorya.WithValidator(validator),
    zorya.WithFormat("application/xml", xmlFormat),
)
```

### Adapters

Adapters connect Zorya to different HTTP routers:

- `adapters.NewChi(router)` - Chi router
- `adapters.NewFiber(app)` - Fiber framework
- `adapters.NewStdlib(mux)` - Go stdlib ServeMux

### Context

The `Context` interface provides access to request information:

```go
type Context interface {
    Context() context.Context
    Request() *http.Request
    RouterParams() map[string]string
    Header(name string) string
    SetReadDeadline(t time.Time) error
    BodyWriter() http.ResponseWriter
}
```

### Schema System

The schema system handles request parsing and OpenAPI generation:

- Extracts metadata from struct tags
- Generates JSON Schema for validation
- Creates OpenAPI parameter definitions
- Applies default values

### Validator

Pluggable validation interface with go-playground/validator adapter:

```go
type Validator interface {
    Validate(ctx context.Context, input any) []error
}
```

## Next Steps

- [Why Zorya](why-zorya.md) - Comparison with alternatives
- [Installation](installation.md) - Get started
- [Quick Start Tutorial](../tutorial/quick-start.md) - Build your first API
