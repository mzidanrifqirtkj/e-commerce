// Package apperror holds the errors every layer may return, so callers can
// match them with errors.Is instead of comparing error strings.
package apperror

// Error is a sentinel domain error that also carries the HTTP status the
// handler should answer with.
type Error struct {
	Code    int
	Message string
}

func (e *Error) Error() string { return e.Message }

var (
	// ErrNotFound means the requested row does not exist.
	ErrNotFound = &Error{Code: 404, Message: "not found"}
	// ErrUnauthorized means a token is expired, already used, or of the wrong type.
	ErrUnauthorized = &Error{Code: 401, Message: "unauthorized"}
	// ErrConflict means the requested operation would violate a unique constraint.
	ErrConflict = &Error{Code: 409, Message: "conflict with existing data"}
	// ErrInvalidRequest means the client sent a malformed or out-of-range value.
	ErrInvalidRequest = &Error{Code: 400, Message: "invalid request"}
)
