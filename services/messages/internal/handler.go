package internal

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jakottelaar/relay-microservices/shared/errors"
	"github.com/jakottelaar/relay-microservices/shared/sonyflake"
	"go.uber.org/zap"
)

type MessageHandler struct {
	service MessageService
	log     *zap.Logger
}

func NewMessageHandler(service MessageService, log *zap.Logger) *MessageHandler {
	return &MessageHandler{
		service: service,
		log: log,
	}
}

func (h *MessageHandler) CreateMessage(c *gin.Context) {
	userID := c.GetInt64("user_id")

	channelID, err := sonyflake.ParseID(c.Param("channel_id"))
    if err != nil {
        c.Error(err)
        return
    }

    var req CreateMessageRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        h.log.Warn("Invalid request body", zap.Error(err),
            zap.String("path", c.Request.URL.Path),
        )
        c.Error(errors.NewValidationError(err))
        return
    }

	h.log.Info("Create message attempt",
		zap.Int64("channel_id", channelID),
		zap.Int64("user_id", userID),
	)

	messageResp, err := h.service.CreateMessage(c.Request.Context(), userID, channelID, &req)
	if err != nil {
		h.log.Error("Failed to create message", zap.Error(err))
		_ = c.Error(err)
		return
	}

	h.log.Info("Message created successfully",
		zap.String("message_id", messageResp.ID),
		zap.Int64("channel_id", channelID),
	)

	c.JSON(http.StatusCreated, messageResp)
}