package messages

import (
	"net/http"

	"github.com/Caqil/vyrall/internal/services/message"
	"github.com/Caqil/vyrall/internal/utils/response"
	"github.com/Caqil/vyrall/internal/utils/validation"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ConversationsHandler handles conversation operations
type ConversationsHandler struct {
	messageService *message.Service
}

// NewConversationsHandler creates a new conversations handler
func NewConversationsHandler(messageService *message.Service) *ConversationsHandler {
	return &ConversationsHandler{
		messageService: messageService,
	}
}

// GetConversations handles the request to get a user's conversations
func (h *ConversationsHandler) GetConversations(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Get pagination parameters
	limit, offset := response.GetPaginationParams(c)

	// Get conversations
	conversations, total, err := h.messageService.GetConversations(c.Request.Context(), userID.(primitive.ObjectID), limit, offset)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get conversations", err)
		return
	}

	// Return paginated response
	response.SuccessWithPagination(c, http.StatusOK, "Conversations retrieved successfully", conversations, limit, offset, total)
}

// GetConversation handles the request to get a specific conversation
func (h *ConversationsHandler) GetConversation(c *gin.Context) {
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

	// Get the conversation
	conversation, err := h.messageService.GetConversation(c.Request.Context(), conversationID)
	if err != nil {
		response.NotFoundError(c, "Conversation not found")
		return
	}

	// Check if user is a participant
	isParticipant := false
	for _, participant := range conversation.Participants {
		if participant.UserID == userID.(primitive.ObjectID) {
			isParticipant = true
			break
		}
	}

	if !isParticipant {
		response.ForbiddenError(c, "You are not a participant in this conversation")
		return
	}

	// Return success response
	response.OK(c, "Conversation retrieved successfully", conversation)
}

