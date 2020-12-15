package errors

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/golang/glog"

	"gitlab.cee.redhat.com/service/managed-services-api/pkg/api/openapi"
)

const (
	ERROR_CODE_PREFIX = "MGD-SERV-API"

	// HREF for API errors
	ERROR_HREF = "/api/managed-services-api/v1/errors/"

	// Forbidden occurs when a user has been blacklisted
	ErrorForbidden ServiceErrorCode = 4

	// Conflict occurs when a database constraint is violated
	ErrorConflict ServiceErrorCode = 6

	// NotFound occurs when a record is not found in the database
	ErrorNotFound ServiceErrorCode = 7

	// Validation occurs when an object fails validation
	ErrorValidation ServiceErrorCode = 8

	// General occurs when an error fails to match any other error code
	ErrorGeneral ServiceErrorCode = 9

	// NotImplemented occurs when an API REST method is not implemented in a handler
	ErrorNotImplemented ServiceErrorCode = 10

	// Unauthorized occurs when the requester is not authorized to perform the specified action
	ErrorUnauthorized ServiceErrorCode = 11

	// Unauthenticated occurs when the provided credentials cannot be validated
	ErrorUnauthenticated ServiceErrorCode = 15

	// MalformedRequest occurs when the request body cannot be read
	ErrorMalformedRequest ServiceErrorCode = 17

	// Bad Request
	ErrorBadRequest ServiceErrorCode = 21

	// Invalid Search Query
	ErrorFailedToParseSearch ServiceErrorCode = 23

	// Synchronous request not supported
	ErrorSyncActionNotSupported ServiceErrorCode = 103

	ErrorFailedToCreateSSOClient ServiceErrorCode = 106
)

type ServiceErrorCode int

type ServiceErrors []ServiceError

func Find(code ServiceErrorCode) (bool, *ServiceError) {
	for _, err := range Errors() {
		if err.Code == code {
			return true, &err
		}
	}
	return false, nil
}

func Errors() ServiceErrors {
	return ServiceErrors{
		ServiceError{ErrorForbidden, "Forbidden to perform this action", http.StatusForbidden},
		ServiceError{ErrorConflict, "An entity with the specified unique values already exists", http.StatusConflict},
		ServiceError{ErrorNotFound, "Resource not found", http.StatusNotFound},
		ServiceError{ErrorValidation, "General validation failure", http.StatusBadRequest},
		ServiceError{ErrorGeneral, "Unspecified error", http.StatusInternalServerError},
		ServiceError{ErrorNotImplemented, "HTTP Method not implemented for this endpoint", http.StatusMethodNotAllowed},
		ServiceError{ErrorUnauthorized, "Account is unauthorized to perform this action", http.StatusForbidden},
		ServiceError{ErrorUnauthenticated, "Account authentication could not be verified", http.StatusUnauthorized},
		ServiceError{ErrorMalformedRequest, "Unable to read request body", http.StatusBadRequest},
		ServiceError{ErrorBadRequest, "Bad request", http.StatusBadRequest},
		ServiceError{ErrorFailedToParseSearch, "Failed to parse search query", http.StatusBadRequest},
		ServiceError{ErrorSyncActionNotSupported, "Synchronous action is not supported", http.StatusBadRequest},
		ServiceError{ErrorFailedToCreateSSOClient, "failed to create kafka client in the mas sso", http.StatusBadRequest},
	}
}

type ServiceError struct {
	// Code is the numeric and distinct ID for the error
	Code ServiceErrorCode
	// Reason is the context-specific reason the error was generated
	Reason string
	// HttopCode is the HttpCode associated with the error when the error is returned as an API response
	HttpCode int
}

// Reason can be a string with format verbs, which will be replace by the specified values
func New(code ServiceErrorCode, reason string, values ...interface{}) *ServiceError {
	// If the code isn't defined, use the general error code
	var err *ServiceError
	exists, err := Find(code)
	if !exists {
		glog.Errorf("Undefined error code used: %d", code)
		err = &ServiceError{ErrorGeneral, "Unspecified error", 500}
	}

	// If the reason is unspecified, use the default
	if reason != "" {
		err.Reason = fmt.Sprintf(reason, values...)
	}

	return err
}

func (e *ServiceError) Error() string {
	return fmt.Sprintf("%s: %s", CodeStr(e.Code), e.Reason)
}

func (e *ServiceError) AsError() error {
	return fmt.Errorf(e.Error())
}

func (e *ServiceError) Is404() bool {
	return e.Code == NotFound("").Code
}

func (e *ServiceError) IsConflict() bool {
	return e.Code == Conflict("").Code
}

func (e *ServiceError) IsForbidden() bool {
	return e.Code == Forbidden("").Code
}

func (e *ServiceError) AsOpenapiError(operationID string) openapi.Error {
	return openapi.Error{
		Kind:        "Error",
		Id:          strconv.Itoa(int(e.Code)),
		Href:        Href(e.Code),
		Code:        CodeStr(e.Code),
		Reason:      e.Reason,
		OperationId: operationID,
	}
}

func CodeStr(code ServiceErrorCode) string {
	return fmt.Sprintf("%s-%d", ERROR_CODE_PREFIX, code)
}

func Href(code ServiceErrorCode) string {
	return fmt.Sprintf("%s%d", ERROR_HREF, code)
}

func NotFound(reason string, values ...interface{}) *ServiceError {
	return New(ErrorNotFound, reason, values...)
}

func GeneralError(reason string, values ...interface{}) *ServiceError {
	return New(ErrorGeneral, reason, values...)
}

func Unauthorized(reason string, values ...interface{}) *ServiceError {
	return New(ErrorUnauthorized, reason, values...)
}

func Unauthenticated(reason string, values ...interface{}) *ServiceError {
	return New(ErrorUnauthenticated, reason, values...)
}

func Forbidden(reason string, values ...interface{}) *ServiceError {
	return New(ErrorForbidden, reason, values...)
}

func NotImplemented(reason string, values ...interface{}) *ServiceError {
	return New(ErrorNotImplemented, reason, values...)
}

func Conflict(reason string, values ...interface{}) *ServiceError {
	return New(ErrorConflict, reason, values...)
}

func Validation(reason string, values ...interface{}) *ServiceError {
	return New(ErrorValidation, reason, values...)
}

func MalformedRequest(reason string, values ...interface{}) *ServiceError {
	return New(ErrorMalformedRequest, reason, values...)
}

func BadRequest(reason string, values ...interface{}) *ServiceError {
	return New(ErrorBadRequest, reason, values...)
}

func FailedToParseSearch(reason string, values ...interface{}) *ServiceError {
	message := fmt.Sprintf("Failed to parse search query: %s", reason)
	return New(ErrorFailedToParseSearch, message, values...)
}

func SyncActionNotSupported(reason string, values ...interface{}) *ServiceError {
	message := fmt.Sprintf("Synchronous action (%s) is unsupported, use async=true parameter", reason)
	return New(ErrorSyncActionNotSupported, message)
}
func NotMultiAzActionNotSupported(reason string, values ...interface{}) *ServiceError {
	message := "only multi_az is supported, use multi_az=true in Kafka requests"
	return New(ErrorBadRequest, message)
}
