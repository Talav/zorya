package zorya

import (
	"encoding"
	"errors"
	"fmt"
	"math/bits"
	"reflect"

	"github.com/talav/schema"
	"github.com/talav/zorya/metadata"
)

// ErrSchemaInvalid is sent when there is a problem building the schema.
var ErrSchemaInvalid = errors.New("schema is invalid")

const (
	formatInt32           = "int32"
	formatInt64           = "int64"
	contentEncodingBase64 = "base64"
)

var (
	// Interface types for efficient implementation checks without allocation.
	schemaTransformerType = reflect.TypeOf((*SchemaTransformer)(nil)).Elem()
	schemaProviderType    = reflect.TypeOf((*SchemaProvider)(nil)).Elem()
	textUnmarshalerType   = reflect.TypeOf((*encoding.TextUnmarshaler)(nil)).Elem()
)

// SchemaBuilder is an interface that can be implemented by types to build a schema for themselves.
type SchemaBuilder interface {
	SchemaFromType(t reflect.Type) (*Schema, error)
}

// SchemaProvider is an interface that can be implemented by types to provide
// a custom schema for themselves, overriding the built-in schema generation.
// This can be used by custom types with their own special serialization rules.
type SchemaProvider interface {
	Schema(r Registry) *Schema
}

// SchemaTransformer is an interface that can be implemented by types
// to transform the generated schema as needed.
// This can be used to leverage the default schema generation for a type,
// and arbitrarily modify parts of it.
type SchemaTransformer interface {
	TransformSchema(r Registry, s *Schema) *Schema
}

type schemaBuilder struct {
	r Registry
	m *schema.Metadata
}

func newSchemaBuilder(r Registry, m *schema.Metadata) SchemaBuilder {
	return &schemaBuilder{r: r, m: m}
}

func (b *schemaBuilder) SchemaFromType(t reflect.Type) (*Schema, error) {
	s, err := b.schemaFromType(t)
	if err != nil {
		return nil, err
	}

	t = deref(t)

	// Transform generated schema if type implements SchemaTransformer
	// Check without allocation first, then allocate only if needed
	if t.Implements(schemaTransformerType) || reflect.PointerTo(t).Implements(schemaTransformerType) {
		v := reflect.New(t).Interface()
		st, ok := v.(SchemaTransformer)
		if ok {
			s = st.TransformSchema(b.r, s)
		}
	}

	return s, nil
}

func (b *schemaBuilder) schemaFromType(t reflect.Type) (*Schema, error) {
	isPointer := t.Kind() == reflect.Pointer
	s := Schema{}
	t = deref(t)

	// Check for interface implementations that override schema generation
	if schema, err := b.schemaFromInterface(t, isPointer); schema != nil || err != nil {
		return schema, err
	}

	// Lookup in maps (type first, then kind)
	if s := b.schemaForSimpleType(t, isPointer); s != nil {
		return s, nil
	}

	//nolint:exhaustive // Only handling supported Go types for OpenAPI schema generation
	switch t.Kind() {
	case reflect.Slice, reflect.Array:
		return b.schemaForSlice(t, isPointer)
	case reflect.Map:
		s.Type = TypeObject
		s.AdditionalProperties = b.r.Schema(t.Elem(), true, t.Name()+"Value")

		return &s, nil
	case reflect.Struct:
		return b.schemaForStruct(t)
	case reflect.Interface:
		// Interfaces mean any object.
		return &s, nil
	default:
		//nolint:nilnil // Returning nil schema for unsupported types is intentional
		return nil, nil
	}
}

