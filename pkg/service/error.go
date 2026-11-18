package service

import "fmt"

type ErrType int

const (
	ErrTypeUnauthorized ErrType = iota
	ErrTypePaymentRequired
	ErrTypeForbidden
	ErrTypeNotFound
	ErrTypeConflict
	ErrTypeGone
	ErrTypePreconditionFailed
	ErrTypeUnprocessableContent
	ErrTypeLocked
	ErrTypeFailedDependency
	ErrTypePreconditionRequired
	ErrTypeTooManyRequests
	ErrTypeUnavailableForLegalReasons
	ErrTypeInternalServerError
	ErrTypeBadGateway
	ErrTypeServiceUnavailable
	ErrTypeGatewayTimeout
	ErrTypeInsufficientStorage
	ErrTypeLoopDetected
)

type Error struct {
	Type ErrType
	Err  error
}

func NewErrorf(errType ErrType, format string, args ...any) *Error {
	return &Error{
		Type: errType,
		Err:  fmt.Errorf(format, args...),
	}
}

func NewError(errType ErrType, err error) *Error {
	return &Error{
		Type: errType,
		Err:  err,
	}
}

func (e *Error) Error() string {
	return e.Err.Error()
}

func (e *Error) Unwrap() error {
	return e.Err
}
