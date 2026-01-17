# Documentation UI

Zorya provides built-in interactive API documentation using Stoplight Elements, automatically generated from your OpenAPI specification.

## Overview

The documentation UI features:

- **Interactive API Explorer** - Try API endpoints directly from the browser
- **Automatic Generation** - Documentation stays in sync with code
- **Stoplight Elements** - Modern, responsive UI with dark mode support
- **Zero Configuration** - Works out of the box with sensible defaults

## Quick Start

### Enable Documentation

```go
package main

import (
    "github.com/go-chi/chi/v5"
    "github.com/talav/talav/pkg/component/zorya"
    "github.com/talav/talav/pkg/component/zorya/adapters"
)

func main() {
    router := chi.NewMux()
    adapter := adapters.NewChi(router)
    
    // Create API with documentation enabled
    api := zorya.NewAPI(adapter,
        zorya.WithConfig(zorya.Config{
            Title:       "My API",
            Version:     "1.0.0",
            Description: "API for managing users and resources",
            DocsPath:    "/docs",        // Documentation UI path
            OpenAPIPath: "/openapi",     // OpenAPI spec path
        }),
    )
    
    // Register routes...
    zorya.Get(api, "/users", listUsers)
    zorya.Post(api, "/users", createUser)
    
    http.ListenAndServe(":8080", router)
}
```

### Access Documentation

Once your server is running, visit:

- **Documentation UI**: `http://localhost:8080/docs`
- **OpenAPI Spec (JSON)**: `http://localhost:8080/openapi.json`
- **OpenAPI Spec (YAML)**: `http://localhost:8080/openapi.yaml`

## Configuration

### Basic Configuration

```go
api := zorya.NewAPI(adapter,
    zorya.WithConfig(zorya.Config{
        // API metadata
        Title:       "My API",
        Version:     "1.0.0",
        Description: "Comprehensive API documentation",
        
        // Documentation paths
        DocsPath:    "/docs",      // UI path
        OpenAPIPath: "/openapi",   // Spec path
        
        // Optional: Custom paths
        // DocsPath: "/api-docs"
        // OpenAPIPath: "/api/schema"
    }),
)
```

### Advanced Configuration

```go
api := zorya.NewAPI(adapter,
    zorya.WithConfig(zorya.Config{
        Title:   "My API",
        Version: "2.0.0",
        Description: `
# My API

This API provides access to user management and resource operations.

## Authentication

All endpoints except /health require authentication via Bearer token.

## Rate Limiting

Requests are limited to 100 per minute per IP address.
        `,
        
        // Contact information
        Contact: &zorya.Contact{
            Name:  "API Support",
            Email: "support@example.com",
            URL:   "https://example.com/support",
        },
        
        // License
        License: &zorya.License{
            Name: "Apache 2.0",
            URL:  "https://www.apache.org/licenses/LICENSE-2.0.html",
        },
        
        // Terms of service
        TermsOfService: "https://example.com/terms",
        
        // Servers
        Servers: []zorya.Server{
            {
                URL:         "https://api.example.com",
                Description: "Production server",
            },
            {
                URL:         "https://staging-api.example.com",
                Description: "Staging server",
            },
            {
                URL:         "http://localhost:8080",
                Description: "Development server",
            },
        },
        
        // External docs
        ExternalDocs: &zorya.ExternalDocs{
            Description: "Full API documentation",
            URL:         "https://docs.example.com",
        },
        
        // Documentation paths
        DocsPath:    "/docs",
        OpenAPIPath: "/openapi",
    }),
)
```

### Disable Documentation

To disable the documentation UI in production:

```go
api := zorya.NewAPI(adapter,
    zorya.WithConfig(zorya.Config{
        Title:       "My API",
        Version:     "1.0.0",
        DocsPath:    "",  // Disable UI
        OpenAPIPath: "/openapi",  // Keep spec available
    }),
)
```

Or disable both:

