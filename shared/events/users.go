package events

import "time"

const (
	SubjectUserAccountCreated = "user.account.created"
)

type UserAccountCreatedEvent struct {
	UserID    int64   `json:"user_id"`
	Username  string   `json:"username"`
	CreatedAt time.Time `json:"created_at"`
}