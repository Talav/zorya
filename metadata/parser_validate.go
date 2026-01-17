package metadata

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/talav/tagparser"
)

// ValidateMetadata represents validation constraints extracted from the validate tag.
// Types match OpenAPI v3.0 specification for schema validation constraints.
// This metadata is used to generate OpenAPI schema constraints by mapping
// go-playground/validator tags to OpenAPI/JSON Schema keywords.
type ValidateMetadata struct {
	// Numeric validation constraints (for number/integer types)
	// OpenAPI v3.0: minimum, maximum, exclusiveMinimum, exclusiveMaximum, multipleOf are numbers
	Minimum          *float64 // inclusive minimum value
	ExclusiveMinimum *float64 // exclusive minimum value (value must be > exclusiveMinimum)
	Maximum          *float64 // inclusive maximum value
	ExclusiveMaximum *float64 // exclusive maximum value (value must be < exclusiveMaximum)
	MultipleOf       *float64 // value must be a multiple of this number

	// String validation constraints (for string types)
	Pattern string // regular expression pattern that string must match
	Format  string // predefined format for string validation (e.g., "email", "date-time", "uri")

	// General validation constraints
	Enum     []any // parsed enum values
	Required *bool // field must be present
}

// ParseValidateTag parses a validate tag in go-playground/validator format and returns ValidateMetadata.
// Tag format: validate:"required,email,min=5,max=100,pattern=^[a-z]+$"
//
// This parser:
// 1. Parses go-playground/validator tag format (comma-separated, key=value pairs)
// 2. Maps validator tags to OpenAPI/JSON Schema constraints
// 3. Converts string values to proper OpenAPI types (int, float64, bool)
// 4. Returns error if value cannot be parsed to expected type
//
// Validator tag -> OpenAPI mapping:
//   - required -> Required=true
//   - min=N -> Minimum=N (as float64)
//   - max=N -> Maximum=N (as float64)
//   - len=N -> Minimum=N, Maximum=N (as float64, sets both to same value)
//   - email -> Format="email"
//   - url -> Format="uri"
//   - pattern=... -> Pattern="..."
//   - oneof=... -> Enum="[...]"
//   - etc.
func ParseValidateTag(field reflect.StructField, index int, tagValue string) (any, error) {
	vm := &ValidateMetadata{}

	// Parse go-playground/validator format using tagparser
	// Format: "required,email,min=5,max=100"
	// Use ParseFunc to handle all items, including flags without values
	allValidators := make(map[string]string)

	tag, err := tagparser.Parse(tagValue)
	if err != nil {
		return nil, fmt.Errorf("field %s: failed to parse validate tag: %w", field.Name, err)
	}

	for key, value := range tag.Options {
		if key == "" {
			// First item without equals sign (flag without value)
			allValidators[value] = ""
		} else {
			// Key=value pair
			allValidators[key] = value
		}
	}

	// Map validator tags to OpenAPI constraints
	for validator, value := range allValidators {
		if err := applyValidatorMapping(vm, validator, value); err != nil {
			return nil, fmt.Errorf("field %s: failed to apply validator %q: %w", field.Name, validator, err)
		}
	}

	return vm, nil
}

// applyValidatorMapping maps a single validator tag to OpenAPI constraint.
// Only includes validators actually supported by go-playground/validator v10.
// Reference: https://pkg.go.dev/github.com/go-playground/validator/v10
//
//nolint:cyclop // Large switch statement for validator mapping - acceptable complexity
func applyValidatorMapping(vm *ValidateMetadata, validator, value string) error {
	switch validator {
	// Boolean flags (empty string means true)
	case "required": // Standard - field must be present
		vm.Required = parseBool(value)

	// Numeric constraints (parse as float64 for OpenAPI)
	case "min": // Standard - minimum value (works for numbers, strings, slices, maps)
		f, err := parseFloat64(value)
		if err != nil {
			return fmt.Errorf("invalid min value %q: %w", value, err)
		}
		vm.Minimum = &f
	case "max": // Standard - maximum value (works for numbers, strings, slices, maps)
		f, err := parseFloat64(value)
		if err != nil {
			return fmt.Errorf("invalid max value %q: %w", value, err)
		}
		vm.Maximum = &f
	case "gte": // Standard - greater than or equal
		f, err := parseFloat64(value)
		if err != nil {
			return fmt.Errorf("invalid gte value %q: %w", value, err)
		}
		vm.Minimum = &f
	case "lte": // Standard - less than or equal
		f, err := parseFloat64(value)
		if err != nil {
			return fmt.Errorf("invalid lte value %q: %w", value, err)
		}
		vm.Maximum = &f
	case "gt": // Standard - greater than
		f, err := parseFloat64(value)
		if err != nil {
			return fmt.Errorf("invalid gt value %q: %w", value, err)
		}
		vm.ExclusiveMinimum = &f
	case "lt": // Standard - less than
		f, err := parseFloat64(value)
		if err != nil {
			return fmt.Errorf("invalid lt value %q: %w", value, err)
		}
		vm.ExclusiveMaximum = &f
	case "multiple_of": // Standard - value must be a multiple of this number
		f, err := parseFloat64(value)
		if err != nil {
			return fmt.Errorf("invalid multiple_of value %q: %w", value, err)
		}
		vm.MultipleOf = &f

	// String length constraints (parse as int for OpenAPI)
	case "len": // Standard - exact length
		i, err := parseFloat64(value)
		if err != nil {
			return fmt.Errorf("invalid len value %q: %w", value, err)
		}
		vm.Minimum = &i
		vm.Maximum = &i

	// String format constraints
	case "email": // Standard - valid email address
		vm.Format = "email"
	case "url": // Standard - valid URL
		vm.Format = "uri"
	case "alpha": // Standard - alphabetic characters only
		vm.Pattern = "^[a-zA-Z]+$"
	case "alphanum": // Standard - alphanumeric characters only
		vm.Pattern = "^[a-zA-Z0-9]+$"
	case "alphaunicode": // Standard - unicode alphabetic characters only
		vm.Pattern = "^[\\p{L}]+$"
	case "alphanumunicode": // Standard - unicode alphanumeric characters only
		vm.Pattern = "^[\\p{L}\\p{N}]+$"
	case "pattern": // Standard - regular expression pattern
		vm.Pattern = value

	// Enum/oneof
	case "oneof": // Standard - value must be one of the specified values
		// oneof=red green blue -> Enum=["red","green","blue"]
		// Parse space-separated values into []any
		value = strings.TrimSpace(value)
		if value == "" {
			return fmt.Errorf("oneof requires at least one value")
		}
		parts := strings.Fields(value)
		enumValues := make([]any, 0, len(parts))
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if part != "" {
				enumValues = append(enumValues, part)
			}
		}
		vm.Enum = enumValues
	}

	return nil
}
