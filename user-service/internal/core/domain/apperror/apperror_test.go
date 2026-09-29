package apperror

import (
	"errors"
	"fmt"
	"testing"
)

// The point of named sentinels is that errors.Is still matches them after a
// layer wraps the error with context, which string comparison would break.
func TestSentinelsSurviveWrapping(t *testing.T) {
	wrapped := fmt.Errorf("UpdatePasswordByID: %w", ErrNotFound)

	if !errors.Is(wrapped, ErrNotFound) {
		t.Error("errors.Is failed on a wrapped ErrNotFound")
	}

	if errors.Is(wrapped, ErrUnauthorized) {
		t.Error("ErrNotFound matched ErrUnauthorized, the sentinels are aliased")
	}
}
