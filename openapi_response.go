package zorya

import (
	"fmt"
	"net/http"
	"reflect"
	"strconv"

	"github.com/talav/schema"
	"github.com/talav/zorya/metadata"
)

// ContentTypeProvider allows you to override the content type for responses,
// allowing you to return a different content type like
// `application/problem+json` after using the `application/json` marshaller.
// This should be implemented by the response body struct.
type ContentTypeProvider interface {
	ContentType(string) string
}

// ResponseSchemaExtractor extracts OpenAPI response schemas from output struct types.
type ResponseSchemaExtractor struct {
	registry Registry
	builder  SchemaBuilder
	metadata *schema.Metadata
}

// NewResponseSchemaExtractor creates a new response schema extractor.
func NewResponseSchemaExtractor(registry Registry, builder SchemaBuilder, metadata *schema.Metadata) *ResponseSchemaExtractor {
	return &ResponseSchemaExtractor{
		registry: registry,
		builder:  builder,
		metadata: metadata,
	}
}

// ExtractResponseSchemas extracts OpenAPI response schemas from an output struct type
// and populates the operation's Responses map.
//
// This handles:
// - Success response: Generated from Body field (one response, default status).
// - Error responses: Generated from route.Errors list.
// - Headers: Generated from fields with "header" tag.
// - Automatic error additions: 422 (if inputs exist) and 500 (always).
func (e *ResponseSchemaExtractor) ResponseFromType(outputType reflect.Type, route *BaseRoute) error {
	structMeta, err := e.metadata.GetStructMetadata(outputType)
	if err != nil {
		return fmt.Errorf("failed to get struct metadata for type %s: %w", outputType, err)
	}

	// Initialize operation and responses
	initializeOperation(route)
	defaultStatus := getDefaultStatus(route)
	resp := getResponse(route.Operation, defaultStatus)

	if resp.Content == nil {
		resp.Content = make(map[string]*MediaType)
	}

	// Find and process body field
	bodyField := FindBodyField(structMeta)
	if bodyField == nil {
		return nil
	}

	// Extract body schema and add to response
	if err := e.extractBodySchema(bodyField, resp, structMeta.Type, route.Operation); err != nil {
		return err
	}

	// Extract header schemas and add to success response
	e.extractHeaderSchemas(structMeta, resp)

	// Process error responses
	hasInputParams := len(route.Operation.Parameters) > 0
	hasInputBody := route.Operation.RequestBody != nil
	e.defineErrorResponses(route.Operation, route, hasInputParams, hasInputBody)

	return nil
}

// initializeOperation ensures the operation and responses map are initialized.
func initializeOperation(route *BaseRoute) {
	if route.Operation == nil {
		route.Operation = &Operation{}
	}
	if route.Operation.Responses == nil {
		route.Operation.Responses = make(map[string]*Response)
	}
}

// getDefaultStatus returns the default HTTP status code for the route.
func getDefaultStatus(route *BaseRoute) int {
	if route.DefaultStatus != 0 {
		return route.DefaultStatus
	}

	return http.StatusOK
}

// extractBodySchema extracts the body schema from a body field and adds it to the response.
func (e *ResponseSchemaExtractor) extractBodySchema(
	bodyField *schema.FieldMetadata,
	resp *Response,
	structType reflect.Type,
	op *Operation,
) error {
	// Get body metadata to determine content type based on body tag
	bodyMeta, ok := schema.GetTagMetadata[*schema.BodyMetadata](bodyField, "body")
	if !ok {
		return fmt.Errorf("body field missing body metadata")
	}

	// Determine content type
	ct := determineContentType(bodyField, bodyMeta)

	// Initialize media type if needed (only if Content is empty)
	if len(resp.Content) == 0 {
		resp.Content[ct] = &MediaType{}
	}

	// Generate schema and set it if successful
	hint := getResponseHint(structType, bodyField.StructFieldName, op.OperationID)
	bodySchema := e.registry.Schema(bodyField.Type, true, hint)
	if bodySchema != nil {
		// Transform schema for file responses: use format: binary instead of contentEncoding: base64
		if bodyMeta.BodyType == schema.BodyTypeFile {
			bodySchema = transformSchemaForFileResponse(bodySchema)
		}
		if resp.Content[ct] != nil && resp.Content[ct].Schema == nil {
			resp.Content[ct].Schema = bodySchema
		}
	}

	return nil
}

// determineContentType determines the content type for a body field.
func determineContentType(bodyField *schema.FieldMetadata, bodyMeta *schema.BodyMetadata) string {
	// Determine content type based on BodyType (same logic as requests)
	ct := getContentType(bodyMeta.BodyType)

	// Fallback to ContentTypeProvider interface if needed
	if ct == contentTypeJSON && reflect.PointerTo(bodyField.Type).Implements(reflect.TypeOf((*ContentTypeProvider)(nil)).Elem()) {
		instance, ok := reflect.New(bodyField.Type).Interface().(ContentTypeProvider)
		if ok {
			ct = instance.ContentType(ct)
		}
	}

	return ct
}

