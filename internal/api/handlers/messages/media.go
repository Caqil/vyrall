package messages

import (
	"net/http"

	"github.com/Caqil/vyrall/internal/services/message"
	"github.com/Caqil/vyrall/internal/utils/response"
	"github.com/Caqil/vyrall/internal/utils/validation"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// MediaHandler handles media message operations
type MediaHandler struct {
	messageService *message.Service
}

// NewMediaHandler creates a new media handler
func NewMediaHandler(messageService *message.Service) *MediaHandler {
	return &MediaHandler{
		messageService: messageService,
	}
}

// SendMediaMessage handles the request to send a message with media
func (h *MediaHandler) SendMediaMessage(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Parse request body
	var req struct {
		ConversationID string   `json:"conversation_id" binding:"required"`
		Caption        string   `json:"caption,omitempty"`
		MediaIDs       []string `json:"media_ids" binding:"required"`
		IsEncrypted    bool     `json:"is_encrypted,omitempty"`
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

	// Validate media IDs
	if len(req.MediaIDs) == 0 {
		response.ValidationError(c, "At least one media ID is required", nil)
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

	if len(mediaIDs) == 0 {
		response.ValidationError(c, "No valid media IDs provided", nil)
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

	// Send media message
	message, err := h.messageService.SendMediaMessage(c.Request.Context(), conversationID, userID.(primitive.ObjectID), req.Caption, mediaIDs, req.IsEncrypted)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to send media message", err)
		return
	}

	// Return success response
	response.Created(c, "Media message sent successfully", message)
}

// GetMediaMessages handles the request to get media messages from a conversation
func (h *MediaHandler) GetMediaMessages(c *gin.Context) {
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

	// Get media type filter (optional)
	mediaType := c.DefaultQuery("type", "") // image, video, audio, document, or empty for all

	// Get pagination parameters
	limit, offset := response.GetPaginationParams(c)

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

	// Get media messages
	messages, total, err := h.messageService.GetMediaMessages(c.Request.Context(), conversationID, mediaType, limit, offset)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get media messages", err)
		return
	}

	// Return paginated response
	response.SuccessWithPagination(c, http.StatusOK, "Media messages retrieved successfully", messages, limit, offset, total)
}

// GetSharedMedia handles the request to get all media shared between users
func (h *MediaHandler) GetSharedMedia(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Get other user ID from URL parameter
	otherUserIDStr := c.Param("userId")
	if !validation.IsValidObjectID(otherUserIDStr) {
		response.ValidationError(c, "Invalid user ID", nil)
		return
	}
	otherUserID, _ := primitive.ObjectIDFromHex(otherUserIDStr)

	// Get media type filter (optional)
	mediaType := c.DefaultQuery("type", "") // image, video, audio, document, or empty for all

	// Get pagination parameters
	limit, offset := response.GetPaginationParams(c)

	// Get shared media
	messages, total, err := h.messageService.GetSharedMedia(c.Request.Context(), userID.(primitive.ObjectID), otherUserID, mediaType, limit, offset)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get shared media", err)
		return
	}

	// Return paginated response
	response.SuccessWithPagination(c, http.StatusOK, "Shared media retrieved successfully", messages, limit, offset, total)
}

// GetMessageMedia handles the request to get media files from a specific message
func (h *MediaHandler) GetMessageMedia(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Get message ID from URL parameter
	messageIDStr := c.Param("id")
	if !validation.IsValidObjectID(messageIDStr) {
		response.ValidationError(c, "Invalid message ID", nil)
		return
	}
	messageID, _ := primitive.ObjectIDFromHex(messageIDStr)

	// Get the message
	message, err := h.messageService.GetMessage(c.Request.Context(), messageID)
	if err != nil {
		response.NotFoundError(c, "Message not found")
		return
	}

	// Check if user is a participant in the conversation
	isParticipant, err := h.messageService.IsParticipant(c.Request.Context(), message.ConversationID, userID.(primitive.ObjectID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to verify conversation participation", err)
		return
	}

	if !isParticipant {
		response.ForbiddenError(c, "You are not a participant in this conversation")
		return
	}

	// Get message media
	mediaFiles, err := h.messageService.GetMessageMedia(c.Request.Context(), messageID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get message media", err)
		return
	}

	// Return success response
	response.OK(c, "Message media retrieved successfully", mediaFiles)
}
