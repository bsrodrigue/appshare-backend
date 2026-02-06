// Package handler contains HTTP handlers (transport layer).
// Handlers are responsible for:
//   - Parsing requests
//   - Calling services
//   - Formatting responses
//   - Translating domain errors to HTTP errors
package handler

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/bsrodrigue/appshare-backend/internal/domain"
	"github.com/danielgtaylor/huma/v2"
)

// ApiResponse is the standard response wrapper for all endpoints.
// Clients can use the 'code' field for programmatic error handling.
type ApiResponse[T any] struct {
	Status  int              `json:"status" doc:"HTTP status code"`
	Code    domain.ErrorCode `json:"code,omitempty" doc:"Machine-readable error code for client-side handling"`
	Message string           `json:"message" doc:"Brief description of the response"`
	Data    T                `json:"data" doc:"The actual response payload"`
}

// ErrorDetail provides additional error information for clients.
// It implements the error interface so it can be passed to huma error functions.
type ErrorDetail struct {
	Code    domain.ErrorCode `json:"code" doc:"Machine-readable error code"`
	Message string           `json:"message" doc:"Human-readable error message"`
	Field   string           `json:"field,omitempty" doc:"Field name for validation errors"`
}

// Error implements the error interface.
func (e *ErrorDetail) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("%s: %s (code: %s)", e.Field, e.Message, e.Code)
	}
	return fmt.Sprintf("%s (code: %s)", e.Message, e.Code)
}

// mapDomainError translates domain errors to huma HTTP errors with proper error codes.
func mapDomainError(err error) error {
	if err == nil {
		return nil
	}

	// Extract error code and message from domain error
	code := domain.GetErrorCode(err)
	message := domain.GetErrorMessage(err)

	// Create the error detail
	detail := &ErrorDetail{Code: code, Message: message}

	// Check for validation errors first (they have field info)
	var valErr *domain.ValidationError
	if errors.As(err, &valErr) {
		detail.Field = valErr.Field
		return huma.Error422UnprocessableEntity(valErr.Message, detail)
	}

	// Check for specific domain errors and map to HTTP status
	var appErr *domain.AppError
	if errors.As(err, &appErr) {
		switch appErr.Code {
		case domain.CodeNotFound, domain.CodeProjectNotFound, domain.CodeApplicationNotFound, domain.CodeReleaseNotFound:
			return huma.Error404NotFound(message, detail)

		case domain.CodeEmailExists, domain.CodeUsernameExists, domain.CodePhoneExists, domain.CodeAlreadyExists, domain.CodePackageNameExists, domain.CodeReleaseExists:
			return huma.Error409Conflict(message, detail)

		case domain.CodeInvalidCredentials, domain.CodeUnauthorized, domain.CodeTokenExpired, domain.CodeTokenInvalid:
			return huma.Error401Unauthorized(message, detail)

		case domain.CodeUserInactive, domain.CodeForbidden, domain.CodeNotProjectOwner, domain.CodeInsufficientRole:
			return huma.Error403Forbidden(message, detail)

		case domain.CodeInvalidInput, domain.CodeValidation:
			return huma.Error400BadRequest(message, detail)

		case domain.CodeInternal:
			return huma.Error500InternalServerError(message, detail)
		}
	}

	// Default to internal server error
	return huma.Error500InternalServerError("an unexpected error occurred", &ErrorDetail{
		Code:    domain.CodeInternal,
		Message: "an unexpected error occurred",
	})
}

// successResponse creates a standard success response.
func successResponse[T any](status int, message string, data T) ApiResponse[T] {
	return ApiResponse[T]{
		Status:  status,
		Message: message,
		Data:    data,
	}
}

// emptyData is used for responses that don't return data.
type emptyData struct{}

// ok creates a 200 OK response.
func ok[T any](message string, data T) ApiResponse[T] {
	return successResponse(http.StatusOK, message, data)
}

// created creates a 201 Created response.
func created[T any](message string, data T) ApiResponse[T] {
	return successResponse(http.StatusCreated, message, data)
}

// noContent creates a 204 No Content response.
func noContent(message string) ApiResponse[emptyData] {
	return ApiResponse[emptyData]{
		Status:  http.StatusNoContent,
		Message: message,
		Data:    emptyData{},
	}
}
