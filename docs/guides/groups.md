# Route Groups

Groups let you share a URL prefix, middleware, security, and response transformers across a set of routes.

## Creating a group

```go
v1 := zorya.NewGroup(api, "/v1")

zorya.Get(v1, "/users",    listUsers)
zorya.Post(v1, "/users",   createUser)
zorya.Get(v1, "/users/{id}", getUser)
```

All three routes are registered with the `/v1` prefix: `/v1/users`, `/v1/users`, `/v1/users/{id}`.

## Shared middleware

```go
v1 := zorya.NewGroup(api, "/v1")
v1.UseMiddleware(authMiddleware, loggingMiddleware)
```

Middleware added to a group runs only for routes in that group, after any API-level middleware.

## Shared security

```go
adminGroup := zorya.NewGroup(api, "/admin")
adminGroup.WithSecurity(&zorya.RouteSecurity{
    Roles: []string{"admin"},
})

zorya.Get(adminGroup, "/users",  listAllUsers)   // requires admin role
zorya.Delete(adminGroup, "/users/{id}", deleteUser) // requires admin role
```

Individual routes can narrow the security requirements by adding `zorya.Secure(...)` options; they cannot loosen group-level requirements.

## Shared transformers

```go
grp := zorya.NewGroup(api, "/api")
grp.UseTransformer(func(r *http.Request, status int, v any) (any, error) {
    // wrap every response in an envelope
    return map[string]any{"data": v}, nil
})
```

See [Middleware](middleware.md) for the transformer type signature.

## Nested groups

Groups can be nested to any depth:

```go
api  := zorya.NewAPI(adapter)
v2   := zorya.NewGroup(api,  "/v2")
admin := zorya.NewGroup(v2,  "/admin")  // prefix: /v2/admin

zorya.Get(admin, "/stats", getStats)    // registered at /v2/admin/stats
```

## Route-level overrides

Options passed directly to `Get`, `Post`, etc. apply only to that route and are processed after group modifiers:

```go
zorya.Get(v1, "/health", healthCheck, func(r *zorya.BaseRoute) {
    r.Middlewares = nil // no middleware for health check
})
```
