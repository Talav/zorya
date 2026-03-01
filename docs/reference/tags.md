# Struct Tag Cheatsheet

Zorya reads several struct tags to extract request parameters, decode bodies, run validation, and generate OpenAPI schemas. This page is a quick reference. Each tag links to its canonical documentation.

## Tag summary

| Tag | Used on | Purpose | Reference |
|---|---|---|---|
| `schema` | Input fields | Parameter name, location, style | [talav/schema](https://talav.github.io/schema/) |
| `body` | Input body field | Marks the request body; value sets decode mode | [talav/schema](https://talav.github.io/schema/) |
| `json` | All struct fields | JSON property name; `-` to hide | [encoding/json](https://pkg.go.dev/encoding/json) |
| `validate` | Input fields | Validation constraints | [go-playground/validator](https://pkg.go.dev/github.com/go-playground/validator/v10) |
| `openapi` | Any field | OpenAPI metadata: title, description, format, examples | [talav/openapi](https://github.com/talav/openapi) |
| `default` | Input fields | Default value when parameter is absent | [talav/openapi](https://github.com/talav/openapi) |
| `requires` | Input fields | Conditional required (JSON Schema `dependentRequired`) | [talav/openapi](https://github.com/talav/openapi) |

## `schema` tag

Syntax: `schema:"<name>,location=<loc>"`

| `location` value | Where the parameter is read from |
|---|---|
| `path` | URL path segment (e.g. `/users/{id}`) |
| `query` | URL query string (e.g. `?page=2`) |
| `header` | HTTP request header |
| `cookie` | Cookie value |

For multipart form fields and file uploads, use `body:"multipart"` with `schema:"name"` (no location). See [File Uploads](../guides/uploads.md).

```go
type Input struct {
    ID      int    `schema:"id,location=path"`
    Page    int    `schema:"page,location=query" default:"1"`
    APIKey  string `schema:"X-API-Key,location=header"`
}
```

## `body` tag

```go
type Input struct {
    Body struct {
        Name string `json:"name" validate:"required"`
    } `body:"structured"`
}
```

The `body` tag value selects the decoding mode. Use `"structured"` for JSON payloads.

## `openapi` tag

```go
type User struct {
    _       struct{} `openapi:"additionalProperties=false"`
    ID      int      `json:"id"    openapi:"readOnly,description=Unique identifier"`
    Email   string   `json:"email" openapi:"description=Contact email,examples=user@example.com"`
    Secret  string   `json:"secret" openapi:"writeOnly"`
    Legacy  string   `json:"legacy" openapi:"deprecated"`
}
```

Common values:

| Value | Effect |
|---|---|
| `readOnly` | `readOnly: true` in schema |
| `writeOnly` | `writeOnly: true` in schema |
| `deprecated` | `deprecated: true` in schema |
| `hidden` | Excluded from OpenAPI schema |
| `required` | Override required status in docs |
| `title=...` | Schema `title` |
| `description=...` | Schema `description` |
| `format=...` | Schema `format` (e.g. `date-time`, `email`, `uri`) |
| `examples=a\|b\|c` | Pipe-separated example values |
| `x-custom=value` | OpenAPI extension |
| `additionalProperties=false` | On `_` blank field: struct-level |
| `nullable=true` | On `_` blank field: struct-level |

## `validate` tag

```go
type Input struct {
    Body struct {
        Name  string `json:"name"  validate:"required,min=2,max=100"`
        Email string `json:"email" validate:"required,email"`
        Age   int    `json:"age"   validate:"gte=0,lte=150"`
        Role  string `json:"role"  validate:"oneof=admin member viewer"`
    } `body:"structured"`
}
```

Full constraint list: [go-playground/validator baked-in tags](https://pkg.go.dev/github.com/go-playground/validator/v10#hdr-Baked_In_Validators_and_Tags).

## `requires` tag

```go
type Payment struct {
    CardNumber     string `json:"card_number"     requires:"billing_address,cvv"`
    BillingAddress string `json:"billing_address"`
    CVV            string `json:"cvv"`
}
```

When `card_number` is present, `billing_address` and `cvv` become required in the JSON Schema (`dependentRequired`).
