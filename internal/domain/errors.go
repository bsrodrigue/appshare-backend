package domain

import "errors"

// ErrorCode represents machine-readable error codes for API clients.
// These codes are stable and can be used for client-side logic.
type ErrorCode string

const (
	// General errors
	CodeInternal      ErrorCode = "INTERNAL_ERROR"
	CodeNotFound      ErrorCode = "NOT_FOUND"
	CodeAlreadyExists ErrorCode = "ALREADY_EXISTS"
	CodeInvalidInput  ErrorCode = "INVALID_INPUT"
	CodeValidation    ErrorCode = "VALIDATION_ERROR"

	// Authentication errors
	CodeUnauthorized       ErrorCode = "UNAUTHORIZED"
	CodeInvalidCredentials ErrorCode = "INVALID_CREDENTIALS"
	CodeTokenExpired       ErrorCode = "TOKEN_EXPIRED"
	CodeTokenInvalid       ErrorCode = "TOKEN_INVALID"

	// Authorization errors
	CodeForbidden        ErrorCode = "FORBIDDEN"
	CodeInsufficientRole ErrorCode = "INSUFFICIENT_ROLE"

	// User-specific errors
	CodeEmailExists    ErrorCode = "EMAIL_ALREADY_EXISTS"
	CodeUsernameExists ErrorCode = "USERNAME_ALREADY_EXISTS"
	CodePhoneExists    ErrorCode = "PHONE_ALREADY_EXISTS"
	CodeUserInactive   ErrorCode = "USER_INACTIVE"

	// Project-specific errors
	CodeProjectNotFound ErrorCode = "PROJECT_NOT_FOUND"
	CodeNotProjectOwner ErrorCode = "NOT_PROJECT_OWNER"

	// Application-specific errors
	CodeApplicationNotFound ErrorCode = "APPLICATION_NOT_FOUND"
	CodePackageNameExists   ErrorCode = "PACKAGE_NAME_EXISTS"

	// Release-specific errors
	CodeReleaseNotFound    ErrorCode = "RELEASE_NOT_FOUND"
	CodeReleaseExists      ErrorCode = "RELEASE_EXISTS"
	CodeInvalidVersionCode ErrorCode = "INVALID_VERSION_CODE"
)

// AppError is the base error type for all domain errors.
// It contains both a machine-readable code and a human-readable message.
type AppError struct {
	Code    ErrorCode
	Message string
	Err     error // Wrapped error (optional)
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Err
}

// Is allows errors.Is() to match by error code.
func (e *AppError) Is(target error) bool {
	t, ok := target.(*AppError)
	if !ok {
		return false
	}
	return e.Code == t.Code
}

// NewAppError creates a new AppError with the given code and message.
func NewAppError(code ErrorCode, message string) *AppError {
	return &AppError{Code: code, Message: message}
}

// WrapError creates an AppError that wraps another error.
func WrapError(code ErrorCode, message string, err error) *AppError {
	return &AppError{Code: code, Message: message, Err: err}
}

// ============================================================================
// Pre-defined domain errors
// These are sentinel errors that can be compared with errors.Is()
// ============================================================================

var (
	// General errors
	ErrNotFound      = &AppError{Code: CodeNotFound, Message: "resource not found"}
	ErrAlreadyExists = &AppError{Code: CodeAlreadyExists, Message: "resource already exists"}
	ErrInvalidInput  = &AppError{Code: CodeInvalidInput, Message: "invalid input"}
	ErrInternal      = &AppError{Code: CodeInternal, Message: "internal server error"}

	// Authentication errors
	ErrUnauthorized       = &AppError{Code: CodeUnauthorized, Message: "unauthorized"}
	ErrInvalidCredentials = &AppError{Code: CodeInvalidCredentials, Message: "invalid credentials"}
	ErrTokenExpired       = &AppError{Code: CodeTokenExpired, Message: "token has expired"}
	ErrTokenInvalid       = &AppError{Code: CodeTokenInvalid, Message: "token is invalid"}

	// Authorization errors
	ErrForbidden = &AppError{Code: CodeForbidden, Message: "you don't have permission to access this resource"}

	// User-specific errors
	ErrEmailAlreadyExists    = &AppError{Code: CodeEmailExists, Message: "email already exists"}
	ErrUsernameAlreadyExists = &AppError{Code: CodeUsernameExists, Message: "username already exists"}
	ErrPhoneAlreadyExists    = &AppError{Code: CodePhoneExists, Message: "phone number already exists"}
	ErrUserInactive          = &AppError{Code: CodeUserInactive, Message: "user account is inactive"}

	// Project-specific errors
	ErrProjectNotFound = &AppError{Code: CodeProjectNotFound, Message: "project not found"}
	ErrNotProjectOwner = &AppError{Code: CodeNotProjectOwner, Message: "you are not the project owner"}

	// Application-specific errors
	ErrApplicationNotFound = &AppError{Code: CodeApplicationNotFound, Message: "application not found"}
	ErrPackageNameExists   = &AppError{Code: CodePackageNameExists, Message: "package name already exists"}

	// Release-specific errors
	ErrReleaseNotFound = &AppError{Code: CodeReleaseNotFound, Message: "release not found"}
	ErrReleaseExists   = &AppError{Code: CodeReleaseExists, Message: "release already exists"}
)

// ValidationError provides field-level validation error information.
type ValidationError struct {
	Code    ErrorCode
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return e.Field + ": " + e.Message
}

func (e *ValidationError) Unwrap() error {
	return ErrInvalidInput
}

// NewValidationError creates a new validation error for a specific field.
func NewValidationError(field, message string) *ValidationError {
	return &ValidationError{
		Code:    CodeValidation,
		Field:   field,
		Message: message,
	}
}

// GetErrorCode extracts the error code from any error.
// Returns CodeInternal if the error doesn't have a code.
func GetErrorCode(err error) ErrorCode {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Code
	}

	var valErr *ValidationError
	if errors.As(err, &valErr) {
		return valErr.Code
	}

	return CodeInternal
}

// GetErrorMessage extracts a user-friendly message from any error.
func GetErrorMessage(err error) string {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Message
	}

	var valErr *ValidationError
	if errors.As(err, &valErr) {
		return valErr.Message
	}

	return "an unexpected error occurred"
}
