# Error Helpers

All helpers return a `StatusError` that Zorya serializes as `application/problem+json` per [RFC 9457](https://www.rfc-editor.org/rfc/rfc9457).

## Helper functions

| Function | Status | Typical use |
|---|---|---|
| `Error400BadRequest(msg, ...error)` | 400 | Malformed JSON, missing required fields |
| `Error401Unauthorized(msg, ...error)` | 401 | No or invalid authentication token |
| `Error403Forbidden(msg, ...error)` | 403 | Authenticated but not authorized |
| `Error404NotFound(msg, ...error)` | 404 | Resource does not exist |
| `Error405MethodNotAllowed(msg, ...error)` | 405 | HTTP method not supported on this path |
| `Error406NotAcceptable(msg, ...error)` | 406 | Cannot serve requested `Accept` type |
| `Error409Conflict(msg, ...error)` | 409 | Duplicate resource, state conflict |
| `Error410Gone(msg, ...error)` | 410 | Resource permanently removed |
| `Error412PreconditionFailed(msg, ...error)` | 412 | Conditional request precondition failed |
| `Error415UnsupportedMediaType(msg, ...error)` | 415 | Unrecognized `Content-Type` |
| `Error422UnprocessableEntity(msg, ...error)` | 422 | Validation failure |
| `Error429TooManyRequests(msg, ...error)` | 429 | Rate limit exceeded |
| `Error500InternalServerError(msg, ...error)` | 500 | Unexpected server fault |
| `Error501NotImplemented(msg, ...error)` | 501 | Feature not yet implemented |
| `Error502BadGateway(msg, ...error)` | 502 | Upstream service error |
| `Error503ServiceUnavailable(msg, ...error)` | 503 | Service temporarily unavailable |
| `Error504GatewayTimeout(msg, ...error)` | 504 | Upstream response timeout |

## Special helpers

| Function | Description |
|---|---|
| `Status304NotModified()` | Returns a 304 status (used by conditional request checks) |
| `ErrorWithHeaders(err, http.Header)` | Wraps any error with additional response headers |
| `WriteErr(api, r, w, status, msg, ...error)` | Writes an error response directly (for middleware use) |

## ErrorModel

```go
type ErrorModel struct {
    Title   string        `json:"title"`
    Status  int           `json:"status"`
    Detail  string        `json:"detail,omitempty"`
    Errors  []error       `json:"errors,omitempty"`
}
```

`ErrorModel` implements `StatusError`, `error`, and `ContentTypeProvider` (returns `application/problem+json`).

## ErrorDetail

```go
type ErrorDetail struct {
    Code     string `json:"code,omitempty"`     // machine-readable code
    Message  string `json:"message,omitempty"`  // human-readable explanation
    Location string `json:"location,omitempty"` // field path, e.g. "body.email", "query.page"
}
```

Attach to any error helper to provide field-level context:

```go
return nil, zorya.Error422UnprocessableEntity("validation failed",
    &zorya.ErrorDetail{Code: "email",  Message: "must be a valid email",         Location: "body.email"},
    &zorya.ErrorDetail{Code: "min",    Message: "must be at least 8 characters", Location: "body.password"},
)
```

## StatusError interface

```go
type StatusError interface {
    GetStatus() int
    Error() string
}
```

Any type satisfying this interface can be returned from a handler and Zorya will serialize it correctly.

## HeadersError interface

```go
type HeadersError interface {
    GetHeaders() http.Header
}
```

If a `StatusError` also implements `HeadersError`, the headers are written to the response. Use `ErrorWithHeaders` to attach headers to existing errors.

## Response format

```json
{
  "title": "Unprocessable Entity",
  "status": 422,
  "errors": [
    {
      "code": "required",
      "message": "field validation failed on 'required' constraint",
      "location": "body.Email"
    },
    {
      "code": "email",
      "message": "field validation failed on 'email' constraint",
      "location": "body.Email"
    }
  ]
}
```