// schemaFromInterface checks if the type implements SchemaProvider or TextUnmarshaler
// and returns the appropriate schema. Returns (nil, nil) if neither is implemented.
func (b *schemaBuilder) schemaFromInterface(t reflect.Type, isPointer bool) (*Schema, error) {
	// Check SchemaProvider without allocation first
	if t.Implements(schemaProviderType) || reflect.PointerTo(t).Implements(schemaProviderType) {
		// Special case: type provides its own schema. Do not try to generate.
		// Only allocate when we know we need to call the method
		v := reflect.New(t).Interface()
		sp, ok := v.(SchemaProvider)
		if !ok {
			return nil, fmt.Errorf("type does not implement SchemaProvider")
		}

		return sp.Schema(b.r), nil
	}

	// Check TextUnmarshaler without allocation (no method call needed)
	if t.Implements(textUnmarshalerType) || reflect.PointerTo(t).Implements(textUnmarshalerType) {
		// Special case: types that implement encoding.TextUnmarshaler are able to
		// be loaded from plain text, and so should be treated as strings.
		// This behavior can be overridden by implementing `SchemaProvider`
		// and returning a custom schema.
		return &Schema{Type: TypeString, Nullable: &isPointer}, nil
	}

	//nolint:nilnil // Returning (nil, nil) signals that no interface implementation was found
	return nil, nil
}

var (
	lookUpByType = map[reflect.Type]*Schema{
		timeType:   {Type: TypeString, Format: "date-time"},
		urlType:    {Type: TypeString, Format: "uri"},
		ipType:     {Type: TypeString, Format: "ipv4"},
		ipAddrType: {Type: TypeString, Format: "ipv4"},
	}

	minZero = 0.0

	lookUpByKind = map[reflect.Kind]*Schema{
		reflect.Bool:    {Type: TypeBoolean},
		reflect.Int8:    {Type: TypeInteger, Format: formatInt32},
		reflect.Int16:   {Type: TypeInteger, Format: formatInt32},
		reflect.Int32:   {Type: TypeInteger, Format: formatInt32},
		reflect.Int64:   {Type: TypeInteger, Format: formatInt64},
		reflect.Uint8:   {Type: TypeInteger, Format: formatInt32, Minimum: &minZero},
		reflect.Uint16:  {Type: TypeInteger, Format: formatInt32, Minimum: &minZero},
		reflect.Uint32:  {Type: TypeInteger, Format: formatInt32, Minimum: &minZero},
		reflect.Uint64:  {Type: TypeInteger, Format: formatInt64, Minimum: &minZero},
		reflect.Float32: {Type: TypeNumber, Format: "float"},
		reflect.Float64: {Type: TypeNumber, Format: "double"},
		reflect.String:  {Type: TypeString},
	}
)

// schemaForSimpleType looks up schema information by type first, then by kind.
// Returns nil if not found.
func (b *schemaBuilder) schemaForSimpleType(t reflect.Type, isPointer bool) *Schema {
	// Try type lookup first (for stdlib types)
	if found, ok := lookUpByType[t]; ok {
		s := *found
		applyNullableForScalar(&s, isPointer)

		return &s
	}

	// Try kind lookup
	kind := t.Kind()
	if kind == reflect.Int || kind == reflect.Uint {
		s := &Schema{Type: TypeInteger}
		if bits.UintSize == 32 {
			s.Format = formatInt32
		} else {
			s.Format = formatInt64
		}
		if kind == reflect.Uint {
			s.Minimum = &minZero
		}
		applyNullableForScalar(s, isPointer)

		return s
	}

	if found, ok := lookUpByKind[kind]; ok {
		s := *found
		applyNullableForScalar(&s, isPointer)

		return &s
	}

	return nil
}

// applyNullableForScalar sets nullable for scalar types if isPointer is true.
func applyNullableForScalar(s *Schema, isPointer bool) {
	if s.Type == TypeBoolean || s.Type == TypeInteger || s.Type == TypeNumber || s.Type == TypeString {
		s.Nullable = &isPointer
	}
}

