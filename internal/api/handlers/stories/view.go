package stories

import (
	"net/http"

	"github.com/Caqil/vyrall/internal/services/story"
	"github.com/Caqil/vyrall/internal/utils/response"
	"github.com/Caqil/vyrall/internal/utils/validation"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ViewHandler handles story viewing operations
type ViewHandler struct {
	storyService *story.Service
}

// NewViewHandler creates a new view handler
func NewViewHandler(storyService *story.Service) *ViewHandler {
	return &ViewHandler{
		storyService: storyService,
	}
}

// MarkAsViewed handles the request to mark a story as viewed
func (h *ViewHandler) MarkAsViewed(c *gin.Context) {
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

	// Parse request body (optional)
	var req struct {
		Device string `json:"device,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		// If there's an error parsing the request body, just ignore it
		req.Device = ""
	}

	// Mark story as viewed
	err := h.storyService.MarkAsViewed(c.Request.Context(), storyID, userID.(primitive.ObjectID), req.Device)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to mark story as viewed", err)
		return
	}

	// Return success response
	response.OK(c, "Story marked as viewed", nil)
}

// GetViewStatus handles the request to get the view status of a story
func (h *ViewHandler) GetViewStatus(c *gin.Context) {
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

	// Get view status
	status, err := h.storyService.GetViewStatus(c.Request.Context(), storyID, userID.(primitive.ObjectID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get view status", err)
		return
	}

	// Return success response
	response.OK(c, "View status retrieved successfully", status)
}

// GetUnviewedStories handles the request to get unviewed stories
func (h *ViewHandler) GetUnviewedStories(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Get unviewed stories
	stories, err := h.storyService.GetUnviewedStories(c.Request.Context(), userID.(primitive.ObjectID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve unviewed stories", err)
		return
	}

	// Return success response
	response.OK(c, "Unviewed stories retrieved successfully", stories)
}

// MarkAllStoriesAsViewed handles the request to mark all stories as viewed
func (h *ViewHandler) MarkAllStoriesAsViewed(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Parse request body (optional)
	var req struct {
		UserIDs []string `json:"user_ids,omitempty"` // If provided, only mark stories from these users
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		// If there's an error parsing the request body, just ignore it
		req.UserIDs = nil
	}

	// Convert user IDs to ObjectIDs if provided
	var targetUserIDs []primitive.ObjectID
	if len(req.UserIDs) > 0 {
		targetUserIDs = make([]primitive.ObjectID, 0, len(req.UserIDs))
		for _, idStr := range req.UserIDs {
			if !validation.IsValidObjectID(idStr) {
				continue // Skip invalid IDs
			}
			targetID, _ := primitive.ObjectIDFromHex(idStr)
			targetUserIDs = append(targetUserIDs, targetID)
		}
	}

	// Mark all stories as viewed
	count, err := h.storyService.MarkAllStoriesAsViewed(c.Request.Context(), userID.(primitive.ObjectID), targetUserIDs)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to mark all stories as viewed", err)
		return
	}

	// Return success response
	response.OK(c, "All stories marked as viewed", gin.H{
		"viewed_count": count,
	})
}
