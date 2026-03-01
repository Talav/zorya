# Zorya

A Go HTTP API framework for building type-safe, RFC-compliant REST APIs with automatic request/response handling, content negotiation, validation, and OpenAPI documentation.

## Features

- **Type-Safe Handlers** – Decode requests and encode responses using plain Go structs
- **Router Adapters** – Works with [Chi, Fiber, and the Go stdlib mux](docs/guides/router-adapters.md)
- **Content Negotiation** – Automatic `Accept`-based negotiation (JSON, CBOR, custom formats)
- **Request Validation** – Pluggable validation with go-playground/validator support
- **Route Security** – Declarative authentication, role-based, permission-based, and resource-based authorization
- **RFC 9457 Error Handling** – Structured error responses with machine-readable codes
- **Conditional Requests** – If-Match, If-None-Match, If-Modified-Since, If-Unmodified-Since
- **Streaming Responses** – Server-Sent Events (SSE) and chunked transfers
- **Response Transformers** – Modify response bodies before serialization
- **Middleware** – Standard `func(http.Handler) http.Handler` chains at API and route level
- **Route Groups** – Shared prefixes, middleware, and transformers
- **Request Limits** – Per-route body size limits and read timeouts
- **Default Values** – Automatic default application via struct tags
- **OpenAPI 3.1** – Spec and interactive docs UI generated automatically

## Installation

```bash
go get github.com/talav/zorya
```

## Quick Start

```go
package main

import (
    "context"
    "log"
    "net/http"

    "github.com/go-chi/chi/v5"
    "github.com/talav/zorya"
    "github.com/talav/zorya/adapters"
)

type GetGreetingInput struct {
    Name string `schema:"name,location=query"`
}

type GetGreetingOutput struct {
    Body struct {
        Message string `json:"message"`
    }
}

func greet(ctx context.Context, input *GetGreetingInput) (*GetGreetingOutput, error) {
    name := input.Name
    if name == "" {
        name = "World"
    }
    out := &GetGreetingOutput{}
    out.Body.Message = "Hello, " + name + "!"
    return out, nil
}

func main() {
    router := chi.NewMux()
    api := zorya.NewAPI(
        adapters.NewChi(router),
        zorya.WithConfig(&zorya.Config{
            OpenAPIPath: "/openapi.json",
            DocsPath:    "/docs",
        }),
    )

    zorya.Get(api, "/greet", greet)

    log.Println("Listening on :8080  —  docs at http://localhost:8080/docs")
    log.Fatal(http.ListenAndServe(":8080", router))
}
```

```bash
curl "http://localhost:8080/greet?name=Alice"
# {"message":"Hello, Alice!"}
```

## Documentation

| Topic | Guide |
|---|---|
| Full quick-start walkthrough | [Tutorial: Quick Start](docs/tutorial/quick-start.md) |
| Defining input structs (path, query, header, body) | [Inputs](docs/guides/inputs.md) |
| Defining output structs (status, headers, body) | [Outputs](docs/guides/outputs.md) |
| Request validation | [Validation](docs/guides/validation.md) |
| Error handling (RFC 9457) | [Errors](docs/guides/errors.md) |
| Content negotiation and custom formats | [Content Negotiation](docs/guides/content-negotiation.md) |
| Middleware | [Middleware](docs/guides/middleware.md) |
| Route groups | [Groups](docs/guides/groups.md) |
| Security and authorization | [Security](docs/guides/security.md) |
| Conditional requests (ETags, If-Match, …) | [Conditional Requests](docs/guides/conditional.md) |
| Streaming / SSE | [Streaming](docs/guides/streaming.md) |
| File uploads | [File Uploads](docs/guides/uploads.md) |
| OpenAPI spec and docs UI | [OpenAPI](docs/guides/openapi.md) |
| Router adapters (Chi, Fiber, stdlib) | [Router Adapters](docs/guides/router-adapters.md) |
| All struct tags reference | [Tags Reference](docs/reference/tags.md) |
| Config reference | [Config Reference](docs/reference/config.md) |
| Error types reference | [Error Reference](docs/reference/errors.md) |
| Example programs | [Examples](docs/examples/index.md) |

## Examples

Runnable examples are in [`examples/`](examples/):

| Example | Description |
|---|---|
| [`basic-crud`](examples/basic-crud/) | CRUD with in-memory store |
| [`auth-jwt`](examples/auth-jwt/) | JWT authentication middleware |
| [`validation`](examples/validation/) | Input validation with go-playground/validator |
| [`file-upload`](examples/file-upload/) | Multipart file uploads |
| [`streaming-sse`](examples/streaming-sse/) | Server-Sent Events |
| [`content-negotiation`](examples/content-negotiation/) | Multiple response formats |
| [`conditional-requests`](examples/conditional-requests/) | ETag-based caching |
| [`route-groups`](examples/route-groups/) | Grouped routes with shared middleware |
| [`fiber-adapter`](examples/fiber-adapter/) | Same handlers on Fiber |
| [`openapi-ui`](examples/openapi-ui/) | OpenAPI spec and Stoplight Elements UI |

## License

MIT
