package ws

import "encoding/json"

// WSMessage represents a request coming from the websocket client.
type WSMessage struct {
	RequestID string          `json:"requestId"`
	Action    string          `json:"action"`
	Data      json.RawMessage `json:"data,omitempty"`
}

// WSResponse represents a message sent back to the client.
type WSResponse struct {
	RequestID string      `json:"requestId,omitempty"`
	Action    string      `json:"action"`
	Status    int         `json:"status"`
	Data      interface{} `json:"data,omitempty"`
}
