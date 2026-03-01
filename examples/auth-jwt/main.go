// auth-jwt demonstrates declarative role-based security using Zorya's Secure() helper.
// A simple bearer-token middleware reads the token from the Authorization header,
// looks up the caller's roles, and Zorya's GetRouteSecurityContext is used to enforce
// the required roles on each route.
//
// Run:
//
//	go run ./examples/auth-jwt
//
// Try:
//
//	# Public route — no token required
//	curl http://localhost:8080/status
//
//	# Protected route — requires "viewer" role
//	curl -H "Authorization: Bearer alice-token" http://localhost:8080/users
//
//	# Admin-only route — requires "admin" role
//	curl -H "Authorization: Bearer bob-token" http://localhost:8080/admin/users
//
//	# Forbidden — alice is not an admin
//	curl -H "Authorization: Bearer alice-token" http://localhost:8080/admin/users
package main

import (
	"context"
	"log"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/talav/zorya"
	"github.com/talav/zorya/adapters"
)

// --- Fake token → roles lookup (replace with real JWT validation) ---

var tokenRoles = map[string][]string{
	"alice-token": {"viewer"},
	"bob-token":   {"viewer", "admin"},
}

// --- Auth middleware ---

func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Resolve security requirements for this route (nil = public)
		sec := zorya.GetRouteSecurityContext(r)
		if sec == nil {
			next.ServeHTTP(w, r)
			return
		}

		// Extract bearer token
		authHeader := r.Header.Get("Authorization")
		token, ok := strings.CutPrefix(authHeader, "Bearer ")
		if !ok || token == "" {
			http.Error(w, `{"title":"Unauthorized","status":401}`, http.StatusUnauthorized)
			return
		}

		// Look up caller roles
		callerRoles, known := tokenRoles[token]
		if !known {
			http.Error(w, `{"title":"Unauthorized","status":401}`, http.StatusUnauthorized)
			return
		}

		// Check required roles (any match grants access)
		if !hasAnyRole(callerRoles, sec.Roles) {
			http.Error(w, `{"title":"Forbidden","status":403}`, http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func hasAnyRole(caller, required []string) bool {
	set := make(map[string]struct{}, len(caller))
	for _, r := range caller {
		set[r] = struct{}{}
	}
	for _, r := range required {
		if _, ok := set[r]; ok {
			return true
		}
	}
	return false
}

// --- Input / output types ---

type StatusOutput struct {
	Body struct {
		Status string `json:"status"`
	}
}

type ListUsersOutput struct {
	Body struct {
		Users []string `json:"users"`
	}
}

type AdminOutput struct {
	Body struct {
		Message string `json:"message"`
	}
}

// --- Handlers ---

func getStatus(_ context.Context, _ *struct{}) (*StatusOutput, error) {
	out := &StatusOutput{}
	out.Body.Status = "ok"
	return out, nil
}

func listUsers(_ context.Context, _ *struct{}) (*ListUsersOutput, error) {
	out := &ListUsersOutput{}
	out.Body.Users = []string{"alice", "bob"}
	return out, nil
}

func adminListUsers(_ context.Context, _ *struct{}) (*AdminOutput, error) {
	out := &AdminOutput{}
	out.Body.Message = "admin: all users visible"
	return out, nil
}

// --- Main ---

func main() {
	router := chi.NewMux()
	api := zorya.NewAPI(
		adapters.NewChi(router),
		zorya.WithConfig(zorya.DefaultConfig()),
	)

	// Register the auth middleware globally
	api.UseMiddleware(authMiddleware)

	// Public route
	zorya.Get(api, "/status", getStatus)

	// Viewer route
	zorya.Get(api, "/users", listUsers,
		zorya.Secure(zorya.Roles("viewer")),
	)

	// Admin route
	zorya.Get(api, "/admin/users", adminListUsers,
		zorya.Secure(zorya.Roles("admin")),
	)

	log.Println("Listening on :8080  —  docs at http://localhost:8080/docs")
	log.Fatal(http.ListenAndServe(":8080", router))
}
