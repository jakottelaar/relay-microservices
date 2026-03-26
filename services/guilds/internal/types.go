package internal

type ChannelType int16

const (
	TextChannel  ChannelType = 1
	VoiceChannel ChannelType = 2
)

type CreateGuildRequest struct {
	Name        string  `form:"name" binding:"required,min=1,max=100"`
	Description *string `form:"description" binding:"omitempty,max=300"`
}

type GuildResponse struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	Icon        *string `json:"icon,omitempty"`
	OwnerID     string  `json:"owner_id"` // change to user object later
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at,omitempty"`
}

type CreateGuildChannelRequest struct {
	Name  string      `json:"name" binding:"required,min=1,max=100"`
	Type  ChannelType `json:"type" binding:"omitempty,oneof=1 2"`
	Topic *string     `json:"topic" binding:"omitempty,max=1024"`
}

type GuildChannelResponse struct {
	ID        string      `json:"id"`
	Name      string      `json:"name"`
	Type      ChannelType `json:"type"`
	Topic     *string     `json:"topic,omitempty"`
	GuildID   string      `json:"guild_id"`
	CreatedAt string      `json:"created_at"`
	UpdatedAt string      `json:"updated_at,omitempty"`
}