// schemaForSlice generates a schema for slice or array types.
func (b *schemaBuilder) schemaForSlice(t reflect.Type, isPointer bool) (*Schema, error) {
	s := Schema{}

	if t.Elem().Kind() == reflect.Uint8 {
		// Special case: []byte will be serialized as a base64 string.
		s.Type = TypeString
		s.ContentEncoding = contentEncodingBase64
		s.ContentMediaType = "application/octet-stream" // Describe decoded content
		s.Nullable = &isPointer
	} else {
		s.Type = TypeArray
		s.Nullable = &DefaultArrayNullable
		s.Items = b.r.Schema(t.Elem(), true, t.Name()+"Item")

		if t.Kind() == reflect.Array {
			l := t.Len()
			s.MinItems = &l
			s.MaxItems = &l
		}
	}

	return &s, nil
}

// schemaForStruct generates a schema for struct types.
func (b *schemaBuilder) schemaForStruct(t reflect.Type) (*Schema, error) {
	// Get struct metadata
	structMeta, err := b.m.GetStructMetadata(t)
	if err != nil {
		return nil, fmt.Errorf("failed to get struct metadata for type %s: %w", t, err)
	}

	var required []string
	requiredMap := map[string]bool{}
	propNames := make([]string, 0, len(structMeta.Fields))
	props := map[string]*Schema{}

	s := Schema{Type: TypeObject}

	// Process each field and build properties
	b.processStructFields(t, *structMeta, &s, props, &propNames, &required, requiredMap)

	// Validate dependent required fields
	if err := validateDependentRequired(&s, props); err != nil {
		return nil, err
	}

	// Handle struct-level metadata (_ field)
	applyStructLevelMetadata(&s, structMeta)

	s.Properties = props
	s.propertyNames = propNames
	s.Required = required
	s.requiredMap = requiredMap

	return &s, nil
}

// processStructFields iterates through struct fields and builds property schemas.
func (b *schemaBuilder) processStructFields(
	t reflect.Type,
	structMeta schema.StructMetadata,
	s *Schema,
	props map[string]*Schema,
	propNames *[]string,
	required *[]string,
	requiredMap map[string]bool,
) {
	// Iterate through metadata fields
	for _, fieldMeta := range structMeta.Fields {
		// Extract field name from metadata
		name := extractFieldName(fieldMeta)

		// Determine required status from metadata
		fieldRequired := determineRequired(fieldMeta)

		reflectField := t.Field(fieldMeta.Index)
		fs := b.r.Schema(reflectField.Type, true, t.Name()+fieldMeta.StructFieldName+"Struct")
		if fs == nil {
			continue
		}

		// Apply OpenAPI metadata
		applyOpenAPIMetadata(fs, fieldMeta)

		// Apply validation metadata
		applyValidateMetadata(fs, fieldMeta)

		// If field is required, it cannot be null (required pointer means non-nil in go-playground/validator)
		// Override nullable to false if field is marked as required
		if fieldRequired && fs.Nullable != nil && *fs.Nullable {
			falseVal := false
			fs.Nullable = &falseVal
		}

		// Apply default value from default tag (runtime behavior, not validation)
		applyDefaultValue(fs, fieldMeta)

		// Apply dependent required metadata (on object schema, not field schema)
		applyDependentRequired(s, fieldMeta, name)

		// Add to properties
		props[name] = fs
		*propNames = append(*propNames, name)

		if fieldRequired {
			*required = append(*required, name)
			requiredMap[name] = true
		}
	}
}

