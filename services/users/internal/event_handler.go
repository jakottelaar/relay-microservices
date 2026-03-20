package internal

import (
	"context"
	"encoding/json"
	"time"

	"github.com/jakottelaar/relay-microservices/shared/events"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

type EventHandler struct {
	service UserService
	nc      *nats.Conn
	log    	*zap.Logger
}

func NewEventHandler(service UserService, nc *nats.Conn, log *zap.Logger) *EventHandler {
	return &EventHandler{
		service: service,
		nc:      nc,
		log: log,
	}
}

func (h *EventHandler) SubscribeToEvents(ctx context.Context) error {
    _, err := h.nc.Subscribe(events.SubjectUserAccountCreated, func(msg *nats.Msg) {
		var eventData CreateUserRequest
		if err := json.Unmarshal(msg.Data, &eventData); err != nil {
			h.log.Error("Failed to unmarshal UserAccountCreatedEvent",
				zap.Error(err),
			)
			return
		}

        h.service.CreateUser(ctx, &eventData)
    })
    
    if err != nil {
        return err
    }

    h.log.Info("Subscribed to NATS events",
        zap.String("subject", events.SubjectUserAccountCreated),
    )

    return nil
}

func (h *EventHandler) CreateUser(ctx context.Context, msg *nats.Msg) {
	var eventData CreateUserRequest
	if err := json.Unmarshal(msg.Data, &eventData); err != nil {
		h.log.Error("Failed to unmarshal UserAccountCreatedEvent",
			zap.Error(err),
		)
		return
	}

	h.log.Info("Received user account created event",
        zap.Int64("user_id", eventData.UserID),
        zap.String("username", eventData.Username),
    )

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := h.service.CreateUser(ctx, &eventData); err != nil {
		h.log.Error("Failed to create user from event",
			zap.Error(err),
			zap.Int64("user_id", eventData.UserID),
			zap.String("username", eventData.Username),
		)
		return
	}

	h.log.Info("User created successfully from event",
		zap.Int64("user_id", eventData.UserID),
		zap.String("username", eventData.Username),
	)
}