package entity

import "time"

// Token types. A token is only accepted by the flow it was issued for.
const (
	TokenTypeEmailVerification = "email_verification"
	TokenTypeResetPassword     = "reset_password"
)

type VerificationTokenEntity struct {
	ID        int64
	UserID    int64
	Token     string
	TokenType string
	ExpiresAt time.Time
	User      UserEntity
}
