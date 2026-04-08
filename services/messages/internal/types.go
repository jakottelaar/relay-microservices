package internal

type CreateMessageRequest struct {
	Content *string `json:"content" binding:"required,min=1,max=2000"`
}

type MessageResponse struct {
	ID        string `json:"id"`
	Content   string `json:"content"`
	AuthorID  string `json:"author_id"` // change to user object later
	ChannelID string `json:"channel_id"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at,omitempty"`
}