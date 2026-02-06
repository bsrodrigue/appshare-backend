package domain

import "errors"

// Domain errors - these represent business-level errors.
// The service layer returns these; the handler layer translates them to HTTP.

var (
	// ErrNotFound is returned when a requested entity doesn't exist.
	ErrNotFound = errors.New("resource not found")

	// ErrAlreadyExists is returned when trying to create a duplicate.
	ErrAlreadyExists = errors.New("resource already exists")

	// ErrEmailAlreadyExists is returned when email is taken.
	ErrEmailAlreadyExists = errors.New("email already exists")

	// ErrUsernameAlreadyExists is returned when username is taken.
	ErrUsernameAlreadyExists = errors.New("username already exists")

	// ErrPhoneAlreadyExists is returned when phone number is taken.
	ErrPhoneAlreadyExists = errors.New("phone number already exists")

	// ErrInvalidCredentials is returned when login fails.
	ErrInvalidCredentials = errors.New("invalid credentials")

	// ErrUserInactive is returned when an inactive user tries to authenticate.
	ErrUserInactive = errors.New("user account is inactive")

	// ErrInvalidInput is returned when input validation fails.
	ErrInvalidInput = errors.New("invalid input")

	// ErrUnauthorized is returned when the user lacks permission.
	ErrUnauthorized = errors.New("unauthorized")

	// ErrForbidden is returned when the user is authenticated but lacks access.
	ErrForbidden = errors.New("forbidden")
)

// ValidationError wraps ErrInvalidInput with specific field information.
type ValidationError struct {
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
	return &ValidationError{Field: field, Message: message}
}
