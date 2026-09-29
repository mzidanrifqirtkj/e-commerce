package entity

import "time"

type JwtUserData struct {
	CreatedAt time.Time `json:"created_at"`
	Email     string    `json:"email"`
	LoggedIn  bool      `json:"logged_in"`
	Name      string    `json:"name"`
	Token     string    `json:"token"`
	UserID    int64     `json:"user_id"`
}
