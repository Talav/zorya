package zorya

import (
	"net/http"
	"time"
)

// DefaultMaxBodyBytes is the default maximum request body size (1MB).
const DefaultMaxBodyBytes int64 = 1024 * 1024

// DefaultBodyReadTimeout is the default timeout for reading request bodies (5 seconds).
const DefaultBodyReadTimeout = 5 * time.Second

// BaseRoute is the base struct for all routes in Fuego.
// It contains the OpenAPI operation and other metadata.
type BaseRoute struct {
	// OpenAPI operation
	Operation *Operation

	// HTTP method (GET, POST, PUT, PATCH, DELETE)
	Method string

	// URL path. Will be prefixed by the base path of the server and the group path if any
	Path string

	// DefaultStatus is the default HTTP status code for this operation. It will
	// be set to 200 or 204 if not specified, depending on whether the handler
	// returns a response body.
	DefaultStatus int

	// Middlewares is a list of middleware functions to run before the handler.
	// This is useful for adding custom logic to operations, such as logging,
	// authentication, or rate limiting.
	Middlewares Middlewares

	// BodyReadTimeout sets a deadline for reading the request body.
	// If > 0, sets read deadline to now + timeout.
	// If == 0, uses DefaultBodyReadTimeout (5 seconds).
	// If < 0, disables any deadline (no timeout).
	BodyReadTimeout time.Duration

	// MaxBodyBytes limits the size of the request body in bytes.
	// If > 0, enforces the specified limit.
	// If == 0, uses DefaultMaxBodyBytes (1MB).
	// If < 0, disables the limit (no size restriction).
	MaxBodyBytes int64

	// Errors is a list of HTTP status codes that the handler may return. If
	// not specified, then a default error response is added to the OpenAPI.
	// This is a convenience for handlers that return a fixed set of errors
	// where you do not wish to provide each one as an OpenAPI response object.
	// Each error specified here is expanded into a response object with the
	// schema generated from the type returned by `NewError()`.
	Errors []int

	// Security holds authorization requirements for this route.
	// Routes without Security are public by default (anonymous access allowed).
	// Adding any security requirement makes the route protected.
	Security *RouteSecurity
}

// RouteSecurity defines authorization requirements for a route.
// Routes without RouteSecurity are public by default (anonymous access allowed).
// Adding any security requirement makes the route protected.
type RouteSecurity struct {
	// Roles required (any match grants access)
	Roles []string

	// Permissions required (all must match)
	Permissions []string

	// Resource identifier for RBAC (static or template)
	Resource string

	// ResourceResolver dynamically resolves the resource string at runtime.
	// If set, this takes precedence over Resource field.
	// Receives the request and path parameters extracted by the router.
	ResourceResolver func(r *http.Request, params map[string]string) string

	// Action for RBAC
	Action string
}

// SecurityOption configures security requirements for a route.
type SecurityOption func(*RouteSecurity)

// Secure wraps security options and automatically injects metadata middleware.
// The metadata middleware resolves resource templates and stores requirements in context.
// Enforcement is done by the global security middleware (registered via api.UseMiddleware).
//
// Secure() requires at least one security option (Roles, Permissions, or Resource).
// Routes without Secure() are public by default.
//
// Usage:
//
//	zorya.Get(api, "/admin/users", handler,
//		zorya.Secure(
//			zorya.Roles("admin"),
//		),
//	)
//
//	zorya.Get(api, "/orgs/{orgId}/projects", handler,
//		zorya.Secure(
//			zorya.Roles("member"),
//			zorya.ResourceFromParams(func(params map[string]string) string {
//				return "orgs/" + params["orgId"] + "/projects"
//			}),
//		),
//	)
func Secure(opts ...SecurityOption) func(*BaseRoute) {
	if len(opts) == 0 {
		panic("zorya.Secure() requires at least one security option. Use Roles(), Permissions(), or Resource() to define security requirements.")
	}

	return func(r *BaseRoute) {
		// Initialize security metadata
		if r.Security == nil {
			r.Security = &RouteSecurity{}
		}

		// Apply all security options
		for _, opt := range opts {
			opt(r.Security)
		}

		// Validate that at least one requirement was set
		if !hasSecurityRequirements(r.Security) {
			panic("zorya.Secure() requires at least one security requirement. Use Roles(), Permissions(), or Resource() to define security requirements.")
		}
	}
}

// Roles requires the user to have at least one of the specified roles.
func Roles(roles ...string) SecurityOption {
	return func(s *RouteSecurity) {
		s.Roles = append(s.Roles, roles...)
	}
}

// Permissions requires the user to have all specified permissions.
func Permissions(perms ...string) SecurityOption {
	return func(s *RouteSecurity) {
		s.Permissions = append(s.Permissions, perms...)
	}
}

// Resource sets a static RBAC resource identifier.
// For dynamic resources with path parameters, use ResourceFromParams or ResourceFromRequest.
func Resource(resource string) SecurityOption {
	return func(s *RouteSecurity) {
		s.Resource = resource
		s.ResourceResolver = nil
	}
}

// ResourceFromParams resolves a resource using path parameters.
// The resolver function receives only the path parameters extracted from the route.
// This is the recommended approach for most dynamic resource cases.
//
// Example:
//
//	zorya.Get(api, "/orgs/{orgId}/projects", handler,
//	    zorya.Secure(
//	        zorya.Roles("member"),
//	        zorya.ResourceFromParams(func(params map[string]string) string {
//	            orgId := params["orgId"]
//	            if !isValidUUID(orgId) {
//	                panic("invalid orgId")
//	            }
//	            return "orgs/" + orgId + "/projects"
//	        }),
//	    ),
//	)
func ResourceFromParams(fn func(params map[string]string) string) SecurityOption {
	return func(s *RouteSecurity) {
		s.Resource = ""
		s.ResourceResolver = func(r *http.Request, params map[string]string) string {
			return fn(params)
		}
	}
}

// ResourceFromRequest resolves a resource using the full HTTP request.
// Use this for complex cases that need query parameters, headers, or other request data.
// For simple path parameter cases, prefer ResourceFromParams.
//
// Example:
//
//	zorya.Get(api, "/reports", handler,
//	    zorya.Secure(
//	        zorya.Roles("analyst"),
//	        zorya.ResourceFromRequest(func(r *http.Request) string {
//	            year := r.URL.Query().Get("year")
//	            dept := r.Header.Get("X-Department")
//	            return fmt.Sprintf("reports/%s/%s", dept, year)
//	        }),
//	    ),
//	)
func ResourceFromRequest(fn func(r *http.Request) string) SecurityOption {
	return func(s *RouteSecurity) {
		s.Resource = ""
		s.ResourceResolver = func(r *http.Request, params map[string]string) string {
			return fn(r)
		}
	}
}

// Action sets the RBAC action.
func Action(action string) SecurityOption {
	return func(s *RouteSecurity) {
		s.Action = action
	}
}

// hasSecurityRequirements checks if RouteSecurity has any requirements defined.
func hasSecurityRequirements(s *RouteSecurity) bool {
	return len(s.Roles) > 0 ||
		len(s.Permissions) > 0 ||
		s.Resource != "" ||
		s.ResourceResolver != nil
}
