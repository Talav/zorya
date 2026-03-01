# Defining Inputs

Every Zorya handler receives a typed input struct. Zorya decodes the incoming HTTP request into that struct automatically.

## Input struct shape

```go
type MyInput struct {
    // Path parameters, query parameters, and headers go at the top level.
    ID   int    `schema:"id,location=path"`
    Page int    `schema:"page,location=query" default:"1"`

    // The request body is a nested struct tagged with `body`.
    Body struct {
        Name  string `json:"name"  validate:"required"`
        Email string `json:"email" validate:"required,email"`
    } `body:"structured"`
}
```

## Path parameters

Declare with `schema:"<name>,location=path"`. The name must match the placeholder in the route pattern.

```go
type GetArticleInput struct {
    Slug string `schema:"slug,location=path"`
}

zorya.Get(api, "/articles/{slug}", handler)
```

## Query parameters

Declare with `schema:"<name>,location=query"`.

```go
type ListUsersInput struct {
    Page  int    `schema:"page,location=query"  default:"1"`
    Limit int    `schema:"limit,location=query" default:"20"`
    Q     string `schema:"q,location=query"`
}
```

Default values are applied when the parameter is absent. See the [talav/schema tag reference](https://talav.github.io/schema/) for all supported types and coercion rules.

## Headers

Declare with `schema:"<Header-Name>,location=header"`.

```go
type AuthorizedInput struct {
    Authorization string `schema:"Authorization,location=header"`
    XRequestID    string `schema:"X-Request-ID,location=header"`
}
```

## Request body

Tag a nested struct field with `` `body:"structured"` `` to mark it as the request body (JSON-decoded).

```go
type CreatePostInput struct {
    Body struct {
        Title   string   `json:"title"   validate:"required,min=3"`
        Content string   `json:"content" validate:"required"`
        Tags    []string `json:"tags"`
    } `body:"structured"`
}
```

Zorya reads the body, decodes it according to the `Content-Type`, and validates it before calling the handler.

## Body size and timeout limits

Set per-route limits via `BaseRoute` options:

```go
zorya.Post(api, "/upload", handler, func(r *zorya.BaseRoute) {
    r.MaxBodyBytes    = 10 * 1024 * 1024 // 10 MB
    r.BodyReadTimeout = 30 * time.Second
})
```

Defaults: `MaxBodyBytes = 1 MB`, `BodyReadTimeout = 5s`. Set either to `-1` to disable.

## Tag reference

Zorya reads these struct tags. For full semantics, follow the links to the canonical documentation:

| Tag | Purpose | Reference |
|---|---|---|
| `schema` | Parameter location, name, style | [talav/schema](https://talav.github.io/schema/) |
| `body` | Marks the request body field | [talav/schema](https://talav.github.io/schema/) |
| `json` | JSON property name | [encoding/json](https://pkg.go.dev/encoding/json) |
| `validate` | Validation constraints | [go-playground/validator](https://pkg.go.dev/github.com/go-playground/validator/v10) |
| `default` | Default value when param is absent | [talav/openapi](https://github.com/talav/openapi) |
| `openapi` | OpenAPI metadata (title, description, examples) | [talav/openapi](https://github.com/talav/openapi) |
| `requires` | Conditional required fields (JSON Schema) | [talav/openapi](https://github.com/talav/openapi) |
