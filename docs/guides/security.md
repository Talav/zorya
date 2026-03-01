# Security

Zorya provides declarative authorization at the route and group level. It separates *declaring* requirements (what roles/permissions a route needs) from *enforcing* them (your auth middleware decides whether the caller satisfies those requirements).

## Declaring security on a route

```go
zorya.Get(api, "/admin/reports", handler,
    zorya.Secure(
        zorya.Roles("admin"),
    ),
)
```

`Secure` requires at least one of `Roles`, `Permissions`, or `Resource`. Calling `Secure()` with no options panics.

## Security options

### Roles

The caller must have **at least one** of the listed roles:

```go
zorya.Secure(zorya.Roles("admin", "super-admin"))
```

### Permissions

The caller must have **all** listed permissions:

```go
zorya.Secure(zorya.Permissions("users:read", "users:write"))
```

### Static resource

Associates an RBAC resource identifier with the route:

```go
zorya.Secure(
    zorya.Roles("member"),
    zorya.Resource("projects"),
)
```

### Dynamic resource from path params

```go
zorya.Get(api, "/orgs/{orgId}/repos", handler,
    zorya.Secure(
        zorya.Roles("member"),
        zorya.ResourceFromParams(func(params map[string]string) string {
            return "orgs/" + params["orgId"]
        }),
    ),
)
```

### Dynamic resource from full request

Use `ResourceFromRequest` when you need query parameters or headers:

```go
zorya.Secure(
    zorya.Roles("analyst"),
    zorya.ResourceFromRequest(func(r *http.Request) string {
        return "reports/" + r.URL.Query().Get("year")
    }),
)
```

### Action

Override the default action (which falls back to the HTTP method):

```go
zorya.Secure(
    zorya.Roles("editor"),
    zorya.Action("publish"),
)
```

## Enforcing security in middleware

Zorya stores the resolved security context in the request. Your auth middleware reads it:

```go
func authMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        sec := zorya.GetRouteSecurityContext(r)
        if sec == nil {
            // No security requirements — public route
            next.ServeHTTP(w, r)
            return
        }

        // Validate the caller's token and extract their roles
        callerRoles := extractRolesFromToken(r.Header.Get("Authorization"))

        // Check if caller has any required role
        if !hasAnyRole(callerRoles, sec.Roles) {
            http.Error(w, "Forbidden", http.StatusForbidden)
            return
        }

        next.ServeHTTP(w, r)
    })
}

api.UseMiddleware(authMiddleware)
```

`RouteSecurityContext` fields:

```go
type RouteSecurityContext struct {
    Roles       []string
    Permissions []string
    Resource    string   // Resolved (e.g. "orgs/abc-123")
    Action      string   // Resolved (explicit or HTTP method)
}
```

## Group-level security

Apply security to all routes in a group:

```go
adminGroup := zorya.NewGroup(api, "/admin")
adminGroup.WithSecurity(&zorya.RouteSecurity{
    Roles: []string{"admin"},
})

zorya.Get(adminGroup, "/users", listUsers)     // admin only
zorya.Delete(adminGroup, "/users/{id}", deleteUser) // admin only
```

## Routes without security

Routes registered without `Secure(...)` are public. `GetRouteSecurityContext` returns `nil` for them.
