package internal

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/jakottelaar/relay-microservices/shared/events"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

type EventHandler struct {
	nc     *nats.Conn
	hub    *Hub
	guilds *GuildsClient
	log    *zap.Logger
	subs   []*nats.Subscription
}

func NewEventHandler(nc *nats.Conn, hub *Hub, guilds *GuildsClient, log *zap.Logger) *EventHandler {
	return &EventHandler{
		nc:     nc,
		hub:    hub,
		guilds: guilds,
		log:    log,
	}
}

func (e *EventHandler) Subscribe() error {
	sub, err := e.nc.Subscribe("messages.channels.>", e.handleChannelMessage)
	if err != nil {
		return err
	}

	e.subs = append(e.subs, sub)
	e.log.Info("subscribed to NATS subject",
		zap.String("pattern", "messages.channels.>"),
	)

	return nil
}

func (e *EventHandler) handleChannelMessage(msg *nats.Msg) {
	parts := strings.Split(msg.Subject, ".")
	if len(parts) < 4 {
		e.log.Error("unexpected subject format", zap.String("subject", msg.Subject))
		return
	}

	action := parts[3]
	switch action {
	case "created":
		e.handleMessageCreated(msg.Data)
	default:
		e.log.Warn("unknown action", zap.String("action", action))
	}
}

func (e *EventHandler) handleMessageCreated(data []byte) {
	var payload events.MessageCreatedEvent
	if err := json.Unmarshal(data, &payload); err != nil {
		e.log.Error("failed to unmarshal message created event", zap.Error(err))
		return
	}

	e.log.Info("received message created event",
		zap.String("message_id", payload.MessageID),
		zap.String("channel_id", payload.ChannelID),
	)

	channelID, err := strconv.ParseInt(payload.ChannelID, 10, 64)
	if err != nil {
		e.log.Error("invalid channel_id in event", zap.String("channel_id", payload.ChannelID))
		return
	}

	recipients, err := e.guilds.GetChannelMembers(channelID)
	if err != nil {
		e.log.Error("failed to get channel members", zap.Error(err))
		return
	}

	e.log.Info("resolved event recipients",
		zap.String("channel_id", payload.ChannelID),
		zap.Int("recipient_count", len(recipients)),
	)

	rawPayload, err := json.Marshal(payload)
	if err != nil {
		e.log.Error("failed to marshal event payload", zap.Error(err))
		return
	}

	event := &Event{
		Type:       EventMessageCreated,
		Payload:    rawPayload,
		Recipients: recipients,
	}

	select {
	case e.hub.broadcast <- event:
	default:
		e.log.Error("hub broadcast channel full, dropping event",
			zap.String("message_id", payload.MessageID),
		)
	}
}