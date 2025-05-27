package websocket

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	// Time allowed to write a message to the peer
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer
	pongWait = 60 * time.Second

	// Send pings to peer with this period (must be less than pongWait)
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer
	maxMessageSize = 10000
)

// Client represents a websocket client connection
type Client struct {
	Hub    *Hub
	Conn   *websocket.Conn
	Send   chan []byte
	UserID primitive.ObjectID
	ctx    context.Context
	active bool
}

// NewClient creates a new websocket client
func NewClient(hub *Hub, conn *websocket.Conn, userID primitive.ObjectID) *Client {
	return &Client{
		Hub:    hub,
		Conn:   conn,
		Send:   make(chan []byte, 256),
		UserID: userID,
		ctx:    context.Background(),
		active: true,
	}
}

// ReadPump pumps messages from the websocket connection to the hub
func (c *Client) ReadPump() {
	defer func() {
		c.Hub.Unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		// Process the message
		c.processMessage(message)
	}
}

// WritePump pumps messages from the hub to the websocket connection
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued messages to the current websocket message
			n := len(c.Send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.Send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// processMessage handles incoming messages based on their type
func (c *Client) processMessage(message []byte) {
	// Try to parse the message to determine its type
	var msg struct {
		Type string `json:"type"`
	}

	if err := json.Unmarshal(message, &msg); err != nil {
		log.Printf("Error parsing message: %v", err)
		c.SendErrorMessage("Invalid message format")
		return
	}

	// Route the message to the appropriate handler
	switch msg.Type {
	case "chat_message":
		c.Hub.chatHandler.ProcessMessage(c, message)
	case "typing_indicator":
		c.Hub.typingHandler.ProcessTypingIndicator(c, message)
	case "read_receipt":
		c.Hub.chatHandler.SendMessageReadReceipt(c, message)
	case "presence_update":
		c.Hub.presenceHandler.ProcessPresenceUpdate(c, message)
	case "join_room":
		c.Hub.roomsHandler.JoinRoom(c, message)
	case "leave_room":
		c.Hub.roomsHandler.LeaveRoom(c, message)
	default:
		log.Printf("Unknown message type: %s", msg.Type)
		c.SendErrorMessage("Unknown message type")
	}
}

// SendMessage sends a message to the client
func (c *Client) SendMessage(message []byte) {
	// Don't block on send, if the channel is full, the message is dropped
	select {
	case c.Send <- message:
	default:
		log.Printf("Client send buffer full, dropping message")
	}
}

// SendErrorMessage sends an error message to the client
func (c *Client) SendErrorMessage(errorMsg string) {
	payload, err := json.Marshal(map[string]interface{}{
		"type":  "error",
		"error": errorMsg,
	})

	if err != nil {
		log.Printf("Error marshaling error message: %v", err)
		return
	}

	c.SendMessage(payload)
}

// SetContext sets the context for the client
func (c *Client) SetContext(ctx context.Context) {
	c.ctx = ctx
}

// IsActive returns whether the client is active
func (c *Client) IsActive() bool {
	return c.active
}

// SetActive sets the client's active status
func (c *Client) SetActive(active bool) {
	c.active = active
}
