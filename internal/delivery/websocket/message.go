package websocket

import (
	"time"

	"github.com/goccy/go-json"
)

// Message types
const (
	MessageTypeText      = "text"
	MessageTypeBroadcast = "broadcast"
	MessageTypePing      = "ping"
	MessageTypePong      = "pong"
	MessageTypeClose     = "close"
	MessageTypeError     = "error"
)

// Message represents a WebSocket message structure.
type Message struct {
	Type      string      `json:"type"`
	Payload   interface{} `json:"payload,omitempty"`
	Timestamp int64       `json:"timestamp"`
}

// NewMessage creates a new Message with the current timestamp.
func NewMessage(msgType string, payload interface{}) *Message {
	return &Message{
		Type:      msgType,
		Payload:   payload,
		Timestamp: time.Now().Unix(),
	}
}

// ToJSON converts the message to JSON bytes.
func (m *Message) ToJSON() ([]byte, error) {
	return json.Marshal(m)
}

// ParseMessage parses JSON bytes into a Message.
func ParseMessage(data []byte) (*Message, error) {
	var msg Message
	if err := json.Unmarshal(data, &msg); err != nil {
		return nil, err
	}
	return &msg, nil
}

// TextPayload represents a simple text message payload.
type TextPayload struct {
	Content string `json:"content"`
	From    string `json:"from,omitempty"`
}

// ErrorPayload represents an error message payload.
type ErrorPayload struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// NewTextMessage creates a new text message.
func NewTextMessage(content, from string) *Message {
	return NewMessage(MessageTypeText, TextPayload{
		Content: content,
		From:    from,
	})
}

// NewErrorMessage creates a new error message.
func NewErrorMessage(code, message string) *Message {
	return NewMessage(MessageTypeError, ErrorPayload{
		Code:    code,
		Message: message,
	})
}
