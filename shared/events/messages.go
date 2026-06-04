package events

type MessageCreatedEvent struct {
	MessageID string `json:"message_id"`
	ChannelID string `json:"channel_id"`
	AuthorID  string `json:"author_id"`
	Content   string `json:"content"`
	CreatedAt string `json:"created_at"`
}