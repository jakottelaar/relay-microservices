package internal

import "encoding/json"

type Event struct {
	Type    string `json:"type"`
	Payload json.RawMessage `json:"payload"`
	Recipients []int64 `json:"recipients"`
}

func marshalEvent(e *Event) ([]byte, error) {
	return json.Marshal(e)
}

const EventMessageCreated = "MESSAGE_CREATE"