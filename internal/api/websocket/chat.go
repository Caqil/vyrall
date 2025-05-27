package websocket

import (
	"encoding/json"
	"log"
	"time"

	"github.com/Caqil/vyrall/internal/models"
	"github.com/Caqil/vyrall/internal/services/message"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ChatHandler manages chat message websocket events
type ChatHandler struct {
	hub            *Hub
	messageService *message.Service
}

// NewChatHandler creates a new chat handler
func NewChatHandler(hub *Hub, messageService *message.Service) *ChatHandler {
	return &ChatHandler{
		hub:            hub,
		messageService: messageService,
	}
}

// ChatMessage represents a chat message payload
type ChatMessage struct {
	Type             string   `json:"type"`
	ConversationID   string   `json:"conversation_id"`
	Content          string   `json:"content"`
	MediaIDs         []string `json:"media_ids,omitempty"`
	ReplyToID        string   `json:"reply_to_id,omitempty"`
	MentionedUserIDs []string `json:"mentioned_user_ids,omitempty"`
}

// ProcessMessage handles incoming chat messages
func (h *ChatHandler) ProcessMessage(client *Client, data []byte) {
	var message ChatMessage
	if err := json.Unmarshal(data, &message); err != nil {
		log.Printf("Error unmarshaling chat message: %v", err)
		client.SendErrorMessage("Invalid message format")
		return
	}

	// Validate message
	if message.ConversationID == "" {
		client.SendErrorMessage("Conversation ID is required")
		return
	}

	if message.Content == "" && len(message.MediaIDs) == 0 {
		client.SendErrorMessage("Message must have content or media")
		return
	}

	// Convert IDs to ObjectIDs
	conversationID, err := primitive.ObjectIDFromHex(message.ConversationID)
	if err != nil {
		client.SendErrorMessage("Invalid conversation ID")
		return
	}

	// Process mentioned users
	var mentionedUserIDs []primitive.ObjectID
	for _, userIDStr := range message.MentionedUserIDs {
		userID, err := primitive.ObjectIDFromHex(userIDStr)
		if err != nil {
			continue
		}
		mentionedUserIDs = append(mentionedUserIDs, userID)
	}

	// Process media IDs
	var mediaIDs []primitive.ObjectID
	for _, mediaIDStr := range message.MediaIDs {
		mediaID, err := primitive.ObjectIDFromHex(mediaIDStr)
		if err != nil {
			continue
		}
		mediaIDs = append(mediaIDs, mediaID)
	}

	// Create message object
	msgObj := &models.Message{
		ConversationID: conversationID,
		SenderID:       client.UserID,
		Content:        message.Content,
		MentionedUsers: mentionedUserIDs,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		MessageType:    "text",
	}

	// Add reply if specified
	if message.ReplyToID != "" {
		replyID, err := primitive.ObjectIDFromHex(message.ReplyToID)
		if err == nil {
			msgObj.ReplyToID = &replyID
		}
	}

	// Save message to database
	savedMsg, err := h.messageService.CreateMessage(client.ctx, msgObj, mediaIDs)
	if err != nil {
		log.Printf("Error saving message: %v", err)
		client.SendErrorMessage("Failed to save message")
		return
	}

	// Broadcast message to conversation participants
	h.broadcastMessageToConversation(savedMsg)
}

// broadcastMessageToConversation sends a message to all participants in a conversation
func (h *ChatHandler) broadcastMessageToConversation(message *models.Message) {
	// Get conversation participants
	conversation, err := h.messageService.GetConversationByID(nil, message.ConversationID)
	if err != nil {
		log.Printf("Error getting conversation: %v", err)
		return
	}

	// Create message payload
	payload, err := json.Marshal(map[string]interface{}{
		"type":      "chat_message",
		"message":   message,
		"sender_id": message.SenderID.Hex(),
		"timestamp": message.CreatedAt,
	})
	if err != nil {
		log.Printf("Error marshaling message payload: %v", err)
		return
	}

	// Get participant user IDs
	var participantIDs []primitive.ObjectID
	for _, participant := range conversation.Participants {
		participantIDs = append(participantIDs, participant.UserID)
	}

	// Broadcast to all participants
	h.hub.BroadcastToUsers(participantIDs, payload)
}

// SendMessageDeliveryStatus sends a delivery status update for a message
func (h *ChatHandler) SendMessageDeliveryStatus(messageID, userID primitive.ObjectID, status string) {
	payload, err := json.Marshal(map[string]interface{}{
		"type":       "message_status",
		"message_id": messageID.Hex(),
		"user_id":    userID.Hex(),
		"status":     status,
		"timestamp":  time.Now(),
	})
	if err != nil {
		log.Printf("Error marshaling message status payload: %v", err)
		return
	}

	// Send to the message sender
	message, err := h.messageService.GetMessageByID(nil, messageID)
	if err != nil {
		log.Printf("Error getting message: %v", err)
		return
	}

	h.hub.SendToUser(message.SenderID, payload)
}

// SendMessageReadReceipt sends a read receipt for a message
func (h *ChatHandler) SendMessageReadReceipt(client *Client, data []byte) {
	var payload struct {
		MessageID      string `json:"message_id"`
		ConversationID string `json:"conversation_id"`
	}

	if err := json.Unmarshal(data, &payload); err != nil {
		log.Printf("Error unmarshaling read receipt: %v", err)
		client.SendErrorMessage("Invalid read receipt format")
		return
	}

	messageID, err := primitive.ObjectIDFromHex(payload.MessageID)
	if err != nil {
		client.SendErrorMessage("Invalid message ID")
		return
	}

	conversationID, err := primitive.ObjectIDFromHex(payload.ConversationID)
	if err != nil {
		client.SendErrorMessage("Invalid conversation ID")
		return
	}

	// Mark message as read in database
	err = h.messageService.MarkMessageAsRead(client.ctx, messageID, conversationID, client.UserID)
	if err != nil {
		log.Printf("Error marking message as read: %v", err)
		client.SendErrorMessage("Failed to mark message as read")
		return
	}

	// Send read receipt to other participants
	readReceiptPayload, err := json.Marshal(map[string]interface{}{
		"type":            "read_receipt",
		"message_id":      payload.MessageID,
		"conversation_id": payload.ConversationID,
		"user_id":         client.UserID.Hex(),
		"timestamp":       time.Now(),
	})
	if err != nil {
		log.Printf("Error marshaling read receipt payload: %v", err)
		return
	}

	// Get conversation participants
	conversation, err := h.messageService.GetConversationByID(nil, conversationID)
	if err != nil {
		log.Printf("Error getting conversation: %v", err)
		return
	}

	// Get participant user IDs
	var participantIDs []primitive.ObjectID
	for _, participant := range conversation.Participants {
		participantIDs = append(participantIDs, participant.UserID)
	}

	// Broadcast to all participants
	h.hub.BroadcastToUsers(participantIDs, readReceiptPayload)
}
