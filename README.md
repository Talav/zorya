# Zorya

A Go HTTP API framework for building type-safe, RFC-compliant REST APIs with automatic request/response handling, content negotiation, validation, and comprehensive error handling.

## Features

- **Type-Safe Request/Response Handling** - Decode requests and encode responses using Go structs
- **Router Adapters** - Works with Chi, Fiber, and Go 1.22+ standard library
- **Content Negotiation** - Automatic content type negotiation (JSON, CBOR, and custom formats)
- **Request Validation** - Pluggable validation with go-playground/validator support
- **Route Security** - Declarative authentication, role-based, permission-based, and resource-based authorization
- **RFC 9457 Error Handling** - Structured error responses with machine-readable codes
- **Conditional Requests** - Support for If-Match, If-None-Match, If-Modified-Since, If-Unmodified-Since
- **Streaming Responses** - Server-Sent Events (SSE) and chunked transfer support
- **Response Transformers** - Modify response bodies before serialization
- **Middleware Support** - API-level and route-level middleware chains
- **Route Groups** - Group routes with shared prefixes, middleware, and transformers
- **Request Limits** - Configurable body size limits and read timeouts
- **Default Parameter Values** - Automatic default value application using struct tags

## Installation

```bash
go get github.com/talav/zorya
```

## Quick Start

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
        Name  string `json:"name" validate:"required"`
        Email string `json:"email" validate:"required,email"`
    }
}

type CreateUserOutput struct {
    Status int `json:"-"`
    Body struct {
        ID    int    `json:"id"`
        Name  string `json:"name"`
        Email string `json:"email"`
}
}

func main() {
    router := chi.NewMux()
    adapter := adapters.NewChi(router)
    api := zorya.NewAPI(adapter)
    
    zorya.Post(api, "/users", func(ctx context.Context, input *CreateUserInput) (*CreateUserOutput, error) {
        // Your business logic here
        return &CreateUserOutput{
            Status: http.StatusCreated,
            Body: struct {
                ID    int    `json:"id"`
                Name  string `json:"name"`
                Email string `json:"email"`
            }{
                ID:    1,
                Name:  input.Body.Name,
                Email: input.Body.Email,
            },
        }, nil
    })
    
    http.ListenAndServe(":8080", router)
}
```

## Router Adapters

Zorya supports multiple router backends through adapters:

### Chi Router

```go
import (
    "github.com/go-chi/chi/v5"
    "github.com/talav/zorya/adapters"
)

router := chi.NewMux()
adapter := adapters.NewChi(router)
api := zorya.NewAPI(adapter)
```

### Fiber

```go
import (
    "github.com/gofiber/fiber/v2"
    "github.com/talav/zorya/adapters"
)

app := fiber.New()
adapter := adapters.NewFiber(app)
api := zorya.NewAPI(adapter)
```

### Standard Library (Go 1.22+)

```go
import (
    "net/http"
    "github.com/talav/zorya/adapters"
)

mux := http.NewServeMux()
adapter := adapters.NewStdlib(mux)
api := zorya.NewAPI(adapter)

// With prefix
adapter := adapters.NewStdlibWithPrefix(mux, "/api")
```

## Request Handling

### Input Structs

Input structs define request parameters and body using struct tags from the `schema` package:

```go
type GetUserInput struct {
    // Path parameter (from router)
    ID string `schema:"id,location=path"`
    
    // Query parameters
    Format string `schema:"format,location=query"`
    Page   int    `schema:"page,location=query"`
    
    // Header parameters
    APIKey string `schema:"X-API-Key,location=header"`
    
    // Cookie parameters
    SessionID string `schema:"session_id,location=cookie"`
    
    // Request body
    Body struct {
        Name  string `json:"name"`
        Email string `json:"email"`
    } `body:"structured"`
}
```

See the [schema package documentation](../schema/README.md) for detailed information on struct tags and parameter locations.

### HTTP Methods

Zorya provides convenience functions for all HTTP methods. These functions panic on errors since route registration happens during startup and errors represent programming/configuration mistakes:

```go
zorya.Get(api, "/users/{id}", handler)
zorya.Post(api, "/users", handler)
zorya.Put(api, "/users/{id}", handler)
zorya.Patch(api, "/users/{id}", handler)
zorya.Delete(api, "/users/{id}", handler)
zorya.Head(api, "/users/{id}", handler)
```

**Note:** For advanced use cases where you need error handling, use `zorya.Register` directly, which returns errors instead of panicking.

### Advanced Route Configuration

```go
zorya.Post(api, "/users", handler,
    func(route *zorya.BaseRoute) {
        // Set body size limit (default: 1MB)
        route.MaxBodyBytes = 10 * 1024 * 1024 // 10MB
        
        // Set body read timeout (default: 5 seconds)
        route.BodyReadTimeout = 10 * time.Second
        
        // Add route-specific middleware
        route.Middlewares = zorya.Middlewares{
            func(ctx zorya.Context, next func(zorya.Context)) {
                // Your middleware logic
                next(ctx)
            },
        }
    },
)
```

## Response Handling

### Output Structs

Output structs define response status, headers, and body:

```go
type GetUserOutput struct {
    // HTTP status code - use 'status' tag to identify the field
    HTTPStatus int `status:""`
    
    // Response headers - use 'schema' tag with location=header
    ContentType string `schema:"Content-Type,location=header"`
    ETag        string `schema:"ETag,location=header"`
    
    // Response body - use 'body' tag (same as input)
    ResponseBody struct {
        ID    int    `json:"id"`
        Name  string `json:"name"`
        Email string `json:"email"`
    } `body:"structured"`
}
```

### Status Codes

Set the status code using a field with the `status` tag:

```go
type CreateUserOutput struct {
    StatusCode int `status:"201"`  // Default status code
    Body        struct {
        ID   int    `json:"id"`
        Name string `json:"name"`
    } `body:"structured"`
}

// In handler:
return &CreateUserOutput{
    StatusCode: http.StatusCreated, // 201 (or use default from tag)
    Body:       userData,
}, nil
```

### Response Headers

Use the `header` tag to set response headers:

```go
type Output struct {
    Location     string   `header:"Location"`           // Single value
    CacheControl []string `header:"Cache-Control"`      // Multiple values (slice)
    CustomHeader string   `header:"X-Custom-Header"`   // Custom header
}
```

## Content Negotiation

Zorya automatically negotiates content types based on the `Accept` header:

```go
// Client requests: Accept: application/cbor
// Response: Content-Type: application/cbor

