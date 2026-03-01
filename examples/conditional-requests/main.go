// conditional-requests demonstrates HTTP conditional request support using
// conditional.Params. GET returns 304 Not Modified when the ETag matches.
// PUT returns 412 Precondition Failed when If-Match does not match.
//
// Run:
//
//	go run ./examples/conditional-requests
//
// Try:
//
//	# First GET — returns full response + ETag header
//	curl -i http://localhost:8080/users/1
//
//	# Conditional GET — returns 304 if ETag matches
//	curl -i -H 'If-None-Match: "<etag-from-above>"' http://localhost:8080/users/1
//
//	# Conditional PUT — succeeds if ETag matches
//	curl -i -X PUT http://localhost:8080/users/1 \
//	     -H 'Content-Type: application/json' \
//	     -H 'If-Match: "<etag-from-above>"' \
//	     -d '{"name":"Alicia"}'
//
//	# Conditional PUT — 412 if ETag is stale
//	curl -i -X PUT http://localhost:8080/users/1 \
//	     -H 'Content-Type: application/json' \
//	     -H 'If-Match: "stale"' \
//	     -d '{"name":"Alicia"}'
package main

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/talav/zorya"
	"github.com/talav/zorya/adapters"
	"github.com/talav/zorya/conditional"
)

// --- Domain model ---

type User struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	ETag string `json:"-"`
}

func computeETag(u *User) string {
	b, _ := json.Marshal(u)
	return fmt.Sprintf(`"%x"`, sha256.Sum256(b))
}

// --- In-memory store ---

var (
	mu    sync.RWMutex
	store = map[int]*User{
		1: {ID: 1, Name: "Alice"},
	}
)

func init() {
	store[1].ETag = computeETag(store[1])
}

// --- Input / output types ---

type GetUserInput struct {
	ID int `schema:"id,location=path"`
	conditional.Params
}

type GetUserOutput struct {
	ETag string `schema:"ETag,location=header"`
	Body User
}

type UpdateUserInput struct {
	ID   int `schema:"id,location=path"`
	conditional.Params
	Body struct {
		Name string `json:"name" validate:"required"`
	} `body:"structured"`
}

type UpdateUserOutput struct {
	ETag string `schema:"ETag,location=header"`
	Body User
}

// --- Handlers ---

func getUser(_ context.Context, input *GetUserInput) (*GetUserOutput, error) {
	mu.RLock()
	u, ok := store[input.ID]
	mu.RUnlock()
	if !ok {
		return nil, zorya.Error404NotFound("user not found")
	}

	// Check If-None-Match (read path: isWrite=false → 304 on match)
	// Pass zero time.Time — a real implementation would store the modification time.
	if err := input.CheckPreconditions(u.ETag, time.Time{}, false); err != nil {
		return nil, err
	}

	return &GetUserOutput{ETag: u.ETag, Body: *u}, nil
}

func updateUser(_ context.Context, input *UpdateUserInput) (*UpdateUserOutput, error) {
	mu.Lock()
	defer mu.Unlock()

	u, ok := store[input.ID]
	if !ok {
		return nil, zorya.Error404NotFound("user not found")
	}

	// Check If-Match (write path: isWrite=true → 412 on mismatch)
	if err := input.CheckPreconditions(u.ETag, time.Time{}, true); err != nil {
		return nil, err
	}

	u.Name = input.Body.Name
	u.ETag = computeETag(u)

	return &UpdateUserOutput{ETag: u.ETag, Body: *u}, nil
}

// --- Main ---

func main() {
	router := chi.NewMux()
	api := zorya.NewAPI(adapters.NewChi(router), zorya.WithConfig(zorya.DefaultConfig()))

	zorya.Get(api, "/users/{id}", getUser)
	zorya.Put(api, "/users/{id}", updateUser)

	log.Println("Listening on :8080  —  docs at http://localhost:8080/docs")
	log.Fatal(http.ListenAndServe(":8080", router))
}
