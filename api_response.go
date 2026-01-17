package zorya

import (
	"fmt"
	"net/http"
	"reflect"

	"github.com/talav/schema"
)

// StatusProvider allows output structs to provide HTTP status code.
// If an output struct implements this interface, the Status() method
// will be called to determine the HTTP status code for the response.
type StatusProvider interface {
	Status() int
}

// writeResponse writes the HTTP response.
func writeResponse[O any](api API, r *http.Request, w http.ResponseWriter, output *O, statusCode int) error {
	vo := reflect.ValueOf(output).Elem()

	// Get struct metadata for response handling
	outputType := reflect.TypeOf(output).Elem()
	structMeta, err := api.Metadata().GetStructMetadata(outputType)
	if err != nil {
		return fmt.Errorf("failed to get struct metadata: %w", err)
	}

	// Extract and write headers
	writeHeaders(w, structMeta, vo)

	// Check if output type implements StatusProvider interface.
	statusProviderType := reflect.TypeOf((*StatusProvider)(nil)).Elem()
	if structMeta.Type.Implements(statusProviderType) {
		if sp, ok := any(output).(StatusProvider); ok {
			statusCode = sp.Status()
		}
	}

	// Find body field by checking for "body" tag
	bodyFieldMeta := FindBodyField(structMeta)
	if bodyFieldMeta == nil {
		// No body field, just set status.
		w.WriteHeader(statusCode)

		return nil
	}

	// Extract and write body.
	writeBody(api, r, w, vo, bodyFieldMeta, statusCode)

	return nil
}

// writeHeaders writes header fields from the output struct to the HTTP response.
// Headers are identified by fields with "schema" tag and location=header.
func writeHeaders(w http.ResponseWriter, structMeta *schema.StructMetadata, vo reflect.Value) {
	// Iterate through metadata fields
	for i := range structMeta.Fields {
		fieldMeta := &structMeta.Fields[i]

		// Only process fields with schema tag and location=header
		schemaMeta, ok := schema.GetTagMetadata[*schema.SchemaMetadata](fieldMeta, "schema")
		if !ok {
			continue
		}

		if schemaMeta.Location != schema.LocationHeader {
			continue
		}

		headerName := schemaMeta.ParamName

		// Get field value from output struct
		field := vo.Field(fieldMeta.Index)
		if !field.IsValid() {
			continue
		}

		field = reflect.Indirect(field)
		if field.Kind() == reflect.Invalid {
			continue
		}

		// Handle slice (multiple header values)
		if field.Kind() == reflect.Slice {
			for i := 0; i < field.Len(); i++ {
				value := field.Index(i)
				headerValue := formatHeaderValue(value)
				w.Header().Add(headerName, headerValue)
			}
		} else {
			// Single header value
			headerValue := formatHeaderValue(field)
			w.Header().Set(headerName, headerValue)
		}
	}
}

// writeBody handles body extraction and writing.
func writeBody(api API, r *http.Request, w http.ResponseWriter, vo reflect.Value, bodyFieldMeta *schema.FieldMetadata, status int) {
	bodyField := vo.Field(bodyFieldMeta.Index)
	if !bodyField.IsValid() {
		w.WriteHeader(status)

		return
	}

	// Check if Body is a function
	if isBodyFunc(bodyFieldMeta.Type) {
		writeBodyFunc(w, r, bodyField, status)

		return
	}

	body := bodyField.Interface()

	// Handle []byte (raw bytes) - no content negotiation.
	if b, ok := body.([]byte); ok {
		writeRawBody(w, status, b)

		return
	}

	writeNegotiatedBody(api, r, w, status, body)
}

// writeBodyFunc executes a body callback function for streaming responses.
// Status and headers from struct fields should already be set before this is called.
func writeBodyFunc(w http.ResponseWriter, r *http.Request, bodyField reflect.Value, status int) {
	// Set status before streaming starts
	w.WriteHeader(status)

	if fn, ok := bodyField.Interface().(func(http.ResponseWriter, *http.Request)); ok {
		fn(w, r)
	}
}

// writeRawBody writes raw bytes without content negotiation.
func writeRawBody(w http.ResponseWriter, status int, data []byte) {
	w.WriteHeader(status)
	_, _ = w.Write(data)
}

// writeNegotiatedBody negotiates content type and marshals the body.
func writeNegotiatedBody(api API, r *http.Request, w http.ResponseWriter, status int, body any) {
	var ct string
	var err error
	// Check if body implements ContentTypeProvider (e.g., ErrorModel).
	if ctp, ok := body.(ContentTypeProvider); ok {
		ct = ctp.ContentType(ct)
	} else {
		ct, err = api.Negotiate(r.Header.Get("Accept"))
		if err != nil {
			WriteErr(api, r, w, http.StatusNotAcceptable, "Not Acceptable", err)

			return
		}
	}

	w.Header().Set("Content-Type", ct)
	w.WriteHeader(status)

	// Marshal body (fallback handled internally by Marshal)
	api.Marshal(w, ct, body)
}

// formatHeaderValue converts a reflect.Value to a string suitable for use as a header value.
func formatHeaderValue(v reflect.Value) string {
	if v.CanInterface() {
		// If the value implements fmt.Stringer, use that.
		if stringer, ok := v.Interface().(fmt.Stringer); ok {
			return stringer.String()
		}

		// For most built-in types, fmt.Sprintf("%v", v) does the right thing.
		// It handles string, int/uint (all sizes), float, bool correctly with good formatting.
		// Special cases like time.Time will also format nicely if they implement Stringer (which they do).
		return fmt.Sprintf("%v", v.Interface())
	}

	// Fallback for values that can't be interfaced (e.g., unexported fields).
	// This is rare in practice for struct tags / header population.
	return ""
}

// isBodyFunc checks if the type is a valid body callback function and validates its signature.
func isBodyFunc(t reflect.Type) bool {
	if t.Kind() != reflect.Func {
		return false
	}

	// Validate function signature: func(http.ResponseWriter, *http.Request).
	if t.NumIn() != 2 || t.NumOut() != 0 {
		return false
	}

	// Check if parameters match http.ResponseWriter and *http.Request.
	// We check by comparing with reflect types
	httpResponseWriterType := reflect.TypeOf((*http.ResponseWriter)(nil)).Elem()
	httpRequestType := reflect.TypeOf((*http.Request)(nil))

	if !t.In(0).Implements(httpResponseWriterType) || t.In(1) != httpRequestType {
		return false
	}

	return true
}