// Client requests: Accept: application/json
// Response: Content-Type: application/json

// Client requests: Accept: */*
// Response: Content-Type: application/json (default)
```

Supported formats:
- `application/json` (default)
- `application/cbor`

Plus-segment matching is supported (e.g., `application/vnd.api+json` matches `json`).

### Custom Formats

You can add custom formats (e.g., XML, YAML, etc.) by implementing a `Format`. Default formats (JSON, CBOR) are automatically included unless you use `WithFormatsReplace`.

#### Adding a Single Format

```go
import (
    "encoding/xml"
    "github.com/talav/zorya"
)

// Create XML format
xmlFormat := zorya.Format{
    Marshal: func(w io.Writer, v any) error {
        enc := xml.NewEncoder(w)
        enc.Indent("", "  ")
        return enc.Encode(v)
    },
}

// Add XML format - defaults (JSON, CBOR) are automatically included
api := zorya.NewAPI(
    adapter,
    zorya.WithFormat("application/xml", xmlFormat),
    zorya.WithFormat("xml", xmlFormat), // For +xml suffix matching
)
```

#### Adding Multiple Formats

```go
// Add multiple formats - defaults are automatically included
formats := map[string]zorya.Format{
    "application/xml": xmlFormat,
    "xml":             xmlFormat,
    "text/plain":      textFormat,
}

api := zorya.NewAPI(
    adapter,
    zorya.WithFormats(formats),
)
```

#### Replacing All Formats

To have complete control and exclude defaults, use `WithFormatsReplace`:

```go
// Only JSON, no CBOR - you must explicitly include JSON
api := zorya.NewAPI(
    adapter,
    zorya.WithFormatsReplace(map[string]zorya.Format{
        "application/json": zorya.JSONFormat(),
        "json":             zorya.JSONFormat(),
    }),
)

// Or only XML - no defaults included
api := zorya.NewAPI(
    adapter,
    zorya.WithFormatsReplace(map[string]zorya.Format{
        "application/xml": xmlFormat,
        "xml":             xmlFormat,
    }),
)
```

Now clients can request XML responses:

```go
// Client request: Accept: application/xml
// Response: Content-Type: application/xml

// Client request: Accept: application/json, application/xml
// Response: Content-Type: application/xml (if XML is preferred)
```

## Request Validation

### Using go-playground/validator

```go
import (
    "github.com/talav/validator"
    "github.com/talav/zorya"
)

// Create validator
validate := validator.New()
zoryaValidator := zorya.NewPlaygroundValidator(validate)

// Configure API
api := zorya.NewAPI(adapter, zorya.WithValidator(zoryaValidator))
```

### Validation Tags

```go
type CreateUserInput struct {
    Body struct {
        Email    string `json:"email" validate:"required,email"`
        Name     string `json:"name" validate:"required,min=3"`
        Age      int    `json:"age" validate:"min=0,max=150"`
        Username string `json:"username" validate:"required,alphanum"`
}
}
```

### Validation Error Response

When validation fails, Zorya returns a 422 Unprocessable Entity response:

```json
{
  "status": 422,
  "title": "Unprocessable Entity",
  "detail": "validation failed",
  "errors": [
    {
      "code": "email",
      "message": "Key: 'CreateUserInput.Body.email' Error:Field validation for 'email' failed on the 'email' tag",
      "location": "body.email"
    },
    {
      "code": "min",
      "message": "Key: 'CreateUserInput.Body.name' Error:Field validation for 'name' failed on the 'min' tag",
      "location": "body.name"
    }
  ]
}
```

The `code` field contains the validation tag for frontend translation.

### Custom Validators

Implement the `Validator` interface to use any validation library:

```go
type MyValidator struct {
    // your validator implementation
}

func (v *MyValidator) Validate(ctx context.Context, input any) []error {
    // Validate and return errors that implement ErrorDetailer
    return []error{
        &zorya.ErrorDetail{
            Code:     "custom_error",
            Message:  "Custom validation failed",
            Location: "body.field",
        },
    }
}
```

## Error Handling

Zorya provides comprehensive error handling based on [RFC 9457 Problem Details for HTTP APIs](https://datatracker.ietf.org/doc/html/rfc9457).

### Error Types

```go
type ErrorModel struct {
    Type     string         `json:"type,omitempty"`
    Title    string         `json:"title,omitempty"`
    Status   int            `json:"status,omitempty"`
    Detail   string         `json:"detail,omitempty"`
    Instance string         `json:"instance,omitempty"`
    Errors   []*ErrorDetail `json:"errors,omitempty"`
}

type ErrorDetail struct {
    Code     string `json:"code,omitempty"`     // Machine-readable error code
    Message  string `json:"message,omitempty"`  // Human-readable message
    Location string `json:"location,omitempty"` // Error location (e.g., "body.email")
}
```

### Convenience Functions

```go
// 4xx Client Errors
zorya.Error400BadRequest(msg string, errs ...error)
zorya.Error401Unauthorized(msg string, errs ...error)
zorya.Error403Forbidden(msg string, errs ...error)
zorya.Error404NotFound(msg string, errs ...error)
zorya.Error422UnprocessableEntity(msg string, errs ...error)
zorya.Error429TooManyRequests(msg string, errs ...error)

// 5xx Server Errors
zorya.Error500InternalServerError(msg string, errs ...error)
zorya.Error503ServiceUnavailable(msg string, errs ...error)
```

### Error Interfaces

```go
// StatusError - Set HTTP status code
type StatusError interface {
    GetStatus() int
    Error() string
}

// HeadersError - Set response headers
type HeadersError interface {
    GetHeaders() http.Header
    Error() string
}

// ErrorDetailer - Provide structured error details
type ErrorDetailer interface {
    ErrorDetail() *ErrorDetail
}
```

### Example

```go
func getUser(ctx context.Context, input *GetUserInput) (*GetUserOutput, error) {
    user, err := db.FindUser(input.ID)
    if err != nil {
        return nil, zorya.Error404NotFound("User not found")
    }
    return &GetUserOutput{Body: user}, nil
}
```

For detailed error handling documentation, see [Error Processing in Zorya](#error-processing).

## Conditional Requests

Zorya supports HTTP conditional requests (If-Match, If-None-Match, If-Modified-Since, If-Unmodified-Since) for caching and concurrency control.

### Usage

Embed `conditional.Params` in your input struct:

```go
import "github.com/talav/zorya/conditional"

