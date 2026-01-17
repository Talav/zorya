package zorya

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ComplexRequestInput is a comprehensive test struct covering all parameter types and scenarios.
type ComplexRequestInput struct {
	// Path parameters (required)
	UserID   int    `schema:"user_id,location=path,required=true" openapi:"title=User ID,description=Unique identifier for the user"`
	PostSlug string `schema:"post_slug,location=path,required=true,style=simple" openapi:"title=Post Slug,description=URL-friendly identifier for the post"`

	// Query parameters (mix of required and optional)
	Page           int      `schema:"page,location=query,required=true" openapi:"title=Page Number,description=Page number for pagination"`
	PageSize       *int     `schema:"page_size,location=query" openapi:"title=Page Size,description=Number of items per page"`
	Search         string   `schema:"search,location=query,style=form,explode=true" openapi:"title=Search Query,description=Search term to filter results"`
	Tags           []string `schema:"tags,location=query,style=form,explode=true" openapi:"title=Tags,description=Filter by tags"`
	IncludeDeleted bool     `schema:"include_deleted,location=query" openapi:"title=Include Deleted,description=Include soft-deleted items"`

	// Header parameters
	APIKey     string   `schema:"X-API-Key,location=header,required=true" openapi:"title=API Key,description=Authentication API key"`
	RequestID  string   `schema:"X-Request-ID,location=header" openapi:"title=Request ID,description=Unique request identifier for tracing"`
	AcceptLang []string `schema:"Accept-Language,location=header,style=simple,explode=false" openapi:"title=Accept Language,description=Preferred languages"`

	// Cookie parameters
	SessionID string `schema:"session_id,location=cookie,required=true" openapi:"title=Session ID,description=User session identifier"`
	Theme     string `schema:"theme,location=cookie" openapi:"title=Theme,description=UI theme preference"`

	// Request body (nested struct)
	Body ComplexRequestBody `body:"structured,required" openapi:"title=Request Body,description=Main request payload"`
}

// ComplexRequestBody represents a complex nested request body structure.
type ComplexRequestBody struct {
	Title       string    `json:"title" validate:"required,min=1,max=200" openapi:"title=Title,description=Main title of the content"`
	Description *string   `json:"description" openapi:"title=Description,description=Detailed description"`
	Content     string    `json:"content" validate:"required" openapi:"title=Content,description=Main content body"`
	Tags        []string  `json:"tags" openapi:"title=Tags,description=Content tags"`
	Metadata    Metadata  `json:"metadata" openapi:"title=Metadata,description=Additional metadata"`
	Settings    *Settings `json:"settings" openapi:"title=Settings,description=Optional settings"`
	IsPublished bool      `json:"is_published" default:"false" openapi:"title=Is Published,description=Publication status"`
}

// Metadata represents nested metadata structure.
type Metadata struct {
	Author      string   `json:"author" validate:"required" openapi:"title=Author,description=Content author"`
	Category    string   `json:"category" openapi:"title=Category,description=Content category"`
	Keywords    []string `json:"keywords" openapi:"title=Keywords,description=SEO keywords"`
	PublishedAt *string  `json:"published_at" openapi:"title=Published At,description=Publication timestamp"`
}

// Settings represents optional nested settings.
type Settings struct {
	AllowComments bool           `json:"allow_comments" default:"true" openapi:"title=Allow Comments,description=Enable comments"`
	Priority      int            `json:"priority" validate:"min=0,max=100" openapi:"title=Priority,description=Content priority (0-100)"`
	CustomFields  map[string]any `json:"custom_fields" openapi:"title=Custom Fields,description=Additional custom fields"`
}

// setupExtractor creates a RequestSchemaExtractor with default configuration for testing.
func setupExtractor() *requestSchemaExtractor {
	metadata := NewMetadata()
	registry := NewMapRegistry("#/components/schemas/", DefaultSchemaNamer, metadata)

	return NewRequestSchemaExtractor(registry, metadata)
}

// assertOperationJSON marshals the operation to JSON and compares it with the expected JSON string.
func assertOperationJSON(t *testing.T, op *Operation, wantedJSON string) {
	t.Helper()
	operationJSON, err := json.MarshalIndent(op, "", "  ")
	require.NoError(t, err, "Should be able to marshal operation with indentation")
	assert.JSONEq(t, wantedJSON, string(operationJSON), "Generated JSON should match expected JSON")
}

