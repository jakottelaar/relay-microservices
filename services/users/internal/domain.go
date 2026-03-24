package internal

import "time"

type User struct {
	ID        int64
	Username  string
	Avatar    string
	Bio       string
	CreatedAt time.Time
	UpdatedAt time.Time
}