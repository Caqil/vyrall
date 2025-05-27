package stories

import (
	"net/http"

	"github.com/Caqil/vyrall/internal/services/story"
	"github.com/Caqil/vyrall/internal/utils/response"
	"github.com/Caqil/vyrall/internal/utils/validation"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ListHandler handles story listing operations
type ListHandler struct {
	storyService *story.Service
}

// NewListHandler creates a new list handler
func NewListHandler(storyService *story.Service) *ListHandler {
	return &ListHandler{
		storyService: storyService,
	}
}

// GetFeedStories handles the request to get stories for the feed
func (h *ListHandler) GetFeedStories(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Get pagination parameters
	limit, offset := response.GetPaginationParams(c)

	// Get feed stories
	stories, total, err := h.storyService.GetFeedStories(c.Request.Context(), userID.(primitive.ObjectID), limit, offset)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve feed stories", err)
		return
	}

	// Return paginated response
	response.SuccessWithPagination(c, http.StatusOK, "Feed stories retrieved successfully", stories, limit, offset, total)
}

// GetUserStories handles the request to get a specific user's stories
func (h *ListHandler) GetUserStories(c *gin.Context) {
	// Get target user ID from URL parameter
	userIDStr := c.Param("userId")
	if !validation.IsValidObjectID(userIDStr) {
		response.ValidationError(c, "Invalid user ID", nil)
		return
	}
	targetUserID, _ := primitive.ObjectIDFromHex(userIDStr)

	// Get authenticated user ID (may be nil for unauthenticated users)
	var authUserID primitive.ObjectID
	if id, exists := c.Get("userID"); exists {
		authUserID = id.(primitive.ObjectID)
	}

	// Get user stories
	stories, err := h.storyService.GetUserStories(c.Request.Context(), targetUserID, authUserID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve user stories", err)
		return
	}

	// Return success response
	response.OK(c, "User stories retrieved successfully", stories)
}

// GetStoryByID handles the request to get a specific story
func (h *ListHandler) GetStoryByID(c *gin.Context) {
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

	// Get the story
	story, err := h.storyService.GetStoryByID(c.Request.Context(), storyID, userID)
	if err != nil {
		response.NotFoundError(c, "Story not found")
		return
	}

	// Return success response
	response.OK(c, "Story retrieved successfully", story)
}

// GetMyStories handles the request to get the authenticated user's stories
func (h *ListHandler) GetMyStories(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Get the user's stories
	stories, err := h.storyService.GetMyStories(c.Request.Context(), userID.(primitive.ObjectID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve your stories", err)
		return
	}

	// Return success response
	response.OK(c, "Your stories retrieved successfully", stories)
}

// GetTrendingStories handles the request to get trending stories
func (h *ListHandler) GetTrendingStories(c *gin.Context) {
	// Get authenticated user ID (may be nil for unauthenticated users)
	var userID primitive.ObjectID
	if id, exists := c.Get("userID"); exists {
		userID = id.(primitive.ObjectID)
	}

	// Get location parameter (optional)
	location := c.DefaultQuery("location", "")

	// Get pagination parameters
	limit, offset := response.GetPaginationParams(c)

	// Get trending stories
	stories, total, err := h.storyService.GetTrendingStories(c.Request.Context(), location, userID, limit, offset)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve trending stories", err)
		return
	}

	// Return paginated response
	response.SuccessWithPagination(c, http.StatusOK, "Trending stories retrieved successfully", stories, limit, offset, total)
}
