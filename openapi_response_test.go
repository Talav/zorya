package zorya

import (
	"encoding/json"
	"net/http"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ComplexResponseOutput is a comprehensive test struct covering all response scenarios.
type ComplexResponseOutput struct {
	// Response headers
	ETag         string   `schema:"ETag,location=header" openapi:"title=ETag,description=Entity tag for cache validation"`
	Location     string   `schema:"Location,location=header" openapi:"title=Location,description=Resource location URL"`
	XRequestID   string   `schema:"X-Request-ID,location=header" openapi:"title=Request ID,description=Request identifier for tracing"`
	XRateLimit   int      `schema:"X-RateLimit-Limit,location=header" openapi:"title=Rate Limit,description=Rate limit per hour"`
	AllowMethods []string `schema:"Allow,location=header" openapi:"title=Allowed Methods,description=HTTP methods allowed for this resource"`
	CacheControl string   `schema:"Cache-Control,location=header" openapi:"title=Cache Control,description=Cache control directives"`

	// Response body (complex nested structure)
	Body ComplexResponseBody `body:"structured" openapi:"title=Response Body,description=Main response payload"`
}

// ComplexResponseBody represents a complex nested response body structure.
type ComplexResponseBody struct {
	ID          int64            `json:"id" openapi:"title=ID,description=Unique identifier"`
	Title       string           `json:"title" validate:"required" openapi:"title=Title,description=Resource title"`
	Description *string          `json:"description" openapi:"title=Description,description=Detailed description"`
	Status      string           `json:"status" openapi:"title=Status,description=Current status"`
	Tags        []string         `json:"tags" openapi:"title=Tags,description=Associated tags"`
	Metadata    ResponseMetadata `json:"metadata" openapi:"title=Metadata,description=Response metadata"`
	Links       *ResponseLinks   `json:"links" openapi:"title=Links,description=Related resource links"`
	Pagination  PaginationInfo   `json:"pagination" openapi:"title=Pagination,description=Pagination information"`
	Items       []ResponseItem   `json:"items" openapi:"title=Items,description=List of items"`
	IsActive    bool             `json:"is_active" default:"true" openapi:"title=Is Active,description=Active status"`
	Score       *float64         `json:"score" openapi:"title=Score,description=Quality score"`
	CustomData  map[string]any   `json:"custom_data" openapi:"title=Custom Data,description=Additional custom data"`
}

// ResponseMetadata represents nested metadata structure.
type ResponseMetadata struct {
	CreatedAt  string   `json:"created_at" openapi:"title=Created At,description=Creation timestamp"`
	UpdatedAt  string   `json:"updated_at" openapi:"title=Updated At,description=Last update timestamp"`
	Version    int      `json:"version" openapi:"title=Version,description=Resource version"`
	Author     Author   `json:"author" openapi:"title=Author,description=Resource author"`
	Categories []string `json:"categories" openapi:"title=Categories,description=Content categories"`
}

// Author represents author information.
type Author struct {
	ID    int64  `json:"id" openapi:"title=Author ID,description=Author identifier"`
	Name  string `json:"name" openapi:"title=Name,description=Author name"`
	Email string `json:"email" openapi:"title=Email,description=Author email"`
}

// ResponseLinks represents optional nested links structure.
type ResponseLinks struct {
	Self    string  `json:"self" openapi:"title=Self,description=Self reference link"`
	Related *string `json:"related" openapi:"title=Related,description=Related resource link"`
	Next    *string `json:"next" openapi:"title=Next,description=Next page link"`
	Prev    *string `json:"prev" openapi:"title=Previous,description=Previous page link"`
}

// PaginationInfo represents pagination details.
type PaginationInfo struct {
	Page       int `json:"page" openapi:"title=Page,description=Current page number"`
	PageSize   int `json:"page_size" openapi:"title=Page Size,description=Items per page"`
	TotalPages int `json:"total_pages" openapi:"title=Total Pages,description=Total number of pages"`
	TotalItems int `json:"total_items" openapi:"title=Total Items,description=Total number of items"`
}

// ResponseItem represents an item in the response list.
type ResponseItem struct {
	ID          int64   `json:"id" openapi:"title=Item ID,description=Item identifier"`
	Name        string  `json:"name" openapi:"title=Name,description=Item name"`
	Description *string `json:"description" openapi:"title=Description,description=Item description"`
	Price       float64 `json:"price" openapi:"title=Price,description=Item price"`
	InStock     bool    `json:"in_stock" openapi:"title=In Stock,description=Stock availability"`
}

func TestResponseFromType_EndToEnd(t *testing.T) {
	// Create metadata with all parsers
	metadata := NewMetadata()

	// Create registry
	registry := NewMapRegistry("#/components/schemas/", DefaultSchemaNamer, metadata)

	// Create schema builder
	builder := newSchemaBuilder(registry, metadata)

	// Create extractor
	extractor := NewResponseSchemaExtractor(registry, builder, metadata)

	// Create operation with some input parameters (to trigger automatic 422)
	op := &Operation{
		OperationID: "getResource",
		Parameters: []*Param{
			{Name: "id", In: "path", Required: true},
		},
		RequestBody: &RequestBody{
			Required: true,
			Content: map[string]*MediaType{
				"application/json": {},
			},
		},
	}

	// Create route with custom error codes
	route := &BaseRoute{
		Operation:     op,
		DefaultStatus: http.StatusOK,
		Errors:        []int{http.StatusNotFound, http.StatusForbidden},
	}

	// Generate response schema from complex output type
	outputType := reflect.TypeOf(ComplexResponseOutput{})
	err := extractor.ResponseFromType(outputType, route)
	require.NoError(t, err, "Should successfully extract response schema")

	// Marshal operation to JSON with indentation for comparison
	operationJSON, err := json.MarshalIndent(op, "", "  ")
	require.NoError(t, err, "Should be able to marshal operation with indentation")

	// Expected JSON string - review and verify this is correct
	// This can be copy-pasted from actual output for validation with external OpenAPI validators
	// Expected structure:
	// - Success response (200) with body schema and headers
	// - Custom error responses (404, 403)
	// - Automatic error responses (422, 500) because Errors is provided and hasInputParams/hasInputBody
	wantedJSON := `{
  "operationId": "getResource",
  "parameters": [
    {
      "in": "path",
      "name": "id",
      "required": true
    }
  ],
  "requestBody": {
    "content": {
      "application/json": {}
    },
    "required": true
  },
  "responses": {
    "200": {
      "content": {
        "application/json": {
          "schema": {
            "$ref": "#/components/schemas/ComplexResponseBody"
          }
        }
      },
      "description": "OK",
      "headers": {
        "Allow": {
          "description": "HTTP methods allowed for this resource",
          "schema": {
            "type": "string"
          }
        },
        "Cache-Control": {
          "description": "Cache control directives",
          "schema": {
            "type": "string"
          }
        },
        "ETag": {
          "description": "Entity tag for cache validation",
          "schema": {
            "type": "string"
          }
        },
        "Location": {
          "description": "Resource location URL",
          "schema": {
            "type": "string"
          }
        },
        "X-RateLimit-Limit": {
          "description": "Rate limit per hour",
          "schema": {
            "format": "int64",
            "type": "integer"
          }
        },
        "X-Request-ID": {
          "description": "Request identifier for tracing",
          "schema": {
            "type": "string"
          }
        }
      }
    },
    "403": {
      "content": {
        "application/problem+json": {
          "schema": {
            "$ref": "#/components/schemas/ErrorModel"
          }
        }
      },
      "description": "Forbidden"
    },
    "404": {
      "content": {
        "application/problem+json": {
          "schema": {
            "$ref": "#/components/schemas/ErrorModel"
          }
        }
      },
      "description": "Not Found"
    },
    "422": {
      "content": {
        "application/problem+json": {
          "schema": {
            "$ref": "#/components/schemas/ErrorModel"
          }
        }
      },
      "description": "Unprocessable Entity"
    },
    "500": {
      "content": {
        "application/problem+json": {
          "schema": {
            "$ref": "#/components/schemas/ErrorModel"
          }
        }
      },
      "description": "Internal Server Error"
    }
  }
}`

	// Compare JSON strings
	assert.JSONEq(t, wantedJSON, string(operationJSON), "Generated JSON should match expected JSON")
}

// TestResponseFromType_FileDownload tests schema generation for file download response.
func TestResponseFromType_FileDownload(t *testing.T) {
	metadata := NewMetadata()
	registry := NewMapRegistry("#/components/schemas/", DefaultSchemaNamer, metadata)
	builder := newSchemaBuilder(registry, metadata)
	extractor := NewResponseSchemaExtractor(registry, builder, metadata)

	// Define file download output
	type FileDownloadOutput struct {
		ContentType        string `schema:"Content-Type,location=header" openapi:"description=MIME type of the file"`
		ContentDisposition string `schema:"Content-Disposition,location=header" openapi:"description=Attachment filename"`
		Body               []byte `body:"file"`
	}

	route := &BaseRoute{
		Method:        "GET",
		Path:          "/files/{file_id}",
		DefaultStatus: 200,
		Errors:        []int{404},
		Operation: &Operation{
			OperationID: "downloadFile",
			Parameters:  []*Param{{Name: "file_id", In: "path", Required: true}},
		},
	}

	outputType := reflect.TypeOf(FileDownloadOutput{})
	err := extractor.ResponseFromType(outputType, route)
	require.NoError(t, err, "Should successfully extract response schema for file download")

	operationJSON, err := json.MarshalIndent(route.Operation, "", "  ")
	require.NoError(t, err, "Should be able to marshal operation with indentation")

	wantedJSON := `{
  "operationId": "downloadFile",
  "parameters": [
    {
      "in": "path",
      "name": "file_id",
      "required": true
    }
  ],
  "responses": {
    "200": {
      "content": {
        "application/octet-stream": {
          "schema": {
            "contentMediaType": "application/octet-stream",
            "format": "binary",
            "type": "string"
          }
        }
      },
      "description": "OK",
      "headers": {
        "Content-Disposition": {
          "description": "Attachment filename",
          "schema": {
            "type": "string"
          }
        },
        "Content-Type": {
          "description": "MIME type of the file",
          "schema": {
            "type": "string"
          }
        }
      }
    },
    "404": {
      "content": {
        "application/problem+json": {
          "schema": {
            "$ref": "#/components/schemas/ErrorModel"
          }
        }
      },
      "description": "Not Found"
    },
    "422": {
      "content": {
        "application/problem+json": {
          "schema": {
            "$ref": "#/components/schemas/ErrorModel"
          }
        }
      },
      "description": "Unprocessable Entity"
    },
    "500": {
      "content": {
        "application/problem+json": {
          "schema": {
            "$ref": "#/components/schemas/ErrorModel"
          }
        }
      },
      "description": "Internal Server Error"
    }
  }
}`

	assert.JSONEq(t, wantedJSON, string(operationJSON), "Generated JSON should match expected JSON for file download")
}

// TestResponseFromType_StreamingResponse tests schema generation for streaming response.
func TestResponseFromType_StreamingResponse(t *testing.T) {
	metadata := NewMetadata()
	registry := NewMapRegistry("#/components/schemas/", DefaultSchemaNamer, metadata)
	builder := newSchemaBuilder(registry, metadata)
	extractor := NewResponseSchemaExtractor(registry, builder, metadata)

	// Define streaming output
	type StreamOutput struct {
		ContentType  string                                   `schema:"Content-Type,location=header" openapi:"description=Content type for streaming"`
		CacheControl string                                   `schema:"Cache-Control,location=header" openapi:"description=Cache control directives"`
		Body         func(http.ResponseWriter, *http.Request) `body:"structured"`
	}

	route := &BaseRoute{
		Method:        "GET",
		Path:          "/stream",
		DefaultStatus: 200,
		Operation: &Operation{
			OperationID: "streamData",
		},
	}

	outputType := reflect.TypeOf(StreamOutput{})
	err := extractor.ResponseFromType(outputType, route)
	require.NoError(t, err, "Should successfully extract response schema for streaming")

	operationJSON, err := json.MarshalIndent(route.Operation, "", "  ")
	require.NoError(t, err, "Should be able to marshal operation with indentation")

	wantedJSON := `{
  "operationId": "streamData",
  "responses": {
    "200": {
      "content": {
        "application/json": {}
      },
      "description": "OK",
      "headers": {
        "Cache-Control": {
          "description": "Cache control directives",
          "schema": {
            "type": "string"
          }
        },
        "Content-Type": {
          "description": "Content type for streaming",
          "schema": {
            "type": "string"
          }
        }
      }
    },
    "500": {
      "content": {
        "application/problem+json": {
          "schema": {
            "$ref": "#/components/schemas/ErrorModel"
          }
        }
      },
      "description": "Internal Server Error"
    }
  }
}`

	assert.JSONEq(t, wantedJSON, string(operationJSON), "Generated JSON should match expected JSON for streaming")
}

// TestResponseFromType_NoErrors tests response generation without custom errors (should add default).
func TestResponseFromType_NoErrors(t *testing.T) {
	metadata := NewMetadata()
	registry := NewMapRegistry("#/components/schemas/", DefaultSchemaNamer, metadata)
	builder := newSchemaBuilder(registry, metadata)
	extractor := NewResponseSchemaExtractor(registry, builder, metadata)

	op := &Operation{
		OperationID: "simpleGet",
	}

	route := &BaseRoute{
		Operation:     op,
		DefaultStatus: http.StatusOK,
		Errors:        []int{}, // No custom errors
	}

	outputType := reflect.TypeOf(ComplexResponseOutput{})
	err := extractor.ResponseFromType(outputType, route)
	require.NoError(t, err)

	operationJSON, err := json.MarshalIndent(op, "", "  ")
	require.NoError(t, err)

	// Should have success response and 500 error (always added)
	wantedJSON := `{
  "operationId": "simpleGet",
  "responses": {
    "200": {
      "content": {
        "application/json": {
          "schema": {
            "$ref": "#/components/schemas/ComplexResponseBody"
          }
        }
      },
      "description": "OK",
      "headers": {
        "Allow": {
          "description": "HTTP methods allowed for this resource",
          "schema": {
            "type": "string"
          }
        },
        "Cache-Control": {
          "description": "Cache control directives",
          "schema": {
            "type": "string"
          }
        },
        "ETag": {
          "description": "Entity tag for cache validation",
          "schema": {
            "type": "string"
          }
        },
        "Location": {
          "description": "Resource location URL",
          "schema": {
            "type": "string"
          }
        },
        "X-RateLimit-Limit": {
          "description": "Rate limit per hour",
          "schema": {
            "format": "int64",
            "type": "integer"
          }
        },
        "X-Request-ID": {
          "description": "Request identifier for tracing",
          "schema": {
            "type": "string"
          }
        }
      }
    },
    "500": {
      "content": {
        "application/problem+json": {
          "schema": {
            "$ref": "#/components/schemas/ErrorModel"
          }
        }
      },
      "description": "Internal Server Error"
    }
  }
}`

	assert.JSONEq(t, wantedJSON, string(operationJSON))
}
