# Zorya

A Go HTTP API framework for building type-safe, RFC-compliant REST APIs with automatic request/response handling, content negotiation, validation, and OpenAPI documentation.

## What is Zorya?

Zorya brings type safety and declarative programming to HTTP APIs. Define your request and response types, add struct tags, and Zorya handles parsing, validation, serialization, and OpenAPI documentation — automatically.

## Key Features

- **Type-Safe Handlers** — Generic `Register[I, O]` ensures input and output types are checked at compile time
- **Router Adapters** — Works with Chi, Fiber, and the Go 1.22+ standard library
- **Content Negotiation** — Automatic JSON/CBOR negotiation via `Accept` header
- **Request Validation** — Pluggable validation; go-playground/validator works out of the box
- **File Upload Support** — Multipart/form-data with `*multipart.FileHeader` fields
- **Route Security** — Declarative role, permission, and resource-based authorization
- **RFC 9457 Error Handling** — Structured error responses with machine-readable codes
- **Conditional Requests** — `If-Match`, `If-None-Match`, `If-Modified-Since`, `If-Unmodified-Since`
- **Streaming Responses** — Server-Sent Events and chunked transfer
- **Response Transformers** — Modify response bodies before serialization
- **Middleware Support** — API-level and route-level middleware chains
- **Route Groups** — Shared prefixes, middleware, and transformers
- **OpenAPI 3.1** — Automatic spec generation via [talav/openapi](https://github.com/talav/openapi)
- **Interactive Docs UI** — Built-in Stoplight Elements UI

## Quick Example

```go
package main

import (
    "context"
    "net/http"

    "github.com/go-chi/chi/v5"
    "github.com/talav/zorya"
    "github.com/talav/zorya/adapters"
)

type CreateUserInput struct {
    Body struct {
        Name  string `json:"name"  validate:"required"`
        Email string `json:"email" validate:"required,email"`
    } `body:"structured"`
}

type CreateUserOutput struct {
    Status int `json:"-"`
    Body   struct {
        ID    int    `json:"id"`
        Name  string `json:"name"`
        Email string `json:"email"`
    }
}

func main() {
    router := chi.NewMux()
    api := zorya.NewAPI(adapters.NewChi(router))

    zorya.Post(api, "/users", func(ctx context.Context, input *CreateUserInput) (*CreateUserOutput, error) {
        return &CreateUserOutput{
            Status: http.StatusCreated,
            Body: struct {
                ID    int    `json:"id"`
                Name  string `json:"name"`
                Email string `json:"email"`
            }{ID: 1, Name: input.Body.Name, Email: input.Body.Email},
        }, nil
    })

    http.ListenAndServe(":8080", router)
}
```

## Installation

```bash
go get github.com/talav/zorya
```

## Getting Started

- **[Quick Start](tutorial/quick-start.md)** — Build your first API in 5 minutes
- **[Router Adapters](guides/router-adapters.md)** — Choose and configure Chi, Fiber, or stdlib
- **[Defining Inputs](guides/inputs.md)** — Path params, query params, headers, and body
- **[Error Handling](guides/errors.md)** — RFC 9457 structured errors

## Why Zorya?

- **Type Safety** — Catch request/response shape errors at compile time
- **Less Boilerplate** — No manual `json.Unmarshal`, no manual status code juggling
- **Self-Documenting** — OpenAPI spec generated from code, always in sync
- **RFC Compliant** — RFC 9457 errors, conditional requests, content negotiation
- **Framework Agnostic** — Bring your own router (Chi, Fiber, stdlib)
