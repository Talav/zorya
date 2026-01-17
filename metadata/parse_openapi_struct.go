package metadata

import (
	"fmt"
	"reflect"

	"github.com/talav/tagparser"
)

// OpenAPIStructMetadata represents OpenAPI-specific struct-level schema metadata extracted from the openapiStruct tag.
// This metadata is used to configure struct-level schema properties (e.g., additionalProperties, nullable).
type OpenAPIStructMetadata struct {
	AdditionalProperties *bool // allow additional properties
	Nullable             *bool // struct is nullable
}

// ParseOpenAPIStructTag parses an openapiStruct tag and returns OpenAPIStructMetadata.
// Tag format: openapiStruct:"additionalProperties=true,nullable=false"
//
// This parser is used for struct-level OpenAPI schema configuration.
// It should be used on the _ field of a struct to configure struct-level schema properties.
//
// Supported options:
//   - additionalProperties=true/false -> AdditionalProperties=bool
//   - nullable=true/false -> Nullable=bool
func ParseOpenAPIStructTag(field reflect.StructField, index int, tagValue string) (any, error) {
	osm := &OpenAPIStructMetadata{}

	// Parse tag using tagparser (options mode - all items are options)
	tag, err := tagparser.Parse(tagValue)
	if err != nil {
		return nil, fmt.Errorf("field %s: failed to parse openapiStruct tag: %w", field.Name, err)
	}

	// Process all options
	for key, value := range tag.Options {
		switch key {
		case "additionalProperties":
			osm.AdditionalProperties = parseBool(value)
		case "nullable":
			osm.Nullable = parseBool(value)
		default:
			// Ignore unknown options for struct-level metadata
		}
	}

	return osm, nil
}
