package zorya

import (
	"context"
	"net/http"
)

type contextKey string

const routeSecurityContextKey contextKey = "route_security_context"

// RouteSecurityContext contains fully resolved security requirements
// It's stored in context by Zorya's security middleware after resolving resources.
type RouteSecurityContext struct {
	Roles       []string
	Permissions []string
	Resource    string
	Action      string // Resolved action (explicit or HTTP method fallback)
}

// GetRouteSecurityContext retrieves resolved security metadata from request context.
func GetRouteSecurityContext(r *http.Request) *RouteSecurityContext {
	sec, ok := r.Context().Value(routeSecurityContextKey).(*RouteSecurityContext)
	if !ok {
		return nil
	}

	return sec
}

// newSecurityMetadataMiddleware creates middleware that resolves security resource identifiers
// and stores metadata in context for the security enforcer to read.
// Returns nil if the route has no security requirements.
func newSecurityMetadataMiddleware(security *RouteSecurity) Middleware {
	if security == nil {
		return nil
	}

	// Create middleware that resolves resources and stores metadata in context
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract params once (injected by newRouterParamsMiddleware)
			params := GetRouterParams(r)

			// Resolve resource identifier
			resolvedResource := ""
			if security.ResourceResolver != nil {
				// Use custom resolver (from ResourceFromParams or ResourceFromRequest)
				// Pass both request and params - resolver uses what it needs
				resolvedResource = security.ResourceResolver(r, params)
			} else if security.Resource != "" {
				// Use static resource
				resolvedResource = security.Resource
			}

			// Resolve action (use explicit action or fallback to HTTP method)
			action := security.Action
			if action == "" {
				action = r.Method
			}

			// Create resolved security metadata
			resolved := &RouteSecurityContext{
				Roles:       security.Roles,
				Permissions: security.Permissions,
				Resource:    resolvedResource,
				Action:      action,
			}

			// Store in context for security middleware to enforce
			r = r.WithContext(context.WithValue(r.Context(), routeSecurityContextKey, resolved))

			next.ServeHTTP(w, r)
		})
	}
}
