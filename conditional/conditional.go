// Package conditional provides utilities for HTTP conditional requests using
// If-Match, If-None-Match, If-Modified-Since, and If-Unmodified-Since headers.
//
// Usage:
//
//	Embed Params in your input struct:
//
//		type GetUserInput struct {
//			ID string `path:"id"`
//			conditional.Params
//		}
//
//	Then check preconditions in your handler:
//
//		func getUser(ctx context.Context, input *GetUserInput) (*UserOutput, error) {
//			user := getUserFromDB(input.ID)
//
//			if err := input.Params.CheckPreconditions(
//				user.ETag,
//				user.Modified,
//				ctx.Request().Method != http.MethodGet,
//			); err != nil {
//				return nil, err
//			}
//
//			return &UserOutput{User: user}, nil
//		}
package conditional

import (
	"net/http"
	"strings"
	"time"

	"github.com/talav/zorya"
)

// trimETag removes quotes and W/ prefix from ETag values for comparison.
func trimETag(value string) string {
	if strings.HasPrefix(value, "W/") && len(value) > 2 {
		value = value[2:]
	}

	return strings.Trim(value, "\"")
}

// Params represents conditional request headers. Embed this struct in your input
// struct to enable conditional request support.
//
// Example:
//
//	type GetUserInput struct {
//		ID string `path:"id"`
//		conditional.Params
//	}
type Params struct {
	IfMatch           []string  `schema:"If-Match,location=header"`
	IfNoneMatch       []string  `schema:"If-None-Match,location=header"`
	IfModifiedSince   time.Time `schema:"If-Modified-Since,location=header"`
	IfUnmodifiedSince time.Time `schema:"If-Unmodified-Since,location=header"`
}

// HasConditionalParams returns true if any conditional request headers are present.
func (p *Params) HasConditionalParams() bool {
	return len(p.IfMatch) > 0 || len(p.IfNoneMatch) > 0 || !p.IfModifiedSince.IsZero() || !p.IfUnmodifiedSince.IsZero()
}

// CheckPreconditions validates conditional request headers against the current
// resource state. Returns an error if preconditions fail, nil otherwise.
//
// For read requests (GET, HEAD), returns 304 Not Modified if conditions fail.
// For write requests (POST, PUT, PATCH, DELETE), returns 412 Precondition Failed.
func (p *Params) CheckPreconditions(currentETag string, currentModified time.Time, isWrite bool) zorya.StatusError {
	return CheckPreconditions(
		p.IfMatch,
		p.IfNoneMatch,
		p.IfModifiedSince,
		p.IfUnmodifiedSince,
		currentETag,
		currentModified,
		isWrite,
	)
}

// checkIfNoneMatch validates If-None-Match header.
func checkIfNoneMatch(ifNoneMatch []string, currentETag string, isWrite bool) (bool, []error) {
	var errs []error
	failed := false

	foundMsg := "found no existing resource"
	if currentETag != "" {
		foundMsg = "found resource with ETag " + currentETag
	}

	for _, match := range ifNoneMatch {
		trimmed := trimETag(match)
		if trimmed == currentETag || (trimmed == "*" && currentETag != "") {
			if isWrite {
				errs = append(errs, &zorya.ErrorDetail{
					Code:     "precondition_failed",
					Message:  "If-None-Match: " + match + " precondition failed, " + foundMsg,
					Location: "headers.If-None-Match",
				})
			}
			failed = true
		}
	}

	return failed, errs
}

// checkIfMatch validates If-Match header.
func checkIfMatch(ifMatch []string, currentETag string, isWrite bool) (bool, []error) {
	var errs []error
	failed := false

	if len(ifMatch) == 0 {
		return false, nil
	}

	foundMsg := "found no existing resource"
	if currentETag != "" {
		foundMsg = "found resource with ETag " + currentETag
	}

	found := false
	for _, match := range ifMatch {
		if trimETag(match) == currentETag {
			found = true

			break
		}
	}

	if !found {
		if isWrite {
			errs = append(errs, &zorya.ErrorDetail{
				Code:     "precondition_failed",
				Message:  "If-Match precondition failed, " + foundMsg,
				Location: "headers.If-Match",
			})
		}
		failed = true
	}

	return failed, errs
}