```go
api := zorya.NewAPI(adapter,
    zorya.WithConfig(zorya.Config{
        Title:       "My API",
        Version:     "1.0.0",
        DocsPath:    "",  // Disable UI
        OpenAPIPath: "",  // Disable spec endpoint
    }),
)
```

## UI Features

### Interactive Try-It

The documentation UI allows testing endpoints directly:

1. Navigate to an endpoint
2. Fill in parameters
3. Click "Send API Request"
4. View the response

### Authentication

Configure security schemes in your API:

```go
// Add JWT security to API
api.AddSecurityScheme("bearerAuth", zorya.SecurityScheme{
    Type:         "http",
    Scheme:       "bearer",
    BearerFormat: "JWT",
    Description:  "JWT access token",
})

// Protected route
zorya.Get(api, "/users", listUsers,
    zorya.Secure(
        zorya.Auth(),
    ),
)
```

In the UI, click the "Auth" button to configure the token.

### Request Examples

Zorya automatically generates request examples from your struct definitions:

```go
type CreateUserInput struct {
    Body struct {
        Name  string `json:"name" validate:"required" openapi:"example=John Doe"`
        Email string `json:"email" validate:"required,email" openapi:"example=john@example.com"`
        Age   int    `json:"age" validate:"min=0" openapi:"example=30"`
    } `body:"structured"`
}
```

### Response Examples

```go
type CreateUserOutput struct {
    Status int `status:"201"`
    Body struct {
        ID        int64     `json:"id" openapi:"example=123"`
        Name      string    `json:"name" openapi:"example=John Doe"`
        Email     string    `json:"email" openapi:"example=john@example.com"`
        CreatedAt time.Time `json:"created_at" openapi:"example=2024-01-15T10:30:00Z"`
    } `body:"structured"`
}
```

## Customization

### Custom UI Theme

Since Zorya uses Stoplight Elements, you can customize the UI by serving your own HTML:

```go
// Disable built-in docs
api := zorya.NewAPI(adapter,
    zorya.WithConfig(zorya.Config{
        OpenAPIPath: "/openapi",
        DocsPath:    "",  // Disable built-in
    }),
)

// Serve custom UI
router.Get("/docs", func(w http.ResponseWriter, r *http.Request) {
    html := `
<!DOCTYPE html>
<html>
<head>
    <title>My Custom API Docs</title>
    <link rel="stylesheet" href="https://unpkg.com/@stoplight/elements/styles.min.css">
    <style>
        /* Custom styles */
        body { font-family: 'Inter', sans-serif; }
    </style>
</head>
<body>
    <elements-api
        apiDescriptionUrl="/openapi.json"
        router="hash"
        layout="sidebar"
        logo="https://example.com/logo.png"
    />
    <script src="https://unpkg.com/@stoplight/elements/web-components.min.js"></script>
</body>
</html>
    `
    w.Header().Set("Content-Type", "text/html")
    w.Write([]byte(html))
})
```

### Alternative UI

You can use any OpenAPI UI tool:

#### Swagger UI

```go
// Serve OpenAPI spec
api := zorya.NewAPI(adapter,
    zorya.WithConfig(zorya.Config{
        OpenAPIPath: "/openapi",
        DocsPath:    "",
    }),
)

// Serve Swagger UI
router.Get("/docs", func(w http.ResponseWriter, r *http.Request) {
    html := `
<!DOCTYPE html>
<html>
<head>
    <title>API Documentation</title>
    <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist/swagger-ui.css">
</head>
<body>
    <div id="swagger-ui"></div>
    <script src="https://unpkg.com/swagger-ui-dist/swagger-ui-bundle.js"></script>
    <script>
        SwaggerUIBundle({
            url: '/openapi.json',
            dom_id: '#swagger-ui',
        })
    </script>
</body>
</html>
    `
    w.Header().Set("Content-Type", "text/html")
    w.Write([]byte(html))
})
```

#### ReDoc

