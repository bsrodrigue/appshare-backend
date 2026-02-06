// Package handler contains HTTP handlers (transport layer).
// Handlers are responsible for:
//   - Parsing requests
//   - Calling services
//   - Formatting responses
//   - Translating domain errors to HTTP errors
package handler

import (
	"errors"
	"net/http"

	"github.com/bsrodrigue/appshare-backend/internal/domain"
	"github.com/danielgtaylor/huma/v2"
)

// ApiResponse is the standard response wrapper for all endpoints.
type ApiResponse[T any] struct {
	Status  int    `json:"status" doc:"HTTP status code"`
	Message string `json:"message" doc:"Brief description of the response"`
	Data    T      `json:"data" doc:"The actual response payload"`
}

// mapDomainError translates domain errors to huma HTTP errors.
func mapDomainError(err error) error {
	if err == nil {
		return nil
	}

	// Check for specific domain errors
	switch {
	case errors.Is(err, domain.ErrNotFound):
		return huma.Error404NotFound("resource not found")

	case errors.Is(err, domain.ErrEmailAlreadyExists):
		return huma.Error409Conflict("email already exists")

	case errors.Is(err, domain.ErrUsernameAlreadyExists):
		return huma.Error409Conflict("username already exists")

	case errors.Is(err, domain.ErrPhoneAlreadyExists):
		return huma.Error409Conflict("phone number already exists")

	case errors.Is(err, domain.ErrAlreadyExists):
		return huma.Error409Conflict("resource already exists")

	case errors.Is(err, domain.ErrInvalidCredentials):
		return huma.Error401Unauthorized("invalid credentials")

	case errors.Is(err, domain.ErrUserInactive):
		return huma.Error403Forbidden("user account is inactive")

	case errors.Is(err, domain.ErrUnauthorized):
		return huma.Error401Unauthorized("unauthorized")

	case errors.Is(err, domain.ErrForbidden):
		return huma.Error403Forbidden("you don't have permission to access this resource")

	case errors.Is(err, domain.ErrInvalidInput):
		// Check if it's a ValidationError with field info
		var validationErr *domain.ValidationError
		if errors.As(err, &validationErr) {
			return huma.Error422UnprocessableEntity(validationErr.Error())
		}
		return huma.Error400BadRequest("invalid input")

	default:
		// Log the actual error here in production
		return huma.Error500InternalServerError("internal server error", err)
	}
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