type GetUserInput struct {
    ID string `schema:"id,location=path"`
    conditional.Params
}
```

Check preconditions in your handler:

```go
func getUser(ctx context.Context, input *GetUserInput) (*GetUserOutput, error) {
    user := getUserFromDB(input.ID)
    
    if err := input.Params.CheckPreconditions(
        user.ETag,
        user.Modified,
        ctx.Request().Method != http.MethodGet,
    ); err != nil {
        return nil, err
    }
    
    return &GetUserOutput{Body: user}, nil
}
```

### Behavior

- **Read requests (GET, HEAD)**: Returns `304 Not Modified` if conditions fail
- **Write requests (POST, PUT, PATCH, DELETE)**: Returns `412 Precondition Failed` if conditions fail

## Streaming Responses

Zorya supports streaming responses via `Body func(Context)` fields for Server-Sent Events (SSE) and chunked transfers.

### Basic Streaming

```go
type StreamOutput struct {
    Status      int    `default:"200"`
    ContentType string `header:"Content-Type"`
    Body        func(ctx zorya.Context)
}

func streamHandler(ctx context.Context, input *Input) (*StreamOutput, error) {
    return &StreamOutput{
        ContentType: "text/plain",
        Body: func(ctx zorya.Context) {
            w := ctx.BodyWriter()
            for i := 0; i < 5; i++ {
                fmt.Fprintf(w, "chunk %d\n", i)
                time.Sleep(time.Second)
            }
        },
    }, nil
}
```

### Server-Sent Events (SSE)

```go
type SSEOutput struct {
    ContentType  string `header:"Content-Type"`
    CacheControl string `header:"Cache-Control"`
    Body         func(ctx zorya.Context)
}

func sseHandler(ctx context.Context, input *Input) (*SSEOutput, error) {
    return &SSEOutput{
        ContentType:  "text/event-stream",
        CacheControl: "no-cache",
        Body: func(ctx zorya.Context) {
            w := ctx.BodyWriter()
            flusher, _ := w.(http.Flusher)
            
            for i := 0; i < 10; i++ {
                select {
                case <-ctx.Context().Done():
                    return
                default:
                    fmt.Fprintf(w, "event: message\n")
                    fmt.Fprintf(w, "data: {\"count\": %d}\n\n", i)
                    if flusher != nil {
                        flusher.Flush()
                    }
                    time.Sleep(time.Second)
                }
            }
        },
    }, nil
}
```

### ResponseWriter Interface

For full response control, type-assert to `ResponseWriter`:

```go
func streamWithHeaders(ctx zorya.Context) {
    if w, ok := ctx.(zorya.ResponseWriter); ok {
        w.SetHeader("X-Custom", "value")
        w.SetStatus(200)
    }
    
    w := ctx.BodyWriter()
    w.Write([]byte("streaming data"))
}
```

## Response Transformers

Transformers modify response bodies before serialization. They run in the order they were added.

### API-Level Transformers

```go
api.UseTransformer(func(ctx zorya.Context, status string, v any) (any, error) {
    // Transform response body
    if data, ok := v.(map[string]any); ok {
        data["transformed"] = true
        return data, nil
    }
    return v, nil
})
```

### Group-Level Transformers

```go
group := zorya.NewGroup(api, "/v1")
group.UseTransformer(func(ctx zorya.Context, status string, v any) (any, error) {
    // Transform only for this group
    return v, nil
})
```

Transformers are chained: group transformers run first, then API transformers.

## Middleware

### API-Level Middleware

```go
api.UseMiddleware(func(ctx zorya.Context, next func(zorya.Context)) {
    // Before handler
    start := time.Now()
    
    next(ctx)
    
    // After handler
    duration := time.Since(start)
    log.Printf("Request took %v", duration)
})
```

### Route-Level Middleware

```go
zorya.Get(api, "/users", handler,
    func(route *zorya.BaseRoute) {
        route.Middlewares = zorya.Middlewares{
            func(ctx zorya.Context, next func(zorya.Context)) {
                // Route-specific middleware
                next(ctx)
            },
        }
    },
)
```

### Middleware Chain

Middleware runs in the order added:
1. API-level middleware
2. Route-level middleware
3. Handler

## Route Groups

Groups allow you to organize routes with shared prefixes, middleware, and transformers.

### Basic Group

```go
group := zorya.NewGroup(api, "/v1")

zorya.Get(group, "/users", getUserHandler)
zorya.Post(group, "/users", createUserHandler)

// Routes registered as:
// GET /v1/users
// POST /v1/users
```

### Group with Middleware

```go
group := zorya.NewGroup(api, "/v1")
group.UseMiddleware(authMiddleware)

zorya.Get(group, "/users", getUserHandler)
// All routes in group require authentication
```

### Multiple Prefixes

```go
group := zorya.NewGroup(api, "/v1", "/api/v1")

zorya.Get(group, "/users", getUserHandler)

// Routes registered as:
// GET /v1/users
// GET /api/v1/users
```

### Group with Transformers

```go
group := zorya.NewGroup(api, "/v1")
group.UseTransformer(func(ctx zorya.Context, status string, v any) (any, error) {
    // Transform responses for this group
    return v, nil
})
```

## Route Security and Authorization

Zorya provides declarative route-based authorization with clean separation of concerns. Security requirements are defined on routes, and enforcement is handled by the security component.

### Overview

Security requirements are declared on routes using the `Secure()` wrapper with composable options:

- **Authentication**: `zorya.Auth()` - Require authenticated user
- **Role-Based**: `zorya.Roles(...)` - Require specific roles
- **Permission-Based**: `zorya.Permissions(...)` - Require specific permissions
- **Resource-Based**: `zorya.Resource(...)` - Resource template for RBAC policies
- **Action**: `zorya.Action(...)` - Custom action (defaults to HTTP method)

By default, routes are **public** (no authentication required). You must explicitly add security requirements.

### Setup

First, register the security enforcement middleware:

```go
import (
    "github.com/talav/security"
    "github.com/talav/zorya"
)

// Create API
api := zorya.NewAPI(adapter)

// Create enforcer
enforcer := security.NewSimpleEnforcer()

// Register security middleware
api.UseMiddleware(security.NewSecurityMiddleware(enforcer,
    security.OnUnauthorized(func(w http.ResponseWriter, r *http.Request) {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
    }),
    security.OnForbidden(func(w http.ResponseWriter, r *http.Request) {
        http.Error(w, "Forbidden", http.StatusForbidden)
    }),
))
```

### Basic Protection

```go
// Public routes (no Secure() = public)
zorya.Get(api, "/health", healthHandler)
zorya.Get(api, "/login", loginHandler)

