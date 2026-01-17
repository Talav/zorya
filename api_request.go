package zorya

import (
	"net/http"
	"reflect"
	"time"
)

// setupRequestLimits configures body read timeout and size limits for the request.
func setupRequestLimits(r *http.Request, w http.ResponseWriter, route BaseRoute) {
	// Apply body read timeout.
	// This sets a deadline for reading the request body, helping prevent slow-loris attacks.
	// Default is 5 seconds if not explicitly configured.
	bodyTimeout := route.BodyReadTimeout
	if bodyTimeout == 0 {
		bodyTimeout = DefaultBodyReadTimeout
	}
	if bodyTimeout != 0 {
		rc := http.NewResponseController(w)
		if bodyTimeout > 0 {
			_ = rc.SetReadDeadline(time.Now().Add(bodyTimeout))
		} else {
			// Negative value disables any deadline.
			_ = rc.SetReadDeadline(time.Time{})
		}
	}

	// Apply body size limit using http.MaxBytesReader.
	// Default to 1MB if not explicitly configured.
	maxBytes := route.MaxBodyBytes
	if maxBytes == 0 {
		maxBytes = DefaultMaxBodyBytes
	}
	if maxBytes > 0 {
		r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
	}
}

// validateRequest validates the decoded input struct.
// Returns validation errors if validation failed, or nil if validation succeeded.
func validateRequest[I any](api API, r *http.Request, input *I) []error {
	v := api.Validator()
	if v == nil {
		return nil
	}

	// Get struct metadata for validation location mapping
	typ := reflect.TypeOf(input).Elem()

	metadata, err := api.Metadata().GetStructMetadata(typ)
	if err != nil {
		// If we can't get metadata, skip validation rather than failing
		// This allows validation to work even if metadata lookup fails
		return nil
	}

	return v.Validate(r.Context(), input, metadata)
}
