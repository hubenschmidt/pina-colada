package errors

import "errors"

// Sentinel errors for not-found conditions.
var (
	ErrNotFound        = errors.New("not found")
	ErrSessionNotFound = errors.New("session not found")
	ErrUserNotFound    = errors.New("user not found")
)