// Authenticated routes (any logged-in user)
zorya.Get(api, "/profile", profileHandler,
    zorya.Secure(
        zorya.Auth(),
    ),
)

// Role-protected routes
zorya.Get(api, "/admin/dashboard", dashboardHandler,
    zorya.Secure(
        zorya.Roles("admin"),
    ),
)

zorya.Post(api, "/posts", createPostHandler,
    zorya.Secure(
        zorya.Roles("editor", "admin"),
    ),
)

// Permission-protected routes
zorya.Delete(api, "/posts/{id}", deletePostHandler,
    zorya.Secure(
        zorya.Permissions("posts:delete"),
    ),
)
```

### Group-Based Protection

Apply security requirements to entire groups of routes:

```go
// Admin group - all routes require admin role
admin := zorya.NewGroup(api, "/admin")
admin.UseRoles("admin")

zorya.Get(admin, "/users", listUsersHandler)
zorya.Delete(admin, "/users/{id}", deleteUserHandler,
    zorya.Secure(
        zorya.Permissions("users:delete"),
    ),
)

// API v1 group - requires authentication
v1 := zorya.NewGroup(api, "/api/v1")
v1.RequireAuthForGroup()

zorya.Get(v1, "/posts", listPostsHandler) // Requires auth
zorya.Post(v1, "/posts", createPostHandler,
    zorya.Secure(
        zorya.Roles("editor"),
    ),
)
```

### Resource-Based Authorization

Zorya provides three ways to define resources for fine-grained access control:

#### 1. Static Resources

For resources that don't depend on request parameters:

```go
zorya.Get(api, "/reports", listReportsHandler,
    zorya.Secure(
        zorya.Roles("analyst"),
        zorya.Resource("reports"), // Static resource
    ),
)
```

#### 2. Dynamic Resources from Path Parameters (Recommended)

For most dynamic resource cases, use `ResourceFromParams`:

```go
zorya.Get(api, "/organizations/{orgId}/projects", listProjectsHandler,
    zorya.Secure(
        zorya.Roles("member"),
        zorya.ResourceFromParams(func(params map[string]string) string {
            orgId := params["orgId"]
            // Optional: Add custom validation
            if !isValidUUID(orgId) {
                panic("invalid orgId")
            }
            return "organizations/" + orgId + "/projects"
        }),
    ),
)

// Multiple path parameters
zorya.Delete(api, "/organizations/{orgId}/projects/{projectId}", deleteProjectHandler,
    zorya.Secure(
        zorya.Roles("admin", "owner"),
        zorya.Permissions("projects:delete"),
        zorya.ResourceFromParams(func(params map[string]string) string {
            return fmt.Sprintf("organizations/%s/projects/%s", 
                params["orgId"], params["projectId"])
        }),
        zorya.Action("delete"),
    ),
)
```

**Built-in Security:**
- Path parameter values are automatically validated (alphanumeric, `-`, `_` only)
- Invalid characters cause panic (caught by middleware)
- Maximum length: 256 characters

#### 3. Dynamic Resources from Full Request

For complex cases requiring query parameters, headers, or other request data:

```go
zorya.Get(api, "/reports/custom", customReportHandler,
    zorya.Secure(
        zorya.Roles("analyst"),
        zorya.ResourceFromRequest(func(r *http.Request) string {
            year := r.URL.Query().Get("year")
            dept := r.Header.Get("X-Department")
            return fmt.Sprintf("reports/%s/%s", dept, year)
        }),
    ),
)
```

**How It Works:**

1. Zorya's `newRouterParamsMiddleware` extracts path parameters and stores them in context
2. `Secure()` injects a middleware that:
   - Calls your resolver function with params or request
   - Validates the resolved resource
   - Stores fully resolved security metadata in context
3. Security middleware reads resolved metadata and enforces it via the enforcer

**Example Flow:**

Request: `GET /organizations/123/projects`
1. Router extracts: `{"orgId": "123"}`
2. Your resolver: `params["orgId"]` → `"123"`
3. Validates: alphanumeric ✓
4. Builds resource: `"organizations/123/projects"`
5. Enforcer checks: Can user access `"organizations/123/projects"`?

### Security Inheritance and Overrides

Route-level security merges with group-level security:

```go
group := zorya.NewGroup(api, "/api")
group.UseRoles("user")                    // Group requires 'user' role
group.UseResource("api-resources")        // Group resource

zorya.Get(group, "/posts", handler,
    zorya.Secure(
        zorya.Roles("editor"),            // Merged: user, editor
        zorya.Permissions("posts:read"),  // Added permission
    ),
)

// Route inherits group's resource unless overridden
zorya.Get(group, "/admin", adminHandler,
    zorya.Secure(
        zorya.Resource("admin-panel"),    // Overrides group resource
    ),
)
```

### Complete Example

```go
package main

import (
    "context"
    "net/http"
    
    "github.com/talav/security"
    "github.com/talav/zorya"
)

