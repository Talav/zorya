# OpenAPI & Docs UI

Zorya generates an OpenAPI 3.1 specification from your registered routes and serves an interactive Stoplight Elements UI — with zero extra code.

## Configuration

Pass a `Config` when creating the API:

```go
api := zorya.NewAPI(adapter, zorya.WithConfig(&zorya.Config{
    OpenAPIPath: "/openapi.json",   // spec endpoint
    DocsPath:    "/docs",           // Stoplight Elements UI
}))
```

Or use `DefaultConfig()` which sets the same defaults:

```go
api := zorya.NewAPI(adapter, zorya.WithConfig(zorya.DefaultConfig()))
```

## API metadata

Set the title, version, and description via the `OpenAPI` object:

```go
api := zorya.NewAPI(adapter, zorya.WithOpenAPI(&zorya.OpenAPI{
    Info: &zorya.Info{
        Title:       "My API",
        Version:     "1.0.0",
        Description: "A brief description of what this API does.",
    },
    Servers: []*zorya.Server{
        {URL: "https://api.example.com", Description: "Production"},
        {URL: "http://localhost:8080",   Description: "Local development"},
    },
}))
```

## Security schemes

Add Bearer JWT authentication to the spec:

```go
openAPISpec := &zorya.OpenAPI{
    Info: &zorya.Info{Title: "Secure API", Version: "1.0.0"},
    Components: &zorya.Components{
        SecuritySchemes: map[string]*zorya.SecurityScheme{
            "bearerAuth": {
                Type:   "http",
                Scheme: "bearer",
                Description: "JWT access token",
            },
        },
    },
}
api := zorya.NewAPI(adapter, zorya.WithOpenAPI(openAPISpec))
```

## Viewing the spec and docs UI

- **Spec JSON**: `GET /openapi.json`
- **Docs UI**: `GET /docs` (Stoplight Elements, served as HTML)

The spec is generated lazily on the first request and cached. Adding routes after the server starts invalidates the cache.

## Operation metadata

Enrich individual operations through `BaseRoute.Operation`:

```go
zorya.Get(api, "/users/{id}", getUser, func(r *zorya.BaseRoute) {
    r.Operation = &zorya.Operation{
        Summary:     "Get a user",
        Description: "Returns a single user by ID.",
        Tags:        []string{"Users"},
        OperationID: "getUser",
    }
})
```

## Advanced: talav/openapi

Zorya uses [talav/openapi](https://github.com/talav/openapi) internally for schema and spec generation. Refer to that library's documentation for:

- Custom schema hooks (`SchemaTransformer`, `SchemaProvider`)
- `openapi` struct tags (title, description, format, examples, readOnly, writeOnly)
- `requires` tag for JSON Schema `dependentRequired`
- Choosing between OpenAPI 3.0.4 and 3.1.2 targets

The tag reference is at [talav/openapi — Metadata guide](https://github.com/talav/openapi/blob/main/docs/guides/metadata.md).

## Conditional requests on the spec endpoint

The spec endpoint supports `ETag` / `If-None-Match` out of the box. A `304 Not Modified` is returned when the spec has not changed since the client last fetched it.
