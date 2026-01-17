package adapters

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/talav/zorya"
)

type contextKey string

const routerParamsKey contextKey = "zorya.routerParams"

// FiberAdapter implements zorya.Adapter for Fiber router.
type FiberAdapter struct {
	app *fiber.App
}

// NewFiber creates a new adapter for the given Fiber app.
func NewFiber(app *fiber.App) *FiberAdapter {
	return &FiberAdapter{app: app}
}

func (a *FiberAdapter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Use Fiber's Test method to handle http.Request
	resp, err := a.app.Test(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}
	defer func() { _ = resp.Body.Close() }()

	// Copy headers
	for k, v := range resp.Header {
		for _, val := range v {
			w.Header().Add(k, val)
		}
	}

	w.WriteHeader(resp.StatusCode)
	_, _ = io.Copy(w, resp.Body)
}

func (a *FiberAdapter) Handle(route *zorya.BaseRoute, handler http.HandlerFunc) {
	// Convert {param} to :param for Fiber
	path := route.Path
	path = strings.ReplaceAll(path, "{", ":")
	path = strings.ReplaceAll(path, "}", "")

	a.app.Add(route.Method, path, func(c *fiber.Ctx) error {
		// Extract path parameters
		routerParams := make(map[string]string)
		if c.Route() != nil {
			for _, param := range c.Route().Params {
				routerParams[param] = c.Params(param)
			}
		}

		// Convert fiber.Ctx to http.Request/ResponseWriter
		r := c.Request()
		w := &fiberResponseWriter{ctx: c}

		// Create http.Request from fiber request
		req, _ := http.NewRequestWithContext(
			c.UserContext(),
			string(r.Header.Method()),
			c.OriginalURL(),
			bytes.NewReader(c.BodyRaw()),
		)
		// Copy headers
		r.Header.VisitAll(func(key, value []byte) {
			req.Header.Set(string(key), string(value))
		})

		// Store router params in request context for ExtractRouterParams
		ctx := context.WithValue(req.Context(), routerParamsKey, routerParams)
		req = req.WithContext(ctx)

		// Call the standard http.HandlerFunc (with middleware already applied)
		handler(w, req)

		return nil
	})
}

func (a *FiberAdapter) ExtractRouterParams(r *http.Request, route *zorya.BaseRoute) map[string]string {
	// For Fiber, params are extracted in Handle from fiber.Ctx
	// If called from Register, try to get from context (stored by Handle)
	if params, ok := r.Context().Value(routerParamsKey).(map[string]string); ok {
		return params
	}

	return make(map[string]string)
}

type fiberResponseWriter struct {
	ctx *fiber.Ctx
}

func (w *fiberResponseWriter) Header() http.Header {
	h := make(http.Header)
	w.ctx.Response().Header.VisitAll(func(key, value []byte) {
		h.Add(string(key), string(value))
	})

	return h
}

func (w *fiberResponseWriter) Write(data []byte) (int, error) {
	return w.ctx.Write(data)
}

func (w *fiberResponseWriter) WriteHeader(statusCode int) {
	w.ctx.Status(statusCode)
}
