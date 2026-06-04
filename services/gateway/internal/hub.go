package internal

import "go.uber.org/zap"

type Hub struct {
	clients map[int64]map[*Client]struct{}

	register   chan *Client
	unregister chan *Client
	broadcast  chan *Event

	log *zap.Logger
}

func NewHub(log *zap.Logger) *Hub {
	return &Hub{
		clients: make(map[int64]map[*Client]struct{}),
		register: make(chan *Client),
		unregister: make(chan *Client),
		broadcast: make(chan *Event),
		log: log,
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			if _, ok := h.clients[client.userID]; !ok {
				h.clients[client.userID] = make(map[*Client]struct{})
			}
			h.clients[client.userID][client] = struct{}{}
			h.log.Info("Client registered", 
				zap.Int64("userID", client.userID),
			)

		case client := <-h.unregister:
			if clients, ok := h.clients[client.userID]; ok {
				if _, ok := clients[client]; ok {
					delete(clients, client)
					close(client.send)
					if len(clients) == 0 {
						delete(h.clients, client.userID)
					}
					h.log.Info("client disconnected",
						zap.Int64("userID", client.userID),
					)
				}
			}
		case event := <-h.broadcast:
			h.dispatch(event)
		}
	}
}

func (h *Hub) dispatch(event *Event) {
	data, err := marshalEvent(event)
	if err != nil {
		h.log.Error("marshal failed", 
			zap.String("event", event.Type),
			zap.Error(err),
		)
		return
	}

	h.log.Info("dispatching websocket event",
		zap.String("event", event.Type),
		zap.Int("recipient_count", len(event.Recipients)),
	)

	for _, userID := range event.Recipients {
		clients, ok := h.clients[userID]
		if !ok {
			h.log.Debug("no connected client for recipient",
				zap.Int64("userID", userID),
			)
			continue
		}
		for client := range clients {
			select {
			case client.send <- data:
				h.log.Debug("websocket event queued",
					zap.Int64("userID", userID),
				)
			default:
				h.log.Warn("client buffer full, disconnecting",
					zap.Int64("userID", userID),
				)

				delete(clients, client)
				close(client.send)
				if len(clients) == 0 {
					delete(h.clients, userID)
				}
			}
		}
	}
}