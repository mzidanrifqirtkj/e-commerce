// Package apperror holds the errors every layer may return, so callers can
// match them with errors.Is instead of comparing error strings.
package apperror

import "errors"

var (
	// ErrNotFound means the requested row does not exist.
	ErrNotFound = errors.New("404, user not found")
	// ErrUnauthorized means a token is expired, already used, or of the wrong type.
	ErrUnauthorized = errors.New("401, unauthorized")
	// ErrInvalidPassword means the supplied password does not match the stored hash.
	ErrInvalidPassword = errors.New("400, password is incorrect")
	// ErrConflict means the requested operation would violate a unique constraint.
	ErrConflict = errors.New("400, conflict with existing data")
)
