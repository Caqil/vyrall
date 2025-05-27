package websocket

import (
	"encoding/json"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// TypingHandler manages typing indicator websocket events
type TypingHandler struct {
	hub *Hub
}

// NewTypingHandler creates a new typing handler
func NewTypingHandler(hub *Hub) *TypingHandler {
	return &TypingHandler{
		hub: hub,
	}
}

// TypingIndicator represents a typing indicator payload
type TypingIndicator struct {
	Type           string `json:"type"`
	ConversationID string `json:"conversation_id"`
	IsTyping       bool   `json:"is_typing"`
}

// ProcessTypingIndicator handles typing indicator events
func (h *TypingHandler) ProcessTypingIndicator(client *Client, data []byte) {
	var indicator TypingIndicator
	if err := json.Unmarshal(data, &indicator); err != nil {
		log.Printf("Error unmarshaling typing indicator: %v", err)
		client.SendErrorMessage("Invalid typing indicator format")
		return
	}

	// Validate conversation ID
	if indicator.ConversationID == "" {
		client.SendErrorMessage("Conversation ID is required")
		return
	}

	conversationID, err := primitive.ObjectIDFromHex(indicator.ConversationID)
	if err != nil {
		client.SendErrorMessage("Invalid conversation ID")
		return
	}

	// Broadcast typing indicator to conversation participants
	h.broadcastTypingIndicator(client.UserID, conversationID, indicator.IsTyping)
}

// broadcastTypingIndicator broadcasts a typing indicator to all participants in a conversation
func (h *TypingHandler) broadcastTypingIndicator(userID primitive.ObjectID, conversationID primitive.ObjectID, isTyping bool) {
	// Create typing indicator payload
	payload, err := json.Marshal(map[string]interface{}{
		"type":            "typing_indicator",
		"user_id":         userID.Hex(),
		"conversation_id": conversationID.Hex(),
		"is_typing":       isTyping,
		"timestamp":       time.Now(),
	})
	if err != nil {
		log.Printf("Error marshaling typing indicator payload: %v", err)
		return
	}

	// Get conversation participants from the database
	// In a real implementation, this would be done through a service
	// For simplicity, we'll broadcast to a room
	roomID := "conversation:" + conversationID.Hex()
	h.hub.roomsHandler.BroadcastToRoom(roomID, payload)
}

// UpdateGroupTypingIndicator sends typing indicators for group chats
func (h *TypingHandler) UpdateGroupTypingIndicator(userID primitive.ObjectID, groupID primitive.ObjectID, isTyping bool) {
	// Create typing indicator payload
	payload, err := json.Marshal(map[string]interface{}{
		"type":      "group_typing_indicator",
		"user_id":   userID.Hex(),
		"group_id":  groupID.Hex(),
		"is_typing": isTyping,
		"timestamp": time.Now(),
	})
	if err != nil {
		log.Printf("Error marshaling group typing indicator payload: %v", err)
		return
	}

	// Broadcast to group room
	roomID := "group:" + groupID.Hex()
	h.hub.roomsHandler.BroadcastToRoom(roomID, payload)
}

// GetActiveTypers gets a list of users currently typing in a conversation
func (h *TypingHandler) GetActiveTypers(conversationID primitive.ObjectID) []primitive.ObjectID {
	// In a real implementation, this would track active typers in a map or database
	// For simplicity, we'll return an empty slice
	return []primitive.ObjectID{}
}
