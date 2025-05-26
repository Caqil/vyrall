package messages

import (
	"net/http"

	"github.com/Caqil/vyrall/internal/services/message"
	"github.com/Caqil/vyrall/internal/utils/response"
	"github.com/Caqil/vyrall/internal/utils/validation"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// DeliveryStatusHandler handles message delivery status operations
type DeliveryStatusHandler struct {
	messageService *message.Service
}

// NewDeliveryStatusHandler creates a new delivery status handler
func NewDeliveryStatusHandler(messageService *message.Service) *DeliveryStatusHandler {
	return &DeliveryStatusHandler{
		messageService: messageService,
	}
}

// MarkAsDelivered handles the request to mark a message as delivered
func (h *DeliveryStatusHandler) MarkAsDelivered(c *gin.Context) {
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

	// Mark the message as delivered
	err := h.messageService.MarkAsDelivered(c.Request.Context(), messageID, userID.(primitive.ObjectID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to mark message as delivered", err)
		return
	}

	// Return success response
	response.OK(c, "Message marked as delivered", nil)
}

// MarkConversationAsDelivered handles the request to mark all messages in a conversation as delivered
func (h *DeliveryStatusHandler) MarkConversationAsDelivered(c *gin.Context) {
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

	// Mark all messages in the conversation as delivered
	count, err := h.messageService.MarkConversationAsDelivered(c.Request.Context(), conversationID, userID.(primitive.ObjectID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to mark conversation as delivered", err)
		return
	}

	// Return success response
	response.OK(c, "Conversation marked as delivered", gin.H{
		"marked_count": count,
	})
}

// GetDeliveryStatus handles the request to get the delivery status of a message
func (h *DeliveryStatusHandler) GetDeliveryStatus(c *gin.Context) {
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

	// Check if user is the sender or a participant
	if message.SenderID != userID.(primitive.ObjectID) {
		isParticipant, err := h.messageService.IsParticipant(c.Request.Context(), message.ConversationID, userID.(primitive.ObjectID))
		if err != nil || !isParticipant {
			response.ForbiddenError(c, "You don't have permission to view this message's delivery status")
			return
		}
	}

	// Get delivery status
	deliveryStatus := message.DeliveryStatus
	if deliveryStatus == nil {
		deliveryStatus = make(map[string]string)
	}

	// Return success response
	response.OK(c, "Delivery status retrieved successfully", gin.H{
		"delivery_status": deliveryStatus,
		"message_id":      message.ID.Hex(),
		"sender_id":       message.SenderID.Hex(),
		"sent_at":         message.CreatedAt,
	})
}
