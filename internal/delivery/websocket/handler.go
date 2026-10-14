package websocket

import (
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// Handler handles WebSocket HTTP requests.
type Handler struct {
	hub *Hub
}

// NewHandler creates a new WebSocket handler.
func NewHandler(hub *Hub) *Handler {
	return &Handler{hub: hub}
}

// Upgrade is a middleware that checks if the request is a WebSocket upgrade request.
func (h *Handler) Upgrade() fiber.Handler {
	return func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	}
}

// HandleWebSocket handles WebSocket connections.
func (h *Handler) HandleWebSocket() fiber.Handler {
	return websocket.New(func(c *websocket.Conn) {
		// Generate a unique client ID
		clientID := uuid.New().String()
		h.hub.serve(c, clientID)
	})
}

// RegisterRoutes registers WebSocket routes to the Fiber app.
func RegisterRoutes(app *fiber.App, hub *Hub) *Handler {
	handler := NewHandler(hub)

	// WebSocket endpoint
	ws := app.Group("/ws")
	ws.Use(handler.Upgrade())
	ws.Get("/", handler.HandleWebSocket())

	return handler
}
