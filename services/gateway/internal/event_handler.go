package internal

import (
	"encoding/json"
	"strings"

	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

type EventHandler struct {
	nc *nats.Conn
	hub *Hub
	log *zap.Logger
	subs []*nats.Subscription
}

func NewEventHandler(nc *nats.Conn, hub *Hub, log *zap.Logger) *EventHandler {
	return &EventHandler{
		nc: nc,
		hub: hub,
		log: log,
	}
}

func (e *EventHandler) Subscribe() error {
	subjects := []struct {
		pattern string
		handler nats.MsgHandler
	}{
		{"messages.guild.>", e.handleChannelMessage},
	}

	for _, subj := range subjects {
		sub, err := e.nc.Subscribe(subj.pattern, subj.handler)
		if err != nil {
			return err
		}
		e.subs = append(e.subs, sub)
		e.log.Info("subscribed to NATS subject", 
			zap.String("pattern", subj.pattern),
		)
	}

	return nil
}

func (e *EventHandler) handleChannelMessage(msg *nats.Msg) {
    parts := strings.Split(msg.Subject, ".")

    if len(parts) < 5 {
        e.log.Error("unexpected subject format", zap.String("subject", msg.Subject))
        return
    }
    action := parts[4] // "created", "updated", "deleted"

    var event Event
    if err := json.Unmarshal(msg.Data, &event); err != nil { 
		e.log.Error("failed to unmarshal message", zap.Error(err))
		return
	}
	
    switch action {
    case "created":
        event.Type = EventMessageCreated
    default:
        e.log.Warn("unknown action", zap.String("action", action))
        return
    }

    select {
    case e.hub.broadcast <- &event:
    default:
        e.log.Error("hub broadcast channel full, dropping event")
    }
}