func main() {
    // Create API
    api := zorya.NewAPI(adapter)
    
    // Create enforcer (simple or Casbin)
    enforcer := security.NewSimpleEnforcer()
    
    // Register global security middleware
    api.UseMiddleware(security.NewSecurityMiddleware(enforcer,
        security.OnUnauthorized(func(w http.ResponseWriter, r *http.Request) {
            http.Error(w, "Unauthorized", http.StatusUnauthorized)
        }),
        security.OnForbidden(func(w http.ResponseWriter, r *http.Request) {
            http.Error(w, "Forbidden", http.StatusForbidden)
        }),
    ))
    
    // Public routes
    zorya.Get(api, "/health", healthHandler)
    
    // Protected routes
    zorya.Get(api, "/profile", profileHandler,
        zorya.Secure(zorya.Auth()),
    )
    
    zorya.Get(api, "/admin/users", adminHandler,
        zorya.Secure(zorya.Roles("admin")),
    )
    
    zorya.Get(api, "/orgs/{orgId}/projects", projectsHandler,
        zorya.Secure(
            zorya.Roles("member"),
            zorya.Resource("orgs/{orgId}/projects"),
        ),
    )
    
    http.ListenAndServe(":8080", adapter)
}
```

For Fx dependency injection, see the [fxsecurity README](../../fx/fxsecurity/README.md).

### Architecture

Zorya implements clean separation of concerns for security:

**Zorya's Responsibility** (metadata only):
- `Secure()` attaches security requirements to routes
- Auto-injects middleware that:
  - Reads router params from context
  - Resolves resource templates (`orgs/{orgId}` → `orgs/123`)
  - Stores resolved metadata in context
- No enforcement logic

**Security Component's Responsibility** (enforcement only):
- `NewSecurityMiddleware()` enforces requirements
- Reads resolved metadata from context
- Calls enforcer to check access
- Denies/allows requests

**Integration via Context**:
1. `newRouterParamsMiddleware` (zorya) - extracts path params → context
2. `Secure()` middleware (zorya) - resolves templates, stores metadata → context
3. `SecurityMiddleware` (security) - reads metadata from context, enforces

This architecture ensures:
- Clean separation of routing and security concerns
- No circular dependencies between packages
- Flexibility to use or skip security as needed
- Testable components with clear responsibilities

### OpenAPI Security Documentation

Security requirements are automatically documented in OpenAPI:

```go
zorya.Get(api, "/admin/users", handler,
    zorya.Secure(
        zorya.Roles("admin"),
        zorya.Permissions("users:read"),
    ),
)
```

Generates OpenAPI security:

```json
{
  "paths": {
    "/admin/users": {
      "get": {
        "security": [
          {
            "bearerAuth": ["admin", "users:read"]
          }
        ]
      }
    }
  }
}
```

### Custom Security Enforcers

Implement the `SecurityEnforcer` interface for custom authorization logic:

```go
type SecurityEnforcer interface {
    // Enforce checks if the user meets all security requirements
    Enforce(ctx context.Context, user *AuthUser, requirements *SecurityRequirements) (bool, error)
}
```

Built-in enforcers:
- `SimpleEnforcer` - Basic role checks from `AuthUser.Roles` (provided by security component)
- `CasbinEnforcer` - Full RBAC with Casbin (provided by fxcasbin module)

## Request Limits

### Body Size Limits

Set per-route body size limits:

```go
zorya.Post(api, "/upload", handler,
    func(route *zorya.BaseRoute) {
        route.MaxBodyBytes = 10 * 1024 * 1024 // 10MB
        // Default: 1MB (DefaultMaxBodyBytes)
        // Negative: No limit
    },
)
```

When the limit is exceeded, Zorya returns `413 Request Entity Too Large`.

### Body Read Timeout

Set per-route body read timeouts to prevent slow-loris attacks:

```go
zorya.Post(api, "/upload", handler,
    func(route *zorya.BaseRoute) {
        route.BodyReadTimeout = 10 * time.Second
        // Default: 5 seconds (DefaultBodyReadTimeout)
        // Negative: No timeout
    },
)
```

## Default Parameter Values

Zorya automatically applies default values to missing fields using the `default` struct tag:

```go
type ListUsersInput struct {
    Page     int    `schema:"page,location=query" default:"1"`
    PageSize int    `schema:"page_size,location=query" default:"20"`
    Sort     string `schema:"sort,location=query" default:"created_at"`
    Order    string `schema:"order,location=query" default:"desc"`
}
```

Default values are parsed using the `mapstructure` converter registry, supporting all built-in types and custom types with registered converters.

## Context Interface

The `Context` interface provides access to request information:

```go
type Context interface {
    Context() context.Context
    Request() *http.Request
    RouterParams() map[string]string
    Header(name string) string
    SetReadDeadline(t time.Time) error
    BodyWriter() http.ResponseWriter
}
```

### ResponseWriter Interface

Adapter contexts implement `ResponseWriter` for response control:

```go
type ResponseWriter interface {
    Context
    SetStatus(status int)
    SetHeader(name, value string)
    AppendHeader(name, value string)
    Write(data []byte) (int, error)
}
```

## API Reference

### Types

- `API` - Main API interface
- `Context` - Request context interface
- `ResponseWriter` - Response writer interface
- `Adapter` - Router adapter interface
- `Validator` - Request validation interface
- `Transformer` - Response transformer function type
- `Group` - Route group
- `BaseRoute` - Route configuration
- `RouteSecurity` - Security requirements for a route
- `ErrorModel` - RFC 9457 error model
- `ErrorDetail` - Error detail with code, message, location

### Functions

- `NewAPI(adapter Adapter, opts ...Option) API` - Create new API instance with options
  - `WithValidator(validator Validator) Option` - Set request validator
  - `WithFormat(contentType string, format Format) Option` - Add a single format (merges with defaults)
  - `WithFormats(formats map[string]Format) Option` - Add multiple formats (merges with defaults)
  - `WithFormatsReplace(formats map[string]Format) Option` - Replace all formats (excludes defaults)
  - `WithCodec(codec *schema.Codec) Option` - Set custom codec
  - `WithDefaultFormat(format string) Option` - Set default content type
- `Get[I, O any](api API, path string, handler, ...options)` - Register GET route (panics on errors)
- `Post[I, O any](api API, path string, handler, ...options)` - Register POST route (panics on errors)
- `Put[I, O any](api API, path string, handler, ...options)` - Register PUT route (panics on errors)
- `Patch[I, O any](api API, path string, handler, ...options)` - Register PATCH route (panics on errors)
- `Delete[I, O any](api API, path string, handler, ...options)` - Register DELETE route (panics on errors)
- `Head[I, O any](api API, path string, handler, ...options)` - Register HEAD route (panics on errors)
- `Register[I, O any](api API, route BaseRoute, handler) error` - Register route with full configuration (returns error)
- `NewGroup(api API, prefixes ...string) *Group` - Create route group
- **Security Options:**
  - `Secure(opts ...SecurityOption) RouteOption` - Wrap security requirements
  - `Auth() SecurityOption` - Require authenticated user
  - `Roles(roles ...string) SecurityOption` - Require specific roles
  - `Permissions(permissions ...string) SecurityOption` - Require specific permissions
  - `Resource(resource string) SecurityOption` - Set static resource for RBAC
  - `ResourceFromParams(fn func(params map[string]string) string) SecurityOption` - Resolve resource from path params
  - `ResourceFromRequest(fn func(r *http.Request) string) SecurityOption` - Resolve resource from full request
  - `Action(action string) SecurityOption` - Set custom action (defaults to HTTP method)
- `NewPlaygroundValidator(v Validator) Validator` - Create validator adapter
- `WriteErr(api API, ctx Context, status int, msg string, errs ...error) error` - Write error response
- `Error400BadRequest(msg string, errs ...error) StatusError` - Create 400 error
- `Error404NotFound(msg string, errs ...error) StatusError` - Create 404 error
- `Error422UnprocessableEntity(msg string, errs ...error) StatusError` - Create 422 error
- (and other error convenience functions)

### Constants

- `DefaultMaxBodyBytes int64` - Default body size limit (1MB)
- `DefaultBodyReadTimeout time.Duration` - Default body read timeout (5 seconds)

## Error Processing

Zorya provides comprehensive error handling based on [RFC 9457 Problem Details for HTTP APIs](https://datatracker.ietf.org/doc/html/rfc9457).

### Error Types

#### ErrorModel

`ErrorModel` is the primary error type implementing RFC 9457 Problem Details:

```go
type ErrorModel struct {
    Type     string         `json:"type,omitempty"`     // URI reference to error documentation
    Title    string         `json:"title,omitempty"`    // Short summary (defaults to HTTP status text)
    Status   int            `json:"status,omitempty"`   // HTTP status code
    Detail   string         `json:"detail,omitempty"`    // Human-readable explanation
    Instance string         `json:"instance,omitempty"` // URI reference identifying this occurrence
    Errors   []*ErrorDetail `json:"errors,omitempty"`  // Validation error details
}
```

#### ErrorDetail

`ErrorDetail` provides specific information about individual validation errors:

```go
type ErrorDetail struct {
    Code     string `json:"code,omitempty"`     // Machine-readable error code (e.g., "required", "email")
    Message  string `json:"message,omitempty"`  // Optional human-readable message
    Location string `json:"location,omitempty"` // Path-like string (e.g., "body.items[3].tags")
}
```

### Error Interfaces

#### StatusError

Errors implementing `StatusError` can specify an HTTP status code:

```go
type StatusError interface {
    GetStatus() int
    Error() string
}
```

#### HeadersError

Errors implementing `HeadersError` can set HTTP response headers:

```go
type HeadersError interface {
    GetHeaders() http.Header
    Error() string
}
```

Useful for:
- `WWW-Authenticate` headers for 401 errors
- `Retry-After` headers for 429 errors
- Custom headers for API-specific metadata

#### ErrorDetailer

Custom error types can implement `ErrorDetailer` to provide structured error details:

```go
type ErrorDetailer interface {
    ErrorDetail() *ErrorDetail
}
```

#### ContentTypeFilter

Response types (including errors) can implement `ContentTypeFilter` to override the negotiated content type:

```go
type ContentTypeFilter interface {
    ContentType(string) string
}
```

`ErrorModel` implements this to return `application/problem+json` for JSON responses and `application/problem+cbor` for CBOR responses, as specified by RFC 9457.

### Creating Errors

#### Convenience Functions

```go
// 4xx Client Errors
zorya.Error400BadRequest(msg string, errs ...error)
zorya.Error401Unauthorized(msg string, errs ...error)
zorya.Error403Forbidden(msg string, errs ...error)
zorya.Error404NotFound(msg string, errs ...error)
zorya.Error405MethodNotAllowed(msg string, errs ...error)
zorya.Error406NotAcceptable(msg string, errs ...error)
zorya.Error409Conflict(msg string, errs ...error)
zorya.Error410Gone(msg string, errs ...error)
zorya.Error412PreconditionFailed(msg string, errs ...error)
zorya.Error415UnsupportedMediaType(msg string, errs ...error)
zorya.Error422UnprocessableEntity(msg string, errs ...error)
zorya.Error429TooManyRequests(msg string, errs ...error)

