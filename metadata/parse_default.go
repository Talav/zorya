package metadata

import (
	"encoding/json"
	"fmt"
	"reflect"
)

// DefaultMetadata represents the default value extracted from the default tag.
// This is used for both runtime default application (via mapstructure) and
// OpenAPI schema generation.
type DefaultMetadata struct {
	Value any // Parsed default value (JSON-compatible types: string, float64, bool, []any, map[string]any)
}

// ParseDefaultTag parses a default tag and returns DefaultMetadata.
// Tag format: default:"value".
//
// Parses the default value based on the field type:
//   - Strings: returned as-is (no quotes needed in tag)
//   - Numbers, booleans, arrays, objects: parsed as JSON (e.g., default:"[1,2,3]" or default:"{\"key\":\"value\"}")
func ParseDefaultTag(field reflect.StructField, index int, tagValue string) (any, error) {
	parsedValue, err := parseDefaultValue(field.Type, tagValue)
	if err != nil {
		return nil, fmt.Errorf("field %s: failed to parse default value %q: %w", field.Name, tagValue, err)
	}

	return &DefaultMetadata{
		Value: parsedValue,
	}, nil
}

// parseDefaultValue parses a default value string based on the Go field type.
func parseDefaultValue(fieldType reflect.Type, value string) (any, error) {
	// Dereference pointer types
	isPointer := fieldType.Kind() == reflect.Ptr
	if isPointer {
		fieldType = fieldType.Elem()
	}

	// Special case: strings don't need quotes
	if fieldType.Kind() == reflect.String {
		return value, nil
	}

	// All other types require JSON format
	var v any
	if err := json.Unmarshal([]byte(value), &v); err != nil {
		return nil, fmt.Errorf("invalid JSON for type %s: %w", fieldType, err)
	}

	// Validate the parsed value matches the field type
	if err := validateDefaultType(fieldType, v); err != nil {
		return nil, fmt.Errorf("value %v does not match type %s: %w", v, fieldType, err)
	}

	return v, nil
}

// validateDefaultType validates that the parsed value matches the Go field type.
func validateDefaultType(fieldType reflect.Type, value any) error {
	//nolint:exhaustive // Only validating types that can have default values
	switch fieldType.Kind() {
	case reflect.Bool:
		if _, ok := value.(bool); !ok {
			return fmt.Errorf("expected bool, got %T", value)
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		// JSON unmarshals numbers as float64
		if _, ok := value.(float64); !ok {
			return fmt.Errorf("expected number, got %T", value)
		}
	case reflect.Slice, reflect.Array:
		if _, ok := value.([]any); !ok {
			return fmt.Errorf("expected array, got %T", value)
		}
	case reflect.Map, reflect.Struct:
		if _, ok := value.(map[string]any); !ok {
			return fmt.Errorf("expected object, got %T", value)
		}
	}

	return nil
}
