package zorya

import (
	"context"
	"errors"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/talav/schema"
)

// PlaygroundValidator adapts go-playground/validator to Zorya's Validator interface.
type PlaygroundValidator struct {
	validate *validator.Validate
}

// NewPlaygroundValidator creates a new validator adapter for go-playground/validator.
func NewPlaygroundValidator(v *validator.Validate) *PlaygroundValidator {
	return &PlaygroundValidator{validate: v}
}

// Validate validates the input struct using go-playground/validator.
// Returns validation errors as ErrorDetail objects with code, message, and location.
func (v *PlaygroundValidator) Validate(ctx context.Context, input any, metadata *schema.StructMetadata) []error {
	err := v.validate.StructCtx(ctx, input)
	if err == nil {
		return nil
	}

	var validationErrors validator.ValidationErrors
	if !errors.As(err, &validationErrors) {
		// Non-validation error (e.g., invalid input type)
		return []error{&ErrorDetail{
			Code:    "validation_error",
			Message: err.Error(),
		}}
	}

	errs := make([]error, len(validationErrors))
	for i, e := range validationErrors {
		location := locationForNamespace(metadata, e.Namespace())

		errs[i] = &ErrorDetail{
			Code:     e.Tag(),   // "required", "email", "min", or custom tag
			Message:  e.Error(), // Human-readable message from validator
			Location: location,  // "query.email", "path.id", "header.auth", "body.User.email"
		}
	}

	return errs
}

// locationForNamespace calculates the full location path for a validator namespace.
// Namespace format is typically "StructName.FieldName" or "FieldName" for top-level fields.
// Returns the full location path (e.g., "query.email", "path.id", "body.User.email").
func locationForNamespace(metadata *schema.StructMetadata, namespace string) string {
	if metadata == nil {
		return "body." + namespace
	}

	// Remove struct name prefix if present (e.g., "User.Email" -> "Email")
	fieldName := namespace
	if parts := strings.Split(namespace, "."); len(parts) > 1 {
		fieldName = parts[len(parts)-1]
	}

	// Look up field by name
	field, ok := metadata.Field(fieldName)
	if !ok {
		return "body." + namespace
	}

	// Get SchemaMetadata from field's TagMetadata
	schemaMeta, ok := schema.GetTagMetadata[*schema.SchemaMetadata](field, "schema")
	if !ok {
		return "body." + namespace
	}

	// Construct full location path
	// Remove struct name prefix from namespace if present for the path component
	namespacePath := namespace
	if parts := strings.Split(namespace, "."); len(parts) > 1 {
		namespacePath = strings.Join(parts[1:], ".")
	}

	return string(schemaMeta.Location) + "." + namespacePath
}
