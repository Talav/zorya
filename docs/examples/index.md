# Examples

Each example is a self-contained, runnable Go program in the `examples/` directory of the repository. Run any example with:

```bash
go run ./examples/<name>
```

## basic-crud

**File**: `examples/basic-crud/main.go`

Full CRUD for a `User` resource (GET list, GET by ID, POST, PUT, DELETE) using the Chi adapter. The starting point for most real APIs.

```bash
go run ./examples/basic-crud
```

## auth-jwt

**File**: `examples/auth-jwt/main.go`

Bearer token authentication with role-based access control using `Secure(Roles(...))`. Shows how `GetRouteSecurityContext` feeds a custom auth middleware.

```bash
go run ./examples/auth-jwt
```

## route-groups

**File**: `examples/route-groups/main.go`

Versioned API (`/v1`, `/v2`) built with `NewGroup`. Each group adds a version header and serves a different response shape from the same handler pattern.

```bash
go run ./examples/route-groups
```

## openapi-ui

**File**: `examples/openapi-ui/main.go`

Full OpenAPI 3.1 configuration: title, description, servers, Bearer auth security scheme, per-operation metadata. The Stoplight Elements UI is served at `/docs`.

```bash
go run ./examples/openapi-ui
# Open http://localhost:8080/docs
```

## validation

**File**: `examples/validation/main.go`

Custom `Validator` wrapping go-playground/validator. Shows how validation errors become structured `ErrorDetail` entries in the RFC 9457 response body.

```bash
go run ./examples/validation
```

## conditional-requests

**File**: `examples/conditional-requests/main.go`

ETag-based conditional reads (304 Not Modified) and optimistic-locking writes (412 Precondition Failed) using `conditional.Params`.

```bash
go run ./examples/conditional-requests
```

## streaming-sse

**File**: `examples/streaming-sse/main.go`

Server-Sent Events using the `func(http.ResponseWriter) error` body field. Includes a small HTML page to view the event stream in a browser.

```bash
go run ./examples/streaming-sse
# Open http://localhost:8080 in a browser
```

## file-upload

**File**: `examples/file-upload/main.go`

Multipart file upload with per-file size validation and content-type detection. Rejects non-image uploads with 415 Unsupported Media Type.

```bash
go run ./examples/file-upload
```

## content-negotiation

**File**: `examples/content-negotiation/main.go`

The same endpoint serves JSON or CBOR depending on the `Accept` header. No handler code changes required — Zorya's format registry handles it.

```bash
go run ./examples/content-negotiation
curl -H "Accept: application/cbor" http://localhost:8080/users/1 | xxd
```

## fiber-adapter

**File**: `examples/fiber-adapter/main.go`

The `basic-crud` example reimplemented with the Fiber adapter. The handlers are identical — only the adapter creation and `app.Listen` call differ.

```bash
go run ./examples/fiber-adapter
```
