package internal

import "time"

type Account struct {
	ID           int64 `json:"id"`
	Email        string `json:"email"`
	PasswordHash string `json:"-"`
	LastLoginAt  *time.Time `json:"last_login_at,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    *time.Time `json:"updated_at,omitempty"`
}

type Session struct {
	ID            int64 `json:"id"`
	UserAccountID int64 `json:"user_account_id"`
	RefreshTokenHash string `json:"-"`
	UserAgent     string `json:"user_agent"`
	ClientIP      string `json:"client_ip"`
	ExpiresAt time.Time `json:"expires_at"`
	RevokedAt *time.Time `json:"revoked_at,omitempty"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}