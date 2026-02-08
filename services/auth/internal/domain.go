package internal

import "time"

type Account struct {
	ID           int64 `json:"id"`
	Email        string `json:"email"`
	PasswordHash string `json:"-"`
	LastLoginAt  *time.Time `json:"last_login_at,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at,omitempty"`
}