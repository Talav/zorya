package zorya

import "net/http"

// Middleware is a standard Go middleware function that takes an http.Handler
// and returns an http.Handler.
type Middleware func(http.Handler) http.Handler

// Middlewares is a list of standard middleware functions that can be attached
// to an API and will be called for all incoming requests.
type Middlewares []Middleware

// Apply applies the middleware chain to an http.Handler and returns the result.
func (m Middlewares) Apply(handler http.Handler) http.Handler {
	if len(m) == 0 {
		return handler
	}

	h := handler
	for i := len(m) - 1; i >= 0; i-- {
		h = m[i](h)
	}

	return h
}
