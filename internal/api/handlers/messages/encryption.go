package messages

import (
	"net/http"

	"github.com/Caqil/vyrall/internal/services/message"
	"github.com/Caqil/vyrall/internal/utils/response"
	"github.com/Caqil/vyrall/internal/utils/validation"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// EncryptionHandler handles message encryption operations
type EncryptionHandler struct {
	messageService *message.Service
}

// NewEncryptionHandler creates a new encryption handler
func NewEncryptionHandler(messageService *message.Service) *EncryptionHandler {
	return &EncryptionHandler{
		messageService: messageService,
	}
}

// EnableEncryption handles the request to enable encryption for a conversation
func (h *EncryptionHandler) EnableEncryption(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Get conversation ID from URL parameter
	conversationIDStr := c.Param("id")
	if !validation.IsValidObjectID(conversationIDStr) {
		response.ValidationError(c, "Invalid conversation ID", nil)
		return
	}
	conversationID, _ := primitive.ObjectIDFromHex(conversationIDStr)

	// Check if user is a participant in the conversation
	isParticipant, err := h.messageService.IsParticipant(c.Request.Context(), conversationID, userID.(primitive.ObjectID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to verify conversation participation", err)
		return
	}

	if !isParticipant {
		response.ForbiddenError(c, "You are not a participant in this conversation")
		return
	}

	// Enable encryption for the conversation
	conversation, err := h.messageService.EnableEncryption(c.Request.Context(), conversationID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to enable encryption", err)
		return
	}

	// Return success response
	response.OK(c, "Encryption enabled successfully", conversation)
}

// DisableEncryption handles the request to disable encryption for a conversation
func (h *EncryptionHandler) DisableEncryption(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Get conversation ID from URL parameter
	conversationIDStr := c.Param("id")
	if !validation.IsValidObjectID(conversationIDStr) {
		response.ValidationError(c, "Invalid conversation ID", nil)
		return
	}
	conversationID, _ := primitive.ObjectIDFromHex(conversationIDStr)

	// Check if user is a participant in the conversation
	isParticipant, err := h.messageService.IsParticipant(c.Request.Context(), conversationID, userID.(primitive.ObjectID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to verify conversation participation", err)
		return
	}

	if !isParticipant {
		response.ForbiddenError(c, "You are not a participant in this conversation")
		return
	}

	// Disable encryption for the conversation
	conversation, err := h.messageService.DisableEncryption(c.Request.Context(), conversationID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to disable encryption", err)
		return
	}

	// Return success response
	response.OK(c, "Encryption disabled successfully", conversation)
}

// UpdateEncryptionKeys handles the request to update encryption keys for a user in a conversation
func (h *EncryptionHandler) UpdateEncryptionKeys(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Parse request body
	var req struct {
		ConversationID string            `json:"conversation_id" binding:"required"`
		PublicKey      string            `json:"public_key" binding:"required"`
		KeyID          string            `json:"key_id" binding:"required"`
		Algorithm      string            `json:"algorithm" binding:"required"`
		RecipientKeys  map[string]string `json:"recipient_keys,omitempty"`
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

	// Check if user is a participant in the conversation
	isParticipant, err := h.messageService.IsParticipant(c.Request.Context(), conversationID, userID.(primitive.ObjectID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to verify conversation participation", err)
		return
	}

	if !isParticipant {
		response.ForbiddenError(c, "You are not a participant in this conversation")
		return
	}

	// Update encryption keys
	err = h.messageService.UpdateEncryptionKeys(
		c.Request.Context(),
		conversationID,
		userID.(primitive.ObjectID),
		req.PublicKey,
		req.KeyID,
		req.Algorithm,
		req.RecipientKeys,
	)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to update encryption keys", err)
		return
	}

	// Return success response
	response.OK(c, "Encryption keys updated successfully", nil)
}

// GetEncryptionKeys handles the request to get encryption keys for a conversation
func (h *EncryptionHandler) GetEncryptionKeys(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Get conversation ID from URL parameter
	conversationIDStr := c.Param("id")
	if !validation.IsValidObjectID(conversationIDStr) {
		response.ValidationError(c, "Invalid conversation ID", nil)
		return
	}
	conversationID, _ := primitive.ObjectIDFromHex(conversationIDStr)

	// Check if user is a participant in the conversation
	isParticipant, err := h.messageService.IsParticipant(c.Request.Context(), conversationID, userID.(primitive.ObjectID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to verify conversation participation", err)
		return
	}

	if !isParticipant {
		response.ForbiddenError(c, "You are not a participant in this conversation")
		return
	}

	// Get encryption keys
	keys, err := h.messageService.GetEncryptionKeys(c.Request.Context(), conversationID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get encryption keys", err)
		return
	}

	// Return success response
	response.OK(c, "Encryption keys retrieved successfully", keys)
}
