# Quick Start

Build a working REST API with automatic OpenAPI documentation in about 5 minutes.

## Prerequisites

Go 1.22 or newer. Verify with:

```bash
go version
```

## 1. Create a new project

```bash
mkdir my-api && cd my-api
go mod init my-api
go get github.com/talav/zorya
go get github.com/go-chi/chi/v5
```

## 2. Write the API

Create `main.go`:

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

// --- Input and output types ---

type GetGreetingInput struct {
    Name string `schema:"name,location=query"`
}

type GetGreetingOutput struct {
    Body struct {
        Message string `json:"message"`
    }
}

// --- Handler ---

func greet(ctx context.Context, input *GetGreetingInput) (*GetGreetingOutput, error) {
    name := input.Name
    if name == "" {
        name = "World"
    }
    out := &GetGreetingOutput{}
    out.Body.Message = "Hello, " + name + "!"
    return out, nil
}

// --- Wire up ---

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

    log.Println("Listening on :8080")
    log.Fatal(http.ListenAndServe(":8080", router))
}
```

## 3. Run it

```bash
go run main.go
```

## 4. Try it

Call the endpoint:

```bash
curl "http://localhost:8080/greet?name=Alice"
# {"message":"Hello, Alice!"}
```

View the OpenAPI spec:

```bash
curl http://localhost:8080/openapi.json
```

Open the interactive docs UI in your browser:

```
http://localhost:8080/docs
```

## What just happened?

Zorya did the following automatically:

1. **Decoded the query parameter** `name` from the URL into `GetGreetingInput.Name`
2. **Serialized the response** struct as JSON with the correct `Content-Type` header
3. **Generated an OpenAPI 3.1 spec** from your input/output types
4. **Served Stoplight Elements** at `/docs` pointing at that spec

## Next steps

- [Defining Inputs](../guides/inputs.md) — add path params, headers, and a request body
- [Defining Outputs](../guides/outputs.md) — custom status codes and response headers
- [Validation](../guides/validation.md) — add `validate` tags for automatic input validation
- [Error Handling](../guides/errors.md) — return structured RFC 9457 errors
- [Router Adapters](../guides/router-adapters.md) — switch to Fiber or the stdlib mux
