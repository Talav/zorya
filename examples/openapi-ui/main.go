// openapi-ui demonstrates full OpenAPI 3.1 configuration with Bearer auth,
// multiple servers, operation metadata, and the Stoplight Elements docs UI.
//
// Run:
//
//	go run ./examples/openapi-ui
//
// Then open http://localhost:8080/docs in your browser.
// The raw spec is at http://localhost:8080/openapi.json.
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

type User struct {
	ID    int    `json:"id"    openapi:"readOnly,description=Unique identifier"`
	Name  string `json:"name"  openapi:"description=Full name of the user"`
	Email string `json:"email" openapi:"description=Contact email address,examples=user@example.com"`
}

type GetUserOutput struct {
	Body User
}

type CreateUserInput struct {
	Body struct {
		Name  string `json:"name"  validate:"required" openapi:"description=Full name"`
		Email string `json:"email" validate:"required,email" openapi:"description=Email address"`
	} `body:"structured"`
}

type CreateUserOutput struct {
	Status int `json:"-"`
	Body   User
}

// --- Handlers ---

func getUser(_ context.Context, input *GetUserInput) (*GetUserOutput, error) {
	return &GetUserOutput{Body: User{ID: input.ID, Name: "Alice", Email: "alice@example.com"}}, nil
}

func createUser(_ context.Context, input *CreateUserInput) (*CreateUserOutput, error) {
	return &CreateUserOutput{
		Status: http.StatusCreated,
		Body:   User{ID: 42, Name: input.Body.Name, Email: input.Body.Email},
	}, nil
}

// --- Main ---

func main() {
	router := chi.NewMux()

	// Full OpenAPI metadata
	openAPISpec := &zorya.OpenAPI{
		Info: &zorya.Info{
			Title:       "User API",
			Version:     "1.0.0",
			Description: "A simple API for managing users. Demonstrates full OpenAPI 3.1 metadata.",
		},
		Servers: []*zorya.Server{
			{URL: "http://localhost:8080", Description: "Local development"},
		},
		Components: &zorya.Components{
			SecuritySchemes: map[string]*zorya.SecurityScheme{
				"bearerAuth": {
					Type:        "http",
					Scheme:      "bearer",
					Description: "JWT access token. Obtain from /auth/token.",
				},
			},
		},
	}

	api := zorya.NewAPI(
		adapters.NewChi(router),
		zorya.WithConfig(zorya.DefaultConfig()),
		zorya.WithOpenAPI(openAPISpec),
	)

	zorya.Get(api, "/users/{id}", getUser, func(r *zorya.BaseRoute) {
		r.Operation = &zorya.Operation{
			Summary:     "Get a user",
			Description: "Returns a single user by their numeric ID.",
			Tags:        []string{"Users"},
			OperationID: "getUser",
		}
	})

	zorya.Post(api, "/users", createUser, func(r *zorya.BaseRoute) {
		r.Operation = &zorya.Operation{
			Summary:     "Create a user",
			Description: "Creates a new user record.",
			Tags:        []string{"Users"},
			OperationID: "createUser",
		}
	})

	log.Println("Listening on :8080")
	log.Println("  Docs UI  → http://localhost:8080/docs")
	log.Println("  Spec     → http://localhost:8080/openapi.json")
	log.Fatal(http.ListenAndServe(":8080", router))
}
