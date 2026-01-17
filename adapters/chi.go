package adapters

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/talav/zorya"
)

// ChiAdapter implements zorya.Adapter for Chi router.
type ChiAdapter struct {
	router chi.Router
}

// NewChi creates a new adapter for the given chi router.
func NewChi(r chi.Router) *ChiAdapter {
	return &ChiAdapter{router: r}
}

func (a *ChiAdapter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.router.ServeHTTP(w, r)
}

func (a *ChiAdapter) Handle(route *zorya.BaseRoute, handler http.HandlerFunc) {
	a.router.MethodFunc(route.Method, route.Path, handler)
}

func (a *ChiAdapter) ExtractRouterParams(r *http.Request, route *zorya.BaseRoute) map[string]string {
	routerParams := make(map[string]string)
	chiCtx := chi.RouteContext(r.Context())
	if chiCtx != nil {
		for i, key := range chiCtx.URLParams.Keys {
			if i < len(chiCtx.URLParams.Values) {
				routerParams[key] = chiCtx.URLParams.Values[i]
			}
		}
	}

	return routerParams
}
