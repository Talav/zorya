package metadata

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/talav/tagparser"
)

// OpenAPIMetadata represents OpenAPI-specific field-level schema metadata extracted from the openapi tag.
// Types match OpenAPI v3.0 specification for schema metadata.
// This metadata is used to generate OpenAPI schema properties that are not validation constraints
// but API contract metadata (e.g., readOnly, writeOnly, deprecated, title, description, examples).
type OpenAPIMetadata struct {
	// API contract metadata (not validation constraints)
	// OpenAPI v3.0: readOnly, writeOnly, deprecated are booleans
	ReadOnly    *bool  // field is read-only
	WriteOnly   *bool  // field is write-only
	Deprecated  *bool  // field is deprecated
	Hidden      *bool  // field is hidden from schema (not included in properties)
	Title       string // title for the schema
	Description string // description for the schema
	Format      string // format for the schema (e.g., "date", "date-time", "time", "email", "uri")
	Examples    []any  // parsed example values

	// Extensions are OpenAPI specification extensions (x-* fields).
	// Keys must start with "x-" per OpenAPI spec requirement.
	Extensions map[string]any
}

// ParseOpenAPITag parses an openapi tag and returns OpenAPIMetadata.
// Tag format: openapi:"readOnly,writeOnly,deprecated,hidden,title=My Title,description=My description,example=123,x-custom=value"
//
// This parser:
// 1. Parses tag format (comma-separated, key=value pairs or flags)
// 2. Converts string values to proper OpenAPI types (bool for readOnly/writeOnly/deprecated/hidden)
// 3. Converts empty string to true for boolean flags (e.g., "readOnly" -> ReadOnly=true)
// 4. Routes x-* prefixed keys to Extensions map (OpenAPI spec requirement)
//
// Supported options:
//   - readOnly -> ReadOnly=true
//   - writeOnly -> WriteOnly=true
//   - deprecated -> Deprecated=true
//   - hidden -> Hidden=true (field excluded from schema properties)
//   - title=... -> Title="..."
//   - description=... -> Description="..."
//   - format=... -> Format="..." (e.g., "date", "date-time", "time", "email", "uri")
//   - example=... -> Examples=[value] (converted to array)
//   - examples=... -> Examples=[...] (parsed from JSON array string)
//   - x-* -> Extensions["x-*"]="..." (OpenAPI extensions, MUST start with x-)
func ParseOpenAPITag(field reflect.StructField, index int, tagValue string) (any, error) {
	om := &OpenAPIMetadata{}

	// Parse tag using tagparser (options mode - all items are options)
	tag, err := tagparser.Parse(tagValue)
	if err != nil {
		return nil, fmt.Errorf("field %s: failed to parse openapi tag: %w", field.Name, err)
	}

	// Process all options
	for key, value := range tag.Options {
		if err := applyOpenAPIMapping(om, key, value); err != nil {
			return nil, fmt.Errorf("field %s: failed to apply openapi mapping: %w", field.Name, err)
		}
	}

	return om, nil
}

// applyOpenAPIMapping maps a single openapi tag option to OpenAPIMetadata field.
// Extensions with x- prefix and length > 3 are processed, others are ignored.
//
//nolint:cyclop // Switch statement for field mapping - acceptable complexity
func applyOpenAPIMapping(om *OpenAPIMetadata, key, value string) error {
	// Process OpenAPI extensions: must have "x-" prefix and be longer than 3 chars
	if strings.HasPrefix(key, "x-") && len(key) > 3 {
		// Initialize Extensions map if needed
		if om.Extensions == nil {
			om.Extensions = make(map[string]any)
		}

		om.Extensions[key] = value

		return nil
	}

	// Handle standard OpenAPI fields
	switch key {
	// Boolean flags (empty string means true)
	case "readOnly":
		om.ReadOnly = parseBool(value) // flag without value -> true (user intent)
	case "writeOnly":
		om.WriteOnly = parseBool(value) // flag without value -> true (user intent)
	case "deprecated":
		om.Deprecated = parseBool(value) // flag without value -> true (user intent)
	case "hidden":
		om.Hidden = parseBool(value) // flag without value -> true (user intent)

	// String values
	case "title":
		om.Title = value
	case "description":
		om.Description = value
	case "format":
		om.Format = value
	case "example":
		// Convert single example to examples array
		// example=25 -> Examples=[25], example="hello" -> Examples=["hello"]
		// Only set if examples hasn't been set yet (examples takes precedence)
		if om.Examples == nil {
			// Try to detect if value is numeric
			if num, err := strconv.ParseFloat(value, 64); err == nil {
				// Valid number - store as float64 (JSON numbers are float64)
				om.Examples = []any{num}
			} else {
				// Not a number - store as string
				om.Examples = []any{value}
			}
		}
	case "examples":
		// Examples should be a JSON array string
		// User provides: examples='["val1","val2"]' or examples='[1,2,3]'
		// Examples takes precedence over example
		var examples []any
		if err := json.Unmarshal([]byte(value), &examples); err != nil {
			return fmt.Errorf("failed to parse examples JSON: %w", err)
		}
		om.Examples = examples
	}

	return nil
}
