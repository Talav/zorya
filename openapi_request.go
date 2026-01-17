package zorya

import (
	"fmt"
	"reflect"

	"github.com/talav/schema"
	"github.com/talav/zorya/metadata"
)

// requestSchemaExtractor extracts OpenAPI request schemas from input struct types.
// It populates operation Parameters and RequestBody based on struct field tags.
type requestSchemaExtractor struct {
	registry Registry
	metadata *schema.Metadata
}

// NewRequestSchemaExtractor creates a new request schema extractor.
func NewRequestSchemaExtractor(registry Registry, metadata *schema.Metadata) *requestSchemaExtractor {
	return &requestSchemaExtractor{
		registry: registry,
		metadata: metadata,
	}
}

// RequestFromType extracts OpenAPI request schemas from an input struct type
// and populates the operation's Parameters and RequestBody.
//
// This handles:
// - Parameters: Generated from fields with "schema" tag (path, query, header, cookie)
// - Request body: Generated from field with "body" tag
// - Content type: Determined from body field type (defaults to application/json)
// - Required fields: Set based on field type (non-pointer = required) and metadata.
func (e *requestSchemaExtractor) RequestFromType(inputType reflect.Type, op *Operation) error {
	// Get struct metadata (parsed and cached)
	structMeta, err := e.metadata.GetStructMetadata(inputType)
	if err != nil {
		return fmt.Errorf("failed to get struct metadata for type %s: %w", inputType, err)
	}

	// Initialize operation fields if needed
	if op.Parameters == nil {
		op.Parameters = make([]*Param, 0)
	}

	// Process parameters (fields with "schema" tag, excluding body)
	// Parameters can be in path, query, header, or cookie locations
	e.extractParameters(structMeta, op, inputType)

	// Process request body (field with "body" tag)
	// Body is handled separately as it's not a parameter
	if err := e.extractRequestBody(structMeta, op, inputType); err != nil {
		return fmt.Errorf("failed to extract request body: %w", err)
	}

	return nil
}

// extractParameters extracts OpenAPI parameters from struct fields with "schema" tag.
// Skips fields with "body" tag (handled separately).
// Only processes valid parameter locations: path, query, header, cookie.
func (e *requestSchemaExtractor) extractParameters(structMeta *schema.StructMetadata, op *Operation, inputType reflect.Type) {
	for i := range structMeta.Fields {
		field := &structMeta.Fields[i]

		// Get schema metadata (must have "schema" tag)
		schemaMeta, ok := schema.GetTagMetadata[*schema.SchemaMetadata](field, "schema")
		if !ok {
			continue
		}

		// Generate schema for parameter type
		hint := getRequestHint(inputType, field.StructFieldName, op.OperationID+"Request")
		paramSchema := e.registry.Schema(field.Type, true, hint)
		if paramSchema == nil {
			continue
		}

		// Get description from openapi metadata if available
		description := ""
		if openAPIMeta, ok := schema.GetTagMetadata[*metadata.OpenAPIMetadata](field, "openapi"); ok {
			description = openAPIMeta.Description
		}

		// Create and add parameter
		op.Parameters = append(op.Parameters, &Param{
			Name:        schemaMeta.ParamName,
			Description: description,
			In:          string(schemaMeta.Location),
			Required:    schemaMeta.Required,
			Schema:      paramSchema,
			Style:       string(schemaMeta.Style),
			Explode:     &schemaMeta.Explode,
		})
	}
}

// extractRequestBody extracts OpenAPI request body from struct field with "body" tag.
// Initializes RequestBody if needed and sets content type and schema.
func (e *requestSchemaExtractor) extractRequestBody(structMeta *schema.StructMetadata, op *Operation, inputType reflect.Type) error {
	// Find body field by checking for "body" tag
	bodyField := FindBodyField(structMeta)
	// No body field - nothing to do
	if bodyField == nil {
		return nil
	}

	// Initialize RequestBody if needed
	initRequestBody(op)

	// Get body metadata
	bodyMeta, ok := schema.GetTagMetadata[*schema.BodyMetadata](bodyField, "body")
	if !ok {
		return fmt.Errorf("body field missing body metadata")
	}

	// Set required based on metadata or field type
	// Non-pointer, non-interface types are required by default
	if bodyMeta.Required || (bodyField.Type.Kind() != reflect.Pointer && bodyField.Type.Kind() != reflect.Interface) {
		op.RequestBody.Required = true
	}

	// Determine content type based on BodyType
	contentType := getContentType(bodyMeta.BodyType)

	// Initialize content map entry if needed
	if op.RequestBody.Content[contentType] == nil {
		op.RequestBody.Content[contentType] = &MediaType{}
	}

	// Generate schema for body type
	hint := getRequestHint(inputType, bodyField.StructFieldName, op.OperationID+"Request")

	// For multipart, always inline the schema (don't use ref) so we can transform it
	// Also mark it as inline-only to prevent it from appearing in components
	forceInline := bodyMeta.BodyType == schema.BodyTypeMultipart
	if forceInline {
		e.registry.MarkInlineOnly(bodyField.Type, hint)
	}
	bodySchema := e.registry.Schema(bodyField.Type, !forceInline, hint)

	if bodySchema != nil {
		// Transform schema for multipart if needed
		if bodyMeta.BodyType == schema.BodyTypeMultipart {
			bodySchema = transformSchemaForMultipart(bodySchema)
			// Add encoding object for binary fields
			op.RequestBody.Content[contentType].Encoding = extractMultipartEncoding(bodySchema)
		}
		op.RequestBody.Content[contentType].Schema = bodySchema
	}

	return nil
}

