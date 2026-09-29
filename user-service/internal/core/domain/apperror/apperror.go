// Package apperror holds the errors every layer may return, so callers can
// match them with errors.Is instead of comparing error strings.
package apperror

import "errors"

var (
	// ErrNotFound means the requested row does not exist.
	ErrNotFound = errors.New("user not found")
	// ErrUnauthorized means a token is expired, already used, or of the wrong type.
	ErrUnauthorized = errors.New("token expired or invalid")
	// ErrInvalidPassword means the supplied password does not match the stored hash.
	ErrInvalidPassword = errors.New("password is incorrect")
)
