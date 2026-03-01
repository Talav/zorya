# Router Adapters

Zorya is router-agnostic. It communicates with the underlying HTTP router through the `Adapter` interface. Three adapters are included in the `adapters` package.

## Chi

[Chi](https://github.com/go-chi/chi) is the recommended adapter for most projects. It supports middleware, route groups, and URL parameters.

```go
import (
    "github.com/go-chi/chi/v5"
    "github.com/talav/zorya"
    "github.com/talav/zorya/adapters"
)

router := chi.NewMux()
api := zorya.NewAPI(adapters.NewChi(router))

// Register routes on api, then serve using the Chi router
http.ListenAndServe(":8080", router)
```

Path parameters use Chi's `:param` syntax in the route and `schema:"param,location=path"` in the input struct:

```go
type GetUserInput struct {
    ID int `schema:"id,location=path"`
}

zorya.Get(api, "/users/{id}", handler)
```

## Fiber

[Fiber](https://gofiber.io/) is a high-performance adapter built on fasthttp. Use it when raw throughput is the primary concern.

```go
import (
    "github.com/gofiber/fiber/v2"
    "github.com/talav/zorya"
    "github.com/talav/zorya/adapters"
)

app := fiber.New()
api := zorya.NewAPI(adapters.NewFiber(app))

// Register routes on api, then serve using the Fiber app
app.Listen(":8080")
```

!!! note
    Fiber uses its own server loop. Call `app.Listen` instead of `http.ListenAndServe`.

## Standard Library (net/http)

The stdlib adapter works with Go's built-in `http.ServeMux` (Go 1.22+ pattern syntax supported).

```go
import (
    "net/http"
    "github.com/talav/zorya"
    "github.com/talav/zorya/adapters"
)

mux := http.NewServeMux()
api := zorya.NewAPI(adapters.NewStdlib(mux))

http.ListenAndServe(":8080", mux)
```

If your mux is mounted at a sub-path (e.g. behind a reverse proxy), use the prefix variant so path parameters are resolved correctly:

```go
api := zorya.NewAPI(adapters.NewStdlibWithPrefix(mux, "/api/v1"))
```

## Implementing a Custom Adapter

Any type that satisfies the `zorya.Adapter` interface can be used:

```go
type Adapter interface {
    Handle(route *zorya.BaseRoute, handler http.HandlerFunc)
    ExtractRouterParams(r *http.Request, route *zorya.BaseRoute) map[string]string
}
```

- `Handle` registers a route with the underlying router.
- `ExtractRouterParams` extracts path parameters from the request at serve time.

See the existing adapters in `adapters/` for reference implementations.
