package messages

import (
	"net/http"
	"time"

	"github.com/Caqil/vyrall/internal/models"
	"github.com/Caqil/vyrall/internal/services/message"
	"github.com/Caqil/vyrall/internal/utils/response"
	"github.com/Caqil/vyrall/internal/utils/validation"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ReadStatusHandler handles message read status operations
type ReadStatusHandler struct {
	messageService *message.Service
}

// NewReadStatusHandler creates a new read status handler
func NewReadStatusHandler(messageService *message.Service) *ReadStatusHandler {
	return &ReadStatusHandler{
		messageService: messageService,
	}
}

// MarkAsRead handles the request to mark a message as read
func (h *ReadStatusHandler) MarkAsRead(c *gin.Context) {
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

	// Get device information from request
	device := c.GetHeader("User-Agent")

	// Create read receipt
	readReceipt := models.ReadReceipt{
		UserID: userID.(primitive.ObjectID),
		ReadAt: time.Now(),
		Device: device,
	}

	// Mark the message as read
	err := h.messageService.MarkAsRead(c.Request.Context(), messageID, readReceipt)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to mark message as read", err)
		return
	}

	// Return success response
	response.OK(c, "Message marked as read", nil)
}

// MarkConversationAsRead handles the request to mark all messages in a conversation as read
func (h *ReadStatusHandler) MarkConversationAsRead(c *gin.Context) {
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

	// Get device information from request
	device := c.GetHeader("User-Agent")

	// Mark all messages in the conversation as read
	count, err := h.messageService.MarkConversationAsRead(c.Request.Context(), conversationID, userID.(primitive.ObjectID), device)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to mark conversation as read", err)
		return
	}

	// Return success response
	response.OK(c, "Conversation marked as read", gin.H{
		"marked_count": count,
	})
}

// GetReadReceipts handles the request to get read receipts for a message
func (h *ReadStatusHandler) GetReadReceipts(c *gin.Context) {
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

	// Get the message to check if the user is the sender or a participant
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

	// Get read receipts for the message
	readReceipts, err := h.messageService.GetReadReceipts(c.Request.Context(), messageID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get read receipts", err)
		return
	}

	// Return success response
	response.OK(c, "Read receipts retrieved successfully", readReceipts)
}
