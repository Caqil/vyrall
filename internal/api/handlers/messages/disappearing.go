package messages

import (
	"net/http"
	"time"

	"github.com/Caqil/vyrall/internal/services/message"
	"github.com/Caqil/vyrall/internal/utils/response"
	"github.com/Caqil/vyrall/internal/utils/validation"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// DisappearingHandler handles disappearing messages operations
type DisappearingHandler struct {
	messageService *message.Service
}

// NewDisappearingHandler creates a new disappearing messages handler
func NewDisappearingHandler(messageService *message.Service) *DisappearingHandler {
	return &DisappearingHandler{
		messageService: messageService,
	}
}

// EnableDisappearing handles the request to enable disappearing messages for a conversation
func (h *DisappearingHandler) EnableDisappearing(c *gin.Context) {
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

	// Parse request body
	var req struct {
		Duration int `json:"duration" binding:"required"` // Duration in seconds
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "Invalid request body", err.Error())
		return
	}

	// Validate duration
	if req.Duration < 60 { // Minimum 1 minute
		response.ValidationError(c, "Disappearing message duration must be at least 60 seconds", nil)
		return
	}

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

	// Enable disappearing messages
	err = h.messageService.EnableDisappearingMessages(c.Request.Context(), conversationID, time.Duration(req.Duration)*time.Second)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to enable disappearing messages", err)
		return
	}

	// Return success response
	response.OK(c, "Disappearing messages enabled successfully", gin.H{
		"conversation_id": conversationID.Hex(),
		"duration":        req.Duration,
		"enabled_by":      userID.(primitive.ObjectID).Hex(),
		"enabled_at":      time.Now(),
	})
}

// DisableDisappearing handles the request to disable disappearing messages for a conversation
func (h *DisappearingHandler) DisableDisappearing(c *gin.Context) {
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

	// Disable disappearing messages
	err = h.messageService.DisableDisappearingMessages(c.Request.Context(), conversationID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to disable disappearing messages", err)
		return
	}

	// Return success response
	response.OK(c, "Disappearing messages disabled successfully", gin.H{
		"conversation_id": conversationID.Hex(),
		"disabled_by":     userID.(primitive.ObjectID).Hex(),
		"disabled_at":     time.Now(),
	})
}

// SendDisappearingMessage handles the request to send a disappearing message
func (h *DisappearingHandler) SendDisappearingMessage(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Parse request body
	var req struct {
		ConversationID string   `json:"conversation_id" binding:"required"`
		Content        string   `json:"content"`
		MediaIDs       []string `json:"media_ids,omitempty"`
		Duration       int      `json:"duration" binding:"required"` // Duration in seconds
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

	// Validate content or media
	if req.Content == "" && len(req.MediaIDs) == 0 {
		response.ValidationError(c, "Message must contain either text content or media", nil)
		return
	}

	// Validate duration
	if req.Duration < 5 { // Minimum 5 seconds
		response.ValidationError(c, "Disappearing message duration must be at least 5 seconds", nil)
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

	// Send the disappearing message
	message, err := h.messageService.SendDisappearingMessage(c.Request.Context(), conversationID, userID.(primitive.ObjectID), req.Content, mediaIDs, time.Duration(req.Duration)*time.Second)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to send disappearing message", err)
		return
	}

	// Return success response
	response.Created(c, "Disappearing message sent successfully", message)
}

// GetDisappearingSettings handles the request to get disappearing message settings for a conversation
func (h *DisappearingHandler) GetDisappearingSettings(c *gin.Context) {
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

	// Get disappearing message settings
	settings, err := h.messageService.GetDisappearingMessageSettings(c.Request.Context(), conversationID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get disappearing message settings", err)
		return
	}

	// Return success response
	response.OK(c, "Disappearing message settings retrieved successfully", settings)
}