```go
router.Get("/docs", func(w http.ResponseWriter, r *http.Request) {
    html := `
<!DOCTYPE html>
<html>
<head>
    <title>API Documentation</title>
</head>
<body>
    <redoc spec-url="/openapi.json"></redoc>
    <script src="https://cdn.jsdelivr.net/npm/redoc/bundles/redoc.standalone.js"></script>
</body>
</html>
    `
    w.Header().Set("Content-Type", "text/html")
    w.Write([]byte(html))
})
```

## Production Considerations

### Security

In production, consider:

1. **Disable docs UI** - Keep only OpenAPI spec for client generation
2. **Authentication** - Protect docs with middleware
3. **CORS** - Configure properly for docs access

```go
// Production: Docs behind auth
if os.Getenv("ENV") == "production" {
    api := zorya.NewAPI(adapter,
        zorya.WithConfig(zorya.Config{
            OpenAPIPath: "/openapi",
            DocsPath:    "",  // Disable public docs
        }),
    )
} else {
    api := zorya.NewAPI(adapter,
        zorya.WithConfig(zorya.Config{
            OpenAPIPath: "/openapi",
            DocsPath:    "/docs",
        }),
    )
}
```

### Performance

The OpenAPI spec is generated once at startup and cached. No performance impact on request handling.

### CDN

For production, consider self-hosting Stoplight Elements instead of CDN:

```bash
npm install @stoplight/elements
# Copy assets to static directory
```

Then reference local files in custom HTML.

## Complete Example

```go
package main

import (
    "net/http"
    "os"
    
    "github.com/go-chi/chi/v5"
    "github.com/talav/talav/pkg/component/zorya"
    "github.com/talav/talav/pkg/component/zorya/adapters"
)

func main() {
    router := chi.NewMux()
    adapter := adapters.NewChi(router)
    
    // Configure based on environment
    docsPath := "/docs"
    if os.Getenv("ENV") == "production" {
        docsPath = "" // Disable in production
    }
    
    api := zorya.NewAPI(adapter,
        zorya.WithConfig(zorya.Config{
            Title:   "User Management API",
            Version: "1.0.0",
            Description: `
# User Management API

RESTful API for managing users and their profiles.

## Features

- User CRUD operations
- Profile management
- Role-based access control

## Authentication

All protected endpoints require a valid JWT token in the Authorization header:

\`\`\`
Authorization: Bearer <token>
\`\`\`
            `,
            Contact: &zorya.Contact{
                Name:  "API Team",
                Email: "api@example.com",
            },
            License: &zorya.License{
                Name: "MIT",
                URL:  "https://opensource.org/licenses/MIT",
            },
            Servers: []zorya.Server{
                {
                    URL:         "https://api.example.com",
                    Description: "Production",
                },
                {
                    URL:         "http://localhost:8080",
                    Description: "Development",
                },
            },
            DocsPath:    docsPath,
            OpenAPIPath: "/openapi",
        }),
    )
    
    // Add security scheme
    api.AddSecurityScheme("bearerAuth", zorya.SecurityScheme{
        Type:         "http",
        Scheme:       "bearer",
        BearerFormat: "JWT",
    })
    
    // Register routes
    zorya.Get(api, "/users", listUsers)
    zorya.Post(api, "/users", createUser,
        zorya.Secure(zorya.Auth()),
    )
    zorya.Get(api, "/users/{id}", getUser)
    zorya.Put(api, "/users/{id}", updateUser,
        zorya.Secure(zorya.Auth()),
    )
    zorya.Delete(api, "/users/{id}", deleteUser,
        zorya.Secure(
            zorya.Roles("admin"),
        ),
    )
    
    http.ListenAndServe(":8080", router)
}
```

## See Also

- [Overview](overview.md) - OpenAPI generation overview
- [Schema Generation](schema-generation.md) - How schemas are created
- [Configuration](../../reference/api.md#config) - Full configuration reference