// checkIfModifiedSince validates If-Modified-Since header.
func checkIfModifiedSince(ifModifiedSince time.Time, currentModified time.Time, isWrite bool) (bool, []error) {
	var errs []error
	failed := false

	if ifModifiedSince.IsZero() {
		return false, nil
	}

	if !currentModified.After(ifModifiedSince) {
		if isWrite {
			errs = append(errs, &zorya.ErrorDetail{
				Code:     "not_modified",
				Message:  "If-Modified-Since: " + ifModifiedSince.Format(http.TimeFormat) + " precondition failed, resource was modified at " + currentModified.Format(http.TimeFormat),
				Location: "headers.If-Modified-Since",
			})
		}
		failed = true
	}

	return failed, errs
}

// checkIfUnmodifiedSince validates If-Unmodified-Since header.
func checkIfUnmodifiedSince(ifUnmodifiedSince time.Time, currentModified time.Time, isWrite bool) (bool, []error) {
	var errs []error
	failed := false

	if ifUnmodifiedSince.IsZero() {
		return false, nil
	}

	if currentModified.After(ifUnmodifiedSince) {
		if isWrite {
			errs = append(errs, &zorya.ErrorDetail{
				Code:     "precondition_failed",
				Message:  "If-Unmodified-Since: " + ifUnmodifiedSince.Format(http.TimeFormat) + " precondition failed, resource was modified at " + currentModified.Format(http.TimeFormat),
				Location: "headers.If-Unmodified-Since",
			})
		}
		failed = true
	}

	return failed, errs
}

// CheckPreconditions validates conditional request headers against the current
// resource state. Returns an error if preconditions fail, nil otherwise.
//
// For read requests (GET, HEAD), returns 304 Not Modified if conditions fail.
// For write requests (POST, PUT, PATCH, DELETE), returns 412 Precondition Failed.
//
// Parameters:
//   - ifMatch: If-Match header values (parsed by schema)
//   - ifNoneMatch: If-None-Match header values (parsed by schema)
//   - ifModifiedSince: If-Modified-Since header value (parsed by schema)
//   - ifUnmodifiedSince: If-Unmodified-Since header value (parsed by schema)
//   - currentETag: Current resource ETag (empty if resource doesn't exist)
//   - currentModified: Current resource last modified time
//   - isWrite: True for write requests (POST, PUT, PATCH, DELETE)
func CheckPreconditions(
	ifMatch []string,
	ifNoneMatch []string,
	ifModifiedSince time.Time,
	ifUnmodifiedSince time.Time,
	currentETag string,
	currentModified time.Time,
	isWrite bool,
) zorya.StatusError {
	var allErrs []error
	failed := false

	failedNoneMatch, errs := checkIfNoneMatch(ifNoneMatch, currentETag, isWrite)
	if failedNoneMatch {
		failed = true
		allErrs = append(allErrs, errs...)
	}

	failedMatch, errs := checkIfMatch(ifMatch, currentETag, isWrite)
	if failedMatch {
		failed = true
		allErrs = append(allErrs, errs...)
	}

	failedModified, errs := checkIfModifiedSince(ifModifiedSince, currentModified, isWrite)
	if failedModified {
		failed = true
		allErrs = append(allErrs, errs...)
	}

	failedUnmodified, errs := checkIfUnmodifiedSince(ifUnmodifiedSince, currentModified, isWrite)
	if failedUnmodified {
		failed = true
		allErrs = append(allErrs, errs...)
	}

	if failed {
		if isWrite {
			return zorya.NewError(
				http.StatusPreconditionFailed,
				http.StatusText(http.StatusPreconditionFailed),
				allErrs...,
			)
		}

		return zorya.Status304NotModified()
	}

	return nil
}
