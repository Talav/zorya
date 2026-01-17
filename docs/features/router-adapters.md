# Router Adapters

Zorya works with multiple HTTP routers through a clean adapter pattern. Your business logic stays the same regardless of which router you use.

## Overview

Zorya supports three router backends out of the box:

- **Chi** - Lightweight, idiomatic Go router with middleware support
- **Fiber** - Express-inspired web framework built on Fasthttp
- **Standard Library** - Go 1.22+ ServeMux with pattern matching

All adapters provide the same functionality through a common interface.

## Chi Router

Chi is a lightweight, idiomatic Go router with excellent middleware support.

### Installation

```bash
go get github.com/go-chi/chi/v5
```

### Basic Usage

```go
import (
    "github.com/go-chi/chi/v5"
    "github.com/talav/talav/pkg/component/zorya"
    "github.com/talav/talav/pkg/component/zorya/adapters"
)

func main() {
    // Create Chi router
    router := chi.NewMux()
    
    // Create Zorya adapter
    adapter := adapters.NewChi(router)
    
    // Create Zorya API
    api := zorya.NewAPI(adapter)
    
    // Register routes
    zorya.Get(api, "/users/{id}", getUserHandler)
    zorya.Post(api, "/users", createUserHandler)
    
    // Start server
    http.ListenAndServe(":8080", router)
}
```

### With Chi Middleware

Chi middleware works seamlessly with Zorya:

```go
router := chi.NewMux()

// Add Chi middleware
router.Use(middleware.Logger)
router.Use(middleware.Recoverer)
router.Use(middleware.RequestID)

adapter := adapters.NewChi(router)
api := zorya.NewAPI(adapter)

// Register Zorya routes
zorya.Get(api, "/users", listUsersHandler)
```

### Path Parameters

Chi uses `{param}` syntax for path parameters:

```go
type GetUserInput struct {
    UserID string `schema:"id,location=path"`
}

// Route: /users/{id}
zorya.Get(api, "/users/{id}", getUserHandler)
```

## Fiber

Fiber is an Express-inspired web framework built on Fasthttp for high performance.

### Installation

```bash
go get github.com/gofiber/fiber/v2
```

### Basic Usage

```go
import (
    "github.com/gofiber/fiber/v2"
    "github.com/talav/talav/pkg/component/zorya/adapters"
)

func main() {
    // Create Fiber app
    app := fiber.New()
    
    // Create Zorya adapter
    adapter := adapters.NewFiber(app)
    
    // Create Zorya API
    api := zorya.NewAPI(adapter)
    
    // Register routes
    zorya.Get(api, "/users/:id", getUserHandler)
    zorya.Post(api, "/users", createUserHandler)
    
    // Start server
    app.Listen(":8080")
}
```

### With Fiber Middleware

```go
app := fiber.New()

// Add Fiber middleware
app.Use(logger.New())
app.Use(recover.New())
app.Use(cors.New())

adapter := adapters.NewFiber(app)
api := zorya.NewAPI(adapter)

// Register Zorya routes
zorya.Get(api, "/users", listUsersHandler)
```

### Path Parameters

Fiber uses `:param` syntax for path parameters:

```go
type GetUserInput struct {
    UserID string `schema:"id,location=path"`
}

// Route: /users/:id
zorya.Get(api, "/users/:id", getUserHandler)
```

## Standard Library (Go 1.22+)

Go 1.22+ includes enhanced pattern matching in `http.ServeMux`.

### Basic Usage

```go
import (
    "net/http"
    "github.com/talav/talav/pkg/component/zorya/adapters"
)

func main() {
    // Create stdlib mux
    mux := http.NewServeMux()
    
    // Create Zorya adapter
    adapter := adapters.NewStdlib(mux)
    
    // Create Zorya API
    api := zorya.NewAPI(adapter)
    
    // Register routes
    zorya.Get(api, "/users/{id}", getUserHandler)
    zorya.Post(api, "/users", createUserHandler)
    
    // Start server
    http.ListenAndServe(":8080", mux)
}
```

### With Prefix

```go
mux := http.NewServeMux()

// All routes will be prefixed with /api
adapter := adapters.NewStdlibWithPrefix(mux, "/api")
api := zorya.NewAPI(adapter)

// Actual route: /api/users/{id}
zorya.Get(api, "/users/{id}", getUserHandler)
```

### Path Parameters

Go 1.22+ uses `{param}` syntax:

```go
type GetUserInput struct {
    UserID string `schema:"id,location=path"`
}

// Route: /users/{id}
zorya.Get(api, "/users/{id}", getUserHandler)
```

## HTTP Methods

All adapters support the same HTTP methods:

```go
zorya.Get(api, "/users/{id}", getUser)
zorya.Post(api, "/users", createUser)
zorya.Put(api, "/users/{id}", updateUser)
zorya.Patch(api, "/users/{id}", patchUser)
zorya.Delete(api, "/users/{id}", deleteUser)
zorya.Head(api, "/users/{id}", headUser)
```

