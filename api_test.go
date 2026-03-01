package zorya

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// GetUserInput represents the request for getting a user.
type GetUserInput struct {
	ID int `schema:"id,location=path,required=true"`
}

// GetUserOutput represents the response for getting a user.
type GetUserOutput struct {
	Body struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `body:"structured"`
}

// testChiAdapter is a minimal adapter implementation for testing to avoid import cycle.
type testChiAdapter struct {
	router chi.Router
}

func (a *testChiAdapter) Handle(route *BaseRoute, handler http.HandlerFunc) {
	a.router.MethodFunc(route.Method, route.Path, handler)
}

func (a *testChiAdapter) ExtractRouterParams(r *http.Request, route *BaseRoute) map[string]string {
	routerParams := make(map[string]string)
	chiCtx := chi.RouteContext(r.Context())
	if chiCtx != nil {
		for i, key := range chiCtx.URLParams.Keys {
			if i < len(chiCtx.URLParams.Values) {
				routerParams[key] = chiCtx.URLParams.Values[i]
			}
		}
	}

	return routerParams
}

func (a *testChiAdapter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.router.ServeHTTP(w, r)
}

func TestOpenAPIEndpoint_EndToEnd(t *testing.T) {
	// 1. Setup: Create router, adapter, and API
	router := chi.NewMux()
	adapter := &testChiAdapter{router: router}
	api := NewAPI(adapter)

	// 2. Register a simple route (GET /users/{id})
	Get(api, "/users/{id}", func(ctx context.Context, input *GetUserInput) (*GetUserOutput, error) {
		output := &GetUserOutput{}
		output.Body.ID = input.ID
		output.Body.Name = "John Doe"

		return output, nil
	})

	// 3. Make HTTP GET request to /openapi.json using httptest
	req := httptest.NewRequest(http.MethodGet, "/openapi.json", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	// 4. Validate response status (200 OK)
	assert.Equal(t, http.StatusOK, recorder.Code, "Response status should be 200 OK")

	// 5. Validate Content-Type header
	assert.Equal(t, "application/vnd.oai.openapi+json", recorder.Header().Get("Content-Type"), "Content-Type should be application/vnd.oai.openapi+json")

	// 6. Parse response body
	var spec map[string]interface{}
	err := json.Unmarshal(recorder.Body.Bytes(), &spec)
	require.NoError(t, err, "Response body should be valid JSON")

	// 7. Marshal with indentation for comparison
	actualJSON, err := json.MarshalIndent(spec, "", "  ")
	require.NoError(t, err, "Should be able to marshal spec with indentation")

	// 8. Expected JSON - copy-pasted from actual output for validation
	// This can be updated from the actual output after first run
	wantedJSON := `{
  "components": {
    "schemas": {
      "ErrorDetail": {
        "properties": {
          "code": {
            "type": "string"
          },
          "location": {
            "type": "string"
          },
          "message": {
            "type": "string"
          }
        },
        "type": "object"
      },
      "ErrorModel": {
        "properties": {
          "detail": {
            "type": "string"
          },
          "errors": {
            "items": {
              "$ref": "#/components/schemas/ErrorDetail"
            },
            "type": "array"
          },
          "instance": {
            "type": "string"
          },
          "status": {
            "format": "int64",
            "type": "integer"
          },
          "title": {
            "type": "string"
          },
          "type": {
            "type": "string"
          }
        },
        "type": "object"
      },
      "GetUserOutputBody": {
        "properties": {
          "id": {
            "format": "int64",
            "type": "integer"
          },
          "name": {
            "type": "string"
          }
        },
        "type": "object"
      }
    }
  },
  "info": {
    "title": "API",
    "version": "1.0.0"
  },
  "openapi": "3.1.2",
  "paths": {
    "/users/{id}": {
      "get": {
        "parameters": [
          {
            "in": "path",
            "name": "id",
            "required": true,
            "schema": {
              "format": "int64",
              "type": "integer"
            },
            "style": "simple"
          }
        ],
        "responses": {
          "200": {
            "content": {
              "application/json": {
                "schema": {
                  "$ref": "#/components/schemas/GetUserOutputBody"
                }
              }
            },
            "description": "OK"
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
      }
    }
  }
}`

	// 9. Compare with expected JSON string
	assert.JSONEq(t, wantedJSON, string(actualJSON), "Generated JSON should match expected JSON")
}

// UploadFileInput represents the request for uploading a file.
type UploadFileInput struct {
	ResourceID string `schema:"resource_id,location=path,required=true" openapi:"title=Resource ID,description=ID of the resource to upload to"`
	Overwrite  bool   `schema:"overwrite,location=query" openapi:"title=Overwrite,description=Whether to overwrite existing file"`
	Body       struct {
		File        []byte `json:"file" openapi:"title=File,description=File content,format=binary"`
		Filename    string `json:"filename" openapi:"title=Filename,description=Name of the file"`
		Description string `json:"description" openapi:"title=Description,description=Optional file description"`
	} `body:"multipart,required"`
}

// UploadFileOutput represents the response for uploading a file.
type UploadFileOutput struct {
	Body struct {
		ID       int64  `json:"id"`
		Filename string `json:"filename"`
		Size     int64  `json:"size"`
	} `body:"structured"`
}

func TestOpenAPIEndpoint_FileUpload(t *testing.T) {
	// 1. Setup: Create router, adapter, and API
	router := chi.NewMux()
	adapter := &testChiAdapter{router: router}
	api := NewAPI(adapter)

	// 2. Register a file upload route (POST /resources/{resource_id}/files)
	Post(api, "/resources/{resource_id}/files", func(ctx context.Context, input *UploadFileInput) (*UploadFileOutput, error) {
		output := &UploadFileOutput{}
		output.Body.ID = 1
		output.Body.Filename = input.Body.Filename
		output.Body.Size = int64(len(input.Body.File))

		return output, nil
	})

	// 3. Make HTTP GET request to /openapi.json using httptest
	req := httptest.NewRequest(http.MethodGet, "/openapi.json", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	// 4. Validate response status (200 OK)
	assert.Equal(t, http.StatusOK, recorder.Code, "Response status should be 200 OK")

	// 5. Validate Content-Type header
	assert.Equal(t, "application/vnd.oai.openapi+json", recorder.Header().Get("Content-Type"), "Content-Type should be application/vnd.oai.openapi+json")

	// 6. Parse response body
	var spec map[string]interface{}
	err := json.Unmarshal(recorder.Body.Bytes(), &spec)
	require.NoError(t, err, "Response body should be valid JSON")

	// 7. Marshal with indentation for comparison
	actualJSON, err := json.MarshalIndent(spec, "", "  ")
	require.NoError(t, err, "Should be able to marshal spec with indentation")

	// 8. Expected JSON - OpenAPI 3.1 compliant multipart/form-data with inline schema
	wantedJSON := `{
  "components": {
    "schemas": {
      "ErrorDetail": {
        "properties": {
          "code": {
            "type": "string"
          },
          "location": {
            "type": "string"
          },
          "message": {
            "type": "string"
          }
        },
        "type": "object"
      },
      "ErrorModel": {
        "properties": {
          "detail": {
            "type": "string"
          },
          "errors": {
            "items": {
              "$ref": "#/components/schemas/ErrorDetail"
            },
            "type": "array"
          },
          "instance": {
            "type": "string"
          },
          "status": {
            "format": "int64",
            "type": "integer"
          },
          "title": {
            "type": "string"
          },
          "type": {
            "type": "string"
          }
        },
        "type": "object"
      },
      "UploadFileOutputBody": {
        "properties": {
          "filename": {
            "type": "string"
          },
          "id": {
            "format": "int64",
            "type": "integer"
          },
          "size": {
            "format": "int64",
            "type": "integer"
          }
        },
        "type": "object"
      }
    }
  },
  "info": {
    "title": "API",
    "version": "1.0.0"
  },
  "openapi": "3.1.2",
  "paths": {
    "/resources/{resource_id}/files": {
      "post": {
        "parameters": [
          {
            "description": "ID of the resource to upload to",
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
                "file": {
                  "contentType": "application/octet-stream"
                }
              },
              "schema": {
                "properties": {
                  "description": {
                    "description": "Optional file description",
                    "title": "Description",
                    "type": "string"
                  },
                  "file": {
                    "description": "File content",
                    "format": "binary",
                    "title": "File",
                    "type": "string"
                  },
                  "filename": {
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
        },
        "responses": {
          "200": {
            "content": {
              "application/json": {
                "schema": {
                  "$ref": "#/components/schemas/UploadFileOutputBody"
                }
              }
            },
            "description": "OK"
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
      }
    }
  }
}`

	// 9. Compare with expected JSON string
	assert.JSONEq(t, wantedJSON, string(actualJSON), "Generated JSON should match expected JSON")
}

// DownloadFileInput represents the request for downloading a file.
type DownloadFileInput struct {
	FileID string `schema:"file_id,location=path,required=true"`
}

// DownloadFileOutput represents the response for downloading a file.
type DownloadFileOutput struct {
	ContentType        string `schema:"Content-Type,location=header" openapi:"description=MIME type of the file"`
	ContentDisposition string `schema:"Content-Disposition,location=header" openapi:"description=Attachment filename"`
	Body               []byte `body:"file"`
}

func TestOpenAPIEndpoint_FileDownload(t *testing.T) {
	// 1. Setup: Create router, adapter, and API
	router := chi.NewMux()
	adapter := &testChiAdapter{router: router}
	api := NewAPI(adapter)

	// 2. Register a file download route (GET /files/{file_id})
	Get(api, "/files/{file_id}", func(ctx context.Context, input *DownloadFileInput) (*DownloadFileOutput, error) {
		output := &DownloadFileOutput{}
		output.ContentType = "application/pdf"
		output.ContentDisposition = "attachment; filename=\"document.pdf\""
		output.Body = []byte("file content")

		return output, nil
	})

	// 3. Make HTTP GET request to /openapi.json using httptest
	req := httptest.NewRequest(http.MethodGet, "/openapi.json", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	// 4. Validate response status (200 OK)
	assert.Equal(t, http.StatusOK, recorder.Code, "Response status should be 200 OK")

	// 5. Validate Content-Type header
	assert.Equal(t, "application/vnd.oai.openapi+json", recorder.Header().Get("Content-Type"), "Content-Type should be application/vnd.oai.openapi+json")

	// 6. Parse response body
	var spec map[string]interface{}
	err := json.Unmarshal(recorder.Body.Bytes(), &spec)
	require.NoError(t, err, "Response body should be valid JSON")

	// 7. Marshal with indentation for comparison
	actualJSON, err := json.MarshalIndent(spec, "", "  ")
	require.NoError(t, err, "Should be able to marshal spec with indentation")

	// 8. Expected JSON - OpenAPI 3.1 compliant with header descriptions
	wantedJSON := `{
  "components": {
    "schemas": {
      "ErrorDetail": {
        "properties": {
          "code": {
            "type": "string"
          },
          "location": {
            "type": "string"
          },
          "message": {
            "type": "string"
          }
        },
        "type": "object"
      },
      "ErrorModel": {
        "properties": {
          "detail": {
            "type": "string"
          },
          "errors": {
            "items": {
              "$ref": "#/components/schemas/ErrorDetail"
            },
            "type": "array"
          },
          "instance": {
            "type": "string"
          },
          "status": {
            "format": "int64",
            "type": "integer"
          },
          "title": {
            "type": "string"
          },
          "type": {
            "type": "string"
          }
        },
        "type": "object"
      }
    }
  },
  "info": {
    "title": "API",
    "version": "1.0.0"
  },
  "openapi": "3.1.2",
  "paths": {
    "/files/{file_id}": {
      "get": {
        "parameters": [
          {
            "in": "path",
            "name": "file_id",
            "required": true,
            "schema": {
              "type": "string"
            },
            "style": "simple"
          }
        ],
        "responses": {
          "200": {
            "content": {
              "application/octet-stream": {
                "schema": {
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
      }
    }
  }
}`

	// 9. Compare with expected JSON string
	assert.JSONEq(t, wantedJSON, string(actualJSON), "Generated JSON should match expected JSON")
}

// GetFileWithMetadataInput represents the request for getting a file with metadata.
type GetFileWithMetadataInput struct {
	FileID string `schema:"file_id,location=path,required=true"`
}

// GetFileWithMetadataOutput represents the response for getting a file with metadata.
type GetFileWithMetadataOutput struct {
	Body struct {
		ID          int64  `json:"id"`
		Filename    string `json:"filename"`
		ContentType string `json:"content_type"`
		Size        int64  `json:"size"`
		File        []byte `json:"file" openapi:"title=File,description=File content"`
	} `body:"structured"`
}

func TestOpenAPIEndpoint_FileInResponse(t *testing.T) {
	// 1. Setup: Create router, adapter, and API
	router := chi.NewMux()
	adapter := &testChiAdapter{router: router}
	api := NewAPI(adapter)

	// 2. Register a route that returns a structured response with a file field (GET /files/{file_id}/with-metadata)
	Get(api, "/files/{file_id}/with-metadata", func(ctx context.Context, input *GetFileWithMetadataInput) (*GetFileWithMetadataOutput, error) {
		output := &GetFileWithMetadataOutput{}
		output.Body.ID = 1
		output.Body.Filename = "document.pdf"
		output.Body.ContentType = "application/pdf"
		output.Body.Size = 1024
		output.Body.File = []byte("file content")

		return output, nil
	})

	// 3. Make HTTP GET request to /openapi.json using httptest
	req := httptest.NewRequest(http.MethodGet, "/openapi.json", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	// 4. Validate response status (200 OK)
	assert.Equal(t, http.StatusOK, recorder.Code, "Response status should be 200 OK")

	// 5. Validate Content-Type header
	assert.Equal(t, "application/vnd.oai.openapi+json", recorder.Header().Get("Content-Type"), "Content-Type should be application/vnd.oai.openapi+json")

	// 6. Parse response body
	var spec map[string]interface{}
	err := json.Unmarshal(recorder.Body.Bytes(), &spec)
	require.NoError(t, err, "Response body should be valid JSON")

	// 7. Marshal with indentation for comparison
	actualJSON, err := json.MarshalIndent(spec, "", "  ")
	require.NoError(t, err, "Should be able to marshal spec with indentation")

	// 8. Expected JSON - Binary file field in JSON response with openapi metadata
	wantedJSON := `{
  "components": {
    "schemas": {
      "ErrorDetail": {
        "properties": {
          "code": {
            "type": "string"
          },
          "location": {
            "type": "string"
          },
          "message": {
            "type": "string"
          }
        },
        "type": "object"
      },
      "ErrorModel": {
        "properties": {
          "detail": {
            "type": "string"
          },
          "errors": {
            "items": {
              "$ref": "#/components/schemas/ErrorDetail"
            },
            "type": "array"
          },
          "instance": {
            "type": "string"
          },
          "status": {
            "format": "int64",
            "type": "integer"
          },
          "title": {
            "type": "string"
          },
          "type": {
            "type": "string"
          }
        },
        "type": "object"
      },
      "GetFileWithMetadataOutputBody": {
        "properties": {
          "content_type": {
            "type": "string"
          },
          "file": {
            "contentEncoding": "base64",
            "contentMediaType": "application/octet-stream",
            "description": "File content",
            "title": "File",
            "type": "string"
          },
          "filename": {
            "type": "string"
          },
          "id": {
            "format": "int64",
            "type": "integer"
          },
          "size": {
            "format": "int64",
            "type": "integer"
          }
        },
        "type": "object"
      }
    }
  },
  "info": {
    "title": "API",
    "version": "1.0.0"
  },
  "openapi": "3.1.2",
  "paths": {
    "/files/{file_id}/with-metadata": {
      "get": {
        "parameters": [
          {
            "in": "path",
            "name": "file_id",
            "required": true,
            "schema": {
              "type": "string"
            },
            "style": "simple"
          }
        ],
        "responses": {
          "200": {
            "content": {
              "application/json": {
                "schema": {
                  "$ref": "#/components/schemas/GetFileWithMetadataOutputBody"
                }
              }
            },
            "description": "OK"
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
      }
    }
  }
}`

	// 9. Compare with expected JSON string
	assert.JSONEq(t, wantedJSON, string(actualJSON), "Generated JSON should match expected JSON")
}

// ComprehensiveValidationInput represents a request with all possible validation types.
type ComprehensiveValidationInput struct {
	ResourceID string                      `schema:"resource_id,location=path,required=true"`
	Body       ComprehensiveValidationBody `body:"structured,required"`
}

// ComprehensiveValidationBody represents a body with all possible validation constraints.
type ComprehensiveValidationBody struct {
	// Required field
	RequiredField string `json:"required_field" validate:"required" openapi:"title=Required Field,description=This field is required"`

	// String validations
	StringMinMax   string `json:"string_min_max" validate:"min=5,max=100" openapi:"title=String Min Max,description=String with min and max length"`
	StringLen      string `json:"string_len" validate:"len=10" openapi:"title=String Len,description=String with exact length"`
	StringEmail    string `json:"string_email" validate:"required,email" openapi:"title=Email,description=Valid email address"`
	StringURL      string `json:"string_url" validate:"url" openapi:"title=URL,description=Valid URL"`
	StringAlpha    string `json:"string_alpha" validate:"alpha" openapi:"title=Alpha,description=Alphabetic characters only"`
	StringAlphanum string `json:"string_alphanum" validate:"alphanum" openapi:"title=Alphanum,description=Alphanumeric characters only"`
	StringPattern  string `json:"string_pattern" validate:"pattern=^[A-Z][a-z]+$" openapi:"title=Pattern,description=Custom pattern validation"`
	StringOneOf    string `json:"string_oneof" validate:"oneof=active inactive pending" openapi:"title=Status,description=Status enum"`

	// Integer validations
	IntMinMax     int  `json:"int_min_max" validate:"min=0,max=100" openapi:"title=Int Min Max,description=Integer with min and max"`
	IntGteLte     int  `json:"int_gte_lte" validate:"gte=10,lte=50" openapi:"title=Int Gte Lte,description=Integer with gte and lte"`
	IntGtLt       int  `json:"int_gt_lt" validate:"gt=5,lt=95" openapi:"title=Int Gt Lt,description=Integer with gt and lt"`
	IntMultipleOf int  `json:"int_multiple_of" validate:"multiple_of=5" openapi:"title=Int Multiple Of,description=Integer multiple of 5"`
	IntRequired   *int `json:"int_required" validate:"required" openapi:"title=Int Required,description=Required integer pointer"`

	// Float validations
	FloatMinMax     float64 `json:"float_min_max" validate:"min=0.0,max=100.0" openapi:"title=Float Min Max,description=Float with min and max"`
	FloatGteLte     float64 `json:"float_gte_lte" validate:"gte=10.5,lte=99.5" openapi:"title=Float Gte Lte,description=Float with gte and lte"`
	FloatGtLt       float64 `json:"float_gt_lt" validate:"gt=0.1,lt=99.9" openapi:"title=Float Gt Lt,description=Float with gt and lt"`
	FloatMultipleOf float64 `json:"float_multiple_of" validate:"multiple_of=0.5" openapi:"title=Float Multiple Of,description=Float multiple of 0.5"`

	// Array validations
	ArrayMinMax []string `json:"array_min_max" validate:"min=1,max=10" openapi:"title=Array Min Max,description=Array with min and max items"`
	ArrayLen    []string `json:"array_len" validate:"len=5" openapi:"title=Array Len,description=Array with exact length"`
	ArrayOneOf  []string `json:"array_oneof" validate:"oneof=red green blue" openapi:"title=Array OneOf,description=Array with enum items"`

	// Nested struct with validations
	Nested NestedValidationStruct `json:"nested" validate:"required" openapi:"title=Nested,description=Nested struct with validations"`
}

// NestedValidationStruct represents a nested struct with validations.
type NestedValidationStruct struct {
	Name  string `json:"name" validate:"required,min=3,max=50" openapi:"title=Name,description=Nested name field"`
	Value int    `json:"value" validate:"min=1,max=100" openapi:"title=Value,description=Nested value field"`
}

// ComprehensiveValidationOutput represents the response for comprehensive validation test.
type ComprehensiveValidationOutput struct {
	Body struct {
		ID      int64  `json:"id"`
		Message string `json:"message"`
	} `body:"structured"`
}

func TestOpenAPIEndpoint_ComprehensiveValidations(t *testing.T) {
	// 1. Setup: Create router, adapter, and API
	router := chi.NewMux()
	adapter := &testChiAdapter{router: router}
	api := NewAPI(adapter)

	// 2. Register a route with comprehensive validations (POST /resources/{resource_id}/validate)
	Post(api, "/resources/{resource_id}/validate", func(ctx context.Context, input *ComprehensiveValidationInput) (*ComprehensiveValidationOutput, error) {
		output := &ComprehensiveValidationOutput{}
		output.Body.ID = 1
		output.Body.Message = "Validation successful"

		return output, nil
	})

	// 3. Make HTTP GET request to /openapi.json using httptest
	req := httptest.NewRequest(http.MethodGet, "/openapi.json", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	// 4. Validate response status (200 OK)
	assert.Equal(t, http.StatusOK, recorder.Code, "Response status should be 200 OK")

	// 5. Validate Content-Type header
	assert.Equal(t, "application/vnd.oai.openapi+json", recorder.Header().Get("Content-Type"), "Content-Type should be application/vnd.oai.openapi+json")

	// 6. Parse response body
	var spec map[string]interface{}
	err := json.Unmarshal(recorder.Body.Bytes(), &spec)
	require.NoError(t, err, "Response body should be valid JSON")

	// 7. Marshal with indentation for comparison
	actualJSON, err := json.MarshalIndent(spec, "", "  ")
	require.NoError(t, err, "Should be able to marshal spec with indentation")

	// 8. Expected JSON - All validation constraints are properly generated and included
	wantedJSON := `{
  "components": {
    "schemas": {
      "ComprehensiveValidationBody": {
        "properties": {
          "array_len": {
            "description": "Array with exact length",
            "items": {
              "type": "string"
            },
            "maxItems": 5,
            "minItems": 5,
            "title": "Array Len",
            "type": "array"
          },
          "array_min_max": {
            "description": "Array with min and max items",
            "items": {
              "type": "string"
            },
            "maxItems": 10,
            "minItems": 1,
            "title": "Array Min Max",
            "type": "array"
          },
          "array_oneof": {
            "description": "Array with enum items",
            "items": {
              "enum": [
                "red",
                "green",
                "blue"
              ],
              "type": "string"
            },
            "title": "Array OneOf",
            "type": "array"
          },
          "float_gt_lt": {
            "description": "Float with gt and lt",
            "exclusiveMaximum": 99.9,
            "exclusiveMinimum": 0.1,
            "title": "Float Gt Lt",
            "type": "number"
          },
          "float_gte_lte": {
            "description": "Float with gte and lte",
            "maximum": 99.5,
            "minimum": 10.5,
            "title": "Float Gte Lte",
            "type": "number"
          },
          "float_min_max": {
            "description": "Float with min and max",
            "maximum": 100,
            "minimum": 0,
            "title": "Float Min Max",
            "type": "number"
          },
          "float_multiple_of": {
            "description": "Float multiple of 0.5",
            "multipleOf": 0.5,
            "title": "Float Multiple Of",
            "type": "number"
          },
          "int_gt_lt": {
            "description": "Integer with gt and lt",
            "exclusiveMaximum": 95,
            "exclusiveMinimum": 5,
            "title": "Int Gt Lt",
            "type": "integer"
          },
          "int_gte_lte": {
            "description": "Integer with gte and lte",
            "maximum": 50,
            "minimum": 10,
            "title": "Int Gte Lte",
            "type": "integer"
          },
          "int_min_max": {
            "description": "Integer with min and max",
            "maximum": 100,
            "minimum": 0,
            "title": "Int Min Max",
            "type": "integer"
          },
          "int_multiple_of": {
            "description": "Integer multiple of 5",
            "multipleOf": 5,
            "title": "Int Multiple Of",
            "type": "integer"
          },
          "int_required": {
            "description": "Required integer pointer",
            "title": "Int Required",
            "type": "integer"
          },
          "nested": {
            "$ref": "#/components/schemas/NestedValidationStruct"
          },
          "required_field": {
            "description": "This field is required",
            "title": "Required Field",
            "type": "string"
          },
          "string_alpha": {
            "description": "Alphabetic characters only",
            "pattern": "^[a-zA-Z]+$",
            "title": "Alpha",
            "type": "string"
          },
          "string_alphanum": {
            "description": "Alphanumeric characters only",
            "pattern": "^[a-zA-Z0-9]+$",
            "title": "Alphanum",
            "type": "string"
          },
          "string_email": {
            "description": "Valid email address",
            "format": "email",
            "title": "Email",
            "type": "string"
          },
          "string_len": {
            "description": "String with exact length",
            "maxLength": 10,
            "minLength": 10,
            "title": "String Len",
            "type": "string"
          },
          "string_min_max": {
            "description": "String with min and max length",
            "maxLength": 100,
            "minLength": 5,
            "title": "String Min Max",
            "type": "string"
          },
          "string_oneof": {
            "description": "Status enum",
            "enum": [
              "active",
              "inactive",
              "pending"
            ],
            "title": "Status",
            "type": "string"
          },
          "string_pattern": {
            "description": "Custom pattern validation",
            "pattern": "^[A-Z][a-z]+$",
            "title": "Pattern",
            "type": "string"
          },
          "string_url": {
            "description": "Valid URL",
            "format": "uri",
            "title": "URL",
            "type": "string"
          }
        },
        "required": [
          "required_field",
          "string_email",
          "int_required",
          "nested"
        ],
        "type": "object"
      },
      "ComprehensiveValidationOutputBody": {
        "properties": {
          "id": {
            "format": "int64",
            "type": "integer"
          },
          "message": {
            "type": "string"
          }
        },
        "type": "object"
      },
      "ErrorDetail": {
        "properties": {
          "code": {
            "type": "string"
          },
          "location": {
            "type": "string"
          },
          "message": {
            "type": "string"
          }
        },
        "type": "object"
      },
      "ErrorModel": {
        "properties": {
          "detail": {
            "type": "string"
          },
          "errors": {
            "items": {
              "$ref": "#/components/schemas/ErrorDetail"
            },
            "type": "array"
          },
          "instance": {
            "type": "string"
          },
          "status": {
            "format": "int64",
            "type": "integer"
          },
          "title": {
            "type": "string"
          },
          "type": {
            "type": "string"
          }
        },
        "type": "object"
      },
      "NestedValidationStruct": {
        "properties": {
          "name": {
            "description": "Nested name field",
            "maxLength": 50,
            "minLength": 3,
            "title": "Name",
            "type": "string"
          },
          "value": {
            "description": "Nested value field",
            "maximum": 100,
            "minimum": 1,
            "title": "Value",
            "type": "integer"
          }
        },
        "required": [
          "name"
        ],
        "type": "object"
      }
    }
  },
  "info": {
    "title": "API",
    "version": "1.0.0"
  },
  "openapi": "3.1.2",
  "paths": {
    "/resources/{resource_id}/validate": {
      "post": {
        "parameters": [
          {
            "in": "path",
            "name": "resource_id",
            "required": true,
            "schema": {
              "type": "string"
            },
            "style": "simple"
          }
        ],
        "requestBody": {
          "content": {
            "application/json": {
              "schema": {
                "$ref": "#/components/schemas/ComprehensiveValidationBody"
              }
            }
          },
          "required": true
        },
        "responses": {
          "200": {
            "content": {
              "application/json": {
                "schema": {
                  "$ref": "#/components/schemas/ComprehensiveValidationOutputBody"
                }
              }
            },
            "description": "OK"
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
      }
    }
  }
}`

	// 9. Compare with expected JSON string
	assert.JSONEq(t, wantedJSON, string(actualJSON), "Generated JSON should match expected JSON")
}

func TestDocsEndpoint_EndToEnd(t *testing.T) {
	// 1. Setup: Create router, adapter, and API
	router := chi.NewMux()
	adapter := &testChiAdapter{router: router}
	api := NewAPI(adapter)

	// 2. Register a simple route (GET /users/{id})
	Get(api, "/users/{id}", func(ctx context.Context, input *GetUserInput) (*GetUserOutput, error) {
		output := &GetUserOutput{}
		output.Body.ID = input.ID
		output.Body.Name = "John Doe"

		return output, nil
	})

	// 3. Make HTTP GET request to /docs using httptest
	req := httptest.NewRequest(http.MethodGet, "/docs", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	// 4. Validate response status (200 OK)
	assert.Equal(t, http.StatusOK, recorder.Code, "Response status should be 200 OK")

	// 5. Validate Content-Type header
	assert.Equal(t, "text/html; charset=utf-8", recorder.Header().Get("Content-Type"), "Content-Type should be text/html; charset=utf-8")

	// 6. Validate response body is HTML
	htmlBody := recorder.Body.String()
	assert.Contains(t, htmlBody, "<!DOCTYPE html>", "Response should contain HTML doctype")
	assert.Contains(t, htmlBody, "<elements-api", "Response should contain Stoplight Elements web component")
	assert.Contains(t, htmlBody, "apiDescriptionUrl", "Response should contain apiDescriptionUrl attribute")
	assert.Contains(t, htmlBody, "/openapi.json", "Response should reference OpenAPI spec at /openapi.json")
	assert.Contains(t, htmlBody, "@stoplight/elements", "Response should include Stoplight Elements from CDN")
}

func TestDocsEndpoint_OpenAPIPathNoDoubleJson(t *testing.T) {
	router := chi.NewMux()
	adapter := &testChiAdapter{router: router}
	cfg := DefaultConfig()
	cfg.OpenAPIPath = "/openapi.json"
	api := NewAPI(adapter, WithConfig(cfg))

	Get(api, "/users/{id}", func(ctx context.Context, input *GetUserInput) (*GetUserOutput, error) {
		return &GetUserOutput{}, nil
	})

	req := httptest.NewRequest(http.MethodGet, "/docs", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	require.Equal(t, http.StatusOK, recorder.Code)
	htmlBody := recorder.Body.String()
	assert.Contains(t, htmlBody, `apiDescriptionUrl="/openapi.json"`, "apiDescriptionUrl must use OpenAPIPath as-is")
	assert.NotContains(t, htmlBody, ".json.json", "apiDescriptionUrl must not have double .json extension")
}

func TestStreamingBodyFunc(t *testing.T) {
	type StreamingOutput struct {
		Body func(w http.ResponseWriter) error
	}

	router := chi.NewMux()
	adapter := &testChiAdapter{router: router}
	api := NewAPI(adapter)

	Get(api, "/stream", func(ctx context.Context, _ *struct{}) (*StreamingOutput, error) {
		out := &StreamingOutput{}
		out.Body = func(w http.ResponseWriter) error {
			w.Header().Set("Content-Type", "text/event-stream")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("data: hello\n\n"))
			return nil
		}
		return out, nil
	})

	req := httptest.NewRequest(http.MethodGet, "/stream", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	require.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, "text/event-stream", recorder.Header().Get("Content-Type"))
	assert.Contains(t, recorder.Body.String(), "data: hello")
}
