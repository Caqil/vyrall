package stories

import (
	"net/http"

	"github.com/Caqil/vyrall/internal/services/story"
	"github.com/Caqil/vyrall/internal/utils/response"
	"github.com/Caqil/vyrall/internal/utils/validation"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// DeleteHandler handles story deletion operations
type DeleteHandler struct {
	storyService *story.Service
}

// NewDeleteHandler creates a new delete handler
func NewDeleteHandler(storyService *story.Service) *DeleteHandler {
	return &DeleteHandler{
		storyService: storyService,
	}
}

// DeleteStory handles the request to delete a story
func (h *DeleteHandler) DeleteStory(c *gin.Context) {
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

	// Delete the story
	err := h.storyService.DeleteStory(c.Request.Context(), storyID, userID.(primitive.ObjectID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to delete story", err)
		return
	}

	// Return success response
	response.OK(c, "Story deleted successfully", nil)
}

// DeleteAllStories handles the request to delete all user's active stories
func (h *DeleteHandler) DeleteAllStories(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Delete all stories
	count, err := h.storyService.DeleteAllStories(c.Request.Context(), userID.(primitive.ObjectID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to delete stories", err)
		return
	}

	// Return success response
	response.OK(c, "Stories deleted successfully", gin.H{
		"deleted_count": count,
	})
}

// ArchiveStory handles the request to archive a story
func (h *DeleteHandler) ArchiveStory(c *gin.Context) {
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

	// Archive the story
	err := h.storyService.ArchiveStory(c.Request.Context(), storyID, userID.(primitive.ObjectID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to archive story", err)
		return
	}

	// Return success response
	response.OK(c, "Story archived successfully", nil)
}

// GetArchivedStories handles the request to get user's archived stories
func (h *DeleteHandler) GetArchivedStories(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Get pagination parameters
	limit, offset := response.GetPaginationParams(c)

	// Get archived stories
	stories, total, err := h.storyService.GetArchivedStories(c.Request.Context(), userID.(primitive.ObjectID), limit, offset)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve archived stories", err)
		return
	}

	// Return paginated response
	response.SuccessWithPagination(c, http.StatusOK, "Archived stories retrieved successfully", stories, limit, offset, total)
}

// UnarchiveStory handles the request to unarchive a story
func (h *DeleteHandler) UnarchiveStory(c *gin.Context) {
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

	// Unarchive the story
	err := h.storyService.UnarchiveStory(c.Request.Context(), storyID, userID.(primitive.ObjectID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to unarchive story", err)
		return
	}

	// Return success response
	response.OK(c, "Story unarchived successfully", nil)
}
