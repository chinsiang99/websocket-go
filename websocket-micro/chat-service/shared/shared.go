package shared

import "encoding/json"

type Message struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type ChatPayload struct {
	Room    string `json:"room"`
	Sender  string `json:"sender"`
	Message string `json:"message"`
}
