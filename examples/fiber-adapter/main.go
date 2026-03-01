// fiber-adapter re-implements the basic CRUD example using the Fiber adapter.
// The handler code is identical — only the adapter and server startup differ,
// demonstrating Zorya's router portability.
//
// Run:
//
//	go run ./examples/fiber-adapter
//
// Try:
//
//	curl -X POST http://localhost:8080/users \
//	     -H 'Content-Type: application/json' \
//	     -d '{"name":"Alice","email":"alice@example.com"}'
//	curl http://localhost:8080/users/1
package main

import (
	"context"
	"log"
	"net/http"
	"sync"
	"sync/atomic"

	"github.com/gofiber/fiber/v2"
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

func nextID() int { return int(counter.Add(1)) }

// --- Input / output types ---

type ListUsersInput struct{}
type ListUsersOutput struct{ Body []User }

type GetUserInput struct {
	ID int `schema:"id,location=path"`
}
type GetUserOutput struct{ Body User }

type CreateUserInput struct {
	Body struct {
		Name  string `json:"name"  validate:"required"`
		Email string `json:"email" validate:"required,email"`
	} `body:"structured"`
}
type CreateUserOutput struct {
	Status int `json:"-"`
	Body   User
}

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
	u := &User{ID: nextID(), Name: input.Body.Name, Email: input.Body.Email}
	mu.Lock()
	store[u.ID] = u
	mu.Unlock()
	return &CreateUserOutput{Status: http.StatusCreated, Body: *u}, nil
}

// --- Main (Fiber) ---

func main() {
	app := fiber.New()

	// The only change from basic-crud: adapters.NewFiber instead of adapters.NewChi
	api := zorya.NewAPI(
		adapters.NewFiber(app),
		zorya.WithConfig(zorya.DefaultConfig()),
	)

	zorya.Get(api, "/users", listUsers)
	zorya.Post(api, "/users", createUser)
	zorya.Get(api, "/users/:id", getUser)

	log.Println("Listening on :8080  —  docs at http://localhost:8080/docs")
	log.Fatal(app.Listen(":8080"))
}
