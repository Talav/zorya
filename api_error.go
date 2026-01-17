package zorya

import (
	"errors"
	"fmt"
	"net/http"
)

const (
	contentTypeJSON        = "application/json"
	contentTypeMultipart   = "multipart/form-data"
	contentTypeOctetStream = "application/octet-stream"
)

// StatusError is an error that has an HTTP status code. When returned from
// an operation handler, this sets the response status code before sending it
// to the client.
type StatusError interface {
	GetStatus() int
	Error() string
}

// HeadersError is an error that has HTTP headers. When returned from an
// operation handler, these headers are set on the response before sending it
// to the client. Use `ErrorWithHeaders` to wrap an error like
// `Error400BadRequest` with additional headers.
type HeadersError interface {
	GetHeaders() http.Header
	Error() string
}

// ErrorDetailer returns error details for responses & debugging. This enables
// the use of custom error types. See `NewError` for more details.
type ErrorDetailer interface {
	ErrorDetail() *ErrorDetail
}

// ErrorModel defines a basic error message model based on RFC 9457 Problem
// Details for HTTP APIs (https://datatracker.ietf.org/doc/html/rfc9457). It
// is augmented with an `errors` field of `ErrorDetail` objects that
// can help provide exhaustive & descriptive errors.
//
//	err := &ErrorModel{
//		Title:  http.StatusText(http.StatusBadRequest),
//		Status: http.StatusBadRequest,
//		Detail: "Validation failed",
//		Errors: []*ErrorDetail{
//			{
//				Code:     "required",
//				Message:  "expected required property id to be present",
//				Location: "body.friends[0]",
//			},
//			{
//				Code:     "type",
//				Message:  "expected boolean",
//				Location: "body.friends[1].active",
//			},
//		},
//	}
//
//nolint:errname // ErrorModel intentionally matches RFC 7807 Problem Details naming
type ErrorModel struct {
	// Type is a URI to get more information about the error type.
	Type string `json:"type,omitempty"`

	// Title provides a short static summary of the problem. Zorya will default this
	// to the HTTP response status code text if not present.
	Title string `json:"title,omitempty"`

	// Status provides the HTTP status code for client convenience. Zorya will
	// default this to the response status code if unset. This SHOULD match the
	// response status code (though proxies may modify the actual status code).
	Status int `json:"status,omitempty"`

	// Detail is an explanation specific to this error occurrence.
	Detail string `json:"detail,omitempty"`

	// Instance is a URI to get more info about this error occurrence.
	Instance string `json:"instance,omitempty"`

	// Errors provides an optional mechanism of passing additional error details
	// as a list.
	Errors []*ErrorDetail `json:"errors,omitempty"`
}

// ErrorDetail provides details about a specific error.
//
//nolint:errname // ErrorDetail is a detail object, not an error type
type ErrorDetail struct {
	// Code is a machine-readable error code (e.g., "required", "email", "min").
	// This enables frontend translation and automated error handling.
	Code string `json:"code,omitempty"`

	// Message is a human-readable explanation of the error (optional).
	// Useful for developers, logs, and debugging.
	Message string `json:"message,omitempty"`

	// Location is a path-like string indicating where the error occurred.
	// It typically begins with `path`, `query`, `header`, or `body`. Example:
	// `body.items[3].tags` or `path.thing-id`.
	Location string `json:"location,omitempty"`
}

//nolint:errname // errWithHeaders is an internal wrapper, not a public error type
type errWithHeaders struct {
	err     error
	headers http.Header
}

// NewError creates a new instance of an error model with the given status code,
// message, and optional error details. If the error details implement the
// `ErrorDetailer` interface, the error details will be used. Otherwise, the
// error string will be used as the message.
//
// Replace this function to use your own error type. Example:
//
//	type MyDetail struct {
//		Message string `json:"message"`
//		Location string `json:"location"`
//	}
//
//	type MyError struct {
//		status  int
//		Message string `json:"message"`
//		Errors  []error `json:"errors"`
//	}
//
//	func (e *MyError) Error() string {
//		return e.Message
//	}
//
//	func (e *MyError) GetStatus() int {
//		return e.status
//	}
//
//	zorya.NewError = func(status int, msg string, errs ...error) StatusError {
//		return &MyError{
//			status:  status,
//			Message: msg,
//			Errors:  errs,
//		}
//	}
var NewError = func(status int, msg string, errs ...error) StatusError {
	details := make([]*ErrorDetail, 0, len(errs))
	for i := range len(errs) {
		if errs[i] == nil {
			continue
		}
		if converted, ok := errs[i].(ErrorDetailer); ok {
			details = append(details, converted.ErrorDetail())
		} else {
			details = append(details, &ErrorDetail{Message: errs[i].Error()})
		}
	}

	title := http.StatusText(status)
	if title == "" {
		title = "Error"
	}

	return &ErrorModel{
		Status: status,
		Title:  title,
		Detail: msg,
		Errors: details,
	}
}

