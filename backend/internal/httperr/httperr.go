// Package httperr provides typed API errors and a JSON error responder.
package httperr

import (
	"encoding/json"
	"errors"
	"net/http"
)

// APIError is a structured error returned by all handlers.
type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Status  int    `json:"-"`
	// Details is optional field-level validation info.
	Details map[string]string `json:"details,omitempty"`
}

func (e *APIError) Error() string { return e.Message }

func New(code, msg string, status int) *APIError {
	return &APIError{Code: code, Message: msg, Status: status}
}

// Convenience constructors.
func BadRequest(msg string) *APIError      { return New("bad_request", msg, http.StatusBadRequest) }
func Unauthorized(msg string) *APIError    { return New("unauthorized", msg, http.StatusUnauthorized) }
func Forbidden(msg string) *APIError       { return New("forbidden", msg, http.StatusForbidden) }
func NotFound(msg string) *APIError        { return New("not_found", msg, http.StatusNotFound) }
func Conflict(msg string) *APIError        { return New("conflict", msg, http.StatusConflict) }
func Unprocessable(msg string) *APIError   { return New("validation_error", msg, http.StatusUnprocessableEntity) }
func Internal(msg string) *APIError        { return New("internal_error", msg, http.StatusInternalServerError) }

// Validation builds a 422 with field-level details.
func Validation(details map[string]string) *APIError {
	return &APIError{
		Code:    "validation_error",
		Message: "request validation failed",
		Status:  http.StatusUnprocessableEntity,
		Details: details,
	}
}

// AsAPIError unwraps an error into an *APIError, defaulting to 500.
func AsAPIError(err error) *APIError {
	var ae *APIError
	if errors.As(err, &ae) {
		return ae
	}
	return Internal(err.Error())
}

// Write sends a structured JSON error response.
func Write(w http.ResponseWriter, err error) {
	ae := AsAPIError(err)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(ae.Status)
	_ = json.NewEncoder(w).Encode(map[string]any{"error": ae})
}

// JSON sends a 200 (or provided status) JSON response.
func JSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
