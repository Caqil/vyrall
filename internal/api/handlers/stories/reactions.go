package stories

import (
	"net/http"

	"github.com/Caqil/vyrall/internal/services/story"
	"github.com/Caqil/vyrall/internal/utils/response"
	"github.com/Caqil/vyrall/internal/utils/validation"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ReactionsHandler handles story reactions operations
type ReactionsHandler struct {
	storyService *story.Service
}

// NewReactionsHandler creates a new reactions handler
func NewReactionsHandler(storyService *story.Service) *ReactionsHandler {
	return &ReactionsHandler{
		storyService: storyService,
	}
}

// AddReaction handles the request to add a reaction to a story
func (h *ReactionsHandler) AddReaction(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Get story ID from URL parameter
	storyIDStr := c.Param("id")
	if !validation.IsValidObjectID(storyIDStr) {
		response.ValidationError(c, "Invalid story ID", nil)
		return
	}
	storyID, _ := primitive.ObjectIDFromHex(storyIDStr)

	// Parse request body
	var req struct {
		Reaction string `json:"reaction" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "Invalid request body", err.Error())
		return
	}

	if req.Reaction == "" {
		response.ValidationError(c, "Reaction cannot be empty", nil)
		return
	}

	// Add reaction to story
	err := h.storyService.AddReaction(c.Request.Context(), storyID, userID.(primitive.ObjectID), req.Reaction)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to add reaction", err)
		return
	}

	// Return success response
	response.OK(c, "Reaction added successfully", nil)
}

// GetReactions handles the request to get reactions to a story
func (h *ReactionsHandler) GetReactions(c *gin.Context) {
	// Get story ID from URL parameter
	storyIDStr := c.Param("id")
	if !validation.IsValidObjectID(storyIDStr) {
		response.ValidationError(c, "Invalid story ID", nil)
		return
	}
	storyID, _ := primitive.ObjectIDFromHex(storyIDStr)

	// Get authenticated user ID (may be nil for unauthenticated users)
	var userID primitive.ObjectID
	if id, exists := c.Get("userID"); exists {
		userID = id.(primitive.ObjectID)
	}

	// Get pagination parameters
	limit, offset := response.GetPaginationParams(c)

	// Get story reactions
	reactions, total, err := h.storyService.GetReactions(c.Request.Context(), storyID, userID, limit, offset)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve reactions", err)
		return
	}

	// Return paginated response
	response.SuccessWithPagination(c, http.StatusOK, "Reactions retrieved successfully", reactions, limit, offset, total)
}

// RemoveReaction handles the request to remove a reaction from a story
func (h *ReactionsHandler) RemoveReaction(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Get story ID from URL parameter
	storyIDStr := c.Param("id")
	if !validation.IsValidObjectID(storyIDStr) {
		response.ValidationError(c, "Invalid story ID", nil)
		return
	}
	storyID, _ := primitive.ObjectIDFromHex(storyIDStr)

	// Remove reaction from story
	err := h.storyService.RemoveReaction(c.Request.Context(), storyID, userID.(primitive.ObjectID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to remove reaction", err)
		return
	}

	// Return success response
	response.OK(c, "Reaction removed successfully", nil)
}

// ReplyToStory handles the request to reply to a story
func (h *ReactionsHandler) ReplyToStory(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Get story ID from URL parameter
	storyIDStr := c.Param("id")
	if !validation.IsValidObjectID(storyIDStr) {
		response.ValidationError(c, "Invalid story ID", nil)
		return
	}
	storyID, _ := primitive.ObjectIDFromHex(storyIDStr)

	// Parse request body
	var req struct {
		Content   string `json:"content" binding:"required"`
		MediaURL  string `json:"media_url,omitempty"`
		IsPrivate bool   `json:"is_private,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "Invalid request body", err.Error())
		return
	}

	if req.Content == "" && req.MediaURL == "" {
		response.ValidationError(c, "Reply must contain either text content or media", nil)
		return
	}

	// Add reply to story
	reply, err := h.storyService.ReplyToStory(
		c.Request.Context(),
		storyID,
		userID.(primitive.ObjectID),
		req.Content,
		req.MediaURL,
		req.IsPrivate,
	)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to reply to story", err)
		return
	}

	// Return success response
	response.Created(c, "Reply sent successfully", reply)
}

// GetStoryReplies handles the request to get replies to a story
func (h *ReactionsHandler) GetStoryReplies(c *gin.Context) {
	// Get story ID from URL parameter
	storyIDStr := c.Param("id")
	if !validation.IsValidObjectID(storyIDStr) {
		response.ValidationError(c, "Invalid story ID", nil)
		return
	}
	storyID, _ := primitive.ObjectIDFromHex(storyIDStr)

	// Get authenticated user ID (may be nil for unauthenticated users)
	var userID primitive.ObjectID
	if id, exists := c.Get("userID"); exists {
		userID = id.(primitive.ObjectID)
	}

	// Get pagination parameters
	limit, offset := response.GetPaginationParams(c)

	// Get story replies
	replies, total, err := h.storyService.GetStoryReplies(c.Request.Context(), storyID, userID, limit, offset)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve replies", err)
		return
	}

	// Return paginated response
	response.SuccessWithPagination(c, http.StatusOK, "Replies retrieved successfully", replies, limit, offset, total)
}
