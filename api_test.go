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
          "Code": {
            "type": "string"
          },
          "Location": {
            "type": "string"
          },
          "Message": {
            "type": "string"
          }
        },
        "type": "object"
      },
      "ErrorModel": {
        "properties": {
          "Detail": {
            "type": "string"
          },
          "Errors": {
            "items": {
              "$ref": "#/components/schemas/ErrorDetail"
            },
            "type": [
              "array",
              "null"
            ]
          },
          "Instance": {
            "type": "string"
          },
          "Status": {
            "format": "int64",
            "type": "integer"
          },
          "Title": {
            "type": "string"
          },
          "Type": {
            "type": "string"
          }
        },
        "type": "object"
      },
      "GetUserOutputBody": {
        "properties": {
          "ID": {
            "format": "int64",
            "type": "integer"
          },
          "Name": {
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
  "openapi": "3.1.0",
  "paths": {
    "/users/{id}": {
      "get": {
        "parameters": [
          {
            "explode": false,
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
          "Code": {
            "type": "string"
          },
          "Location": {
            "type": "string"
          },
          "Message": {
            "type": "string"
          }
        },
        "type": "object"
      },
      "ErrorModel": {
        "properties": {
          "Detail": {
            "type": "string"
          },
          "Errors": {
            "items": {
              "$ref": "#/components/schemas/ErrorDetail"
            },
            "type": [
              "array",
              "null"
            ]
          },
          "Instance": {
            "type": "string"
          },
          "Status": {
            "format": "int64",
            "type": "integer"
          },
          "Title": {
            "type": "string"
          },
          "Type": {
            "type": "string"
          }
        },
        "type": "object"
      },
      "UploadFileOutputBody": {
        "properties": {
          "Filename": {
            "type": "string"
          },
          "ID": {
            "format": "int64",
            "type": "integer"
          },
          "Size": {
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
  "openapi": "3.1.0",
  "paths": {
    "/resources/{resource_id}/files": {
      "post": {
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
          "Code": {
            "type": "string"
          },
          "Location": {
            "type": "string"
          },
          "Message": {
            "type": "string"
          }
        },
        "type": "object"
      },
      "ErrorModel": {
        "properties": {
          "Detail": {
            "type": "string"
          },
          "Errors": {
            "items": {
              "$ref": "#/components/schemas/ErrorDetail"
            },
            "type": [
              "array",
              "null"
            ]
          },
          "Instance": {
            "type": "string"
          },
          "Status": {
            "format": "int64",
            "type": "integer"
          },
          "Title": {
            "type": "string"
          },
          "Type": {
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
  "openapi": "3.1.0",
  "paths": {
    "/files/{file_id}": {
      "get": {
        "parameters": [
          {
            "explode": false,
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
          "Code": {
            "type": "string"
          },
          "Location": {
            "type": "string"
          },
          "Message": {
            "type": "string"
          }
        },
        "type": "object"
      },
      "ErrorModel": {
        "properties": {
          "Detail": {
            "type": "string"
          },
          "Errors": {
            "items": {
              "$ref": "#/components/schemas/ErrorDetail"
            },
            "type": [
              "array",
              "null"
            ]
          },
          "Instance": {
            "type": "string"
          },
          "Status": {
            "format": "int64",
            "type": "integer"
          },
          "Title": {
            "type": "string"
          },
          "Type": {
            "type": "string"
          }
        },
        "type": "object"
      },
      "GetFileWithMetadataOutputBody": {
        "properties": {
          "ContentType": {
            "type": "string"
          },
          "File": {
            "contentEncoding": "base64",
            "contentMediaType": "application/octet-stream",
            "description": "File content",
            "title": "File",
            "type": "string"
          },
          "Filename": {
            "type": "string"
          },
          "ID": {
            "format": "int64",
            "type": "integer"
          },
          "Size": {
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
  "openapi": "3.1.0",
  "paths": {
    "/files/{file_id}/with-metadata": {
      "get": {
        "parameters": [
          {
            "explode": false,
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
          "ArrayLen": {
            "description": "Array with exact length",
            "items": {
              "type": "string"
            },
            "maxItems": 5,
            "minItems": 5,
            "title": "Array Len",
            "type": [
              "array",
              "null"
            ]
          },
          "ArrayMinMax": {
            "description": "Array with min and max items",
            "items": {
              "type": "string"
            },
            "maxItems": 10,
            "minItems": 1,
            "title": "Array Min Max",
            "type": [
              "array",
              "null"
            ]
          },
          "ArrayOneOf": {
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
            "type": [
              "array",
              "null"
            ]
          },
          "FloatGtLt": {
            "description": "Float with gt and lt",
            "exclusiveMaximum": 99.9,
            "exclusiveMinimum": 0.1,
            "title": "Float Gt Lt",
            "type": "number"
          },
          "FloatGteLte": {
            "description": "Float with gte and lte",
            "maximum": 99.5,
            "minimum": 10.5,
            "title": "Float Gte Lte",
            "type": "number"
          },
          "FloatMinMax": {
            "description": "Float with min and max",
            "maximum": 100,
            "minimum": 0,
            "title": "Float Min Max",
            "type": "number"
          },
          "FloatMultipleOf": {
            "description": "Float multiple of 0.5",
            "multipleOf": 0.5,
            "title": "Float Multiple Of",
            "type": "number"
          },
          "IntGtLt": {
            "description": "Integer with gt and lt",
            "exclusiveMaximum": 95,
            "exclusiveMinimum": 5,
            "title": "Int Gt Lt",
            "type": "integer"
          },
          "IntGteLte": {
            "description": "Integer with gte and lte",
            "maximum": 50,
            "minimum": 10,
            "title": "Int Gte Lte",
            "type": "integer"
          },
          "IntMinMax": {
            "description": "Integer with min and max",
            "maximum": 100,
            "minimum": 0,
            "title": "Int Min Max",
            "type": "integer"
          },
          "IntMultipleOf": {
            "description": "Integer multiple of 5",
            "multipleOf": 5,
            "title": "Int Multiple Of",
            "type": "integer"
          },
          "IntRequired": {
            "description": "Required integer pointer",
            "title": "Int Required",
            "type": "integer"
          },
          "Nested": {
            "$ref": "#/components/schemas/NestedValidationStruct",
            "description": "Nested struct with validations",
            "title": "Nested"
          },
          "RequiredField": {
            "description": "This field is required",
            "title": "Required Field",
            "type": "string"
          },
          "StringAlpha": {
            "description": "Alphabetic characters only",
            "pattern": "^[a-zA-Z]+$",
            "title": "Alpha",
            "type": "string"
          },
          "StringAlphanum": {
            "description": "Alphanumeric characters only",
            "pattern": "^[a-zA-Z0-9]+$",
            "title": "Alphanum",
            "type": "string"
          },
          "StringEmail": {
            "description": "Valid email address",
            "format": "email",
            "title": "Email",
            "type": "string"
          },
          "StringLen": {
            "description": "String with exact length",
            "maxLength": 10,
            "minLength": 10,
            "title": "String Len",
            "type": "string"
          },
          "StringMinMax": {
            "description": "String with min and max length",
            "maxLength": 100,
            "minLength": 5,
            "title": "String Min Max",
            "type": "string"
          },
          "StringOneOf": {
            "description": "Status enum",
            "enum": [
              "active",
              "inactive",
              "pending"
            ],
            "title": "Status",
            "type": "string"
          },
          "StringPattern": {
            "description": "Custom pattern validation",
            "pattern": "^[A-Z][a-z]+$",
            "title": "Pattern",
            "type": "string"
          },
          "StringURL": {
            "description": "Valid URL",
            "format": "uri",
            "title": "URL",
            "type": "string"
          }
        },
        "required": [
          "RequiredField",
          "StringEmail",
          "IntRequired",
          "Nested"
        ],
        "type": "object"
      },
      "ComprehensiveValidationOutputBody": {
        "properties": {
          "ID": {
            "format": "int64",
            "type": "integer"
          },
          "Message": {
            "type": "string"
          }
        },
        "type": "object"
      },
      "ErrorDetail": {
        "properties": {
          "Code": {
            "type": "string"
          },
          "Location": {
            "type": "string"
          },
          "Message": {
            "type": "string"
          }
        },
        "type": "object"
      },
      "ErrorModel": {
        "properties": {
          "Detail": {
            "type": "string"
          },
          "Errors": {
            "items": {
              "$ref": "#/components/schemas/ErrorDetail"
            },
            "type": [
              "array",
              "null"
            ]
          },
          "Instance": {
            "type": "string"
          },
          "Status": {
            "format": "int64",
            "type": "integer"
          },
          "Title": {
            "type": "string"
          },
          "Type": {
            "type": "string"
          }
        },
        "type": "object"
      },
      "NestedValidationStruct": {
        "properties": {
          "Name": {
            "description": "Nested name field",
            "maxLength": 50,
            "minLength": 3,
            "title": "Name",
            "type": "string"
          },
          "Value": {
            "description": "Nested value field",
            "maximum": 100,
            "minimum": 1,
            "title": "Value",
            "type": "integer"
          }
        },
        "required": [
          "Name"
        ],
        "type": "object"
      }
    }
  },
  "info": {
    "title": "API",
    "version": "1.0.0"
  },
  "openapi": "3.1.0",
  "paths": {
    "/resources/{resource_id}/validate": {
      "post": {
        "parameters": [
          {
            "explode": false,
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
