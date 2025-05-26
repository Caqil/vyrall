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

// ListHandler handles message listing operations
type ListHandler struct {
	messageService *message.Service
}

// NewListHandler creates a new list handler
func NewListHandler(messageService *message.Service) *ListHandler {
	return &ListHandler{
		messageService: messageService,
	}
}

// GetMessages handles the request to get messages for a conversation
func (h *ListHandler) GetMessages(c *gin.Context) {
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

	// Get pagination parameters
	limit, offset := response.GetPaginationParams(c)

	// Get additional filter parameters
	beforeTimestamp := c.DefaultQuery("before", "")
	afterTimestamp := c.DefaultQuery("after", "")
	messageType := c.DefaultQuery("type", "")

	var beforeTime, afterTime time.Time
	if beforeTimestamp != "" {
		if ts, err := time.Parse(time.RFC3339, beforeTimestamp); err == nil {
			beforeTime = ts
		}
	}
	if afterTimestamp != "" {
		if ts, err := time.Parse(time.RFC3339, afterTimestamp); err == nil {
			afterTime = ts
		}
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

	// Get messages
	messages, total, err := h.messageService.GetMessages(c.Request.Context(), conversationID, beforeTime, afterTime, messageType, limit, offset)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get messages", err)
		return
	}

	// Return paginated response
	response.SuccessWithPagination(c, http.StatusOK, "Messages retrieved successfully", messages, limit, offset, total)
}

// GetMessage handles the request to get a specific message
func (h *ListHandler) GetMessage(c *gin.Context) {
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

	// Return success response
	response.OK(c, "Message retrieved successfully", message)
}

// SearchMessages handles the request to search messages
func (h *ListHandler) SearchMessages(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Get conversation ID from URL parameter (optional)
	conversationIDStr := c.Query("conversation_id")
	var conversationID *primitive.ObjectID
	if conversationIDStr != "" && validation.IsValidObjectID(conversationIDStr) {
		id, _ := primitive.ObjectIDFromHex(conversationIDStr)
		conversationID = &id
	}

	// Get search query
	query := c.Query("q")
	if query == "" {
		response.ValidationError(c, "Search query is required", nil)
		return
	}

	// Get pagination parameters
	limit, offset := response.GetPaginationParams(c)

	// If conversation ID is provided, check if user is a participant
	if conversationID != nil {
		isParticipant, err := h.messageService.IsParticipant(c.Request.Context(), *conversationID, userID.(primitive.ObjectID))
		if err != nil {
			response.Error(c, http.StatusInternalServerError, "Failed to verify conversation participation", err)
			return
		}

		if !isParticipant {
			response.ForbiddenError(c, "You are not a participant in this conversation")
			return
		}
	}

	// Search messages
	messages, total, err := h.messageService.SearchMessages(c.Request.Context(), userID.(primitive.ObjectID), conversationID, query, limit, offset)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to search messages", err)
		return
	}

	// Return paginated response
	response.SuccessWithPagination(c, http.StatusOK, "Messages searched successfully", messages, limit, offset, total)
}

// GetUnreadCount handles the request to get unread message count
func (h *ListHandler) GetUnreadCount(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Get conversation ID from URL parameter (optional)
	conversationIDStr := c.Query("conversation_id")
	var conversationID *primitive.ObjectID
	if conversationIDStr != "" && validation.IsValidObjectID(conversationIDStr) {
		id, _ := primitive.ObjectIDFromHex(conversationIDStr)
		conversationID = &id
	}

	var unreadCount int
	var err error

	// Get unread count for a specific conversation or all conversations
	if conversationID != nil {
		// Check if user is a participant in the conversation
		isParticipant, err := h.messageService.IsParticipant(c.Request.Context(), *conversationID, userID.(primitive.ObjectID))
		if err != nil {
			response.Error(c, http.StatusInternalServerError, "Failed to verify conversation participation", err)
			return
		}

		if !isParticipant {
			response.ForbiddenError(c, "You are not a participant in this conversation")
			return
		}

		// Get unread count for specific conversation
		unreadCount, err = h.messageService.GetUnreadMessagesCount(c.Request.Context(), userID.(primitive.ObjectID), *conversationID)
	} else {
		// Get total unread count across all conversations
		unreadCount, err = h.messageService.GetTotalUnreadMessagesCount(c.Request.Context(), userID.(primitive.ObjectID))
	}

	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get unread message count", err)
		return
	}

	// Return success response
	response.OK(c, "Unread message count retrieved successfully", gin.H{
		"unread_count": unreadCount,
	})
}
