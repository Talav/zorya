//go:build go1.22

package adapters

import (
	"net/http"
	"strings"

	"github.com/talav/zorya"
)

// Mux is an interface for HTTP muxes that support Go 1.22+ routing.
// This includes http.ServeMux and any compatible mux implementation.
type Mux interface {
	HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request))
	ServeHTTP(http.ResponseWriter, *http.Request)
}

// StdlibAdapter implements zorya.Adapter for net/http (Go 1.22+) router.
type StdlibAdapter struct {
	mux    Mux
	prefix string
}

// NewStdlib creates a new adapter for the given HTTP mux (Go 1.22+).
//
//	mux := http.NewServeMux()
//	adapter := adapters.NewStdlib(mux)
//	api := zorya.NewAPI(adapter)
func NewStdlib(mux Mux) *StdlibAdapter {
	return &StdlibAdapter{mux: mux, prefix: ""}
}

// NewStdlibWithPrefix creates a new adapter with a URL prefix.
// This behaves similar to router groups, adding the prefix before each route path.
//
//	mux := http.NewServeMux()
//	adapter := adapters.NewStdlibWithPrefix(mux, "/api")
//	api := zorya.NewAPI(adapter)
func NewStdlibWithPrefix(mux Mux, prefix string) *StdlibAdapter {
	if !strings.HasPrefix(prefix, "/") {
		prefix = "/" + prefix
	}
	if !strings.HasSuffix(prefix, "/") && prefix != "" {
		prefix += "/"
	}

	return &StdlibAdapter{mux: mux, prefix: prefix}
}

func (a *StdlibAdapter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.mux.ServeHTTP(w, r)
}

func (a *StdlibAdapter) Handle(route *zorya.BaseRoute, handler http.HandlerFunc) {
	// Go 1.22+ ServeMux uses "METHOD PATH" pattern format
	pattern := strings.ToUpper(route.Method) + " " + a.prefix + route.Path
	a.mux.HandleFunc(pattern, handler)
}

func (a *StdlibAdapter) ExtractRouterParams(r *http.Request, route *zorya.BaseRoute) map[string]string {
	routerParams := make(map[string]string)
	// Extract parameter names from route pattern
	pathSegments := strings.Split(route.Path, "/")
	for _, segment := range pathSegments {
		if strings.HasPrefix(segment, "{") && strings.HasSuffix(segment, "}") {
			paramName := strings.TrimSuffix(strings.TrimPrefix(segment, "{"), "}")
			// Go 1.22+ PathValue extracts path parameters automatically
			if val := r.PathValue(paramName); val != "" {
				routerParams[paramName] = val
			}
		}
	}

	return routerParams
}