// CreateConversation handles the request to create a new direct conversation
func (h *ConversationsHandler) CreateConversation(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Parse request body
	var req struct {
		ParticipantID string `json:"participant_id" binding:"required"`
		IsEncrypted   bool   `json:"is_encrypted,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "Invalid request body", err.Error())
		return
	}

	// Validate participant ID
	if !validation.IsValidObjectID(req.ParticipantID) {
		response.ValidationError(c, "Invalid participant ID", nil)
		return
	}
	participantID, _ := primitive.ObjectIDFromHex(req.ParticipantID)

	// Verify that the participant ID is not the same as the user ID
	if participantID == userID.(primitive.ObjectID) {
		response.ValidationError(c, "Cannot create a conversation with yourself", nil)
		return
	}

	// Create the conversation
	conversation, err := h.messageService.CreateDirectConversation(c.Request.Context(), userID.(primitive.ObjectID), participantID, req.IsEncrypted)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to create conversation", err)
		return
	}

	// Return success response
	response.Created(c, "Conversation created successfully", conversation)
}

// ArchiveConversation handles the request to archive a conversation
func (h *ConversationsHandler) ArchiveConversation(c *gin.Context) {
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

	// Check if user is a participant
	isParticipant, err := h.messageService.IsParticipant(c.Request.Context(), conversationID, userID.(primitive.ObjectID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to verify conversation participation", err)
		return
	}

	if !isParticipant {
		response.ForbiddenError(c, "You are not a participant in this conversation")
		return
	}

	// Archive the conversation
	err = h.messageService.ArchiveConversation(c.Request.Context(), conversationID, userID.(primitive.ObjectID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to archive conversation", err)
		return
	}

	// Return success response
	response.OK(c, "Conversation archived successfully", nil)
}

// UnarchiveConversation handles the request to unarchive a conversation
func (h *ConversationsHandler) UnarchiveConversation(c *gin.Context) {
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

	// Check if user is a participant
	isParticipant, err := h.messageService.IsParticipant(c.Request.Context(), conversationID, userID.(primitive.ObjectID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to verify conversation participation", err)
		return
	}

	if !isParticipant {
		response.ForbiddenError(c, "You are not a participant in this conversation")
		return
	}

	// Unarchive the conversation
	err = h.messageService.UnarchiveConversation(c.Request.Context(), conversationID, userID.(primitive.ObjectID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to unarchive conversation", err)
		return
	}

	// Return success response
	response.OK(c, "Conversation unarchived successfully", nil)
}

// MuteConversation handles the request to mute a conversation
func (h *ConversationsHandler) MuteConversation(c *gin.Context) {
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
		Duration int `json:"duration"` // Duration in seconds, 0 means indefinitely
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		// Default to indefinite mute
		req.Duration = 0
	}

	// Check if user is a participant
	isParticipant, err := h.messageService.IsParticipant(c.Request.Context(), conversationID, userID.(primitive.ObjectID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to verify conversation participation", err)
		return
	}

	if !isParticipant {
		response.ForbiddenError(c, "You are not a participant in this conversation")
		return
	}

	// Mute the conversation
	expirationTime, err := h.messageService.MuteConversation(c.Request.Context(), conversationID, userID.(primitive.ObjectID), req.Duration)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to mute conversation", err)
		return
	}

	// Return success response
	response.OK(c, "Conversation muted successfully", gin.H{
		"muted_until": expirationTime,
	})
}

// UnmuteConversation handles the request to unmute a conversation
func (h *ConversationsHandler) UnmuteConversation(c *gin.Context) {
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

	// Check if user is a participant
	isParticipant, err := h.messageService.IsParticipant(c.Request.Context(), conversationID, userID.(primitive.ObjectID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to verify conversation participation", err)
		return
	}

	if !isParticipant {
		response.ForbiddenError(c, "You are not a participant in this conversation")
		return
	}

	// Unmute the conversation
	err = h.messageService.UnmuteConversation(c.Request.Context(), conversationID, userID.(primitive.ObjectID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to unmute conversation", err)
		return
	}

	// Return success response
	response.OK(c, "Conversation unmuted successfully", nil)
}

// UpdateConversationSettings handles the request to update a user's conversation settings
func (h *ConversationsHandler) UpdateConversationSettings(c *gin.Context) {
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
		Nickname            string `json:"nickname,omitempty"`
		NotificationSetting string `json:"notification_setting,omitempty"` // all, mentions, none
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "Invalid request body", err.Error())
		return
	}

	// Validate notification setting
	if req.NotificationSetting != "" && req.NotificationSetting != "all" && req.NotificationSetting != "mentions" && req.NotificationSetting != "none" {
		response.ValidationError(c, "Invalid notification setting. Must be 'all', 'mentions', or 'none'", nil)
		return
	}

	// Check if user is a participant
	isParticipant, err := h.messageService.IsParticipant(c.Request.Context(), conversationID, userID.(primitive.ObjectID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to verify conversation participation", err)
		return
	}

	if !isParticipant {
		response.ForbiddenError(c, "You are not a participant in this conversation")
		return
	}

	// Update conversation settings
	updates := make(map[string]interface{})
	if req.Nickname != "" {
		updates["nickname"] = req.Nickname
	}
	if req.NotificationSetting != "" {
		updates["notification_setting"] = req.NotificationSetting
	}

	if len(updates) == 0 {
		response.ValidationError(c, "No updates provided", nil)
		return
	}

	// Apply the updates
	err = h.messageService.UpdateParticipantSettings(c.Request.Context(), conversationID, userID.(primitive.ObjectID), updates)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to update conversation settings", err)
		return
	}

	// Return success response
	response.OK(c, "Conversation settings updated successfully", nil)
}

// SearchConversations handles the request to search for conversations
func (h *ConversationsHandler) SearchConversations(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Get search query
	query := c.Query("q")
	if query == "" {
		response.ValidationError(c, "Search query is required", nil)
		return
	}

	// Get pagination parameters
	limit, offset := response.GetPaginationParams(c)

	// Search conversations
	conversations, total, err := h.messageService.SearchConversations(c.Request.Context(), userID.(primitive.ObjectID), query, limit, offset)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to search conversations", err)
		return
	}

	// Return paginated response
	response.SuccessWithPagination(c, http.StatusOK, "Conversations searched successfully", conversations, limit, offset, total)
}
