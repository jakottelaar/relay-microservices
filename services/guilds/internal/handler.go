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
    userID := c.GetInt64("user_id")

    var req CreateGuildRequest
    if err := c.ShouldBind(&req); err != nil { // Use ShouldBind to handle both JSON and form data
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

func (h *GuildHandler) CreateGuildMember(c *gin.Context) {
    guildID, err := sonyflake.ParseID(c.Param("id"))
    if err != nil {
        c.Error(err)
        return
    }

    userID, err := sonyflake.ParseID(c.Param("user_id"))
    if err != nil {
        c.Error(err)
        return
    }

    var req GuildMemberRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        h.log.Warn("Invalid request body", zap.Error(err),
            zap.String("path", c.Request.URL.Path),
        )
        c.Error(errors.NewValidationError(err))
        return
    }

    h.log.Info("Create guild member attempt",
        zap.Int64("guild_id", guildID),
        zap.Int64("user_id", userID),
    )

    memberResp, err := h.service.CreateGuildMember(c.Request.Context(), guildID, userID, req.Nick)
    if err != nil {
        h.log.Error("Failed to create guild member", zap.Error(err))
        _ = c.Error(err)
        return
    }

    h.log.Info("Guild member created successfully",
        zap.Int64("guild_id", guildID),
        zap.Int64("user_id", userID),
    )

    c.JSON(http.StatusCreated, memberResp)
}

func (h *GuildHandler) GetGuildChannels(c *gin.Context) {
    guildID, err := sonyflake.ParseID(c.Param("id"))
    if err != nil {
        c.Error(err)
        return
    }

    h.log.Info("Get guild channels attempt",
        zap.Int64("guild_id", guildID),
    )

    channelsResp, err := h.service.GetGuildChannels(c.Request.Context(), guildID)
    if err != nil {
        h.log.Error("Failed to get guild channels", zap.Error(err))
        _ = c.Error(err)
        return
    }

    h.log.Info("Guild channels retrieved successfully",
        zap.Int64("guild_id", guildID),
        zap.Int("channel_count", len(channelsResp)),
    )

    c.JSON(http.StatusOK, channelsResp)
}