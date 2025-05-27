package stories

import (
	"net/http"

	"github.com/Caqil/vyrall/internal/services/story"
	"github.com/Caqil/vyrall/internal/utils/response"
	"github.com/Caqil/vyrall/internal/utils/validation"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ViewersHandler handles story viewers operations
type ViewersHandler struct {
	storyService *story.Service
}

// NewViewersHandler creates a new viewers handler
func NewViewersHandler(storyService *story.Service) *ViewersHandler {
	return &ViewersHandler{
		storyService: storyService,
	}
}

// GetStoryViewers handles the request to get viewers of a story
func (h *ViewersHandler) GetStoryViewers(c *gin.Context) {
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

	// Get sort parameter
	sortBy := c.DefaultQuery("sort_by", "recent") // recent, friends_first

	// Get pagination parameters
	limit, offset := response.GetPaginationParams(c)

	// Get story viewers
	viewers, total, err := h.storyService.GetStoryViewers(c.Request.Context(), storyID, userID.(primitive.ObjectID), sortBy, limit, offset)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve story viewers", err)
		return
	}

	// Return paginated response
	response.SuccessWithPagination(c, http.StatusOK, "Story viewers retrieved successfully", viewers, limit, offset, total)
}

// GetStoryViewerStats handles the request to get viewer statistics for a story
func (h *ViewersHandler) GetStoryViewerStats(c *gin.Context) {
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

	// Get viewer statistics
	stats, err := h.storyService.GetStoryViewerStats(c.Request.Context(), storyID, userID.(primitive.ObjectID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve viewer statistics", err)
		return
	}

	// Return success response
	response.OK(c, "Viewer statistics retrieved successfully", stats)
}

// GetStoryReplays handles the request to get users who replayed a story
func (h *ViewersHandler) GetStoryReplays(c *gin.Context) {
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

	// Get pagination parameters
	limit, offset := response.GetPaginationParams(c)

	// Get story replays
	replays, total, err := h.storyService.GetStoryReplays(c.Request.Context(), storyID, userID.(primitive.ObjectID), limit, offset)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve story replays", err)
		return
	}

	// Return paginated response
	response.SuccessWithPagination(c, http.StatusOK, "Story replays retrieved successfully", replays, limit, offset, total)
}

// GetUserStoriesViewCount handles the request to get view count for all user stories
func (h *ViewersHandler) GetUserStoriesViewCount(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Get time range parameter
	timeRange := c.DefaultQuery("time_range", "all") // day, week, month, all

	// Get view count
	viewCount, err := h.storyService.GetUserStoriesViewCount(c.Request.Context(), userID.(primitive.ObjectID), timeRange)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve view count", err)
		return
	}

	// Return success response
	response.OK(c, "View count retrieved successfully", viewCount)
}

// BlockViewerFromStories handles the request to block a user from viewing stories
func (h *ViewersHandler) BlockViewerFromStories(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Parse request body
	var req struct {
		BlockedUserID string `json:"blocked_user_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "Invalid request body", err.Error())
		return
	}

	if !validation.IsValidObjectID(req.BlockedUserID) {
		response.ValidationError(c, "Invalid user ID", nil)
		return
	}
	blockedUserID, _ := primitive.ObjectIDFromHex(req.BlockedUserID)

	// Block viewer from stories
	err := h.storyService.BlockViewerFromStories(c.Request.Context(), userID.(primitive.ObjectID), blockedUserID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to block user from viewing stories", err)
		return
	}

	// Return success response
	response.OK(c, "User blocked from viewing stories", nil)
}
