package zorya

import (
	"context"
	"net/http"
)

type paramsContextKey string

const routerParamsKey paramsContextKey = "zorya.routerParams"

// GetRouterParams retrieves router parameters from the request context.
// Returns nil if no parameters are available.
//
// Example:
//
//	params := zorya.GetRouterParams(r)
//	orgID := params["orgId"]
func GetRouterParams(r *http.Request) map[string]string {
	if params, ok := r.Context().Value(routerParamsKey).(map[string]string); ok {
		return params
	}

	return make(map[string]string)
}

// newRouterParamsMiddleware creates middleware that extracts router parameters from the request
// and stores them in the request context. This makes path parameters available to downstream
// middleware (e.g., security enforcement, logging) and handlers.
//
// The middleware uses the adapter's ExtractRouterParams method to handle
// router-specific parameter extraction (e.g., Chi's URL params, Fiber's Params,
// Go 1.22+ PathValue).
//
// Router parameters are stored in context and can be retrieved using GetRouterParams.
func newRouterParamsMiddleware(adapter Adapter, route *BaseRoute) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract params using adapter-specific logic
			params := adapter.ExtractRouterParams(r, route)

			// Store in context for downstream use
			ctx := context.WithValue(r.Context(), routerParamsKey, params)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
