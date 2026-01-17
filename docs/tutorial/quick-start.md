# Quick Start

Build your first Zorya API in about 5 minutes.

## Prerequisites

- Go 1.22 or newer
- Basic knowledge of Go and HTTP APIs
- A router (we'll use Chi in this tutorial)

## Installation

```bash
# Install Zorya
go get github.com/talav/talav/pkg/component/zorya

# Install Chi router
go get github.com/go-chi/chi/v5
```

## Step 1: Create a Simple API

Create a new file `main.go`:

```go
package main

import (
    "context"
    "net/http"
    
    "github.com/go-chi/chi/v5"
    "github.com/talav/talav/pkg/component/zorya"
    "github.com/talav/talav/pkg/component/zorya/adapters"
)

func main() {
    // 1. Create router
    router := chi.NewMux()
    
    // 2. Create Zorya adapter
    adapter := adapters.NewChi(router)
    
    // 3. Create Zorya API
    api := zorya.NewAPI(adapter)
    
    // 4. Register a route
    zorya.Get(api, "/hello", helloHandler)
    
    // 5. Start server
    http.ListenAndServe(":8080", router)
}

// Input: no parameters needed
type HelloInput struct{}

// Output: simple JSON response
type HelloOutput struct {
    Body struct {
        Message string `json:"message"`
    } `body:"structured"`
}

// Handler: takes Input, returns Output
func helloHandler(ctx context.Context, input *HelloInput) (*HelloOutput, error) {
    return &HelloOutput{
        Body: struct {
            Message string `json:"message"`
        }{
            Message: "Hello, Zorya!",
        },
    }, nil
}
```

Run it:

```bash
go run main.go
```

Test it:

```bash
curl http://localhost:8080/hello
# {"message":"Hello, Zorya!"}
```

That's it! You have a working API.

## Step 2: Add Path Parameters

Let's add a route that greets a user by name:

```go
type GreetInput struct {
    Name string `schema:"name,location=path"`
}

type GreetOutput struct {
    Body struct {
        Message string `json:"message"`
    } `body:"structured"`
}

func greetHandler(ctx context.Context, input *GreetInput) (*GreetOutput, error) {
    return &GreetOutput{
        Body: struct {
            Message string `json:"message"`
        }{
            Message: "Hello, " + input.Name + "!",
        },
    }, nil
}

// In main():
zorya.Get(api, "/greet/{name}", greetHandler)
```

Test it:

```bash
curl http://localhost:8080/greet/John
# {"message":"Hello, John!"}
```

## Step 3: Add Query Parameters

Add pagination to a user list:

```go
type ListUsersInput struct {
    Page     int `schema:"page,location=query" default:"1"`
    PageSize int `schema:"page_size,location=query" default:"10"`
}

type ListUsersOutput struct {
    Body struct {
        Users []struct {
            ID   int    `json:"id"`
            Name string `json:"name"`
        } `json:"users"`
        Page     int `json:"page"`
        PageSize int `json:"page_size"`
    } `body:"structured"`
}

func listUsersHandler(ctx context.Context, input *ListUsersInput) (*ListUsersOutput, error) {
    // Mock data
    users := []struct {
        ID   int    `json:"id"`
        Name string `json:"name"`
    }{
        {ID: 1, Name: "Alice"},
        {ID: 2, Name: "Bob"},
        {ID: 3, Name: "Charlie"},
    }
    
    return &ListUsersOutput{
        Body: struct {
            Users []struct {
                ID   int    `json:"id"`
                Name string `json:"name"`
            } `json:"users"`
            Page     int `json:"page"`
            PageSize int `json:"page_size"`
        }{
            Users:    users,
            Page:     input.Page,
            PageSize: input.PageSize,
        },
    }, nil
}

// In main():
zorya.Get(api, "/users", listUsersHandler)
```

Test it:

```bash
curl http://localhost:8080/users
# {"users":[{"id":1,"name":"Alice"}...],"page":1,"page_size":10}

curl http://localhost:8080/users?page=2&page_size=5
# {"users":[...],"page":2,"page_size":5}
```

## Step 4: Add POST with Request Body

Create a user:

```go
type CreateUserInput struct {
    Body struct {
        Name  string `json:"name" validate:"required"`
        Email string `json:"email" validate:"required,email"`
    } `body:"structured"`
}

type CreateUserOutput struct {
    Status int `status:"201"`
    Body struct {
        ID    int    `json:"id"`
        Name  string `json:"name"`
        Email string `json:"email"`
    } `body:"structured"`
}

func createUserHandler(ctx context.Context, input *CreateUserInput) (*CreateUserOutput, error) {
    // In real app, save to database
    return &CreateUserOutput{
        Status: http.StatusCreated,
        Body: struct {
            ID    int    `json:"id"`
            Name  string `json:"name"`
            Email string `json:"email"`
        }{
            ID:    123,
            Name:  input.Body.Name,
            Email: input.Body.Email,
        },
    }, nil
}

// In main():
zorya.Post(api, "/users", createUserHandler)
```

Test it:

```bash
curl -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{"name":"Jane","email":"jane@example.com"}'
# {"id":123,"name":"Jane","email":"jane@example.com"}
```

## Step 5: Add Validation

Zorya automatically validates your input if you add a validator:

```go
import (
    "github.com/talav/talav/pkg/component/validator"
)

func main() {
    router := chi.NewMux()
    adapter := adapters.NewChi(router)
    
    // Create validator
    validate := validator.New()
    zoryaValidator := zorya.NewPlaygroundValidator(validate)
    
    // Create API with validator
    api := zorya.NewAPI(adapter, zorya.WithValidator(zoryaValidator))
    
    // Register routes
    zorya.Post(api, "/users", createUserHandler)
    
    http.ListenAndServe(":8080", router)
}
```

Now try invalid data:

```bash
curl -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{"name":"Jane","email":"invalid-email"}'

# Response:
# {
#   "status": 422,
#   "title": "Unprocessable Entity",
#   "detail": "validation failed",
#   "errors": [
#     {
#       "code": "email",
#       "message": "Field validation for 'email' failed on the 'email' tag",
#       "location": "body.email"
#     }
#   ]
# }
```

## Step 6: Add Error Handling

Return errors from your handler:

```go
func getUserHandler(ctx context.Context, input *GetUserInput) (*GetUserOutput, error) {
    // Simulate user not found
    if input.ID == "0" {
        return nil, zorya.Error404NotFound("User not found")
    }
    
    // Return user
    return &GetUserOutput{
        Body: struct {
            ID   string `json:"id"`
            Name string `json:"name"`
        }{
            ID:   input.ID,
            Name: "John Doe",
        },
    }, nil
}

// In main():
zorya.Get(api, "/users/{id}", getUserHandler)
```

Test it:

```bash
curl http://localhost:8080/users/0
# {
#   "status": 404,
#   "title": "Not Found",
#   "detail": "User not found"
# }
```

## Complete Example

Here's the complete working example:

```go
package main

import (
    "context"
    "net/http"
    
    "github.com/go-chi/chi/v5"
    "github.com/talav/talav/pkg/component/validator"
    "github.com/talav/talav/pkg/component/zorya"
    "github.com/talav/talav/pkg/component/zorya/adapters"
)

func main() {
    // Setup router
    router := chi.NewMux()
    adapter := adapters.NewChi(router)
    
    // Setup validator
    validate := validator.New()
    zoryaValidator := zorya.NewPlaygroundValidator(validate)
    
    // Create API
    api := zorya.NewAPI(adapter, zorya.WithValidator(zoryaValidator))
    
    // Register routes
    zorya.Get(api, "/hello", helloHandler)
    zorya.Get(api, "/greet/{name}", greetHandler)
    zorya.Get(api, "/users", listUsersHandler)
    zorya.Get(api, "/users/{id}", getUserHandler)
    zorya.Post(api, "/users", createUserHandler)
    
    // Start server
    http.ListenAndServe(":8080", router)
}

// Handlers...
```

## What You Learned

- ✅ Install Zorya and create an API
- ✅ Define input and output structs
- ✅ Parse path parameters
- ✅ Parse query parameters
- ✅ Handle request bodies
- ✅ Add validation
- ✅ Return errors

## Next Steps

- [First API Tutorial](first-api.md) - Build a complete CRUD API
- [Input Structs](../features/requests/input-structs.md) - Learn all about request parsing
- [Validation](../features/requests/validation.md) - Deep dive into validation
- [Error Handling](../features/responses/errors.md) - Master error responses
- [Security](security.md) - Add authentication and authorization