// 5xx Server Errors
zorya.Error500InternalServerError(msg string, errs ...error)
zorya.Error501NotImplemented(msg string, errs ...error)
zorya.Error502BadGateway(msg string, errs ...error)
zorya.Error503ServiceUnavailable(msg string, errs ...error)
zorya.Error504GatewayTimeout(msg string, errs ...error)
```

#### Custom Error Creation

Replace `NewError` or `NewErrorWithContext` to customize error creation:

```go
zorya.NewErrorWithContext = func(ctx zorya.Context, status int, msg string, errs ...error) zorya.StatusError {
    err := zorya.NewError(status, msg, errs...)
    if em, ok := err.(*zorya.ErrorModel); ok {
        em.Instance = ctx.Request().URL.Path
    }
    return err
}
```

### Error Headers

Wrap errors with response headers using `ErrorWithHeaders()`:

```go
err := zorya.Error401Unauthorized("Authentication required")
err = zorya.ErrorWithHeaders(err, http.Header{
    "WWW-Authenticate": []string{`Bearer realm="example"`},
})
```

Headers are merged if an error already has headers.

### Custom Error Types

#### Implementing StatusError

```go
type BusinessError struct {
    Code    string
    Message string
    Status  int
}

func (e *BusinessError) Error() string {
    return e.Message
}

func (e *BusinessError) GetStatus() int {
    return e.Status
}

// Usage
        return nil, &BusinessError{
            Code:    "INVALID_AMOUNT",
            Message: "Amount must be positive",
            Status:  http.StatusBadRequest,
}
```

#### Implementing HeadersError

```go
type RateLimitError struct {
    RetryAfter int
    Limit      int
    Remaining  int
}

func (e *RateLimitError) Error() string {
    return "Rate limit exceeded"
}

func (e *RateLimitError) GetStatus() int {
    return http.StatusTooManyRequests
}

