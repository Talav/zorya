# Config Options

## Config struct

```go
type Config struct {
    OpenAPIPath      string
    DocsPath         string
    SchemasPath      string
    DefaultFormat    string
    NoFormatFallback bool
}
```

| Field | Default | Description |
|---|---|---|
| `OpenAPIPath` | `/openapi.json` | Path that serves the OpenAPI specification as JSON |
| `DocsPath` | `/docs` | Path that serves the Stoplight Elements docs UI |
| `SchemasPath` | `/schemas` | Path prefix for individual schema JSON files |
| `DefaultFormat` | `application/json` | Content type used when the `Accept` header is absent or `*/*` |
| `NoFormatFallback` | `false` | When `true`, return `406` instead of falling back to JSON for unknown `Accept` types |

Use `zorya.DefaultConfig()` as a starting point:

```go
cfg := zorya.DefaultConfig()
cfg.OpenAPIPath = "/api/openapi.json"
api := zorya.NewAPI(adapter, zorya.WithConfig(cfg))
```

## Option functions

These are the functional options accepted by `zorya.NewAPI(adapter, ...Option)`.

| Option | Description |
|---|---|
| `WithConfig(cfg *Config)` | Set all config fields at once |
| `WithValidator(v Validator)` | Replace the default validator |
| `WithFormat(ct string, f Format)` | Add or replace a single content format |
| `WithFormats(m map[string]Format)` | Merge a map of formats with the defaults |
| `WithFormatsReplace(m map[string]Format)` | Replace *all* formats (disables JSON/CBOR defaults) |
| `WithDefaultFormat(ct string)` | Set the fallback format when no `Accept` matches |
| `WithMetadata(m *schema.Metadata)` | Provide a custom tag parser registry |
| `WithCodec(c *schema.Codec)` | Provide a custom request decoder |
| `WithOpenAPI(spec *OpenAPI)` | Set API title, version, description, servers, security schemes |

## Route options

Options accepted as the final variadic arguments of `Get`, `Post`, `Put`, `Patch`, `Delete`:

```go
zorya.Get(api, "/path", handler,
    func(r *zorya.BaseRoute) {
        r.Operation      = &zorya.Operation{Summary: "...", Tags: []string{"..."}}
        r.MaxBodyBytes    = 10 * 1024 * 1024
        r.BodyReadTimeout = 30 * time.Second
        r.Errors          = []int{400, 404}
    },
    zorya.Secure(zorya.Roles("admin")),
)
```

| Field | Default | Description |
|---|---|---|
| `Operation` | nil | OpenAPI operation metadata (summary, description, tags, operationID) |
| `MaxBodyBytes` | 1 MB | Request body size limit; `-1` disables |
| `BodyReadTimeout` | 5s | Deadline for reading request body; `-1` disables |
| `Errors` | nil | Extra status codes to document in the OpenAPI spec |
| `Security` | nil | Authorization requirements (use `Secure(...)` helper) |

## Register function

For full control, use `Register` directly:

```go
err := zorya.Register[InputType, OutputType](api, zorya.BaseRoute{
    Method: http.MethodGet,
    Path:   "/users/{id}",
    Operation: &zorya.Operation{Summary: "Get user"},
}, handler)
```
