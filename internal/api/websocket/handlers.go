package websocket

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/Caqil/vyrall/internal/services"
	"github.com/Caqil/vyrall/internal/services/auth"
	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// WebSocketHandler manages the websocket connections and handlers
type WebSocketHandler struct {
	hub         *Hub
	authService *auth.Service
	services    *services.Services
	upgrader    websocket.Upgrader
}

// NewWebSocketHandler creates a new websocket handler
func NewWebSocketHandler(services *services.Services) *WebSocketHandler {
	// Create the hub
	hub := NewHub()

	// Initialize the handler
	handler := &WebSocketHandler{
		hub:         hub,
		authService: services.AuthService,
		services:    services,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				// In production, this should check if the origin is allowed
				return true
			},
		},
	}

	// Set up the handlers for the hub
	handler.initializeHandlers()

	// Start the hub
	go hub.Run()

	return handler
}

// initializeHandlers sets up the different event handlers
func (h *WebSocketHandler) initializeHandlers() {
	// Initialize the handlers
	h.hub.chatHandler = NewChatHandler(h.hub, h.services.MessageService)
	h.hub.typingHandler = NewTypingHandler(h.hub)
	h.hub.presenceHandler = NewPresenceHandler(h.hub, h.services.UserService)
	h.hub.roomsHandler = NewRoomsHandler(h.hub)
	h.hub.notificationsHandler = NewNotificationsHandler(h.hub, h.services.NotificationService)
	h.hub.liveStreamHandler = NewLiveStreamHandler(h.hub, h.services.LiveStreamService)
}

// HandleWebSocket handles the websocket connections
func (h *WebSocketHandler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Extract the user ID from the request
	userID, err := h.extractUserID(r)
	if err != nil {
		log.Printf("Error extracting user ID: %v", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Upgrade the HTTP connection to a WebSocket connection
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Error upgrading connection: %v", err)
		return
	}

	// Create a new client
	client := NewClient(h.hub, conn, userID)

	// Register the client
	h.hub.Register <- client

	// Start the client's read and write pumps
	go client.ReadPump()
	go client.WritePump()

	// Notify presence
	h.hub.presenceHandler.NotifyUserOnline(userID)

	// Send initial data to the client
	h.sendInitialData(client)
}

// extractUserID extracts the user ID from the request
func (h *WebSocketHandler) extractUserID(r *http.Request) (primitive.ObjectID, error) {
	// Get the token from the Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		// Try to get the token from the URL query parameters
		token := r.URL.Query().Get("token")
		if token == "" {
			return primitive.NilObjectID, ErrUnauthorized
		}
		authHeader = "Bearer " + token
	}

	// Check if the header has the Bearer prefix
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return primitive.NilObjectID, ErrUnauthorized
	}

	token := parts[1]
	if token == "" {
		return primitive.NilObjectID, ErrUnauthorized
	}

	// Validate the token and get the user ID
	userID, err := h.authService.ValidateToken(r.Context(), token)
	if err != nil {
		return primitive.NilObjectID, ErrUnauthorized
	}

	return userID, nil
}

// sendInitialData sends initial data to the client
func (h *WebSocketHandler) sendInitialData(client *Client) {
	// Send the client's online friends
	h.hub.presenceHandler.SendOnlineFriends(client)

	// Send unread notifications count
	h.hub.notificationsHandler.SendUnreadNotificationsCount(client)

	// Send connection success message
	connectionInfo, err := json.Marshal(map[string]interface{}{
		"type":      "connection_success",
		"user_id":   client.UserID.Hex(),
		"timestamp": h.hub.GetServerTime(),
	})

	if err != nil {
		log.Printf("Error marshaling connection info: %v", err)
		return
	}

	client.SendMessage(connectionInfo)
}
