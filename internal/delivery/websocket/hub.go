// Package websocket provides WebSocket functionality for real-time communication.
package websocket

import (
	"fmt"
	"sync"
	"time"

	"github.com/gofiber/contrib/websocket"
)

const shutdownTimeout = 2 * time.Second

type registerRequest struct {
	client   *Client
	accepted chan bool
}

// Hub maintains the set of active clients and broadcasts messages to clients.
type Hub struct {
	// Registered clients
	clients map[*Client]bool

	// Inbound messages from the clients
	broadcast chan []byte

	// Register requests from the clients
	register chan registerRequest

	// Unregister requests from clients
	unregister chan *Client

	// Mutex for thread-safe operations
	done     chan struct{}
	runDone  chan struct{}
	mu       sync.RWMutex
	started  bool
	stopping bool
	pumps    sync.WaitGroup
}

// NewHub creates a new Hub instance.
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte, 256),
		register:   make(chan registerRequest),
		unregister: make(chan *Client),
		done:       make(chan struct{}),
		runDone:    make(chan struct{}),
	}
}

// Start starts the hub's event loop once.
func (h *Hub) Start() {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.started || h.stopping {
		return
	}
	h.started = true
	go h.run()
}

func (h *Hub) run() {
	defer close(h.runDone)
	for {
		select {
		case request := <-h.register:
			h.addClient(request)

		case client := <-h.unregister:
			h.removeClient(client)

		case message := <-h.broadcast:
			h.broadcastToClients(message)

		case <-h.done:
			h.closeClients()
			return
		}
	}
}

func (h *Hub) addClient(request registerRequest) {
	accepted := false
	h.mu.Lock()
	if !h.stopping {
		h.clients[request.client] = true
		accepted = true
	}
	h.mu.Unlock()
	select {
	case request.accepted <- accepted:
	case <-h.done:
	}
}

func (h *Hub) removeClient(client *Client) {
	h.mu.Lock()
	removed := h.removeClientLocked(client)
	h.mu.Unlock()
	if removed {
		client.stop()
	}
}

func (h *Hub) removeClientLocked(client *Client) bool {
	if _, ok := h.clients[client]; !ok {
		return false
	}
	delete(h.clients, client)
	close(client.send)
	return true
}

func (h *Hub) broadcastToClients(message []byte) {
	var evicted []*Client
	h.mu.Lock()
	if !h.stopping {
		for client := range h.clients {
			select {
			case client.send <- message:
			default:
				if h.removeClientLocked(client) {
					evicted = append(evicted, client)
				}
			}
		}
	}
	h.mu.Unlock()
	for _, client := range evicted {
		client.stop()
	}
}

func (h *Hub) closeClients() {
	var clients []*Client
	h.mu.Lock()
	for client := range h.clients {
		if h.removeClientLocked(client) {
			clients = append(clients, client)
		}
	}
	h.mu.Unlock()
	for _, client := range clients {
		client.stop()
	}
}

func (h *Hub) serve(conn *websocket.Conn, id string) {
	client := NewClient(h, conn, id)
	h.mu.Lock()
	if !h.started || h.stopping {
		h.mu.Unlock()
		_ = conn.Close()
		return
	}
	h.pumps.Add(2)
	h.mu.Unlock()
	defer h.pumps.Done()
	go func() {
		defer h.pumps.Done()
		defer close(client.writeDone)
		client.WritePump()
	}()

	if !h.registerClient(client) {
		client.stop()
		<-client.writeDone
		_ = conn.Close()
		return
	}
	client.ReadPump()
}

func (h *Hub) registerClient(client *Client) bool {
	request := registerRequest{client: client, accepted: make(chan bool)}
	select {
	case h.register <- request:
	case <-h.done:
		return false
	}
	select {
	case accepted := <-request.accepted:
		return accepted
	case <-h.done:
		return false
	}
}

func (h *Hub) unregisterClient(client *Client) {
	select {
	case h.unregister <- client:
	case <-h.done:
	}
}

func (h *Hub) queueBroadcast(message []byte) bool {
	h.mu.RLock()
	running := h.started && !h.stopping
	h.mu.RUnlock()
	if !running {
		return false
	}
	select {
	case h.broadcast <- message:
		return true
	case <-h.done:
		return false
	}
}

// Broadcast sends a message to all connected clients.
func (h *Hub) Broadcast(message []byte) {
	h.queueBroadcast(message)
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

// Shutdown stops the hub and its client pumps within a bounded time.
func (h *Hub) Shutdown() error {
	h.mu.Lock()
	if !h.stopping {
		h.stopping = true
		close(h.done)
		if !h.started {
			close(h.runDone)
		}
	}
	h.mu.Unlock()

	timer := time.NewTimer(shutdownTimeout)
	defer timer.Stop()
	select {
	case <-h.runDone:
	case <-timer.C:
		return fmt.Errorf("websocket hub did not stop within %s", shutdownTimeout)
	}
	pumpsDone := make(chan struct{})
	go func() {
		h.pumps.Wait()
		close(pumpsDone)
	}()
	select {
	case <-pumpsDone:
		return nil
	case <-timer.C:
		return fmt.Errorf("websocket client pumps did not stop within %s", shutdownTimeout)
	}
}
