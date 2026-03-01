// validation demonstrates a custom Validator wrapping go-playground/validator.
// Validation errors are returned as structured ErrorDetail entries in the RFC 9457 response body.
//
// Run:
//
//	go run ./examples/validation
//
// Try:
//
//	curl -X POST http://localhost:8080/validate -H 'Content-Type: application/json' -d '{}'
//	curl -X POST http://localhost:8080/validate -H 'Content-Type: application/json' -d '{"name":"x","email":"bad"}'
//	curl -X POST http://localhost:8080/validate -H 'Content-Type: application/json' -d '{"name":"Alice","email":"alice@example.com"}'
package main

import (
	"context"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/talav/zorya"
	"github.com/talav/zorya/adapters"
)

// --- Input / output types ---

type ValidateInput struct {
	Body struct {
		Name  string `json:"name"  validate:"required,min=2,max=100"`
		Email string `json:"email" validate:"required,email"`
	} `body:"structured"`
}

type ValidateOutput struct {
	Body struct {
		OK bool `json:"ok"`
	} `body:"structured"`
}

// --- Handler ---

func validateHandler(_ context.Context, _ *ValidateInput) (*ValidateOutput, error) {
	return &ValidateOutput{
		Body: struct {
			OK bool `json:"ok"`
		}{OK: true},
	}, nil
}

// --- Main ---

func main() {
	router := chi.NewMux()
	v := validator.New()
	api := zorya.NewAPI(
		adapters.NewChi(router),
		zorya.WithValidator(zorya.NewPlaygroundValidator(v)),
	)

	zorya.Post(api, "/validate", validateHandler)

	log.Println("Listening on :8080  —  POST /validate with JSON body (name, email) to see validation")
	log.Fatal(http.ListenAndServe(":8080", router))
}
