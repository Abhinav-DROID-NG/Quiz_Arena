package utils

import (
	"errors"
	"fmt"
)

// Sentinel errors used across the application.
var (
	ErrNotFound       = errors.New("not found")
	ErrUnauthorized   = errors.New("unauthorized")
	ErrForbidden      = errors.New("forbidden")
	ErrConflict       = errors.New("conflict")
	ErrBadRequest     = errors.New("bad request")
	ErrTimerExpired   = errors.New("question timer expired")
	ErrSessionClosed  = errors.New("session is not active")
	ErrSessionLocked  = errors.New("session belongs to another user")
	ErrAlreadyAnswered = errors.New("question already answered in this session")
)

// AppError wraps a sentinel error with a human-readable message.
type AppError struct {
	Err     error
	Message string
}

// Error implements the error interface.
func (e *AppError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("%v: %s", e.Err, e.Message)
	}
	return e.Err.Error()
}

// Unwrap allows errors.Is / errors.As to work on AppError.
func (e *AppError) Unwrap() error { return e.Err }

// Wrap creates an AppError from an existing error with additional context.
func Wrap(err error, msg string) *AppError {
	return &AppError{Err: err, Message: msg}
}
