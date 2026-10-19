package service

import "fmt"

type ErrorCode int

// TODO: Add more generic errors

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
	Code ErrorCode
	Err  error
}

func NewServiceErrorf(code ErrorCode, format string, args ...any) *ServiceError {
	return NewServiceError(code, fmt.Errorf(format, args...))
}

func NewServiceError(code ErrorCode, err error) *ServiceError {
	return &ServiceError{
		Code: code,
		Err:  err,
	}
}

func (e *ServiceError) Error() string {
	return e.Err.Error()
}

func (e *ServiceError) Unwrap() error {
	return e.Err
}