// initRequestBody initializes the RequestBody on the operation if it's nil.
// Creates an empty Content map ready for media type entries.
func initRequestBody(op *Operation) {
	if op.RequestBody == nil {
		op.RequestBody = &RequestBody{
			Content: make(map[string]*MediaType),
		}
	}
}

// getRequestHint generates a hint for schema naming from input type and field name.
// Used by the schema registry to name schemas for anonymous/unnamed types.
// Priority:
//  1. inputType.Name() + fieldName (e.g., "CreateUserRequestName")
//  2. operationID + fieldName (e.g., "createUserRequestName")
//  3. fieldName (fallback)
func getRequestHint(inputType reflect.Type, fieldName, operationID string) string {
	if inputType.Name() != "" {
		return inputType.Name() + fieldName
	}

	if operationID != "" {
		return operationID + fieldName
	}

	return fieldName
}

// getContentType maps BodyType to HTTP content-type.
func getContentType(bodyType schema.BodyType) string {
	switch bodyType {
	case schema.BodyTypeMultipart:
		return contentTypeMultipart
	case schema.BodyTypeFile:
		return contentTypeOctetStream
	case schema.BodyTypeStructured:
		fallthrough
	default:
		return contentTypeJSON
	}
}

// transformSchemaForMultipart transforms a JSON schema for multipart/form-data.
// For multipart:
// - Binary fields ([]byte) should use format: binary (not contentEncoding: base64).
// - This represents raw octet-stream upload, not base64-encoded JSON.
func transformSchemaForMultipart(s *Schema) *Schema {
	if s == nil || s.Properties == nil {
		return s
	}

	// Create a copy to avoid modifying the original cached schema
	transformed := *s
	transformed.Properties = make(map[string]*Schema, len(s.Properties))

	for name, prop := range s.Properties {
		propCopy := *prop

		// Transform binary fields for multipart
		// In JSON: []byte -> {type: string, contentEncoding: base64}
		// In multipart: []byte -> {type: string, format: binary}
		if propCopy.Type == TypeString && propCopy.ContentEncoding == contentEncodingBase64 {
			// Remove contentEncoding (not useful for multipart per spec)
			propCopy.ContentEncoding = ""
			// Set format to binary for raw octet-stream
			propCopy.Format = formatBinary
			// Note: contentMediaType will be set to application/octet-stream
			// automatically during JSON marshaling when format is binary
		}

		transformed.Properties[name] = &propCopy
	}

	return &transformed
}

// extractMultipartEncoding creates an encoding object for multipart/form-data.
// Per OpenAPI spec, the encoding object specifies content-type for each part.
func extractMultipartEncoding(s *Schema) map[string]*Encoding {
	if s == nil || s.Properties == nil {
		return nil
	}

	encoding := make(map[string]*Encoding)

	for name, prop := range s.Properties {
		// Only add encoding for binary fields (format: binary)
		if prop.Type == TypeString && prop.Format == formatBinary {
			encoding[name] = &Encoding{
				ContentType: contentTypeOctetStream,
			}
		}
	}

	if len(encoding) == 0 {
		return nil
	}

	return encoding
}

// ExtractSecurity populates the operation's Security field from RouteSecurity.
// Only sets Security if route has security requirements and operation.Security is empty.
func (e *requestSchemaExtractor) ExtractSecurity(route *BaseRoute, op *Operation) {
	if route.Security == nil {
		return // Public route - no security
	}

	if len(op.Security) > 0 {
		return // Security already set manually
	}

	// Convert RouteSecurity to OpenAPI format
	// If Security exists, it means the route is protected
	scopes := make([]string, 0, len(route.Security.Roles)+len(route.Security.Permissions))
	scopes = append(scopes, route.Security.Roles...)
	scopes = append(scopes, route.Security.Permissions...)
	op.Security = []map[string][]string{
		{"bearerAuth": scopes},
	}
}
