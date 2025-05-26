package messages

import (
	"net/http"

	"github.com/Caqil/vyrall/internal/services/message"
	"github.com/Caqil/vyrall/internal/utils/response"
	"github.com/Caqil/vyrall/internal/utils/validation"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ForwardHandler handles message forwarding operations
type ForwardHandler struct {
	messageService *message.Service
}

// NewForwardHandler creates a new forward handler
func NewForwardHandler(messageService *message.Service) *ForwardHandler {
	return &ForwardHandler{
		messageService: messageService,
	}
}

// ForwardMessage handles the request to forward a message to another conversation
func (h *ForwardHandler) ForwardMessage(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Parse request body
	var req struct {
		MessageID       string   `json:"message_id" binding:"required"`
		ConversationIDs []string `json:"conversation_ids" binding:"required"`
		AdditionalText  string   `json:"additional_text,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "Invalid request body", err.Error())
		return
	}

	// Validate message ID
	if !validation.IsValidObjectID(req.MessageID) {
		response.ValidationError(c, "Invalid message ID", nil)
		return
	}
	messageID, _ := primitive.ObjectIDFromHex(req.MessageID)

	// Get the message to check if the user is a participant
	message, err := h.messageService.GetMessage(c.Request.Context(), messageID)
	if err != nil {
		response.NotFoundError(c, "Message not found")
		return
	}

	// Check if user is a participant in the source conversation
	isParticipant, err := h.messageService.IsParticipant(c.Request.Context(), message.ConversationID, userID.(primitive.ObjectID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to verify conversation participation", err)
		return
	}

	if !isParticipant {
		response.ForbiddenError(c, "You are not a participant in this conversation")
		return
	}

	// Convert conversation IDs to ObjectIDs
	conversationIDs := make([]primitive.ObjectID, 0, len(req.ConversationIDs))
	for _, idStr := range req.ConversationIDs {
		if !validation.IsValidObjectID(idStr) {
			continue // Skip invalid IDs
		}
		convID, _ := primitive.ObjectIDFromHex(idStr)
		conversationIDs = append(conversationIDs, convID)
	}

	if len(conversationIDs) == 0 {
		response.ValidationError(c, "No valid conversation IDs provided", nil)
		return
	}

	// Check if user is a participant in all destination conversations
	validConversationIDs := make([]primitive.ObjectID, 0, len(conversationIDs))
	for _, convID := range conversationIDs {
		isParticipant, err := h.messageService.IsParticipant(c.Request.Context(), convID, userID.(primitive.ObjectID))
		if err != nil || !isParticipant {
			continue // Skip conversations where user is not a participant
		}
		validConversationIDs = append(validConversationIDs, convID)
	}

	if len(validConversationIDs) == 0 {
		response.ForbiddenError(c, "You are not a participant in any of the destination conversations")
		return
	}

	// Forward the message to each valid conversation
	forwardedMessages := make(map[string]interface{})
	for _, convID := range validConversationIDs {
		forwardedMsg, err := h.messageService.ForwardMessage(
			c.Request.Context(),
			messageID,
			convID,
			userID.(primitive.ObjectID),
			req.AdditionalText,
		)
		if err != nil {
			continue // Skip failed forwards
		}
		forwardedMessages[convID.Hex()] = forwardedMsg
	}

	if len(forwardedMessages) == 0 {
		response.Error(c, http.StatusInternalServerError, "Failed to forward message to any conversation", nil)
		return
	}

	// Return success response
	response.OK(c, "Message forwarded successfully", gin.H{
		"forwarded_messages": forwardedMessages,
		"success_count":      len(forwardedMessages),
		"total_count":        len(validConversationIDs),
	})
}

// ForwardMultipleMessages handles the request to forward multiple messages
func (h *ForwardHandler) ForwardMultipleMessages(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Parse request body
	var req struct {
		MessageIDs     []string `json:"message_ids" binding:"required"`
		ConversationID string   `json:"conversation_id" binding:"required"`
		AdditionalText string   `json:"additional_text,omitempty"`
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

	// Check if user is a participant in the destination conversation
	isParticipant, err := h.messageService.IsParticipant(c.Request.Context(), conversationID, userID.(primitive.ObjectID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to verify conversation participation", err)
		return
	}

	if !isParticipant {
		response.ForbiddenError(c, "You are not a participant in the destination conversation")
		return
	}

	// Convert message IDs to ObjectIDs
	messageIDs := make([]primitive.ObjectID, 0, len(req.MessageIDs))
	for _, idStr := range req.MessageIDs {
		if !validation.IsValidObjectID(idStr) {
			continue // Skip invalid IDs
		}
		msgID, _ := primitive.ObjectIDFromHex(idStr)
		messageIDs = append(messageIDs, msgID)
	}

	if len(messageIDs) == 0 {
		response.ValidationError(c, "No valid message IDs provided", nil)
		return
	}

	// Get the messages to check if the user is a participant in each source conversation
	messages, err := h.messageService.GetMessagesByIDs(c.Request.Context(), messageIDs)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve messages", err)
		return
	}

	// Filter out messages from conversations where the user is not a participant
	validMessageIDs := make([]primitive.ObjectID, 0, len(messages))
	for _, msg := range messages {
		isParticipant, err := h.messageService.IsParticipant(c.Request.Context(), msg.ConversationID, userID.(primitive.ObjectID))
		if err != nil || !isParticipant {
			continue // Skip messages where user is not a participant
		}
		validMessageIDs = append(validMessageIDs, msg.ID)
	}

	if len(validMessageIDs) == 0 {
		response.ForbiddenError(c, "You are not a participant in any of the source conversations")
		return
	}

	// Forward the messages
	forwardedMessages, err := h.messageService.ForwardMultipleMessages(
		c.Request.Context(),
		validMessageIDs,
		conversationID,
		userID.(primitive.ObjectID),
		req.AdditionalText,
	)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to forward messages", err)
		return
	}

	// Return success response
	response.OK(c, "Messages forwarded successfully", gin.H{
		"forwarded_messages": forwardedMessages,
		"success_count":      len(forwardedMessages),
		"total_count":        len(validMessageIDs),
	})
}
