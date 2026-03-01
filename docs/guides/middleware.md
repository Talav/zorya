# Middleware

Zorya middleware is standard Go `func(http.Handler) http.Handler`. Any Chi, Gorilla, or stdlib-compatible middleware works directly.

## Adding middleware

```go
// API-level: runs for every request
api.UseMiddleware(loggingMiddleware, recoveryMiddleware)

// Group-level: runs for routes in the group only
grp := zorya.NewGroup(api, "/api")
grp.UseMiddleware(authMiddleware)
```

## Middleware type

```go
type Middleware func(http.Handler) http.Handler
```

## Execution order

Middleware is applied in the order it is added. API-level middleware runs first, then group-level, then route-level.

```
Request → API middleware → Group middleware → Route middleware → Handler
```

## Writing middleware

```go
func loggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        next.ServeHTTP(w, r)
        log.Printf("%s %s %v", r.Method, r.URL.Path, time.Since(start))
    })
}
```

## Response transformers

Transformers modify the response body struct *before* it is serialized. They run after the handler returns.

```go
type Transformer func(r *http.Request, status int, result any) (any, error)
```

Register a transformer:

```go
api.UseTransformer(func(r *http.Request, status int, v any) (any, error) {
    return map[string]any{
        "data":      v,
        "timestamp": time.Now().Unix(),
    }, nil
})
```

!!! note
    Transformers only run for struct body types. `[]byte` and streaming (`func(http.ResponseWriter) error`) bodies bypass transformers.

## Accessing the request context

Middleware can attach values to the request context using standard `context.WithValue`. Handlers retrieve them via the `context.Context` passed as the first argument:

```go
type ctxKey string

func tenantMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        tenantID := r.Header.Get("X-Tenant-ID")
        ctx := context.WithValue(r.Context(), ctxKey("tenant"), tenantID)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

func myHandler(ctx context.Context, input *MyInput) (*MyOutput, error) {
    tenant := ctx.Value(ctxKey("tenant")).(string)
    // ...
}
```
