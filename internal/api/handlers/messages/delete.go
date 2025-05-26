package messages

import (
	"net/http"

	"github.com/Caqil/vyrall/internal/services/message"
	"github.com/Caqil/vyrall/internal/utils/response"
	"github.com/Caqil/vyrall/internal/utils/validation"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// DeleteHandler handles message deletion operations
type DeleteHandler struct {
	messageService *message.Service
}

// NewDeleteHandler creates a new delete handler
func NewDeleteHandler(messageService *message.Service) *DeleteHandler {
	return &DeleteHandler{
		messageService: messageService,
	}
}

// DeleteMessage handles the request to delete a message
func (h *DeleteHandler) DeleteMessage(c *gin.Context) {
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

	// Get the message to check ownership
	message, err := h.messageService.GetMessage(c.Request.Context(), messageID)
	if err != nil {
		response.NotFoundError(c, "Message not found")
		return
	}

	// Check if delete mode is "for_everyone" or "for_me"
	deleteMode := c.DefaultQuery("mode", "for_me")
	if deleteMode == "for_everyone" {
		// Only the sender can delete a message for everyone
		if message.SenderID != userID.(primitive.ObjectID) {
			response.ForbiddenError(c, "Only the sender can delete a message for everyone")
			return
		}

		// Delete the message for everyone
		err = h.messageService.DeleteMessage(c.Request.Context(), messageID)
	} else {
		// Delete the message for the current user only
		err = h.messageService.DeleteMessageForUser(c.Request.Context(), messageID, userID.(primitive.ObjectID))
	}

	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to delete message", err)
		return
	}

	// Return success response
	response.OK(c, "Message deleted successfully", nil)
}

// BulkDeleteMessages handles the request to delete multiple messages
func (h *DeleteHandler) BulkDeleteMessages(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Parse request body
	var req struct {
		MessageIDs []string `json:"message_ids" binding:"required"`
		Mode       string   `json:"mode"` // "for_everyone" or "for_me"
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "Invalid request body", err.Error())
		return
	}

	if len(req.MessageIDs) == 0 {
		response.ValidationError(c, "No message IDs provided", nil)
		return
	}

	// Set default mode if not provided
	if req.Mode == "" {
		req.Mode = "for_me"
	}

	// Convert message IDs to ObjectIDs
	messageIDs := make([]primitive.ObjectID, 0, len(req.MessageIDs))
	for _, idStr := range req.MessageIDs {
		if !validation.IsValidObjectID(idStr) {
			continue // Skip invalid IDs
		}
		messageID, _ := primitive.ObjectIDFromHex(idStr)
		messageIDs = append(messageIDs, messageID)
	}

	if len(messageIDs) == 0 {
		response.ValidationError(c, "No valid message IDs provided", nil)
		return
	}

	var deletedCount int
	var err error

	if req.Mode == "for_everyone" {
		// For "for_everyone" mode, we need to check if the user is the sender of all messages
		messages, err := h.messageService.GetMessagesByIDs(c.Request.Context(), messageIDs)
		if err != nil {
			response.Error(c, http.StatusInternalServerError, "Failed to retrieve messages", err)
			return
		}

		// Filter out messages that don't belong to the user
		userMessageIDs := make([]primitive.ObjectID, 0, len(messages))
		for _, msg := range messages {
			if msg.SenderID == userID.(primitive.ObjectID) {
				userMessageIDs = append(userMessageIDs, msg.ID)
			}
		}

		if len(userMessageIDs) == 0 {
			response.ForbiddenError(c, "You don't have permission to delete any of these messages for everyone")
			return
		}

		// Delete the messages for everyone
		deletedCount, err = h.messageService.BulkDeleteMessages(c.Request.Context(), userMessageIDs)
	} else {
		// Delete the messages for the current user only
		deletedCount, err = h.messageService.BulkDeleteMessagesForUser(c.Request.Context(), messageIDs, userID.(primitive.ObjectID))
	}

	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to delete messages", err)
		return
	}

	// Return success response
	response.OK(c, "Messages deleted successfully", gin.H{
		"deleted_count": deletedCount,
	})
}

// ClearConversation handles the request to delete all messages in a conversation for the current user
func (h *DeleteHandler) ClearConversation(c *gin.Context) {
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

	// Clear the conversation for the current user
	deletedCount, err := h.messageService.ClearConversationForUser(c.Request.Context(), conversationID, userID.(primitive.ObjectID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to clear conversation", err)
		return
	}

	// Return success response
	response.OK(c, "Conversation cleared successfully", gin.H{
		"deleted_count": deletedCount,
	})
}
