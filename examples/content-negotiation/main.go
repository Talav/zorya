// content-negotiation demonstrates automatic JSON/CBOR format selection based on
// the Accept header. The same handler serves both formats without any extra code.
//
// Run:
//
//	go run ./examples/content-negotiation
//
// Try:
//
//	# Default (JSON)
//	curl http://localhost:8080/users/1
//
//	# Explicit JSON
//	curl -H "Accept: application/json" http://localhost:8080/users/1
//
//	# CBOR (binary — pipe through xxd to inspect)
//	curl -s -H "Accept: application/cbor" http://localhost:8080/users/1 | xxd
//
//	# Unknown type falls back to JSON (NoFormatFallback is false)
//	curl -H "Accept: application/xml" http://localhost:8080/users/1
package main

import (
	"context"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/talav/zorya"
	"github.com/talav/zorya/adapters"
)

// --- Types ---

type GetUserInput struct {
	ID int `schema:"id,location=path"`
}

type GetUserOutput struct {
	Body struct {
		ID    int    `json:"id"`
		Name  string `json:"name"`
		Email string `json:"email"`
	}
}

// --- Handler ---

func getUser(_ context.Context, input *GetUserInput) (*GetUserOutput, error) {
	out := &GetUserOutput{}
	out.Body.ID = input.ID
	out.Body.Name = "Alice"
	out.Body.Email = "alice@example.com"
	return out, nil
}

// --- Main ---

func main() {
	router := chi.NewMux()

	// DefaultFormats() includes both application/json and application/cbor.
	// WithConfig sets NoFormatFallback=false so unknown Accept types fall back to JSON.
	api := zorya.NewAPI(
		adapters.NewChi(router),
		zorya.WithConfig(&zorya.Config{
			OpenAPIPath:      "/openapi.json",
			DocsPath:         "/docs",
			DefaultFormat:    "application/json",
			NoFormatFallback: false, // fall back to JSON for unknown Accept types
		}),
	)

	zorya.Get(api, "/users/{id}", getUser)

	log.Println("Listening on :8080  —  docs at http://localhost:8080/docs")
	log.Println("Try: curl -H 'Accept: application/cbor' http://localhost:8080/users/1 | xxd")
	log.Fatal(http.ListenAndServe(":8080", router))
}
