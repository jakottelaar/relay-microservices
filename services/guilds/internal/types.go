package internal

type CreateGuildRequest struct {
	Name        string `json:"name" binding:"required,min=3,max=100"`
	Description string `json:"description" binding:"max=500"`
}

type GuildResponse struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description,omitempty"`
	Icon        *string `json:"icon,omitempty"`
	OwnerID     string  `json:"owner_id"` // change to user object later
	CreatedAt   string  `json:"created_at"`
}