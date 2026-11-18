package transport

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/jljl1337/gostarter/pkg/log"
	"github.com/jljl1337/gostarter/pkg/service"
)

var statusMap = map[service.ErrType]int{
	service.ErrTypeUnauthorized:               http.StatusUnauthorized,
	service.ErrTypePaymentRequired:            http.StatusPaymentRequired,
	service.ErrTypeForbidden:                  http.StatusForbidden,
	service.ErrTypeNotFound:                   http.StatusNotFound,
	service.ErrTypeConflict:                   http.StatusConflict,
	service.ErrTypeGone:                       http.StatusGone,
	service.ErrTypePreconditionFailed:         http.StatusPreconditionFailed,
	service.ErrTypeUnprocessableContent:       http.StatusUnprocessableEntity,
	service.ErrTypeLocked:                     http.StatusLocked,
	service.ErrTypeFailedDependency:           http.StatusFailedDependency,
	service.ErrTypePreconditionRequired:       http.StatusPreconditionRequired,
	service.ErrTypeTooManyRequests:            http.StatusTooManyRequests,
	service.ErrTypeUnavailableForLegalReasons: http.StatusUnavailableForLegalReasons,
	service.ErrTypeInternalServerError:        http.StatusInternalServerError,
	service.ErrTypeBadGateway:                 http.StatusBadGateway,
	service.ErrTypeServiceUnavailable:         http.StatusServiceUnavailable,
	service.ErrTypeGatewayTimeout:             http.StatusGatewayTimeout,
	service.ErrTypeInsufficientStorage:        http.StatusInsufficientStorage,
	service.ErrTypeLoopDetected:               http.StatusLoopDetected,
}

// WriteServiceError writes a JSON response with the appropriate HTTP status
// code, JSON error code, and error message based on the provided service
// error.
//
// If the error is not a service error, it writes a 500 Internal Server Error
// response with a generic error message.
//
// The jsonCodeMap parameter is a map that allows you to specify custom JSON
// error codes for specific service errors. If a matching service error is
// found in the map, the corresponding JSON error code will be used in the
// response.
func WriteServiceError(w http.ResponseWriter, jsonCodeMap map[*service.Error]string, err error) {
	var serviceErr *service.Error

	if !errors.As(err, &serviceErr) {
		WriteInternalErrorf(w, "non-service-error from service: %v", err)
		return
	}

	statusCode := http.StatusInternalServerError
	if mappedStatusCode, exists := statusMap[serviceErr.Type]; exists {
		statusCode = mappedStatusCode
	}

	jsonCode := strconv.Itoa(statusCode)
	for specificError, errJSONCode := range jsonCodeMap {
		if errors.Is(serviceErr, specificError) {
			jsonCode = errJSONCode
			break
		}
	}

	message := serviceErr.Error()

	WriteJSONCodeMessage(w, statusCode, jsonCode, message)
}

// WriteInternalErrorf writes a JSON response with a 500 Internal Server Error
// status code and a formatted error message to the HTTP response writer.
func WriteInternalErrorf(w http.ResponseWriter, format string, args ...any) {
	WriteMessagef(w, http.StatusInternalServerError, format, args...)
}

// WriteMessagef writes a JSON response with the specified status code and a
// formatted message to the HTTP response writer.
func WriteMessagef(w http.ResponseWriter, statusCode int, format string, args ...any) {
	jsonCode := strconv.Itoa(statusCode)
	message := fmt.Sprintf(format, args...)
	WriteJSONCodeMessage(w, statusCode, jsonCode, message)
}

// WriteJSONCodeMessage writes a JSON response with the specified status code,
// json code and message to the HTTP response writer.
//
// If the status code is 500 (Internal Server Error), it logs the error
// message.
func WriteJSONCodeMessage(w http.ResponseWriter, statusCode int, jsonCode string, message string) {
	if statusCode == http.StatusInternalServerError {
		log.Errorf("Internal server error with message: %v", message)
		message = "Internal server error"
	}

	response := map[string]string{
		"code":    jsonCode,
		"message": message,
	}
	writeJSON(w, statusCode, response)
}

// WriteJSON writes a JSON response with the specified status code and
// data to the HTTP response writer.
func WriteJSON(w http.ResponseWriter, statusCode int, data any) {
	if statusCode == http.StatusInternalServerError {
		log.Errorf("Internal server error with data: %v", data)
	}

	writeJSON(w, statusCode, data)
}

func writeJSON(w http.ResponseWriter, statusCode int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Errorf("Failed to write JSON response: %v", err)
	}
}
