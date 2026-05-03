package internal

import (
	"context"

	"github.com/coder/websocket"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Handler struct {
	hub *Hub
	log *zap.Logger
}

func NewHandler(hub *Hub, log *zap.Logger) *Handler {
	return &Handler{
		hub: hub,
		log: log,
	}
}

func (h *Handler) ServeWS(c *gin.Context) {
	userID := c.GetInt64("user_id")

	conn, err := websocket.Accept(c.Writer, c.Request, &websocket.AcceptOptions{
		OriginPatterns: []string{"*"}, // TODO: Add via env vars
	})
	if err != nil {
		h.log.Error("failed to accept websocket connection", zap.Error(err))
		return
	}

	client := &Client{
		hub: h.hub,
		conn: conn,
		send: make(chan []byte, 256),
		userID: userID,
	}
	
	h.hub.register <- client
	defer func() { h.hub.unregister <- client }()

	ctx, cancel := context.WithCancel(c.Request.Context())
	defer cancel()

	go client.writePump(ctx)

	closeCtx := conn.CloseRead(ctx)
	<-closeCtx.Done()
}