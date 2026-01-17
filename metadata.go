package zorya

import (
	"reflect"

	"github.com/talav/schema"
	"github.com/talav/zorya/metadata"
)

func NewMetadata() *schema.Metadata {
	return schema.NewMetadata(schema.NewTagParserRegistry(
		schema.WithTagParser("schema", schema.ParseSchemaTag, conditionalSchemaDefault),
		schema.WithTagParser("body", schema.ParseBodyTag),
		schema.WithTagParser("openapi", metadata.ParseOpenAPITag),
		schema.WithTagParser("openapiStruct", metadata.ParseOpenAPIStructTag),
		schema.WithTagParser("validate", metadata.ParseValidateTag),
		schema.WithTagParser("default", metadata.ParseDefaultTag),
		schema.WithTagParser("dependentRequired", metadata.ParseDependentRequiredTag),
	))
}

// conditionalSchemaDefault applies schema default metadata only if the field doesn't have a body tag.
// Business rule: fields with body tags should not receive default schema metadata.
func conditionalSchemaDefault(field reflect.StructField, index int) any {
	// Don't apply schema default if field has body tag
	if _, ok := field.Tag.Lookup("body"); ok {
		return nil
	}

	return schema.DefaultSchemaMetadata(field, index)
}

// FindBodyField finds the field with "body" tag in the struct metadata.
// Returns nil if no body field is found.
func FindBodyField(structMeta *schema.StructMetadata) *schema.FieldMetadata {
	for i := range structMeta.Fields {
		if structMeta.Fields[i].HasTag("body") {
			return &structMeta.Fields[i]
		}
	}

	return nil
}
