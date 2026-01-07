// Package websocket provides WebSocket functionality for real-time communication.
package websocket

import (
	"sync"
	"time"
)

// Hub maintains the set of active clients and broadcasts messages to clients.
type Hub struct {
	// Registered clients
	clients map[*Client]bool

	// Inbound messages from the clients
	broadcast chan []byte

	// Register requests from the clients
	register chan *Client

	// Unregister requests from clients
	unregister chan *Client

	// Mutex for thread-safe operations
	mu sync.RWMutex
}

// NewHub creates a new Hub instance.
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

// Run starts the hub's main loop.
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			h.mu.Unlock()

		case message := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					h.mu.RUnlock()
					h.mu.Lock()
					close(client.send)
					delete(h.clients, client)
					h.mu.Unlock()
					h.mu.RLock()
				}
			}
			h.mu.RUnlock()
		}
	}
}

// Broadcast sends a message to all connected clients.
func (h *Hub) Broadcast(message []byte) {
	h.broadcast <- message
}

// ClientCount returns the number of connected clients.
func (h *Hub) ClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

// BroadcastMessage broadcasts a Message struct to all clients.
func (h *Hub) BroadcastMessage(msgType string, payload interface{}) {
	msg := NewMessage(msgType, payload)
	data, err := msg.ToJSON()
	if err != nil {
		return
	}
	h.Broadcast(data)
}

// Shutdown gracefully closes all client connections.
func (h *Hub) Shutdown() {
	h.mu.Lock()
	defer h.mu.Unlock()

	for client := range h.clients {
		// Send close message
		closeMsg := NewMessage(MessageTypeClose, "server shutting down")
		if data, err := closeMsg.ToJSON(); err == nil {
			select {
			case client.send <- data:
			default:
			}
		}
		// Give client time to receive close message
		time.Sleep(100 * time.Millisecond)
		close(client.send)
		delete(h.clients, client)
	}
}
