# Error Handling

Zorya implements [RFC 9457 Problem Details](https://www.rfc-editor.org/rfc/rfc9457) for HTTP error responses. All errors use a consistent JSON structure with a `application/problem+json` content type.

## Error response shape

```json
{
  "title": "Not Found",
  "status": 404,
  "detail": "user 42 does not exist",
  "errors": []
}
```

## Returning errors from a handler

Return a `StatusError` from the handler. The built-in helpers create the right status code and message:

```go
func getUser(ctx context.Context, input *GetUserInput) (*UserOutput, error) {
    user, ok := db[input.ID]
    if !ok {
        return nil, zorya.Error404NotFound("user not found")
    }
    return &UserOutput{Body: user}, nil
}
```

## Error helper functions

| Function | Status | When to use |
|---|---|---|
| `Error400BadRequest` | 400 | Malformed request, missing required fields |
| `Error401Unauthorized` | 401 | Missing or invalid authentication |
| `Error403Forbidden` | 403 | Authenticated but not authorized |
| `Error404NotFound` | 404 | Resource does not exist |
| `Error405MethodNotAllowed` | 405 | HTTP method not supported |
| `Error406NotAcceptable` | 406 | Cannot serve requested `Accept` type |
| `Error409Conflict` | 409 | State conflict (e.g. duplicate) |
| `Error410Gone` | 410 | Resource permanently deleted |
| `Error412PreconditionFailed` | 412 | Conditional request precondition failed |
| `Error415UnsupportedMediaType` | 415 | Unrecognized `Content-Type` |
| `Error422UnprocessableEntity` | 422 | Validation failure |
| `Error429TooManyRequests` | 429 | Rate limit exceeded |
| `Error500InternalServerError` | 500 | Unexpected server error |
| `Error501NotImplemented` | 501 | Feature not implemented |
| `Error502BadGateway` | 502 | Upstream error |
| `Error503ServiceUnavailable` | 503 | Service temporarily down |
| `Error504GatewayTimeout` | 504 | Upstream timeout |

All helpers accept an optional variadic list of additional errors:

```go
zorya.Error400BadRequest("invalid input", validationErr1, validationErr2)
```

## Adding field-level detail

Wrap individual errors with `ErrorDetail` to include a machine-readable code and location:

```go
return nil, zorya.Error422UnprocessableEntity("validation failed",
    &zorya.ErrorDetail{
        Code:     "email",
        Message:  "must be a valid email address",
        Location: "body.email",
    },
)
```

The `ErrorDetail` struct:

```go
type ErrorDetail struct {
    Code     string `json:"code,omitempty"`     // machine-readable code, e.g. "required", "email"
    Message  string `json:"message,omitempty"`  // human-readable explanation
    Location string `json:"location,omitempty"` // path to the field, e.g. "body.email", "query.page"
}
```

## Adding response headers to errors

Use `ErrorWithHeaders` to attach headers to any error:

```go
return nil, zorya.ErrorWithHeaders(
    zorya.Error401Unauthorized("token expired"),
    http.Header{"WWW-Authenticate": []string{`Bearer realm="api"`}},
)
```

## Writing errors directly

In low-level handlers (e.g. middleware), use `WriteErr`:

```go
zorya.WriteErr(api, r, w, http.StatusServiceUnavailable, "database offline")
```

## Implementing StatusError

You can return any type that implements `StatusError` from a handler:

```go
type StatusError interface {
    GetStatus() int
    Error() string
}
```

If the error also implements `ErrorDetailer`, its `ErrorDetail()` result is included in the `errors` array of the response.
