package internal

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jakottelaar/relay-microservices/shared/errors"
	"go.uber.org/zap"
)

type GuildHandler struct {
	service GuildService
	log     *zap.Logger
}

func NewGuildHandler(service GuildService, log *zap.Logger) *GuildHandler {
	return &GuildHandler{
		service: service, 
		log: log,
	}
}

func (h *GuildHandler) CreateGuild(c *gin.Context) {
	userID := c.GetInt64("userID")

	var req CreateGuildRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Warn("Invalid request body", zap.Error(err),
            zap.String("path", c.Request.URL.Path),
        )
        c.Error(errors.NewValidationError(err))
        return
	}

	h.log.Info("Create guild attempt",
		zap.String("guild_name", req.Name),
		zap.Int64("user_id", userID),
	)

	guildResp, err := h.service.CreateGuild(c.Request.Context(), userID, &req)
	if err != nil {
		h.log.Error("Failed to create guild", zap.Error(err))
		_ = c.Error(err)
		return
	}

	h.log.Info("Guild created successfully",
		zap.String("guild_id", guildResp.ID),
		zap.String("guild_name", guildResp.Name),
	)

	c.JSON(http.StatusCreated, guildResp)
}