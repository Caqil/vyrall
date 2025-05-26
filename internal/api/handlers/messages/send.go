package messages

import (
	"net/http"

	"github.com/Caqil/vyrall/internal/models"
	"github.com/Caqil/vyrall/internal/services/message"
	"github.com/Caqil/vyrall/internal/utils/response"
	"github.com/Caqil/vyrall/internal/utils/validation"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// SendHandler handles message sending operations
type SendHandler struct {
	messageService *message.Service
}

// NewSendHandler creates a new send handler
func NewSendHandler(messageService *message.Service) *SendHandler {
	return &SendHandler{
		messageService: messageService,
	}
}

// SendMessageRequest represents the request body for sending a message
type SendMessageRequest struct {
	ConversationID    string                    `json:"conversation_id"`
	Content           string                    `json:"content"`
	MediaIDs          []string                  `json:"media_ids,omitempty"`
	ReplyToID         string                    `json:"reply_to_id,omitempty"`
	MentionedUserIDs  []string                  `json:"mentioned_user_ids,omitempty"`
	IsEncrypted       bool                      `json:"is_encrypted,omitempty"`
	EncryptionDetails *models.EncryptionDetails `json:"encryption_details,omitempty"`
	MessageType       string                    `json:"message_type,omitempty"` // Default: "text"
	IsImportant       bool                      `json:"is_important,omitempty"`
}

// SendMessage handles the request to send a new message
func (h *SendHandler) SendMessage(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Parse request body
	var req SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "Invalid request body", err.Error())
		return
	}

	// Validate conversation ID
	if !validation.IsValidObjectID(req.ConversationID) {
		response.ValidationError(c, "Invalid conversation ID", nil)
		return
	}
	conversationID, _ := primitive.ObjectIDFromHex(req.ConversationID)

	// Validate content or media (at least one must be provided)
	if req.Content == "" && len(req.MediaIDs) == 0 {
		response.ValidationError(c, "Message must contain either text content or media", nil)
		return
	}

	// Convert media IDs to ObjectIDs
	mediaIDs := make([]primitive.ObjectID, 0, len(req.MediaIDs))
	for _, idStr := range req.MediaIDs {
		if !validation.IsValidObjectID(idStr) {
			continue // Skip invalid IDs
		}
		mediaID, _ := primitive.ObjectIDFromHex(idStr)
		mediaIDs = append(mediaIDs, mediaID)
	}

	// Convert mentioned user IDs to ObjectIDs
	mentionedUserIDs := make([]primitive.ObjectID, 0, len(req.MentionedUserIDs))
	for _, idStr := range req.MentionedUserIDs {
		if !validation.IsValidObjectID(idStr) {
			continue // Skip invalid IDs
		}
		mentionID, _ := primitive.ObjectIDFromHex(idStr)
		mentionedUserIDs = append(mentionedUserIDs, mentionID)
	}

	// Convert reply-to ID if provided
	var replyToID *primitive.ObjectID
	if req.ReplyToID != "" {
		if validation.IsValidObjectID(req.ReplyToID) {
			id, _ := primitive.ObjectIDFromHex(req.ReplyToID)
			replyToID = &id
		}
	}

	// Set default message type if not provided
	if req.MessageType == "" {
		req.MessageType = "text"
	}

	// Create message object
	message := &models.Message{
		ConversationID: conversationID,
		SenderID:       userID.(primitive.ObjectID),
		Content:        req.Content,
		ReplyToID:      replyToID,
		MentionedUsers: mentionedUserIDs,
		IsEncrypted:    req.IsEncrypted,
		MessageType:    req.MessageType,
		IsImportant:    req.IsImportant,
	}

	// Send the message
	sentMessage, err := h.messageService.SendMessage(c.Request.Context(), message, mediaIDs, req.EncryptionDetails)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to send message", err)
		return
	}

	// Return success response
	response.Created(c, "Message sent successfully", sentMessage)
}

// SendSystemMessage handles sending system-generated messages
func (h *SendHandler) SendSystemMessage(c *gin.Context) {
	// This endpoint might require admin permissions
	isAdmin, exists := c.Get("isAdmin")
	if !exists || !isAdmin.(bool) {
		response.ForbiddenError(c, "Admin access required")
		return
	}

	// Parse request body
	var req struct {
		ConversationID string                 `json:"conversation_id" binding:"required"`
		Type           string                 `json:"type" binding:"required"` // user_added, user_left, etc.
		Parameters     map[string]interface{} `json:"parameters"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "Invalid request body", err.Error())
		return
	}

	// Validate conversation ID
	if !validation.IsValidObjectID(req.ConversationID) {
		response.ValidationError(c, "Invalid conversation ID", nil)
		return
	}
	conversationID, _ := primitive.ObjectIDFromHex(req.ConversationID)

	// Create system message
	systemMessage := &models.SystemMessage{
		Type:       req.Type,
		Parameters: req.Parameters,
	}

	// Send the system message
	message, err := h.messageService.SendSystemMessage(c.Request.Context(), conversationID, systemMessage)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to send system message", err)
		return
	}

	// Return success response
	response.Created(c, "System message sent successfully", message)
}