func (e *RateLimitError) GetHeaders() http.Header {
    h := make(http.Header)
    h.Set("Retry-After", strconv.Itoa(e.RetryAfter))
    h.Set("X-RateLimit-Limit", strconv.Itoa(e.Limit))
    h.Set("X-RateLimit-Remaining", strconv.Itoa(e.Remaining))
    return h
}
```

### Error Response Format

Errors are automatically serialized using content negotiation:

**JSON Response:**
```json
{
  "type": "about:blank",
  "title": "Not Found",
  "status": 404,
  "detail": "User not found",
  "errors": [
    {
      "code": "required",
      "message": "email is required",
      "location": "body.email"
    }
  ]
}
```

**Content-Type:** `application/problem+json` (for JSON) or `application/problem+cbor` (for CBOR)

## Examples

### Complete Example

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
        Name  string `json:"name" validate:"required"`
        Email string `json:"email" validate:"required,email"`
    }
}

type CreateUserOutput struct {
    Status int `json:"-"`
    Body struct {
        ID    int    `json:"id"`
        Name  string `json:"name"`
        Email string `json:"email"`
    }
}

type GetUserInput struct {
    ID string `schema:"id,location=path"`
}

type GetUserOutput struct {
    Body struct {
        ID    int    `json:"id"`
        Name  string `json:"name"`
        Email string `json:"email"`
    }
}

func main() {
    router := chi.NewMux()
    adapter := adapters.NewChi(router)
    
    // Configure validator
    validate := validator.New()
    zoryaValidator := zorya.NewPlaygroundValidator(validate)
    
    api := zorya.NewAPI(adapter, zorya.WithValidator(zoryaValidator))
    
    // Add API-level middleware
    api.UseMiddleware(func(ctx zorya.Context, next func(zorya.Context)) {
        log.Printf("Request: %s %s", ctx.Request().Method, ctx.Request().URL.Path)
        next(ctx)
    })
    
    // Register routes
    zorya.Post(api, "/users", createUser)
    zorya.Get(api, "/users/{id}", getUser)
    
    http.ListenAndServe(":8080", router)
}

func createUser(ctx context.Context, input *CreateUserInput) (*CreateUserOutput, error) {
    // Business logic
    user := &CreateUserOutput{
        Status: http.StatusCreated,
        Body: struct {
            ID    int    `json:"id"`
            Name  string `json:"name"`
            Email string `json:"email"`
        }{
            ID:    1,
            Name:  input.Body.Name,
            Email: input.Body.Email,
        },
}
    return user, nil
}

func getUser(ctx context.Context, input *GetUserInput) (*GetUserOutput, error) {
    // Business logic
    if input.ID == "0" {
        return nil, zorya.Error404NotFound("User not found")
    }
    
    return &GetUserOutput{
        Body: struct {
            ID    int    `json:"id"`
            Name  string `json:"name"`
            Email string `json:"email"`
        }{
            ID:    1,
            Name:  "John Doe",
            Email: "john@example.com",
        },
    }, nil
}
```

## Integration

### With Dependency Injection

```go
// Using fx
func NewAPI(fx.Lifecycle, adapter zorya.Adapter) zorya.API {
    api := zorya.NewAPI(adapter)
    // Register routes
    return api
}
```

### With Graceful Shutdown

```go
server := &http.Server{
    Addr:    ":8080",
    Handler: router,
}

go func() {
    if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
        log.Fatal(err)
    }
}()

// Graceful shutdown
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()
server.Shutdown(ctx)
```

## Metadata Generation and Supported Tags

Zorya automatically generates OpenAPI schemas from Go struct types using reflection and struct tags. The metadata system parses struct tags to extract schema information, validation constraints, and OpenAPI-specific metadata.

### Metadata Generation Process

1. **Type Analysis**: Uses Go's `reflect` package to analyze struct types at runtime
2. **Tag Parsing**: Extracts metadata from struct tags using registered parsers
3. **Schema Building**: Generates OpenAPI/JSON Schema from parsed metadata
4. **Caching**: Struct metadata is cached per type for efficient reuse

The metadata system supports multiple tag types that work together to generate complete OpenAPI schemas:

### Supported Tags

#### `schema` Tag

Defines request parameter location and serialization style. See the [schema package documentation](../schema/README.md) for detailed information.

**Format:**
```
schema:"name,location:path|query|header|cookie,style:form|simple|...,explode:true|false,required:true|false"
```

**Example:**
```go
type GetUserInput struct {
    ID      string `schema:"id,location:path,required:true"`
    Format  string `schema:"format,location:query"`
    APIKey  string `schema:"X-API-Key,location:header"`
    Session string `schema:"session_id,location:cookie"`
}
```

#### `body` Tag

Defines request body type and requirements.

**Format:**
```
body:"structured|file|multipart,required:true|false"
```

**Body Types:**
- `structured`: JSON, XML (default)
- `file`: File upload
- `multipart`: Multipart form data

**Example:**
```go
type CreateUserInput struct {
    Body struct {
        Name  string `json:"name"`
        Email string `json:"email"`
    } `body:"structured,required:true"`
}
```

#### `status` Tag

Identifies a field as the HTTP status code for response structs. The field must be of type `int`.

**Format:**
```
status:""
```

**Notes:**
- The field must be of type `int`
- The status code becomes the key in the OpenAPI `responses` map (e.g., `"200"`, `"201"`, `"404"`)
- Default values can be set directly in struct field initialization or using the `default` tag

**Example:**
```go
type CreateUserOutput struct {
    // Status code field - any field name works with 'status' tag
    HTTPStatus int `status:""`
    
    Body struct {
        ID   int    `json:"id"`
        Name string `json:"name"`
    } `body:"structured"`
}

// In handler, set status directly:
return &CreateUserOutput{
    HTTPStatus: http.StatusCreated, // 201
    Body:       userData,
}, nil

// Or use default tag for default value:
type GetUserOutput struct {
    StatusCode int `status:"" default:"200"`
    
    Body struct {
        ID   int    `json:"id"`
        Name string `json:"name"`
    } `body:"structured"`
}
```

#### `openapi` Tag

Provides OpenAPI-specific field-level schema metadata (API contract metadata, not validation constraints).

**Format:**
```
openapi:"readOnly,writeOnly,deprecated,hidden,title=...,description=...,format=...,example=...,examples=[...],x-custom=..."
```

**Supported Options:**

- **Boolean Flags** (empty value means `true`):
  - `readOnly` - Field is read-only
  - `writeOnly` - Field is write-only
  - `deprecated` - Field is deprecated
  - `hidden` - Field is hidden from schema (excluded from properties)

- **String Values**:
  - `title=...` - Title for the schema
  - `description=...` - Description for the schema
  - `format=...` - Format for the schema (e.g., "date", "date-time", "time", "email", "uri")

- **Examples**:
  - `example=...` - Single example value (converted to array)
  - `examples=[...]` - Multiple examples (JSON array string, takes precedence over `example`)

- **Extensions**:
  - `x-*` - OpenAPI extensions (MUST start with `x-`)

**Example:**
```go
type User struct {
    ID          int    `json:"id" openapi:"readOnly,title=User ID,description=Unique identifier"`
    Email       string `json:"email" openapi:"format=email,example=user@example.com"`
    Password    string `json:"password" openapi:"writeOnly,hidden"`
    CreatedAt   string `json:"created_at" openapi:"readOnly,format=date-time"`
    Metadata    string `json:"metadata" openapi:"x-custom-feature=enabled"`
}
```

