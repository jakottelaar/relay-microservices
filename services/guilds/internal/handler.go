package internal

import (
	"mime/multipart"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jakottelaar/relay-microservices/shared/errors"
	"github.com/jakottelaar/relay-microservices/shared/sonyflake"
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
    if err := c.ShouldBind(&req); err != nil {
        h.log.Warn("Invalid request body", zap.Error(err),
            zap.String("path", c.Request.URL.Path),
        )
        c.Error(errors.NewValidationError(err))
        return
    }

    // Icon is optional — nil if not provided
    var iconFile *multipart.FileHeader
    file, err := c.FormFile("icon")
    if err == nil {
        iconFile = file
    }

    h.log.Info("Create guild attempt",
        zap.String("guild_name", req.Name),
        zap.Int64("user_id", userID),
    )

    guildResp, err := h.service.CreateGuild(c.Request.Context(), userID, &req, iconFile)
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

func (h *GuildHandler) GetGuild(c *gin.Context) {
    guildID, err := sonyflake.ParseID(c.Param("id"))
    if err != nil {
        c.Error(err)
        return
    }

    h.log.Info("Get guild attempt",
        zap.Int64("guild_id", guildID),
    )

    guildResp, err := h.service.GetGuild(c.Request.Context(), guildID)
    if err != nil {
        h.log.Error("Failed to get guild",
            zap.Int64("guild_id", guildID),
            zap.Error(err),
        )
        _ = c.Error(err)
        return
    }

    h.log.Info("Guild retrieved successfully",
        zap.String("guild_id", guildResp.ID),
        zap.String("guild_name", guildResp.Name),
    )

    c.JSON(http.StatusOK, guildResp)
}

func (h *GuildHandler) CreateGuildChannel(c *gin.Context) {
    guildID, err := sonyflake.ParseID(c.Param("id"))
    if err != nil {
        c.Error(err)
        return
    }

    var req CreateGuildChannelRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        h.log.Warn("Invalid request body", zap.Error(err),
            zap.String("path", c.Request.URL.Path),
        )
        c.Error(errors.NewValidationError(err))
        return
    }

    h.log.Info("Create guild channel attempt",
        zap.String("channel_name", req.Name),
        zap.Int64("guild_id", guildID),
    )

    channelResp, err := h.service.CreateGuildChannel(c.Request.Context(), guildID, &req)
    if err != nil {
        h.log.Error("Failed to create guild channel", zap.Error(err))
        _ = c.Error(err)
        return
    }

    h.log.Info("Guild channel created successfully",
        zap.String("channel_id", channelResp.ID),
        zap.String("channel_name", channelResp.Name),
    )

    c.JSON(http.StatusCreated, channelResp)
}