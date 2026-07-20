package service

import "fmt"

type ErrorCode int

// Generic error codes
const (
	ErrCodeBadRequest ErrorCode = -1 - iota
	ErrCodeUnauthorized
	ErrCodeForbidden
	ErrCodeNotFound
	ErrCodeConflict
	ErrCodeUnprocessable
	ErrCodeInternal
)

// Specific error codes
const (
	ErrCodeUsernameTaken ErrorCode = -1025 - iota
	ErrCodeInvalidCredentials
)

type ServiceError struct {
	Code    ErrorCode
	Message string
}

func NewServiceErrorf(code ErrorCode, format string, args ...any) *ServiceError {
	return NewServiceError(code, fmt.Sprintf(format, args...))
}

func NewServiceError(code ErrorCode, message string) *ServiceError {
	return &ServiceError{
		Code:    code,
		Message: message,
	}
}

func (e *ServiceError) Error() string {
	return e.Message
}
