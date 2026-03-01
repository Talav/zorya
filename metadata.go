package zorya

import (
	"net/http"
	"reflect"

	"github.com/talav/openapi/metadata"
	"github.com/talav/schema"
)

// NewMetadata returns a schema.Metadata configured with Zorya's tag parsers (schema, body, openapi, validate, default, requires).
func NewMetadata() *schema.Metadata {
	return schema.NewMetadata(schema.NewTagParserRegistry(
		schema.WithTagParser("schema", schema.ParseSchemaTag, conditionalSchemaDefault),
		schema.WithTagParser("body", schema.ParseBodyTag),
		schema.WithTagParser("openapi", metadata.ParseOpenAPITag),
		schema.WithTagParser("validate", metadata.ParseValidateTag),
		schema.WithTagParser("default", metadata.ParseDefaultTag),
		schema.WithTagParser("requires", metadata.ParseRequiresTag),
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
// If none, falls back to a field named "Body" with type func(http.ResponseWriter) error (streaming).
// Returns nil if no body field is found.
func FindBodyField(structMeta *schema.StructMetadata) *schema.FieldMetadata {
	for i := range structMeta.Fields {
		if structMeta.Fields[i].HasTag("body") {
			return &structMeta.Fields[i]
		}
	}
	// Fallback: streaming body func(w http.ResponseWriter) error without body tag
	for i := range structMeta.Fields {
		f := &structMeta.Fields[i]
		if f.StructFieldName == "Body" && isStreamingBodyFunc(f.Type) {
			return f
		}
	}
	return nil
}

// isStreamingBodyFunc reports whether t is func(http.ResponseWriter) error.
func isStreamingBodyFunc(t reflect.Type) bool {
	if t == nil || t.Kind() != reflect.Func {
		return false
	}
	if t.NumIn() != 1 || t.NumOut() != 1 {
		return false
	}
	httpResponseWriterType := reflect.TypeOf((*http.ResponseWriter)(nil)).Elem()
	errorType := reflect.TypeOf((*error)(nil)).Elem()
	return t.In(0).Implements(httpResponseWriterType) && t.Out(0).Implements(errorType)
}