func TestRequestFromType_EndToEnd(t *testing.T) {
	extractor := setupExtractor()

	// Create operation
	op := &Operation{
		OperationID: "createPost",
		Parameters:  make([]*Param, 0),
	}

	// Generate request schema from complex input type
	inputType := reflect.TypeOf(ComplexRequestInput{})
	err := extractor.RequestFromType(inputType, op)
	require.NoError(t, err, "Should successfully extract request schema")

	// Expected JSON string - review and verify this is correct
	// This can be copy-pasted from actual output for validation
	// Parameters are ordered: path, query, header, cookie
	wantedJSON := `{
  "operationId": "createPost",
  "parameters": [
    {
      "description": "Unique identifier for the user",
      "explode": false,
      "in": "path",
      "name": "user_id",
      "required": true,
      "schema": {
        "format": "int64",
        "type": "integer"
      },
      "style": "simple"
    },
    {
      "description": "URL-friendly identifier for the post",
      "explode": false,
      "in": "path",
      "name": "post_slug",
      "required": true,
      "schema": {
        "type": "string"
      },
      "style": "simple"
    },
    {
      "description": "Page number for pagination",
      "explode": true,
      "in": "query",
      "name": "page",
      "required": true,
      "schema": {
        "format": "int64",
        "type": "integer"
      },
      "style": "form"
    },
    {
      "description": "Number of items per page",
      "explode": true,
      "in": "query",
      "name": "page_size",
      "schema": {
        "format": "int64",
        "type": [
          "integer",
          "null"
        ]
      },
      "style": "form"
    },
    {
      "description": "Search term to filter results",
      "explode": true,
      "in": "query",
      "name": "search",
      "schema": {
        "type": "string"
      },
      "style": "form"
    },
    {
      "description": "Filter by tags",
      "explode": true,
      "in": "query",
      "name": "tags",
      "schema": {
        "items": {
          "type": "string"
        },
        "type": [
          "array",
          "null"
        ]
      },
      "style": "form"
    },
    {
      "description": "Include soft-deleted items",
      "explode": true,
      "in": "query",
      "name": "include_deleted",
      "schema": {
        "type": "boolean"
      },
      "style": "form"
    },
    {
      "description": "Authentication API key",
      "explode": false,
      "in": "header",
      "name": "X-API-Key",
      "required": true,
      "schema": {
        "type": "string"
      },
      "style": "simple"
    },
    {
      "description": "Unique request identifier for tracing",
      "explode": false,
      "in": "header",
      "name": "X-Request-ID",
      "schema": {
        "type": "string"
      },
      "style": "simple"
    },
    {
      "description": "Preferred languages",
      "explode": false,
      "in": "header",
      "name": "Accept-Language",
      "schema": {
        "items": {
          "type": "string"
        },
        "type": [
          "array",
          "null"
        ]
      },
      "style": "simple"
    },
    {
      "description": "User session identifier",
      "explode": true,
      "in": "cookie",
      "name": "session_id",
      "required": true,
      "schema": {
        "type": "string"
      },
      "style": "form"
    },
    {
      "description": "UI theme preference",
      "explode": true,
      "in": "cookie",
      "name": "theme",
      "schema": {
        "type": "string"
      },
      "style": "form"
    }
  ],
  "requestBody": {
    "content": {
      "application/json": {
        "schema": {
          "$ref": "#/components/schemas/ComplexRequestBody"
        }
      }
    },
    "required": true
  }
}`

	assertOperationJSON(t, op, wantedJSON)
}

// FileUploadInput represents a multipart file upload request.
type FileUploadInput struct {
	// Path parameter
	ResourceID string `schema:"resource_id,location=path,required=true" openapi:"title=Resource ID,description=ID of the resource to upload to"`

	// Query parameters
	Overwrite bool `schema:"overwrite,location=query" openapi:"title=Overwrite,description=Whether to overwrite existing file"`

	// Multipart body with file
	Body struct {
		File        []byte `json:"file" openapi:"title=File,description=File content,format=binary"`
		Filename    string `json:"filename" openapi:"title=Filename,description=Name of the file"`
		Description string `json:"description" openapi:"title=Description,description=Optional file description"`
	} `body:"multipart,required"`
}

func TestRequestFromType_MultipartFileUpload(t *testing.T) {
	extractor := setupExtractor()

	// Create operation
	op := &Operation{
		OperationID: "uploadFile",
		Parameters:  make([]*Param, 0),
	}

	// Generate request schema from file upload input type
	inputType := reflect.TypeOf(FileUploadInput{})
	err := extractor.RequestFromType(inputType, op)
	require.NoError(t, err, "Should successfully extract request schema")

	// Expected JSON string - multipart/form-data with binary file
	wantedJSON := `{
  "operationId": "uploadFile",
  "parameters": [
    {
      "description": "ID of the resource to upload to",
      "explode": false,
      "in": "path",
      "name": "resource_id",
      "required": true,
      "schema": {
        "type": "string"
      },
      "style": "simple"
    },
    {
      "description": "Whether to overwrite existing file",
      "explode": true,
      "in": "query",
      "name": "overwrite",
      "schema": {
        "type": "boolean"
      },
      "style": "form"
    }
  ],
  "requestBody": {
    "content": {
      "multipart/form-data": {
        "encoding": {
          "File": {
            "contentType": "application/octet-stream"
          }
        },
        "schema": {
          "properties": {
            "Description": {
              "description": "Optional file description",
              "title": "Description",
              "type": "string"
            },
            "File": {
              "contentMediaType": "application/octet-stream",
              "description": "File content",
              "format": "binary",
              "title": "File",
              "type": "string"
            },
            "Filename": {
              "description": "Name of the file",
              "title": "Filename",
              "type": "string"
            }
          },
          "type": "object"
        }
      }
    },
    "required": true
  }
}`

	assertOperationJSON(t, op, wantedJSON)
}

