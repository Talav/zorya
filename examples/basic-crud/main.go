// basic-crud demonstrates a complete CRUD API for a User resource using the Chi adapter.
// It shows the full input/output struct pattern, error helpers, and route registration.
//
// Run:
//
//	go run ./examples/basic-crud
//
// Try:
//
//	curl -X POST http://localhost:8080/users -H 'Content-Type: application/json' \
//	     -d '{"name":"Alice","email":"alice@example.com"}'
//	curl http://localhost:8080/users
//	curl http://localhost:8080/users/1
//	curl -X PUT http://localhost:8080/users/1 -H 'Content-Type: application/json' \
//	     -d '{"name":"Alicia"}'
//	curl -X DELETE http://localhost:8080/users/1
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"
	"sync/atomic"

	"github.com/go-chi/chi/v5"
	"github.com/talav/zorya"
	"github.com/talav/zorya/adapters"
)

// --- Domain model ---

type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// --- In-memory store ---

var (
	mu      sync.RWMutex
	store   = map[int]*User{}
	counter atomic.Int32
)

func nextID() int {
	return int(counter.Add(1))
}

// --- Input / output types ---

type ListUsersInput struct{}

type ListUsersOutput struct {
	Body []User
}

type GetUserInput struct {
	ID int `schema:"id,location=path"`
}

type GetUserOutput struct {
	Body User
}

type CreateUserInput struct {
	Body struct {
		Name  string `json:"name"  validate:"required"`
		Email string `json:"email" validate:"required,email"`
	} `body:"structured"`
}

type CreateUserOutput struct {
	Status   int    `json:"-"`
	Location string `schema:"Location,location=header"`
	Body     User
}

type UpdateUserInput struct {
	ID   int `schema:"id,location=path"`
	Body struct {
		Name  string `json:"name"  validate:"required"`
		Email string `json:"email" validate:"omitempty,email"`
	} `body:"structured"`
}

type UpdateUserOutput struct {
	Body User
}

type DeleteUserInput struct {
	ID int `schema:"id,location=path"`
}

type DeleteUserOutput struct{}

// --- Handlers ---

func listUsers(_ context.Context, _ *ListUsersInput) (*ListUsersOutput, error) {
	mu.RLock()
	defer mu.RUnlock()

	users := make([]User, 0, len(store))
	for _, u := range store {
		users = append(users, *u)
	}
	return &ListUsersOutput{Body: users}, nil
}

func getUser(_ context.Context, input *GetUserInput) (*GetUserOutput, error) {
	mu.RLock()
	defer mu.RUnlock()

	u, ok := store[input.ID]
	if !ok {
		return nil, zorya.Error404NotFound("user not found")
	}
	return &GetUserOutput{Body: *u}, nil
}

func createUser(_ context.Context, input *CreateUserInput) (*CreateUserOutput, error) {
	u := &User{
		ID:    nextID(),
		Name:  input.Body.Name,
		Email: input.Body.Email,
	}

	mu.Lock()
	store[u.ID] = u
	mu.Unlock()

	return &CreateUserOutput{
		Status:   http.StatusCreated,
		Location: "/users/" + fmt.Sprint(u.ID),
		Body:     *u,
	}, nil
}

func updateUser(_ context.Context, input *UpdateUserInput) (*UpdateUserOutput, error) {
	mu.Lock()
	defer mu.Unlock()

	u, ok := store[input.ID]
	if !ok {
		return nil, zorya.Error404NotFound("user not found")
	}

	u.Name = input.Body.Name
	if input.Body.Email != "" {
		u.Email = input.Body.Email
	}
	return &UpdateUserOutput{Body: *u}, nil
}

func deleteUser(_ context.Context, input *DeleteUserInput) (*DeleteUserOutput, error) {
	mu.Lock()
	defer mu.Unlock()

	if _, ok := store[input.ID]; !ok {
		return nil, zorya.Error404NotFound("user not found")
	}
	delete(store, input.ID)
	return &DeleteUserOutput{}, nil
}

// --- Main ---

func main() {
	router := chi.NewMux()
	api := zorya.NewAPI(
		adapters.NewChi(router),
		zorya.WithConfig(zorya.DefaultConfig()),
	)

	zorya.Get(api, "/users", listUsers)
	zorya.Post(api, "/users", createUser)
	zorya.Get(api, "/users/{id}", getUser)
	zorya.Put(api, "/users/{id}", updateUser)
	zorya.Delete(api, "/users/{id}", deleteUser)

	log.Println("Listening on :8080  —  docs at http://localhost:8080/docs")
	log.Fatal(http.ListenAndServe(":8080", router))
}
