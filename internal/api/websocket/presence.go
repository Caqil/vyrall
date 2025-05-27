package websocket

import (
	"encoding/json"
	"log"
	"time"

	"github.com/Caqil/vyrall/internal/services/user"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// PresenceHandler manages user presence websocket events
type PresenceHandler struct {
	hub         *Hub
	userService *user.Service
}

// NewPresenceHandler creates a new presence handler
func NewPresenceHandler(hub *Hub, userService *user.Service) *PresenceHandler {
	return &PresenceHandler{
		hub:         hub,
		userService: userService,
	}
}

// ProcessPresenceUpdate handles presence updates from clients
func (h *PresenceHandler) ProcessPresenceUpdate(client *Client, data []byte) {
	var payload struct {
		Status string `json:"status"`
	}

	if err := json.Unmarshal(data, &payload); err != nil {
		log.Printf("Error unmarshaling presence update: %v", err)
		client.SendErrorMessage("Invalid presence update format")
		return
	}

	// Update user's presence status
	h.updateUserPresence(client.UserID, payload.Status)
}

// updateUserPresence updates a user's presence status
func (h *PresenceHandler) updateUserPresence(userID primitive.ObjectID, status string) {
	// Update last active time in the database
	err := h.userService.UpdateLastActive(nil, userID)
	if err != nil {
		log.Printf("Error updating last active time: %v", err)
	}

	// Broadcast presence update to friends
	h.broadcastPresenceToFriends(userID, status)
}

// NotifyUserOnline notifies friends that a user is online
func (h *PresenceHandler) NotifyUserOnline(userID primitive.ObjectID) {
	// Update last active time in the database
	err := h.userService.UpdateLastActive(nil, userID)
	if err != nil {
		log.Printf("Error updating last active time: %v", err)
	}

	// Broadcast online status to friends
	h.broadcastPresenceToFriends(userID, "online")
}

// NotifyUserOffline notifies friends that a user is offline
func (h *PresenceHandler) NotifyUserOffline(userID primitive.ObjectID) {
	// Broadcast offline status to friends
	h.broadcastPresenceToFriends(userID, "offline")
}

// broadcastPresenceToFriends broadcasts a user's presence status to their friends
func (h *PresenceHandler) broadcastPresenceToFriends(userID primitive.ObjectID, status string) {
	// Get user's friends
	friends, err := h.userService.GetUserFriends(nil, userID)
	if err != nil {
		log.Printf("Error getting user friends: %v", err)
		return
	}

	// Create presence payload
	payload, err := json.Marshal(map[string]interface{}{
		"type":      "presence_update",
		"user_id":   userID.Hex(),
		"status":    status,
		"timestamp": time.Now(),
	})
	if err != nil {
		log.Printf("Error marshaling presence update: %v", err)
		return
	}

	// Send to all online friends
	for _, friend := range friends {
		h.hub.SendToUser(friend.ID, payload)
	}
}

// SendOnlineFriends sends a list of online friends to a client
func (h *PresenceHandler) SendOnlineFriends(client *Client) {
	// Get user's friends
	friends, err := h.userService.GetUserFriends(client.ctx, client.UserID)
	if err != nil {
		log.Printf("Error getting user friends: %v", err)
		return
	}

	// Filter for online friends
	var onlineFriends []map[string]interface{}
	for _, friend := range friends {
		if h.hub.IsUserOnline(friend.ID) {
			onlineFriends = append(onlineFriends, map[string]interface{}{
				"user_id": friend.ID.Hex(),
				"status":  "online",
			})
		}
	}

	// Create online friends payload
	payload, err := json.Marshal(map[string]interface{}{
		"type":           "online_friends",
		"online_friends": onlineFriends,
		"timestamp":      time.Now(),
	})
	if err != nil {
		log.Printf("Error marshaling online friends: %v", err)
		return
	}

	client.SendMessage(payload)
}

// CheckUserOnline checks if a user is online
func (h *PresenceHandler) CheckUserOnline(userID primitive.ObjectID) bool {
	return h.hub.IsUserOnline(userID)
}

// GetUserStatusForClient sends a specific user's online status to a client
func (h *PresenceHandler) GetUserStatusForClient(client *Client, targetUserID primitive.ObjectID) {
	// Get user's online status
	isOnline := h.hub.IsUserOnline(targetUserID)
	status := "offline"
	if isOnline {
		status = "online"
	}

	// Create status payload
	payload, err := json.Marshal(map[string]interface{}{
		"type":      "user_status",
		"user_id":   targetUserID.Hex(),
		"status":    status,
		"timestamp": time.Now(),
	})
	if err != nil {
		log.Printf("Error marshaling user status: %v", err)
		return
	}

	client.SendMessage(payload)
}
