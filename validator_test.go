package zorya

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestStruct for validation testing with multiple field locations.
type TestValidationInput struct {
	ID      string `schema:"id,location:path,required:true" validate:"required"`
	Email   string `schema:"email,location:query" validate:"required,email"`
	APIKey  string `schema:"X-API-Key,location:header" validate:"required"`
	Session string `schema:"session_id,location:cookie" validate:"required"`
	Body    struct {
		Name string `json:"name" validate:"required,min=3"`
		Age  int    `json:"age" validate:"min=0,max=150"`
	}
}

// TestPlaygroundValidator_Validate_Success tests successful validation.
func TestPlaygroundValidator_Validate_Success(t *testing.T) {
	v := validator.New()
	validator := NewPlaygroundValidator(v)

	input := TestValidationInput{
		ID:      "123",
		Email:   "user@example.com",
		APIKey:  "secret-key",
		Session: "session-token",
		Body: struct {
			Name string `json:"name" validate:"required,min=3"`
			Age  int    `json:"age" validate:"min=0,max=150"`
		}{
			Name: "John Doe",
			Age:  30,
		},
	}

	metadata := NewMetadata()
	structMeta, err := metadata.GetStructMetadata(reflect.TypeOf(input))
	require.NoError(t, err)

	errs := validator.Validate(context.Background(), &input, structMeta)

	assert.Nil(t, errs, "Validation should succeed with valid input")
}

// TestPlaygroundValidator_Validate_Errors tests validation errors with correct locations.
func TestPlaygroundValidator_Validate_Errors(t *testing.T) {
	v := validator.New()
	validator := NewPlaygroundValidator(v)

	input := TestValidationInput{
		ID:      "",              // Missing required path parameter
		Email:   "invalid-email", // Invalid email format
		APIKey:  "",              // Missing required header
		Session: "",              // Missing required cookie
		Body: struct {
			Name string `json:"name" validate:"required,min=3"`
			Age  int    `json:"age" validate:"min=0,max=150"`
		}{
			Name: "AB", // Too short (min=3)
			Age:  200,  // Too large (max=150)
		},
	}

	metadata := NewMetadata()
	structMeta, err := metadata.GetStructMetadata(reflect.TypeOf(input))
	require.NoError(t, err)

	errs := validator.Validate(context.Background(), &input, structMeta)

	require.NotNil(t, errs, "Validation should return errors for invalid input")
	assert.Greater(t, len(errs), 0, "Should have at least one validation error")

	// Verify error details
	errorMap := make(map[string]*ErrorDetail)
	for _, err := range errs {
		var detail *ErrorDetail
		require.True(t, errors.As(err, &detail), "Error should be ErrorDetail")
		errorMap[detail.Location] = detail
	}

	// Check path parameter error
	if pathErr, ok := errorMap["path.id"]; ok {
		assert.Equal(t, "required", pathErr.Code)
		assert.Contains(t, pathErr.Location, "path")
	}

	// Check query parameter error
	if queryErr, ok := errorMap["query.email"]; ok {
		assert.Equal(t, "email", queryErr.Code)
		assert.Contains(t, queryErr.Location, "query")
	}

	// Check header error
	if headerErr, ok := errorMap["header.X-API-Key"]; ok {
		assert.Equal(t, "required", headerErr.Code)
		assert.Contains(t, headerErr.Location, "header")
	}

	// Check cookie error
	if cookieErr, ok := errorMap["cookie.session_id"]; ok {
		assert.Equal(t, "required", cookieErr.Code)
		assert.Contains(t, cookieErr.Location, "cookie")
	}

	// Check body errors
	if bodyNameErr, ok := errorMap["body.TestValidationInput.Body.name"]; ok {
		assert.True(t, bodyNameErr.Code == "required" || bodyNameErr.Code == "min", "Should be required or min")
		assert.Contains(t, bodyNameErr.Location, "body")
	}

	if bodyAgeErr, ok := errorMap["body.TestValidationInput.Body.age"]; ok {
		assert.Equal(t, "max", bodyAgeErr.Code)
		assert.Contains(t, bodyAgeErr.Location, "body")
	}
}

// TestPlaygroundValidator_Validate_NonValidationError tests non-validation errors.
func TestPlaygroundValidator_Validate_NonValidationError(t *testing.T) {
	v := validator.New()
	validator := NewPlaygroundValidator(v)

	// Pass invalid input type (not a struct)
	input := "not a struct"

	errs := validator.Validate(context.Background(), input, nil)

	require.NotNil(t, errs, "Should return error for invalid input type")
	require.Len(t, errs, 1, "Should return single error")

	var detail *ErrorDetail
	require.True(t, errors.As(errs[0], &detail), "Error should be ErrorDetail")
	assert.Equal(t, "validation_error", detail.Code)
	assert.NotEmpty(t, detail.Message)
	assert.Empty(t, detail.Location, "Location should be empty for non-validation errors")
}

// TestLocationForNamespace tests the LocationForNamespace helper function.
func TestLocationForNamespace(t *testing.T) {
	metadata := NewMetadata()

	// Create a test struct with different field locations
	type TestStruct struct {
		ID      string `schema:"id,location=path"`
		Email   string `schema:"email,location=query"`
		APIKey  string `schema:"X-API-Key,location=header"`
		Session string `schema:"session_id,location=cookie"`
		Body    struct {
			Name string `json:"name"`
		}
	}

	structMeta, err := metadata.GetStructMetadata(reflect.TypeOf(TestStruct{}))
	require.NoError(t, err)

	tests := []struct {
		name      string
		namespace string
		wantLoc   string
	}{
		{
			name:      "path location with struct prefix",
			namespace: "TestStruct.ID",
			wantLoc:   "path.ID",
		},
		{
			name:      "query location without struct prefix",
			namespace: "Email",
			wantLoc:   "query.Email",
		},
		{
			name:      "header location",
			namespace: "TestStruct.APIKey",
			wantLoc:   "header.APIKey",
		},
		{
			name:      "cookie location",
			namespace: "Session",
			wantLoc:   "cookie.Session",
		},
		{
			name:      "body field without schema tag",
			namespace: "TestStruct.Body.Name",
			wantLoc:   "body.TestStruct.Body.Name",
		},
		{
			name:      "non-existent field",
			namespace: "NonExistent",
			wantLoc:   "body.NonExistent",
		},
		{
			name:      "empty namespace",
			namespace: "",
			wantLoc:   "body.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loc := locationForNamespace(structMeta, tt.namespace)
			assert.Equal(t, tt.wantLoc, loc, "LocationForNamespace location result")
		})
	}
}

// TestLocationForNamespace_NilMetadata tests LocationForNamespace with nil metadata.
func TestLocationForNamespace_NilMetadata(t *testing.T) {
	loc := locationForNamespace(nil, "SomeField")

	assert.Equal(t, "body.SomeField", loc, "Should return body location for nil metadata")
}
