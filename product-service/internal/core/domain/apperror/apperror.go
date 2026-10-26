// Package apperror holds the errors every layer may return, so callers can
// match them with errors.Is instead of comparing error strings.
package apperror

import "errors"

var (
	// ErrNotFound means the requested row does not exist.
	ErrNotFound = errors.New("404, not found")
	// ErrUnauthorized means a token is expired, already used, or of the wrong type.
	ErrUnauthorized = errors.New("401, unauthorized")
	// ErrConflict means the requested operation would violate a unique constraint.
	ErrConflict = errors.New("400, conflict with existing data")
	// ErrInvalidRequest means the client sent a malformed or out-of-range value.
	ErrInvalidRequest = errors.New("400, invalid request")
)