// WriteErr writes an error response with the given request and response writer.
// It can be called in two ways:
//   - With status and message: WriteErr(api, r, w, status, msg, errs...)
//   - With an existing error: WriteErr(api, r, w, 0, "", err) where err is the error to write
//
// If status is 0 and msg is empty, the first error in errs is used as the main error.
// Otherwise, a new error is created using NewError(status, msg, errs...).
// The error is marshaled using the API's content negotiation methods.
func WriteErr(api API, r *http.Request, w http.ResponseWriter, status int, msg string, errs ...error) {
	// Determine the error to write and its status
	errToWrite, status := determineErrorToWrite(status, msg, errs)

	// Set headers if error implements HeadersError
	applyErrorHeaders(w, errToWrite)

	// Negotiate and set content type
	ct := negotiateContentType(api, r, errToWrite)
	w.Header().Set("Content-Type", ct)
	w.WriteHeader(status)

	// Marshal and write error (fallback handled internally by Marshal)
	api.Marshal(w, ct, errToWrite)
}

// determineErrorToWrite determines the error to write and its status code.
func determineErrorToWrite(status int, msg string, errs []error) (StatusError, int) {
	// If status is 0 and msg is empty, use the first error directly
	if status == 0 && msg == "" && len(errs) > 0 && errs[0] != nil {
		return processExistingError(errs[0])
	}

	// Create new error from status and message
	errToWrite := NewError(status, msg, errs...)

	return errToWrite, errToWrite.GetStatus()
}

// processExistingError processes an existing error and returns it as StatusError.
func processExistingError(err error) (StatusError, int) {
	// Determine status code
	status := http.StatusInternalServerError
	var statusErr StatusError
	var maxBytesErr *http.MaxBytesError

	if errors.As(err, &statusErr) {
		status = statusErr.GetStatus()
	} else if errors.As(err, &maxBytesErr) {
		status = http.StatusRequestEntityTooLarge
	}

	// Convert to StatusError if needed
	if errors.As(err, &statusErr) {
		return statusErr, status
	}

	msg := err.Error()
	// Special message for MaxBytesError
	if maxBytesErr != nil {
		msg = fmt.Sprintf("request body too large (limit: %d bytes)", maxBytesErr.Limit)
	}

	return NewError(status, msg), status
}

// applyErrorHeaders sets headers from HeadersError if present.
func applyErrorHeaders(w http.ResponseWriter, errToWrite StatusError) {
	var he HeadersError
	if errors.As(errToWrite, &he) {
		for k, values := range he.GetHeaders() {
			for _, v := range values {
				w.Header().Add(k, v)
			}
		}
	}
}

// negotiateContentType negotiates the content type for the error response.
func negotiateContentType(api API, r *http.Request, errToWrite StatusError) string {
	// Negotiate content type
	ct, err := api.Negotiate(r.Header.Get("Accept"))
	if err != nil || ct == "" {
		// Fallback to JSON if negotiation fails or returns empty
		ct = contentTypeJSON
	}

	// Check ContentTypeProvider
	if ctp, ok := errToWrite.(ContentTypeProvider); ok {
		ct = ctp.ContentType(ct)
	}

	return ct
}

// ErrorWithHeaders wraps an error with additional headers to be sent to the
// client. This is useful for e.g. caching, rate limiting, or other metadata.
func ErrorWithHeaders(err error, headers http.Header) error {
	var he HeadersError
	if errors.As(err, &he) {
		// There is already a headers error, so we need to merge the headers. This
		// lets you chain multiple calls together and have all the headers set.
		orig := he.GetHeaders()
		for k, values := range headers {
			for _, v := range values {
				orig.Add(k, v)
			}
		}

		return err
	}

	return &errWithHeaders{err: err, headers: headers}
}

// Status304NotModified returns a 304. This is not really an error, but
// provides a way to send non-default responses.
func Status304NotModified() StatusError {
	return NewError(http.StatusNotModified, "")
}

// Error400BadRequest returns a 400.
func Error400BadRequest(msg string, errs ...error) StatusError {
	return NewError(http.StatusBadRequest, msg, errs...)
}

// Error401Unauthorized returns a 401.
func Error401Unauthorized(msg string, errs ...error) StatusError {
	return NewError(http.StatusUnauthorized, msg, errs...)
}

// Error403Forbidden returns a 403.
func Error403Forbidden(msg string, errs ...error) StatusError {
	return NewError(http.StatusForbidden, msg, errs...)
}

// Error404NotFound returns a 404.
func Error404NotFound(msg string, errs ...error) StatusError {
	return NewError(http.StatusNotFound, msg, errs...)
}

