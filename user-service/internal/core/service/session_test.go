package service

import (
	"encoding/json"
	"testing"
	"time"
	"user-service/internal/core/domain/entity"
)

// The session string stored in Redis is read back by CheckToken and unmarshaled
// into entity.JwtUserData, so that exact round-trip must not fail.
func TestToSessionRoundTrip(t *testing.T) {
	user := &entity.UserEntity{ID: 7, Name: "Budi", Email: "budi@example.com"}

	session, err := toSession(user, "token-abc")
	if err != nil {
		t.Fatalf("toSession: %v", err)
	}

	var got entity.JwtUserData
	if err := json.Unmarshal([]byte(session), &got); err != nil {
		t.Fatalf("unmarshal session: %v", err)
	}

	if got.UserID != user.ID {
		t.Errorf("UserID = %d, want %d", got.UserID, user.ID)
	}
	if got.Email != user.Email {
		t.Errorf("Email = %q, want %q", got.Email, user.Email)
	}
	if got.Token != "token-abc" {
		t.Errorf("Token = %q, want %q", got.Token, "token-abc")
	}
	if !got.LoggedIn {
		t.Error("LoggedIn = false, want true")
	}
	if got.CreatedAt.IsZero() {
		t.Error("CreatedAt is zero, the created_at field did not survive the round-trip")
	}
	if time.Since(got.CreatedAt) > time.Minute {
		t.Errorf("CreatedAt = %v, expected about now", got.CreatedAt)
	}
}
