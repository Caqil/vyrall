package websocket

import (
	"encoding/json"
	"errors"
	"log"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Error definitions
var (
	ErrUnauthorized = errors.New("unauthorized")
	ErrNotFound     = errors.New("not found")
)

// Hub maintains the set of active clients and broadcasts messages
type Hub struct {
	// Registered clients
	clients map[*Client]bool

	// User ID to clients mapping
	userClients map[primitive.ObjectID][]*Client

	// Inbound messages from the clients
	Broadcast chan []byte

	// Register requests from the clients
	Register chan *Client

	// Unregister requests from clients
	Unregister chan *Client

	// Mutex to protect the maps
	mu sync.RWMutex

	// Handlers
	chatHandler          *ChatHandler
	typingHandler        *TypingHandler
	presenceHandler      *PresenceHandler
	roomsHandler         *RoomsHandler
	notificationsHandler *NotificationsHandler
	liveStreamHandler    *LiveStreamHandler
}

// NewHub creates a new hub
func NewHub() *Hub {
	return &Hub{
		Broadcast:   make(chan []byte),
		Register:    make(chan *Client),
		Unregister:  make(chan *Client),
		clients:     make(map[*Client]bool),
		userClients: make(map[primitive.ObjectID][]*Client),
	}
}

// Run starts the hub
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.registerClient(client)
		case client := <-h.Unregister:
			h.unregisterClient(client)
		case message := <-h.Broadcast:
			h.broadcastMessage(message)
		}
	}
}

// registerClient registers a client
func (h *Hub) registerClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Register the client
	h.clients[client] = true

	// Add to user clients mapping
	h.userClients[client.UserID] = append(h.userClients[client.UserID], client)

	log.Printf("Client registered: %s", client.UserID.Hex())
}

// unregisterClient unregisters a client
func (h *Hub) unregisterClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Check if the client is registered
	if _, ok := h.clients[client]; ok {
		// Remove from clients map
		delete(h.clients, client)

		// Remove from user clients mapping
		clients := h.userClients[client.UserID]
		for i, c := range clients {
			if c == client {
				// Remove the client from the slice
				h.userClients[client.UserID] = append(clients[:i], clients[i+1:]...)
				break
			}
		}

		// If no more clients for this user, remove the user entry
		if len(h.userClients[client.UserID]) == 0 {
			delete(h.userClients, client.UserID)
			// Notify that the user went offline
			h.presenceHandler.NotifyUserOffline(client.UserID)
		}

		// Close the send channel
		close(client.Send)

		log.Printf("Client unregistered: %s", client.UserID.Hex())
	}
}

// broadcastMessage broadcasts a message to all clients
func (h *Hub) broadcastMessage(message []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for client := range h.clients {
		select {
		case client.Send <- message:
		default:
			// If the client's send buffer is full, assume it's dead and unregister
			go func(c *Client) {
				h.Unregister <- c
			}(client)
		}
	}
}

// SendToUser sends a message to a specific user's clients
func (h *Hub) SendToUser(userID primitive.ObjectID, message []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	clients, exists := h.userClients[userID]
	if !exists {
		return
	}

	for _, client := range clients {
		select {
		case client.Send <- message:
		default:
			// If the client's send buffer is full, assume it's dead and unregister
			go func(c *Client) {
				h.Unregister <- c
			}(client)
		}
	}
}

// BroadcastToUsers sends a message to multiple users
func (h *Hub) BroadcastToUsers(userIDs []primitive.ObjectID, message []byte) {
	for _, userID := range userIDs {
		h.SendToUser(userID, message)
	}
}

// IsUserOnline checks if a user is online
func (h *Hub) IsUserOnline(userID primitive.ObjectID) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()

	clients, exists := h.userClients[userID]
	return exists && len(clients) > 0
}

// GetUserClients gets all clients for a user
func (h *Hub) GetUserClients(userID primitive.ObjectID) []*Client {
	h.mu.RLock()
	defer h.mu.RUnlock()

	return h.userClients[userID]
}

// GetOnlineUsers gets all online users
func (h *Hub) GetOnlineUsers() []primitive.ObjectID {
	h.mu.RLock()
	defer h.mu.RUnlock()

	userIDs := make([]primitive.ObjectID, 0, len(h.userClients))
	for userID := range h.userClients {
		userIDs = append(userIDs, userID)
	}

	return userIDs
}

// GetServerTime returns the current server time
func (h *Hub) GetServerTime() time.Time {
	return time.Now()
}

// SendSystemMessage sends a system message to a client
func (h *Hub) SendSystemMessage(client *Client, messageType string, data interface{}) {
	payload := map[string]interface{}{
		"type":      messageType,
		"data":      data,
		"timestamp": h.GetServerTime(),
	}

	message, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshaling system message: %v", err)
		return
	}

	client.SendMessage(message)
}
