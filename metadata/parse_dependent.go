package metadata

import (
	"fmt"
	"reflect"

	"github.com/talav/tagparser"
)

// DependentRequiredMetadata represents dependent required fields extracted from the dependentRequired tag.
// This is used for OpenAPI schema generation to specify that certain fields must be present
// when this field is present (JSON Schema dependentRequired keyword).
type DependentRequiredMetadata struct {
	Dependents []string // List of field names that become required when this field is present
}

// ParseDependentRequiredTag parses a dependentRequired tag and returns DependentRequiredMetadata.
// Tag format: dependentRequired:"field1,field2,field3"
//
// Parses comma-separated list of dependent field names. Empty strings and whitespace are filtered out.
// Example:
//   - dependentRequired:"billing_address,cardholder_name" -> Dependents=["billing_address", "cardholder_name"]
//   - dependentRequired:"field1" -> Dependents=["field1"]
//   - dependentRequired:"" -> Dependents=[] (empty, will be ignored)
func ParseDependentRequiredTag(field reflect.StructField, index int, tagValue string) (any, error) {
	tag, err := tagparser.Parse(tagValue)
	if err != nil {
		return nil, fmt.Errorf("field %s: failed to parse dependent required tag: %w", field.Name, err)
	}

	dependents := make([]string, 0, len(tag.Options))
	for key := range tag.Options {
		dependents = append(dependents, key)
	}

	return &DependentRequiredMetadata{
		Dependents: dependents,
	}, nil
}
