package internal

import "go.uber.org/zap"

type Hub struct {
	clients map[int64]*Client

	register   chan *Client
	unregister chan *Client
	broadcast  chan *Event

	log *zap.Logger
}

func NewHub(log *zap.Logger) *Hub {
	return &Hub{
		clients: make(map[int64]*Client),
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
			h.clients[client.userID] = client
			h.log.Info("Client registered", 
				zap.Int64("userID", client.userID),
			)

		case client := <-h.unregister:
			if _, ok := h.clients[client.userID]; ok {
				delete(h.clients, client.userID)
				close(client.send)
				h.log.Info("client disconnected",
					zap.Int64("userID", client.userID),
				)
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

	for _, userID := range event.Recipients {
		client, ok := h.clients[userID]
		if !ok {
			continue
		}
		select {
		case client.send <- data:
		default:
			h.log.Warn("client buffer full, disconnecting",
			zap.Int64("userID", userID))
			
			delete(h.clients, userID)
			close(client.send)
		}
	}
}