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

// EditHandler handles message editing operations
type EditHandler struct {
	messageService *message.Service
}

// NewEditHandler creates a new edit handler
func NewEditHandler(messageService *message.Service) *EditHandler {
	return &EditHandler{
		messageService: messageService,
	}
}

// EditMessage handles the request to edit a message
func (h *EditHandler) EditMessage(c *gin.Context) {
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

	// Parse request body
	var req struct {
		Content string `json:"content" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "Invalid request body", err.Error())
		return
	}

	// Validate content
	if req.Content == "" {
		response.ValidationError(c, "Message content cannot be empty", nil)
		return
	}

	// Get the message to check ownership
	message, err := h.messageService.GetMessage(c.Request.Context(), messageID)
	if err != nil {
		response.NotFoundError(c, "Message not found")
		return
	}

	// Only the sender can edit a message
	if message.SenderID != userID.(primitive.ObjectID) {
		response.ForbiddenError(c, "Only the sender can edit a message")
		return
	}

	// Check if the message is editable
	if message.IsDeleted {
		response.ValidationError(c, "Deleted messages cannot be edited", nil)
		return
	}

	// Certain message types might not be editable
	if message.MessageType != "text" {
		response.ValidationError(c, "Only text messages can be edited", nil)
		return
	}

	// Create edit record
	editRecord := models.MessageEdit{
		PreviousContent: message.Content,
		EditedAt:        time.Now(),
	}

	// Edit the message
	updatedMessage, err := h.messageService.EditMessage(c.Request.Context(), messageID, req.Content, editRecord)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to edit message", err)
		return
	}

	// Return success response
	response.OK(c, "Message edited successfully", updatedMessage)
}

// GetEditHistory handles the request to get the edit history of a message
func (h *EditHandler) GetEditHistory(c *gin.Context) {
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

	// Return edit history
	if !message.IsEdited || len(message.EditHistory) == 0 {
		response.OK(c, "Message has not been edited", []models.MessageEdit{})
		return
	}

	// Return success response
	response.OK(c, "Edit history retrieved successfully", message.EditHistory)
}