// Error405MethodNotAllowed returns a 405.
func Error405MethodNotAllowed(msg string, errs ...error) StatusError {
	return NewError(http.StatusMethodNotAllowed, msg, errs...)
}

// Error406NotAcceptable returns a 406.
func Error406NotAcceptable(msg string, errs ...error) StatusError {
	return NewError(http.StatusNotAcceptable, msg, errs...)
}

// Error409Conflict returns a 409.
func Error409Conflict(msg string, errs ...error) StatusError {
	return NewError(http.StatusConflict, msg, errs...)
}

// Error410Gone returns a 410.
func Error410Gone(msg string, errs ...error) StatusError {
	return NewError(http.StatusGone, msg, errs...)
}

// Error412PreconditionFailed returns a 412.
func Error412PreconditionFailed(msg string, errs ...error) StatusError {
	return NewError(http.StatusPreconditionFailed, msg, errs...)
}

// Error415UnsupportedMediaType returns a 415.
func Error415UnsupportedMediaType(msg string, errs ...error) StatusError {
	return NewError(http.StatusUnsupportedMediaType, msg, errs...)
}

// Error422UnprocessableEntity returns a 422.
func Error422UnprocessableEntity(msg string, errs ...error) StatusError {
	return NewError(http.StatusUnprocessableEntity, msg, errs...)
}

// Error429TooManyRequests returns a 429.
func Error429TooManyRequests(msg string, errs ...error) StatusError {
	return NewError(http.StatusTooManyRequests, msg, errs...)
}

// Error500InternalServerError returns a 500.
func Error500InternalServerError(msg string, errs ...error) StatusError {
	return NewError(http.StatusInternalServerError, msg, errs...)
}

// Error501NotImplemented returns a 501.
func Error501NotImplemented(msg string, errs ...error) StatusError {
	return NewError(http.StatusNotImplemented, msg, errs...)
}

// Error502BadGateway returns a 502.
func Error502BadGateway(msg string, errs ...error) StatusError {
	return NewError(http.StatusBadGateway, msg, errs...)
}

// Error503ServiceUnavailable returns a 503.
func Error503ServiceUnavailable(msg string, errs ...error) StatusError {
	return NewError(http.StatusServiceUnavailable, msg, errs...)
}

// Error504GatewayTimeout returns a 504.
func Error504GatewayTimeout(msg string, errs ...error) StatusError {
	return NewError(http.StatusGatewayTimeout, msg, errs...)
}

// Error satisfies the `error` interface. It returns the error's detail field.
func (e *ErrorModel) Error() string {
	return e.Detail
}

// Add an error to the `Errors` slice. If passed a struct that satisfies the
// `ErrorDetailer` interface, then it is used, otherwise the error
// string is used as the error detail message.
//
//	err := &ErrorModel{ /* ... */ }
//	err.Add(&ErrorDetail{
//		Code:     "type",
//		Message:  "expected boolean",
//		Location: "body.friends[1].active",
//	})
func (e *ErrorModel) Add(err error) {
	if converted, ok := err.(ErrorDetailer); ok {
		e.Errors = append(e.Errors, converted.ErrorDetail())

		return
	}

	if err != nil {
		e.Errors = append(e.Errors, &ErrorDetail{Message: err.Error()})
	}
}

// GetStatus returns the HTTP status that should be returned to the client
// for this error.
func (e *ErrorModel) GetStatus() int {
	return e.Status
}

// ContentType provides a filter to adjust response content types. This is
// used to ensure e.g. `application/problem+json` content types defined in
// RFC 9457 Problem Details for HTTP APIs are used in responses to clients.
func (e *ErrorModel) ContentType(ct string) string {
	if ct == contentTypeJSON {
		return "application/problem+json"
	}
	if ct == "application/cbor" {
		return "application/problem+cbor"
	}

	return ct
}

// Error returns the error message / satisfies the `error` interface.
func (e *ErrorDetail) Error() string {
	if e.Message != "" {
		if e.Location != "" {
			return fmt.Sprintf("%s (%s)", e.Message, e.Location)
		}

		return e.Message
	}

	if e.Code != "" && e.Location != "" {
		return fmt.Sprintf("%s: %s", e.Code, e.Location)
	}

	if e.Code != "" {
		return e.Code
	}

	if e.Location != "" {
		return e.Location
	}

	return "validation error"
}

// ErrorDetail satisfies the `ErrorDetailer` interface.
func (e *ErrorDetail) ErrorDetail() *ErrorDetail {
	return e
}

func (e *errWithHeaders) Error() string {
	return e.err.Error()
}

func (e *errWithHeaders) Unwrap() error {
	return e.err
}

func (e *errWithHeaders) GetHeaders() http.Header {
	return e.headers
}
