package internal

import "time"

type CreateUserRequest struct {
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
}

type ProfileResponse struct {
	ID          int64     `json:"id"`
	Username    string    `json:"username"`
	AvatarURL   string    `json:"avatar_url"`
	Bio         string    `json:"bio"`
	CreatedAt   time.Time `json:"created_at"`
}