func TestExtractSecurity_PublicRoute(t *testing.T) {
	extractor := setupExtractor()

	// Create operation
	op := &Operation{
		OperationID: "publicEndpoint",
	}

	// Extract security from public route
	route := &BaseRoute{
		Security: nil, // Public route
	}
	extractor.ExtractSecurity(route, op)

	// Expected JSON string - public route has no security
	wantedJSON := `{
  "operationId": "publicEndpoint"
}`

	assertOperationJSON(t, op, wantedJSON)
}

func TestExtractSecurity_RequireAuthOnly(t *testing.T) {
	extractor := setupExtractor()

	// Create operation
	op := &Operation{
		OperationID: "authRequired",
	}

	// Extract security from route requiring auth only (using Roles("authenticated"))
	route := &BaseRoute{
		Security: &RouteSecurity{
			Roles: []string{"authenticated"},
		},
	}
	extractor.ExtractSecurity(route, op)

	// Expected JSON string - Auth() adds Roles("authenticated")
	wantedJSON := `{
  "operationId": "authRequired",
  "security": [
    {
      "bearerAuth": ["authenticated"]
    }
  ]
}`

	assertOperationJSON(t, op, wantedJSON)
}

func TestExtractSecurity_WithRoles(t *testing.T) {
	extractor := setupExtractor()

	// Create operation
	op := &Operation{
		OperationID: "adminOnly",
	}

	// Extract security from route with roles
	route := &BaseRoute{
		Security: &RouteSecurity{
			Roles: []string{"admin", "editor"},
		},
	}
	extractor.ExtractSecurity(route, op)

	// Expected JSON string - roles as scopes
	wantedJSON := `{
  "operationId": "adminOnly",
  "security": [
    {
      "bearerAuth": [
        "admin",
        "editor"
      ]
    }
  ]
}`

	assertOperationJSON(t, op, wantedJSON)
}

func TestExtractSecurity_WithPermissions(t *testing.T) {
	extractor := setupExtractor()

	// Create operation
	op := &Operation{
		OperationID: "postsAccess",
	}

	// Extract security from route with permissions
	route := &BaseRoute{
		Security: &RouteSecurity{
			Permissions: []string{"posts:read", "posts:write"},
		},
	}
	extractor.ExtractSecurity(route, op)

	// Expected JSON string - permissions as scopes
	wantedJSON := `{
  "operationId": "postsAccess",
  "security": [
    {
      "bearerAuth": [
        "posts:read",
        "posts:write"
      ]
    }
  ]
}`

	assertOperationJSON(t, op, wantedJSON)
}

func TestExtractSecurity_WithRolesAndPermissions(t *testing.T) {
	extractor := setupExtractor()

	// Create operation
	op := &Operation{
		OperationID: "adminDelete",
	}

	// Extract security from route with roles and permissions
	route := &BaseRoute{
		Security: &RouteSecurity{
			Roles:       []string{"admin"},
			Permissions: []string{"posts:delete"},
		},
	}
	extractor.ExtractSecurity(route, op)

	// Expected JSON string - roles and permissions combined as scopes
	wantedJSON := `{
  "operationId": "adminDelete",
  "security": [
    {
      "bearerAuth": [
        "admin",
        "posts:delete"
      ]
    }
  ]
}`

	assertOperationJSON(t, op, wantedJSON)
}

func TestExtractSecurity_DoesNotOverrideExisting(t *testing.T) {
	extractor := setupExtractor()

	// Create operation with existing security
	op := &Operation{
		OperationID: "customAuth",
		Security: []map[string][]string{
			{"customAuth": {"scope1"}},
		},
	}

	// Extract security from route - should not override existing
	route := &BaseRoute{
		Security: &RouteSecurity{
			Roles: []string{"admin"},
		},
	}
	extractor.ExtractSecurity(route, op)

	// Expected JSON string - existing security preserved
	wantedJSON := `{
  "operationId": "customAuth",
  "security": [
    {
      "customAuth": [
        "scope1"
      ]
    }
  ]
}`

	assertOperationJSON(t, op, wantedJSON)
}
