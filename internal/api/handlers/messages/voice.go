package messages

import (
	"net/http"

	"github.com/Caqil/vyrall/internal/services/message"
	"github.com/Caqil/vyrall/internal/utils/response"
	"github.com/Caqil/vyrall/internal/utils/validation"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// VoiceHandler handles voice message operations
type VoiceHandler struct {
	messageService *message.Service
}

// NewVoiceHandler creates a new voice handler
func NewVoiceHandler(messageService *message.Service) *VoiceHandler {
	return &VoiceHandler{
		messageService: messageService,
	}
}

// SendVoiceMessage handles the request to send a voice message
func (h *VoiceHandler) SendVoiceMessage(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Parse multipart form with voice file
	if err := c.Request.ParseMultipartForm(32 << 20); err != nil { // 32MB max memory
		response.ValidationError(c, "Failed to parse form", err.Error())
		return
	}

	// Get conversation ID from form
	conversationIDStr := c.PostForm("conversation_id")
	if !validation.IsValidObjectID(conversationIDStr) {
		response.ValidationError(c, "Invalid conversation ID", nil)
		return
	}
	conversationID, _ := primitive.ObjectIDFromHex(conversationIDStr)

	// Get voice file from form
	file, header, err := c.Request.FormFile("voice_file")
	if err != nil {
		response.ValidationError(c, "No voice file uploaded", err.Error())
		return
	}
	defer file.Close()

	// Check file type
	fileType := validation.GetFileType(header.Filename)
	if fileType != "audio" {
		response.ValidationError(c, "File must be an audio file", nil)
		return
	}

	// Get additional parameters from form
	duration := 0
	durationStr := c.PostForm("duration")
	if durationStr != "" {
		if val, err := primitive.ParseInt32(durationStr); err == nil {
			duration = int(val)
		}
	}

	// Is encrypted?
	isEncrypted := c.PostForm("is_encrypted") == "true"

	// Waveform data (optional)
	waveform := c.PostForm("waveform")

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

	// Send voice message
	message, err := h.messageService.SendVoiceMessage(c.Request.Context(), conversationID, userID.(primitive.ObjectID), file, header, duration, waveform, isEncrypted)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to send voice message", err)
		return
	}

	// Return success response
	response.Created(c, "Voice message sent successfully", message)
}

// GetVoiceMessages handles the request to get voice messages from a conversation
func (h *VoiceHandler) GetVoiceMessages(c *gin.Context) {
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

	// Get voice messages
	messages, total, err := h.messageService.GetVoiceMessages(c.Request.Context(), conversationID, limit, offset)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get voice messages", err)
		return
	}

	// Return paginated response
	response.SuccessWithPagination(c, http.StatusOK, "Voice messages retrieved successfully", messages, limit, offset, total)
}

// GetVoiceMessageTranscription handles the request to get transcription of a voice message
func (h *VoiceHandler) GetVoiceMessageTranscription(c *gin.Context) {
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

	// Check if message is a voice message
	if message.MessageType != "voice" {
		response.ValidationError(c, "Not a voice message", nil)
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

	// Get voice message transcription
	transcription, err := h.messageService.GetVoiceMessageTranscription(c.Request.Context(), messageID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get voice message transcription", err)
		return
	}

	// Return success response
	response.OK(c, "Voice message transcription retrieved successfully", gin.H{
		"message_id":    messageID.Hex(),
		"transcription": transcription,
	})
}
