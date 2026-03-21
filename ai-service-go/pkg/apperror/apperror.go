package apperror

import "net/http"

type AppError struct {
	Code       string
	Message    string
	Details    string
	Err        error
	HTTPStatus int
}

func (e *AppError) Error() string { return e.Message }

func BadRequest(msg string) *AppError {
	return &AppError{Code: "BAD_REQUEST", Message: msg, HTTPStatus: http.StatusBadRequest}
}

func NotFound(msg string) *AppError {
	return &AppError{Code: "NOT_FOUND", Message: msg, HTTPStatus: http.StatusNotFound}
}

func ValidationFailed(msg, details string) *AppError {
	return &AppError{Code: "VALIDATION_FAILED", Message: msg, Details: details, HTTPStatus: http.StatusBadRequest}
}

func Internal(err error, msg string) *AppError {
	return &AppError{Code: "INTERNAL_ERROR", Message: msg, Err: err, HTTPStatus: http.StatusInternalServerError}
}
