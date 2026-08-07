package transport

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/jljl1337/gostarter/pkg/core/service"
	"github.com/jljl1337/gostarter/pkg/shared/log"
)

/*
GenericErrorMap maps specific service error codes to generic service error
codes. It is used to standardize error handling across the application. This
map is intended to be extended to include more specific error codes as needed.
*/
var GenericErrorMap = map[service.ErrorCode]service.ErrorCode{
	service.ErrCodeBadRequest:    service.ErrCodeBadRequest,
	service.ErrCodeUnauthorized:  service.ErrCodeUnauthorized,
	service.ErrCodeForbidden:     service.ErrCodeForbidden,
	service.ErrCodeNotFound:      service.ErrCodeNotFound,
	service.ErrCodeConflict:      service.ErrCodeConflict,
	service.ErrCodeUnprocessable: service.ErrCodeUnprocessable,
	service.ErrCodeInternal:      service.ErrCodeInternal,

	service.ErrCodeUsernameTaken:      service.ErrCodeConflict,
	service.ErrCodeInvalidCredentials: service.ErrCodeUnauthorized,
}

/*
JsonCodeMap maps specific service error codes to JSON error codes. It is used
to provide consistent JSON error responses across the application. Similar to
GenericErrorMap, this map can be extended to include more specific error codes
as needed. The JSON error codes are intended to be used by the frontend or any
API consumers to handle errors in a standardized way.
*/
var JsonCodeMap = map[service.ErrorCode]string{
	service.ErrCodeUsernameTaken:      "gsUsernameTaken",
	service.ErrCodeInvalidCredentials: "gsInvalidCredentials",
}

/*
ResponseHandler is a struct that handles HTTP responses for the application.
It maps service error codes to generic error codes, HTTP status codes, and
JSON error codes. It provides methods to write JSON responses and error
responses to the HTTP response writer.
*/
type ResponseHandler struct {
	genericErrorMap map[service.ErrorCode]service.ErrorCode
	httpStatusMap   map[service.ErrorCode]int

	jsonCodeMap map[service.ErrorCode]string
}

/*
NewDefaultResponseHandler creates a new ResponseHandler with default mappings
for generic error codes and JSON error codes. It returns a pointer to the
ResponseHandler.

This function should not be used in application code. Instead, use
NewResponseHandler with custom mappings to ensure that error handling is
tailored to the specific needs of your application.
*/
func NewDefaultResponseHandler() *ResponseHandler {
	return NewResponseHandler(GenericErrorMap, JsonCodeMap)
}

/*
NewResponseHandler creates a new ResponseHandler with the provided mappings
for generic error codes and JSON error codes. It returns a pointer to the
ResponseHandler.
*/
func NewResponseHandler(
	genericErrorMap map[service.ErrorCode]service.ErrorCode,
	jsonCodeMap map[service.ErrorCode]string,
) *ResponseHandler {
	var HTTPStatusMap = map[service.ErrorCode]int{
		service.ErrCodeBadRequest:    http.StatusBadRequest,
		service.ErrCodeUnauthorized:  http.StatusUnauthorized,
		service.ErrCodeForbidden:     http.StatusForbidden,
		service.ErrCodeNotFound:      http.StatusNotFound,
		service.ErrCodeConflict:      http.StatusConflict,
		service.ErrCodeUnprocessable: http.StatusUnprocessableEntity,
		service.ErrCodeInternal:      http.StatusInternalServerError,
	}

	return &ResponseHandler{
		genericErrorMap: genericErrorMap,
		httpStatusMap:   HTTPStatusMap,
		jsonCodeMap:     jsonCodeMap,
	}
}

/*
WriteMessage writes a JSON response with a message and status code to
the HTTP response writer. It also includes a JSON error code in the response.
*/
func (rh *ResponseHandler) WriteMessage(w http.ResponseWriter, message string, statusCode int) {
	jsonCode := strconv.Itoa(statusCode)
	rh.WriteJSONCodeMessage(w, message, statusCode, jsonCode)
}

// WriteErrorf writes an error response to the HTTP response writer with a
// formatted message. It creates a new internal error with the provided format
// and arguments, and then calls [*ResponseHandler.WriteServiceError] to handle
// the response.
func (rh *ResponseHandler) WriteErrorf(w http.ResponseWriter, format string, args ...any) {
	message := fmt.Sprintf(format, args...)
	rh.WriteServiceError(w, errors.New(message))
}

/*
WriteServiceError writes an error response to the HTTP response writer. It
maps the provided error to a generic service error, determines the appropriate
HTTP status code and JSON error code, and writes a JSON response with the
error message.

It is also safe to write any internal errors with this function, it will log
the error and return a generic internal server error message to the client.
*/
func (rh *ResponseHandler) WriteServiceError(w http.ResponseWriter, err error) {
	var serviceErr *service.ServiceError

	var statusCode int
	var jsonCode string
	var message string

	if errors.As(err, &serviceErr) {
		genericErr := rh.mapToGenericServiceError(serviceErr)
		statusCode = rh.mapToHTTPStatus(genericErr)
		if *genericErr == *serviceErr {
			jsonCode = strconv.Itoa(statusCode)
		} else {
			jsonCode = rh.mapToJSONCode(serviceErr)
		}
		message = genericErr.Message

		if statusCode == http.StatusInternalServerError {
			log.Errorf("Internal server error: service error: %v", genericErr)
			jsonCode = "500"
			message = "Internal server error"
		}
	} else {
		log.Errorf("Internal server error: unknown error: %v", err)
		statusCode = http.StatusInternalServerError
		jsonCode = "500"
		message = "Internal server error"
	}

	rh.WriteJSONCodeMessage(w, message, statusCode, jsonCode)
}

func (rh *ResponseHandler) mapToGenericServiceError(err error) *service.ServiceError {
	var serviceErr *service.ServiceError
	if errors.As(err, &serviceErr) {
		if genericCode, exists := rh.genericErrorMap[serviceErr.Code]; exists {
			return service.NewServiceError(genericCode, serviceErr.Message)
		}
	}
	return service.NewServiceError(service.ErrCodeInternal, err.Error())
}

func (rh *ResponseHandler) mapToHTTPStatus(err *service.ServiceError) int {
	if status, exists := rh.httpStatusMap[err.Code]; exists {
		return status
	}
	return http.StatusInternalServerError
}

func (rh *ResponseHandler) mapToJSONCode(err *service.ServiceError) string {
	if code, exists := rh.jsonCodeMap[err.Code]; exists {
		return code
	}
	return "500"
}

/*
WriteJSONCodeMessage writes a JSON response with a message, status code,
and JSON error code to the HTTP response writer.
*/
func (rh *ResponseHandler) WriteJSONCodeMessage(w http.ResponseWriter, message string, statusCode int, code string) {
	response := map[string]string{
		"code":    code,
		"message": message,
	}
	rh.WriteJSON(w, statusCode, response)
}

/*
WriteJSON writes a JSON response with the specified status code and
data to the HTTP response writer.
*/
func (rh *ResponseHandler) WriteJSON(w http.ResponseWriter, statusCode int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Errorf("Failed to write JSON response: %v", err)
	}
}
