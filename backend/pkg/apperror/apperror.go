package apperror

import "net/http"

const (
	ErrCodeNotFound          = "NOT_FOUND"
	ErrCodeBadRequest        = "BAD_REQUEST"
	ErrCodeValidationFailed  = "VALIDATION_FAILED"
	ErrCodeInternal          = "INTERNAL_ERROR"
	ErrCodeConflict          = "CONFLICT"
	ErrCodeServiceUnavailable = "SERVICE_UNAVAILABLE"
)

type AppError struct {
	Code       string
	Message    string
	Details    string
	Err        error
	HTTPStatus int
}

func (e *AppError) Error() string { return e.Message }

func NotFound(msg string) *AppError {
	return &AppError{Code: ErrCodeNotFound, Message: msg, HTTPStatus: http.StatusNotFound}
}

func BadRequest(msg string) *AppError {
	return &AppError{Code: ErrCodeBadRequest, Message: msg, HTTPStatus: http.StatusBadRequest}
}

func ValidationFailed(msg, details string) *AppError {
	return &AppError{Code: ErrCodeValidationFailed, Message: msg, Details: details, HTTPStatus: http.StatusBadRequest}
}

func Internal(err error, msg string) *AppError {
	return &AppError{Code: ErrCodeInternal, Message: msg, Err: err, HTTPStatus: http.StatusInternalServerError}
}

func Conflict(msg string) *AppError {
	return &AppError{Code: ErrCodeConflict, Message: msg, HTTPStatus: http.StatusConflict}
}

func ServiceUnavailable(msg string) *AppError {
	return &AppError{Code: ErrCodeServiceUnavailable, Message: msg, HTTPStatus: http.StatusServiceUnavailable}
}

func Unauthorized(msg string) *AppError {
	return &AppError{Code: "UNAUTHORIZED", Message: msg, HTTPStatus: http.StatusUnauthorized}
}

func Forbidden(msg string) *AppError {
	return &AppError{Code: "FORBIDDEN", Message: msg, HTTPStatus: http.StatusForbidden}
}
