package response

import (
	"encoding/json"
	"net/http"
)

// Response represents a standard API response
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *ErrorInfo  `json:"error,omitempty"`
}

// ErrorInfo contains error details
type ErrorInfo struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// JSON sends a JSON response
func JSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	response := Response{
		Success: status >= 200 && status < 300,
		Data:    data,
	}

	json.NewEncoder(w).Encode(response)
}

// Error sends an error response
func Error(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	response := Response{
		Success: false,
		Error: &ErrorInfo{
			Code:    code,
			Message: message,
		},
	}

	json.NewEncoder(w).Encode(response)
}

// BadRequest sends a 400 response
func BadRequest(w http.ResponseWriter, message string) {
	Error(w, http.StatusBadRequest, "BAD_REQUEST", message)
}

// Unauthorized sends a 401 response
func Unauthorized(w http.ResponseWriter, message string) {
	Error(w, http.StatusUnauthorized, "UNAUTHORIZED", message)
}

// Forbidden sends a 403 response
func Forbidden(w http.ResponseWriter, message string) {
	Error(w, http.StatusForbidden, "FORBIDDEN", message)
}

// NotFound sends a 404 response
func NotFound(w http.ResponseWriter, message string) {
	Error(w, http.StatusNotFound, "NOT_FOUND", message)
}

// Conflict sends a 409 response
func Conflict(w http.ResponseWriter, message string) {
	Error(w, http.StatusConflict, "CONFLICT", message)
}

// InternalError sends a 500 response
func InternalError(w http.ResponseWriter, message string) {
	Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", message)
}

// Created sends a 201 response with data
func Created(w http.ResponseWriter, data interface{}) {
	JSON(w, http.StatusCreated, data)
}

// OK sends a 200 response with data
func OK(w http.ResponseWriter, data interface{}) {
	JSON(w, http.StatusOK, data)
}

// NoContent sends a 204 response
func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}