#### `openapiStruct` Tag

Provides OpenAPI-specific struct-level schema configuration. Must be used on the `_` field of a struct.

**Format:**
```
openapiStruct:"additionalProperties=true|false,nullable=true|false"
```

**Supported Options:**
- `additionalProperties=true|false` - Allow additional properties
- `nullable=true|false` - Struct is nullable

**Example:**
```go
type Config struct {
    _ struct{} `openapiStruct:"additionalProperties=false,nullable=false"`
    
    Name  string `json:"name"`
    Value string `json:"value"`
}
```

#### `validate` Tag

Provides validation constraints using go-playground/validator format. These constraints are mapped to OpenAPI/JSON Schema validation keywords.

**Format:**
```
validate:"required,email,min=5,max=100,pattern=^[a-z]+$,oneof=red green blue"
```

**Supported Validators:**

**Boolean Flags:**
- `required` - Field must be present

**Numeric Constraints:**
- `min=N` - Minimum value (as float64)
- `max=N` - Maximum value (as float64)
- `gte=N` - Greater than or equal (maps to `minimum`)
- `lte=N` - Less than or equal (maps to `maximum`)
- `gt=N` - Greater than (maps to `exclusiveMinimum`)
- `lt=N` - Less than (maps to `exclusiveMaximum`)
- `multiple_of=N` - Value must be a multiple of N

**String Length Constraints:**
- `len=N` - Exact length (sets both `minLength` and `maxLength`)

**String Format Constraints:**
- `email` - Valid email address (maps to `format: "email"`)
- `url` - Valid URL (maps to `format: "uri"`)
- `alpha` - Alphabetic characters only (maps to `pattern: "^[a-zA-Z]+$"`)
- `alphanum` - Alphanumeric characters only (maps to `pattern: "^[a-zA-Z0-9]+$"`)
- `alphaunicode` - Unicode alphabetic characters only (maps to `pattern: "^[\\p{L}]+$"`)
- `alphanumunicode` - Unicode alphanumeric characters only (maps to `pattern: "^[\\p{L}\\p{N}]+$"`)
- `pattern=...` - Regular expression pattern

**Enum/OneOf:**
- `oneof=...` - Value must be one of the specified space-separated values (maps to `enum`)

**Example:**
```go
type CreateUserInput struct {
    Body struct {
        Email    string `json:"email" validate:"required,email"`
        Name     string `json:"name" validate:"required,min=3,max=100"`
        Age      int    `json:"age" validate:"min=0,max=150"`
        Username string `json:"username" validate:"required,alphanum,len=8"`
        Status   string `json:"status" validate:"oneof=active inactive pending"`
        Score    float64 `json:"score" validate:"gte=0,lte=100"`
    }
}
```

#### `default` Tag

Provides default values for fields. Used for both runtime default application and OpenAPI schema generation.

**Format:**
```
default:"value"
```

**Value Parsing:**
- **Strings**: Returned as-is (no quotes needed in tag)
- **Numbers, booleans, arrays, objects**: Parsed as JSON

**Example:**
```go
type ListUsersInput struct {
    Page     int    `schema:"page,location:query" default:"1"`
    PageSize int    `schema:"page_size,location:query" default:"20"`
    Sort     string `schema:"sort,location:query" default:"created_at"`
    Order    string `schema:"order,location:query" default:"desc"`
    Active   bool   `schema:"active,location:query" default:"true"`
    Tags     []string `schema:"tags,location:query" default:"[\"default\"]"`
}
```

#### `dependentRequired` Tag

Specifies that certain fields must be present when this field is present (JSON Schema `dependentRequired` keyword).

**Format:**
```
dependentRequired:"field1,field2,field3"
```

**Example:**
```go
type PaymentInput struct {
    PaymentMethod string `json:"payment_method" dependentRequired:"billing_address,cardholder_name"`
    
    BillingAddress string `json:"billing_address"`
    CardholderName string `json:"cardholder_name"`
}
```

When `payment_method` is present, `billing_address` and `cardholder_name` become required.

### Metadata Application Order

When generating schemas, metadata is applied in the following order:

1. **Base Schema Generation**: Type-based schema generation (primitives, arrays, structs)
2. **OpenAPI Metadata**: Field-level OpenAPI metadata (`openapi` tag)
3. **Validation Metadata**: Validation constraints (`validate` tag)
4. **Default Values**: Default value application (`default` tag)
5. **Dependent Required**: Object-level dependent required constraints
6. **Struct-Level Metadata**: Struct-level configuration (`openapiStruct` tag on `_` field)

### Complete Example

```go
type CreateUserInput struct {
    // Path parameter
    ID string `schema:"id,location:path,required:true"`
    
    // Query parameters with defaults
    Format string `schema:"format,location:query" default:"json"`
    
    // Request body
    Body struct {
        // Required fields with validation
        Email    string `json:"email" validate:"required,email" openapi:"format=email,example=user@example.com"`
        Name     string `json:"name" validate:"required,min=3,max=100" openapi:"description=User's full name"`
        
        // Optional fields
        Age      *int   `json:"age,omitempty" validate:"min=0,max=150" openapi:"description=User's age in years"`
        Username string `json:"username" validate:"required,alphanum,len=8" openapi:"title=Username"`
        
        // Read-only field (set by server)
        CreatedAt string `json:"created_at" openapi:"readOnly,format=date-time"`
        
        // Write-only field (password)
        Password  string `json:"password" openapi:"writeOnly,hidden"`
        
        // Enum field
        Status    string `json:"status" validate:"oneof=active inactive pending" openapi:"description=Account status"`
        
        // Field with dependent required
        PaymentMethod string `json:"payment_method" dependentRequired:"billing_address"`
        BillingAddress string `json:"billing_address"`
    } `body:"structured,required:true"`
}
```

## Performance

- **Metadata Caching**: Struct metadata is cached per type for efficient request decoding
- **Zero Allocations**: Minimal allocations in hot paths
- **Efficient Adapters**: Lightweight adapter implementations

## See Also

- [Schema Package](../schema/README.md) - Request parameter decoding
- [Mapstructure Package](../mapstructure/README.md) - Map to struct conversion
- [Negotiation Package](../negotiation/README.md) - Content negotiation
- [Validator Package](../validator/README.md) - Request validation
- [RFC 9457: Problem Details for HTTP APIs](https://datatracker.ietf.org/doc/html/rfc9457)
