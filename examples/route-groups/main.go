// route-groups demonstrates versioned API routing with NewGroup.
// Two groups (/v1, /v2) share common middleware but serve different response shapes,
// illustrating how groups decouple routing concerns from handler logic.
//
// Run:
//
//	go run ./examples/route-groups
//
// Try:
//
//	curl http://localhost:8080/v1/users/1
//	curl http://localhost:8080/v2/users/1
//	curl http://localhost:8080/health
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/talav/zorya"
	"github.com/talav/zorya/adapters"
)

// --- Logging middleware ---

func requestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("[%s] %s %s", r.Method, r.URL.Path, time.Since(start))
	})
}

// --- Version header middleware ---

func versionHeader(version string) zorya.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("API-Version", version)
			next.ServeHTTP(w, r)
		})
	}
}

// --- V1 types (flat structure) ---

type GetUserV1Input struct {
	ID int `schema:"id,location=path"`
}

type GetUserV1Output struct {
	Body struct {
		ID       int    `json:"id"`
		FullName string `json:"full_name"`
	}
}

func getUserV1(_ context.Context, input *GetUserV1Input) (*GetUserV1Output, error) {
	out := &GetUserV1Output{}
	out.Body.ID = input.ID
	out.Body.FullName = fmt.Sprintf("User %d", input.ID)
	return out, nil
}

// --- V2 types (richer structure) ---

type GetUserV2Input struct {
	ID int `schema:"id,location=path"`
}

type GetUserV2Output struct {
	Body struct {
		ID        int    `json:"id"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Version   string `json:"_version"`
	}
}

func getUserV2(_ context.Context, input *GetUserV2Input) (*GetUserV2Output, error) {
	out := &GetUserV2Output{}
	out.Body.ID = input.ID
	out.Body.FirstName = fmt.Sprintf("First%d", input.ID)
	out.Body.LastName = fmt.Sprintf("Last%d", input.ID)
	out.Body.Version = "v2"
	return out, nil
}

// --- Health (no group, no version) ---

type HealthOutput struct {
	Body struct {
		Status string `json:"status"`
	}
}

func health(_ context.Context, _ *struct{}) (*HealthOutput, error) {
	out := &HealthOutput{}
	out.Body.Status = "ok"
	return out, nil
}

// --- Main ---

func main() {
	router := chi.NewMux()
	api := zorya.NewAPI(adapters.NewChi(router), zorya.WithConfig(zorya.DefaultConfig()))

	// API-level middleware (runs for all routes)
	api.UseMiddleware(requestLogger)

	// v1 group
	v1 := zorya.NewGroup(api, "/v1")
	v1.UseMiddleware(versionHeader("1.0"))
	zorya.Get(v1, "/users/{id}", getUserV1)

	// v2 group
	v2 := zorya.NewGroup(api, "/v2")
	v2.UseMiddleware(versionHeader("2.0"))
	zorya.Get(v2, "/users/{id}", getUserV2)

	// No-group route
	zorya.Get(api, "/health", health)

	log.Println("Listening on :8080  —  docs at http://localhost:8080/docs")
	log.Fatal(http.ListenAndServe(":8080", router))
}
