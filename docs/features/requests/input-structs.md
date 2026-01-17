# Input Structs

Input structs define how requests are parsed into typed Go structures. Zorya uses struct tags to automatically extract data from path parameters, query strings, headers, cookies, and request bodies.

## Overview

An input struct can contain:
- **Path parameters** - from URL path (e.g., `/users/{id}`)
- **Query parameters** - from URL query string (e.g., `?page=1&limit=10`)
- **Headers** - from HTTP request headers
- **Cookies** - from HTTP cookies
- **Body** - from request body (JSON, CBOR, multipart, etc.)

## Basic Example

```go
type GetUserInput struct {
    // Path parameter
    ID string `schema:"id,location=path"`
    
    // Query parameters
    Format string `schema:"format,location=query"`
    Page   int    `schema:"page,location=query"`
    
    // Header parameters
    APIKey string `schema:"X-API-Key,location=header"`
    
    // Cookie parameters
    SessionID string `schema:"session_id,location=cookie"`
    
    // Request body
    Body struct {
        Name  string `json:"name"`
        Email string `json:"email"`
    } `body:"structured"`
}
```

## Path Parameters

Path parameters are extracted from the URL path based on your router's pattern syntax.

### Chi and Stdlib (Go 1.22+)

```go
type GetUserInput struct {
    UserID string `schema:"id,location=path"`
}

// Route definition
zorya.Get(api, "/users/{id}", getUserHandler)

// Request: GET /users/123
// input.UserID = "123"
```

### Fiber

```go
type GetUserInput struct {
    UserID string `schema:"id,location=path"`
}

// Route definition (note :id syntax)
zorya.Get(api, "/users/:id", getUserHandler)

// Request: GET /users/123
// input.UserID = "123"
```

### Multiple Path Parameters

```go
type GetProjectInput struct {
    OrgID     string `schema:"orgId,location=path"`
    ProjectID string `schema:"projectId,location=path"`
}

// Route: /orgs/{orgId}/projects/{projectId}
zorya.Get(api, "/orgs/{orgId}/projects/{projectId}", getProjectHandler)

// Request: GET /orgs/acme/projects/website
// input.OrgID = "acme"
// input.ProjectID = "website"
```

## Query Parameters

Query parameters are parsed from the URL query string.

### Basic Query Parameters

```go
type ListUsersInput struct {
    Page     int    `schema:"page,location=query"`
    PageSize int    `schema:"page_size,location=query"`
    Search   string `schema:"search,location=query"`
}

// Request: GET /users?page=2&page_size=20&search=john
// input.Page = 2
// input.PageSize = 20
// input.Search = "john"
```

### Array Query Parameters

```go
type FilterInput struct {
    Tags   []string `schema:"tags,location=query"`
    Status []string `schema:"status,location=query"`
}

// Request: GET /posts?tags=golang&tags=api&status=published
// input.Tags = ["golang", "api"]
// input.Status = ["published"]
```

### Nested Structs

```go
type ListUsersInput struct {
    Pagination struct {
        Page     int `schema:"page,location=query"`
        PageSize int `schema:"page_size,location=query"`
    }
    Filters struct {
        Status string   `schema:"status,location=query"`
        Tags   []string `schema:"tags,location=query"`
    }
}
```

## Header Parameters

Headers are case-insensitive and can be accessed using the schema tag.

### Standard Headers

```go
type RequestInput struct {
    ContentType   string `schema:"Content-Type,location=header"`
    Authorization string `schema:"Authorization,location=header"`
    UserAgent     string `schema:"User-Agent,location=header"`
}

// Request Headers:
// Content-Type: application/json
// Authorization: Bearer token123
// User-Agent: MyApp/1.0
```

### Custom Headers

```go
type RequestInput struct {
    APIKey    string `schema:"X-API-Key,location=header"`
    RequestID string `schema:"X-Request-ID,location=header"`
    Version   string `schema:"X-API-Version,location=header"`
}

// Request Headers:
// X-API-Key: secret123
// X-Request-ID: req-456
// X-API-Version: 2.0
```

### Array Headers

```go
type RequestInput struct {
    AcceptLanguage []string `schema:"Accept-Language,location=header"`
}

// Request Header:
// Accept-Language: en-US,en;q=0.9,fr;q=0.8
// input.AcceptLanguage = ["en-US", "en", "fr"]
```

## Cookie Parameters

Cookies are extracted using the schema tag.

```go
type RequestInput struct {
    SessionID string `schema:"session_id,location=cookie"`
    Theme     string `schema:"theme,location=cookie"`
    UserPref  string `schema:"user_pref,location=cookie"`
}

// Request Cookies:
// session_id=abc123; theme=dark; user_pref=compact
// input.SessionID = "abc123"
// input.Theme = "dark"
// input.UserPref = "compact"
```

## Request Body

The request body can be structured (JSON, CBOR) or multipart (file uploads).

### Structured Body (JSON)

```go
type CreateUserInput struct {
    Body struct {
        Name     string `json:"name"`
        Email    string `json:"email"`
        Age      int    `json:"age"`
        IsActive bool   `json:"is_active"`
    } `body:"structured"`
}

// Request Body (JSON):
// {
//   "name": "John Doe",
//   "email": "john@example.com",
//   "age": 30,
//   "is_active": true
// }
```

