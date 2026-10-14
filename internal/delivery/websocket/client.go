package websocket

import (
	"net"
	"sync"
	"time"

	"github.com/gofiber/contrib/websocket"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

// Client represents a WebSocket client connection.
type Client struct {
	hub *Hub

	// The websocket connection
	conn *websocket.Conn
	// Keep the underlying connection to interrupt blocked pumps without touching Fiber's pooled wrapper.
	rawConn net.Conn

	// Buffered channel of outbound messages
	send chan []byte

	// Client identifier (optional, for tracking)
	ID string

	done       chan struct{}
	writeDone  chan struct{}
	stopOnce   sync.Once
	deadlineMu sync.Mutex
}

// NewClient creates a new Client instance.
func NewClient(hub *Hub, conn *websocket.Conn, id string) *Client {
	var rawConn net.Conn
	if conn != nil {
		rawConn = conn.NetConn()
	}
	return &Client{
		hub:       hub,
		conn:      conn,
		rawConn:   rawConn,
		send:      make(chan []byte, 256),
		ID:        id,
		done:      make(chan struct{}),
		writeDone: make(chan struct{}),
	}
}

func (c *Client) stop() {
	c.stopOnce.Do(func() {
		close(c.done)
		if c.rawConn == nil {
			return
		}
		c.deadlineMu.Lock()
		deadline := time.Now()
		_ = c.rawConn.SetReadDeadline(deadline)
		_ = c.rawConn.SetWriteDeadline(deadline)
		c.deadlineMu.Unlock()
	})
}

func (c *Client) setReadDeadline(deadline time.Time) error {
	c.deadlineMu.Lock()
	defer c.deadlineMu.Unlock()
	select {
	case <-c.done:
		return net.ErrClosed
	default:
		return c.conn.SetReadDeadline(deadline)
	}
}

func (c *Client) setWriteDeadline(deadline time.Time) error {
	c.deadlineMu.Lock()
	defer c.deadlineMu.Unlock()
	select {
	case <-c.done:
		return net.ErrClosed
	default:
		return c.conn.SetWriteDeadline(deadline)
	}
}

// ReadPump pumps messages from the websocket connection to the hub.
func (c *Client) ReadPump() {
	defer func() {
		c.hub.unregisterClient(c)
		<-c.writeDone
		_ = c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	if err := c.setReadDeadline(time.Now().Add(pongWait)); err != nil {
		return
	}
	c.conn.SetPongHandler(func(string) error {
		return c.setReadDeadline(time.Now().Add(pongWait))
	})

	for {
		select {
		case <-c.done:
			return
		default:
		}
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			break
		}
		if !c.hub.queueBroadcast(message) {
			return
		}
	}
}

// WritePump pumps messages from the hub to the websocket connection.
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case <-c.done:
			return

		case message, ok := <-c.send:
			if err := c.setWriteDeadline(time.Now().Add(writeWait)); err != nil {
				return
			}
			if !ok {
				// The hub closed the channel
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}

		case <-ticker.C:
			if err := c.setWriteDeadline(time.Now().Add(writeWait)); err != nil {
				return
			}
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
