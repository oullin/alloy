package errorsx

import (
	"errors"
	"fmt"
)

// UnauthenticatedError is returned when authentication fails.
type UnauthenticatedError struct {
	Message      string
	Guards       []string
	RedirectPath string
}

// NewUnauthenticatedError creates an UnauthenticatedError.

// UnauthorizedError is returned when authorization fails.
type UnauthorizedError struct {
	Message    string
	StatusCode int
}

func (e *UnauthenticatedError) Error() string {
	if e.Message != "" {
		return e.Message
	}

	return "unauthenticated"
}

func NewUnauthenticatedError(guards []string, redirectPath string) *UnauthenticatedError {
	return &UnauthenticatedError{
		Message:      "unauthenticated",
		Guards:       guards,
		RedirectPath: redirectPath,
	}
}

func (e *UnauthorizedError) Error() string {
	if e.Message != "" {
		return e.Message
	}

	return fmt.Sprintf("this action is unauthorized (HTTP %d)", e.StatusCode)
}

// NewUnauthorizedError creates an UnauthorizedError.
func NewUnauthorizedError(message string, statusCode int) *UnauthorizedError {
	if statusCode == 0 {
		statusCode = 403
	}

	return &UnauthorizedError{Message: message, StatusCode: statusCode}
}

var (
	// ErrInvalidProvider is returned when a user provider driver is not registered.
	ErrInvalidProvider = errors.New("auth: invalid user provider driver")
	// ErrInvalidGuard is returned when a guard driver is not registered.
	ErrInvalidGuard = errors.New("auth: invalid guard driver")
	// ErrUserNotFound is returned when no user matches the given credentials or ID.
	ErrUserNotFound = errors.New("auth: user not found")
)