### Nested Body Structures

```go
type CreateOrderInput struct {
    Body struct {
        Customer struct {
            Name    string `json:"name"`
            Email   string `json:"email"`
            Address struct {
                Street  string `json:"street"`
                City    string `json:"city"`
                ZipCode string `json:"zip_code"`
            } `json:"address"`
        } `json:"customer"`
        Items []struct {
            ProductID string  `json:"product_id"`
            Quantity  int     `json:"quantity"`
            Price     float64 `json:"price"`
        } `json:"items"`
        Total float64 `json:"total"`
    } `body:"structured"`
}
```

### Multipart Body (File Uploads)

```go
type UploadFileInput struct {
    Body struct {
        File        []byte `json:"file" openapi:"format=binary"`
        Filename    string `json:"filename"`
        Description string `json:"description"`
    } `body:"multipart"`
}
```

See [File Uploads](file-uploads.md) for detailed information.

## Required Fields

Mark fields as required using the `schema` tag:

```go
type GetUserInput struct {
    // Required path parameter
    ID string `schema:"id,location=path,required=true"`
    
    // Optional query parameter
    Format string `schema:"format,location=query"`
    
    // Required header
    APIKey string `schema:"X-API-Key,location=header,required=true"`
}
```

**Note:** Required validation is enforced before calling your handler. Missing required fields return a 422 error.

## Type Conversions

Zorya automatically converts string values to appropriate Go types:

```go
type Input struct {
    // String to int
    Page int `schema:"page,location=query"`
    
    // String to float
    Price float64 `schema:"price,location=query"`
    
    // String to bool
    Active bool `schema:"active,location=query"`
    
    // String to time.Time
    CreatedAfter time.Time `schema:"created_after,location=query"`
    
    // String to slice
    Tags []string `schema:"tags,location=query"`
}

// Request: GET /items?page=5&price=19.99&active=true&tags=new&tags=sale
// input.Page = 5
// input.Price = 19.99
// input.Active = true
// input.Tags = ["new", "sale"]
```

## Complete Example

```go
package main

import (
    "context"
    "net/http"
    
    "github.com/go-chi/chi/v5"
    "github.com/talav/talav/pkg/component/zorya"
    "github.com/talav/talav/pkg/component/zorya/adapters"
)

// Input struct with all parameter types
type SearchProductsInput struct {
    // Path parameters
    CategoryID string `schema:"category_id,location=path,required=true"`
    
    // Query parameters
    Query    string   `schema:"q,location=query"`
    Page     int      `schema:"page,location=query" default:"1"`
    PageSize int      `schema:"page_size,location=query" default:"20"`
    Sort     string   `schema:"sort,location=query" default:"relevance"`
    Tags     []string `schema:"tags,location=query"`
    MinPrice float64  `schema:"min_price,location=query"`
    MaxPrice float64  `schema:"max_price,location=query"`
    
    // Headers
    APIKey       string `schema:"X-API-Key,location=header,required=true"`
    UserAgent    string `schema:"User-Agent,location=header"`
    
    // Cookies
    SessionID string `schema:"session_id,location=cookie"`
}

type SearchProductsOutput struct {
    Body struct {
        Results []struct {
            ID    string  `json:"id"`
            Name  string  `json:"name"`
            Price float64 `json:"price"`
        } `json:"results"`
        TotalCount int `json:"total_count"`
        Page       int `json:"page"`
        PageSize   int `json:"page_size"`
    } `body:"structured"`
}

func main() {
    router := chi.NewMux()
    adapter := adapters.NewChi(router)
    api := zorya.NewAPI(adapter)
    
    zorya.Get(api, "/categories/{category_id}/products", searchProducts)
    
    http.ListenAndServe(":8080", router)
}

func searchProducts(ctx context.Context, input *SearchProductsInput) (*SearchProductsOutput, error) {
    // Access all parsed parameters
    _ = input.CategoryID  // From path
    _ = input.Query       // From query
    _ = input.Page        // From query with default
    _ = input.Tags        // From query array
    _ = input.APIKey      // From header
    _ = input.SessionID   // From cookie
    
    // Your business logic here
    return &SearchProductsOutput{
        Body: struct {
            Results []struct {
                ID    string  `json:"id"`
                Name  string  `json:"name"`
                Price float64 `json:"price"`
            } `json:"results"`
            TotalCount int `json:"total_count"`
            Page       int `json:"page"`
            PageSize   int `json:"page_size"`
        }{
            Results:    []struct{...}{},
            TotalCount: 100,
            Page:       input.Page,
            PageSize:   input.PageSize,
        },
    }, nil
}
```

## See Also

- [Schema Package](../../packages/schema.md) - Detailed schema tag documentation
- [Validation](validation.md) - Input validation
- [File Uploads](file-uploads.md) - Multipart file uploads
- [Default Values](../defaults.md) - Default parameter values
- [Metadata Tags Reference](../metadata/tags-reference.md) - All struct tags