### Error Handling

Convenience functions (`Get`, `Post`, etc.) panic on errors since route registration happens during startup:

```go
// This panics if there's an error
zorya.Get(api, "/users/{id}", handler)
```

For advanced error handling, use `Register` directly:

```go
route := &zorya.BaseRoute{
    Method:  http.MethodGet,
    Path:    "/users/{id}",
    Handler: handler,
}

err := zorya.Register(api, route, handler)
if err != nil {
    log.Fatalf("Failed to register route: %v", err)
}
```

## Route Configuration

Advanced route configuration works the same across all adapters:

```go
zorya.Post(api, "/upload", uploadHandler,
    func(route *zorya.BaseRoute) {
        // Body size limit
        route.MaxBodyBytes = 10 * 1024 * 1024 // 10MB
        
        // Read timeout
        route.BodyReadTimeout = 30 * time.Second
        
        // Route-specific middleware
        route.Middlewares = zorya.Middlewares{
            authMiddleware,
            loggingMiddleware,
        }
    },
)
```

## Adapter Interface

All adapters implement the `Adapter` interface:

```go
type Adapter interface {
    Handle(route *BaseRoute, handler http.HandlerFunc)
}
```

## Custom Adapters

You can implement custom adapters for other routers:

```go
type MyAdapter struct {
    router *MyRouter
}

func (a *MyAdapter) Handle(route *zorya.BaseRoute, handler http.HandlerFunc) {
    // Convert Zorya route to your router's format
    a.router.HandleFunc(route.Path, route.Method, handler)
}

// Use it
adapter := &MyAdapter{router: myRouter}
api := zorya.NewAPI(adapter)
```

## Comparison

| Feature | Chi | Fiber | Stdlib |
|---------|-----|-------|--------|
| Performance | Fast | Very Fast | Fast |
| Middleware ecosystem | Large | Large | Growing |
| Memory usage | Low | Low | Very Low |
| Path syntax | `{id}` | `:id` | `{id}` |
| Context API | Chi context | Fiber context | Stdlib context |
| Dependencies | Few | Many | None |

## Best Practices

### Choose Based On Needs

- **Chi**: Best for most Go projects, excellent middleware ecosystem
- **Fiber**: When you need maximum performance and Express-like API
- **Stdlib**: When you want zero dependencies and Go 1.22+ features

### Consistent Path Syntax

Use the router's native path syntax:

```go
// Chi and Stdlib
zorya.Get(api, "/users/{id}", handler)

// Fiber
zorya.Get(api, "/users/:id", handler)
```

### Middleware Ordering

Place router middleware before Zorya API creation:

```go
router := chi.NewMux()
router.Use(middleware.Logger)  // Router middleware first

adapter := adapters.NewChi(router)
api := zorya.NewAPI(adapter)
api.UseMiddleware(jwtMiddleware)  // Zorya middleware second
```

### Testing

Use the adapter in tests:

```go
func TestGetUser(t *testing.T) {
    router := chi.NewMux()
    adapter := adapters.NewChi(router)
    api := zorya.NewAPI(adapter)
    
    zorya.Get(api, "/users/{id}", getUserHandler)
    
    req := httptest.NewRequest("GET", "/users/123", nil)
    w := httptest.NewRecorder()
    
    router.ServeHTTP(w, req)
    
    assert.Equal(t, 200, w.Code)
}
```

## Complete Example

```go
package main

import (
    "context"
    "net/http"
    
    "github.com/go-chi/chi/v5"
    "github.com/go-chi/chi/v5/middleware"
    "github.com/talav/talav/pkg/component/zorya"
    "github.com/talav/talav/pkg/component/zorya/adapters"
)

type GetUserInput struct {
    ID string `schema:"id,location=path"`
}

type GetUserOutput struct {
    Body struct {
        ID   string `json:"id"`
        Name string `json:"name"`
    } `body:"structured"`
}

func main() {
    // Create router with middleware
    router := chi.NewMux()
    router.Use(middleware.Logger)
    router.Use(middleware.Recoverer)
    
    // Create Zorya API
    adapter := adapters.NewChi(router)
    api := zorya.NewAPI(adapter)
    
    // Register routes
    zorya.Get(api, "/users/{id}", getUser)
    zorya.Post(api, "/users", createUser)
    
    // Start server
    http.ListenAndServe(":8080", router)
}

func getUser(ctx context.Context, input *GetUserInput) (*GetUserOutput, error) {
    return &GetUserOutput{
        Body: struct {
            ID   string `json:"id"`
            Name string `json:"name"`
        }{
            ID:   input.ID,
            Name: "John Doe",
        },
    }, nil
}
```

## See Also

- [Quick Start Tutorial](../tutorial/quick-start.md) - Get started with Zorya
- [Middleware](middleware.md) - Add middleware to your API
- [Route Groups](groups.md) - Organize routes with groups
- [Testing](../how-to/testing.md) - Test your API