// extractHeaderSchemas extracts header schemas from fields with "schema" tag and location=header
// and adds them to the success response.
func (e *ResponseSchemaExtractor) extractHeaderSchemas(structMeta *schema.StructMetadata, response *Response) {
	if response.Headers == nil {
		response.Headers = make(map[string]*Param)
	}

	// Iterate through metadata fields
	for _, fieldMeta := range structMeta.Fields {
		// Only process fields with schema tag and location=header
		schemaMeta, ok := schema.GetTagMetadata[*schema.SchemaMetadata](&fieldMeta, "schema")
		if !ok {
			continue
		}

		if schemaMeta.Location != schema.LocationHeader {
			continue
		}

		headerName := schemaMeta.ParamName

		// Get field type for schema generation
		fieldType := fieldMeta.Type
		if fieldType.Kind() == reflect.Slice {
			fieldType = fieldType.Elem()
		}

		// Check if field implements fmt.Stringer (will be serialized as string)
		if reflect.PointerTo(fieldType).Implements(reflect.TypeOf((*fmt.Stringer)(nil)).Elem()) {
			fieldType = reflect.TypeOf("")
		}

		// Generate schema for header
		hint := getResponseHint(structMeta.Type, fieldMeta.StructFieldName, headerName)
		headerSchema := e.registry.Schema(fieldType, true, hint)

		// Get description from openapi metadata if available
		description := ""
		if openAPIMeta, ok := schema.GetTagMetadata[*metadata.OpenAPIMetadata](&fieldMeta, "openapi"); ok {
			description = openAPIMeta.Description
		}

		// Create header parameter
		response.Headers[headerName] = &Header{
			Schema:      headerSchema,
			Description: description,
		}
	}
}

// defineErrorResponses defines error responses based on route.Errors and automatic additions.
func (e *ResponseSchemaExtractor) defineErrorResponses(
	op *Operation,
	route *BaseRoute,
	hasInputParams bool,
	hasInputBody bool,
) {
	// Get error schema (ErrorModel type)
	errorType := reflect.TypeOf((*ErrorModel)(nil)).Elem()
	errorSchema := e.registry.Schema(errorType, true, "Error")

	// Determine error content type
	exampleErr := NewError(0, "")
	errContentType := contentTypeJSON
	if ctf, ok := exampleErr.(ContentTypeProvider); ok {
		errContentType = ctf.ContentType(errContentType)
	}

	// Process user-specified errors
	errorsToAdd := make([]int, 0, len(route.Errors)+2)
	errorsToAdd = append(errorsToAdd, route.Errors...)

	// Automatically add 422 if there are input parameters or body
	if hasInputParams || hasInputBody {
		errorsToAdd = append(errorsToAdd, http.StatusUnprocessableEntity)
	}

	// Always add 500 - every endpoint can produce server errors
	errorsToAdd = append(errorsToAdd, http.StatusInternalServerError)

	// Create response objects for each error code
	for _, code := range errorsToAdd {
		response := getResponse(op, code)
		if response.Content == nil {
			response.Content = map[string]*MediaType{
				errContentType: {
					Schema: errorSchema,
				},
			}
		}
	}
}

// getResponse ensures a response exists for the given status code.
// If the response doesn't exist, it creates one with the provided description.
// If description is empty, it uses the HTTP status text.
// Returns the response (existing or newly created).
func getResponse(op *Operation, statusCode int) *Response {
	statusStr := strconv.Itoa(statusCode)
	if op.Responses[statusStr] == nil {
		op.Responses[statusStr] = &Response{
			Description: http.StatusText(statusCode),
		}
	}

	return op.Responses[statusStr]
}

// getResponseHint generates a hint for schema naming from output type and field name.
func getResponseHint(outputType reflect.Type, fieldName, operationID string) string {
	if outputType.Name() != "" {
		return outputType.Name() + fieldName
	}

	if operationID != "" {
		return operationID + fieldName
	}

	return fieldName
}

// transformSchemaForFileResponse transforms a schema for file/binary responses.
// For file responses, []byte should use format: binary (not contentEncoding: base64).
func transformSchemaForFileResponse(s *Schema) *Schema {
	if s == nil {
		return s
	}

	// For []byte fields, change from base64 encoding to binary format
	if s.Type == TypeString && s.ContentEncoding == contentEncodingBase64 {
		sCopy := *s
		sCopy.ContentEncoding = ""
		sCopy.Format = formatBinary

		return &sCopy
	}

	return s
}