// validateDependentRequired validates that all dependent required fields exist.
func validateDependentRequired(s *Schema, props map[string]*Schema) error {
	var errs []error
	for field, dependents := range s.DependentRequired {
		for _, dependent := range dependents {
			if _, ok := props[dependent]; !ok {
				errs = append(errs, fmt.Errorf("dependent field '%s' for field '%s' does not exist", dependent, field))
			}
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("dependent required validation failed: %w", errors.Join(errs...))
	}

	return nil
}

// extractFieldName extracts the field name from metadata.
// Priority: SchemaMetadata.ParamName → StructFieldName.
func extractFieldName(fieldMeta schema.FieldMetadata) string {
	// SchemaMetadata is always present, use ParamName
	if schemaMeta, ok := schema.GetTagMetadata[*schema.SchemaMetadata](&fieldMeta, "schema"); ok {
		return schemaMeta.ParamName
	}

	// Fallback to struct field name
	return fieldMeta.StructFieldName
}

// determineRequired determines if a field is required based on metadata.
// Priority: Hidden fields are never required → SchemaMetadata.Required → ValidateMetadata.Required → default false.
func determineRequired(fieldMeta schema.FieldMetadata) bool {
	// Hidden fields are never required
	if openAPIMeta, ok := schema.GetTagMetadata[*metadata.OpenAPIMetadata](&fieldMeta, "openapi"); ok {
		if openAPIMeta.Hidden != nil && *openAPIMeta.Hidden {
			return false
		}
	}

	// Check SchemaMetadata.Required
	if schemaMeta, ok := schema.GetTagMetadata[*schema.SchemaMetadata](&fieldMeta, "schema"); ok {
		if schemaMeta.Required {
			return true
		}
	}

	// Check ValidateMetadata.Required
	if validateMeta, ok := schema.GetTagMetadata[*metadata.ValidateMetadata](&fieldMeta, "validate"); ok {
		if validateMeta.Required != nil && *validateMeta.Required {
			return true
		}
	}

	// Default to not required
	return false
}

// applyOpenAPIMetadata applies OpenAPI metadata to a schema.
func applyOpenAPIMetadata(fs *Schema, fieldMeta schema.FieldMetadata) {
	openAPIMeta, ok := schema.GetTagMetadata[*metadata.OpenAPIMetadata](&fieldMeta, "openapi")
	if !ok {
		return
	}

	fs.Title = openAPIMeta.Title
	fs.Description = openAPIMeta.Description
	fs.Format = openAPIMeta.Format
	fs.Examples = openAPIMeta.Examples

	// Handle boolean flags
	fs.ReadOnly = openAPIMeta.ReadOnly
	fs.WriteOnly = openAPIMeta.WriteOnly
	fs.Deprecated = openAPIMeta.Deprecated
	if openAPIMeta.Hidden != nil {
		fs.hidden = *openAPIMeta.Hidden
	}
	fs.Extensions = openAPIMeta.Extensions
}

// applyStructLevelMetadata extracts struct-level metadata from the _ field.
// First checks if the field is already parsed in structMeta, otherwise falls back to reflection.
func applyStructLevelMetadata(s *Schema, structMeta *schema.StructMetadata) {
	// Try to get from parsed metadata first (if _ field is exported)
	fieldMeta, ok := structMeta.Field("_")
	if !ok {
		return
	}

	openAPIStructMeta, ok := schema.GetTagMetadata[*metadata.OpenAPIStructMetadata](fieldMeta, "openapiStruct")
	if !ok {
		return
	}

	// Apply struct-level options from parsed metadata
	s.AdditionalProperties = *openAPIStructMeta.AdditionalProperties
	s.Nullable = openAPIStructMeta.Nullable
}

// applyDefaultValue reads the default tag from metadata and applies it to the schema.
// This is separate from validation metadata because defaults are runtime behavior, not validation constraints.
// The value is already parsed by the metadata parser, so we just copy it to the schema.
func applyDefaultValue(fs *Schema, fieldMeta schema.FieldMetadata) {
	defaultMeta, ok := schema.GetTagMetadata[*metadata.DefaultMetadata](&fieldMeta, "default")
	if !ok || defaultMeta.Value == nil {
		return
	}

	fs.Default = defaultMeta.Value
}

// applyValidateMetadata applies validation constraints from ValidateMetadata to a schema.
func applyValidateMetadata(fs *Schema, fieldMeta schema.FieldMetadata) {
	validateMeta, ok := schema.GetTagMetadata[*metadata.ValidateMetadata](&fieldMeta, "validate")
	if !ok {
		return
	}

	// Handle minimum/maximum based on type (go-playground/validator uses same syntax for different semantics)
	applyMinMaxConstraints(fs, validateMeta)

	// Exclusive numeric constraints (only for numbers)
	fs.ExclusiveMinimum = validateMeta.ExclusiveMinimum
	fs.ExclusiveMaximum = validateMeta.ExclusiveMaximum
	fs.MultipleOf = validateMeta.MultipleOf

	// String-specific constraints (from explicit length validators)
	fs.Pattern = validateMeta.Pattern
	// Only set format from validate tag if not already set by openapi tag
	if fs.Format == "" {
		fs.Format = validateMeta.Format
	}

	// Handle enum (already parsed by parser)
	applyEnumConstraints(fs, validateMeta)
}

// applyMinMaxConstraints applies minimum and maximum constraints based on schema type.
func applyMinMaxConstraints(fs *Schema, validateMeta *metadata.ValidateMetadata) {
	switch fs.Type {
	case TypeString:
		applyStringMinMax(fs, validateMeta)
	case TypeInteger, TypeNumber:
		applyNumericMinMax(fs, validateMeta)
	case TypeArray:
		applyArrayMinMax(fs, validateMeta)
	case TypeObject:
		applyObjectMinMax(fs, validateMeta)
	}
}

// applyStringMinMax applies min/max length constraints for string types.
func applyStringMinMax(fs *Schema, validateMeta *metadata.ValidateMetadata) {
	if validateMeta.Minimum != nil {
		minLen := int(*validateMeta.Minimum)
		fs.MinLength = &minLen
	}
	if validateMeta.Maximum != nil {
		maxLen := int(*validateMeta.Maximum)
		fs.MaxLength = &maxLen
	}
}

// applyNumericMinMax applies min/max value constraints for numeric types.
func applyNumericMinMax(fs *Schema, validateMeta *metadata.ValidateMetadata) {
	fs.Minimum = validateMeta.Minimum
	fs.Maximum = validateMeta.Maximum
}

// applyArrayMinMax applies min/max item count constraints for array types.
func applyArrayMinMax(fs *Schema, validateMeta *metadata.ValidateMetadata) {
	if validateMeta.Minimum != nil {
		minItems := int(*validateMeta.Minimum)
		fs.MinItems = &minItems
	}
	if validateMeta.Maximum != nil {
		maxItems := int(*validateMeta.Maximum)
		fs.MaxItems = &maxItems
	}
}

// applyObjectMinMax applies min/max property count constraints for object types.
func applyObjectMinMax(fs *Schema, validateMeta *metadata.ValidateMetadata) {
	if validateMeta.Minimum != nil {
		minProps := int(*validateMeta.Minimum)
		fs.MinProperties = &minProps
	}
	if validateMeta.Maximum != nil {
		maxProps := int(*validateMeta.Maximum)
		fs.MaxProperties = &maxProps
	}
}

// applyEnumConstraints applies enum or const constraints to the schema.
// If enum has exactly one value, use const instead (OpenAPI 3.1 semantic).
// For array types, enum applies to items.
func applyEnumConstraints(fs *Schema, validateMeta *metadata.ValidateMetadata) {
	// Determine target schema (array items or field itself)
	target := fs
	if fs.Type == TypeArray && fs.Items != nil {
		target = fs.Items
	}

	// Apply enum or const based on count
	if len(validateMeta.Enum) == 1 {
		target.Const = validateMeta.Enum[0]
	} else {
		target.Enum = validateMeta.Enum
	}
}

// applyDependentRequired applies dependent required metadata to the object schema.
// dependentRequired is a property of the object schema, not individual field schemas.
// Key: property name that, when present, makes other properties required.
// Value: array of property names that become required when the key property is present.
func applyDependentRequired(s *Schema, fieldMeta schema.FieldMetadata, fieldName string) {
	depMeta, ok := schema.GetTagMetadata[*metadata.DependentRequiredMetadata](&fieldMeta, "dependentRequired")
	if !ok || len(depMeta.Dependents) == 0 {
		return
	}

	if s.DependentRequired == nil {
		s.DependentRequired = make(map[string][]string)
	}
	s.DependentRequired[fieldName] = depMeta.Dependents
}
