package admin

import (
	"net/http"

	"github.com/Caqil/vyrall/internal/utils/response"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ContentService interface for content management operations
type ContentService interface {
	GetContentStats() (map[string]interface{}, error)
	GetContentByType(contentType string, status string, limit, offset int) ([]map[string]interface{}, int, error)
	GetContentByID(contentType string, id primitive.ObjectID) (interface{}, error)
	UpdateContentStatus(contentType string, id primitive.ObjectID, status string) error
	DeleteContent(contentType string, id primitive.ObjectID) error
	RestoreContent(contentType string, id primitive.ObjectID) error
	GetFlaggedContent(contentType string, limit, offset int) ([]map[string]interface{}, int, error)
	GetTrendingContent(contentType string, limit int) ([]map[string]interface{}, error)
	GetContentHistory(contentType string, id primitive.ObjectID) ([]map[string]interface{}, error)
	GetTags() ([]string, error)
	ManageTags(action string, tag string) error
}

// GetContentOverview returns content statistics for the admin dashboard
func GetContentOverview(c *gin.Context) {
	contentService := c.MustGet("contentService").(ContentService)

	stats, err := contentService.GetContentStats()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve content statistics", err)
		return
	}

	response.Success(c, http.StatusOK, "Content statistics retrieved successfully", stats)
}

// ListContent returns a list of content by type and status
func ListContent(c *gin.Context) {
	contentType := c.DefaultQuery("type", "post")
	status := c.DefaultQuery("status", "active")
	limit, offset := getPaginationParams(c)

	contentService := c.MustGet("contentService").(ContentService)

	// Validate content type
	validTypes := map[string]bool{
		"post":        true,
		"comment":     true,
		"story":       true,
		"message":     true,
		"event":       true,
		"group":       true,
		"live_stream": true,
	}

	if !validTypes[contentType] {
		response.Error(c, http.StatusBadRequest, "Invalid content type", nil)
		return
	}

	// Validate status
	validStatuses := map[string]bool{
		"active":    true,
		"hidden":    true,
		"deleted":   true,
		"flagged":   true,
		"reported":  true,
		"all":       true,
		"scheduled": true,
		"archived":  true,
	}

	if !validStatuses[status] {
		response.Error(c, http.StatusBadRequest, "Invalid status", nil)
		return
	}

	content, total, err := contentService.GetContentByType(contentType, status, limit, offset)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve content", err)
		return
	}

	response.SuccessWithPagination(c, http.StatusOK, "Content retrieved successfully", content, limit, offset, total)
}

// GetContentDetail returns details about a specific content
func GetContentDetail(c *gin.Context) {
	contentType := c.Param("type")
	idStr := c.Param("id")

	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid content ID", err)
		return
	}

	contentService := c.MustGet("contentService").(ContentService)

	// Validate content type
	validTypes := map[string]bool{
		"post":        true,
		"comment":     true,
		"story":       true,
		"message":     true,
		"event":       true,
		"group":       true,
		"live_stream": true,
	}

	if !validTypes[contentType] {
		response.Error(c, http.StatusBadRequest, "Invalid content type", nil)
		return
	}

	content, err := contentService.GetContentByID(contentType, id)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve content details", err)
		return
	}

	response.Success(c, http.StatusOK, "Content details retrieved successfully", content)
}

// UpdateContentStatus updates the status of content (hide, restore, etc.)
func UpdateContentStatus(c *gin.Context) {
	contentType := c.Param("type")
	idStr := c.Param("id")

	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid content ID", err)
		return
	}

	var req struct {
		Status string `json:"status" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	contentService := c.MustGet("contentService").(ContentService)

	// Validate content type
	validTypes := map[string]bool{
		"post":        true,
		"comment":     true,
		"story":       true,
		"message":     true,
		"event":       true,
		"group":       true,
		"live_stream": true,
	}

	if !validTypes[contentType] {
		response.Error(c, http.StatusBadRequest, "Invalid content type", nil)
		return
	}

	// Validate status
	validStatuses := map[string]bool{
		"active":   true,
		"hidden":   true,
		"archived": true,
		"featured": true,
	}

	if !validStatuses[req.Status] {
		response.Error(c, http.StatusBadRequest, "Invalid status", nil)
		return
	}

	if err := contentService.UpdateContentStatus(contentType, id, req.Status); err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to update content status", err)
		return
	}

	response.Success(c, http.StatusOK, "Content status updated successfully", nil)
}

// DeleteContent permanently deletes content
func DeleteContent(c *gin.Context) {
	contentType := c.Param("type")
	idStr := c.Param("id")

	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid content ID", err)
		return
	}

	contentService := c.MustGet("contentService").(ContentService)

	// Validate content type
	validTypes := map[string]bool{
		"post":        true,
		"comment":     true,
		"story":       true,
		"message":     true,
		"event":       true,
		"group":       true,
		"live_stream": true,
	}

	if !validTypes[contentType] {
		response.Error(c, http.StatusBadRequest, "Invalid content type", nil)
		return
	}

	if err := contentService.DeleteContent(contentType, id); err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to delete content", err)
		return
	}

	response.Success(c, http.StatusOK, "Content deleted successfully", nil)
}

// RestoreContent restores deleted content
func RestoreContent(c *gin.Context) {
	contentType := c.Param("type")
	idStr := c.Param("id")

	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid content ID", err)
		return
	}

	contentService := c.MustGet("contentService").(ContentService)

	// Validate content type
	validTypes := map[string]bool{
		"post":        true,
		"comment":     true,
		"story":       true,
		"message":     true,
		"event":       true,
		"group":       true,
		"live_stream": true,
	}

	if !validTypes[contentType] {
		response.Error(c, http.StatusBadRequest, "Invalid content type", nil)
		return
	}

	if err := contentService.RestoreContent(contentType, id); err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to restore content", err)
		return
	}

	response.Success(c, http.StatusOK, "Content restored successfully", nil)
}

// GetFlaggedContent returns content that has been flagged as inappropriate
func GetFlaggedContent(c *gin.Context) {
	contentType := c.DefaultQuery("type", "all")
	limit, offset := getPaginationParams(c)

	contentService := c.MustGet("contentService").(ContentService)

	// Validate content type
	if contentType != "all" {
		validTypes := map[string]bool{
			"post":        true,
			"comment":     true,
			"story":       true,
			"message":     true,
			"event":       true,
			"group":       true,
			"live_stream": true,
		}

		if !validTypes[contentType] {
			response.Error(c, http.StatusBadRequest, "Invalid content type", nil)
			return
		}
	}

	content, total, err := contentService.GetFlaggedContent(contentType, limit, offset)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve flagged content", err)
		return
	}

	response.SuccessWithPagination(c, http.StatusOK, "Flagged content retrieved successfully", content, limit, offset, total)
}

// GetTrendingContent returns trending content
func GetTrendingContent(c *gin.Context) {
	contentType := c.DefaultQuery("type", "post")
	limitStr := c.DefaultQuery("limit", "10")

	limit, err := parseInt(limitStr)
	if err != nil || limit < 1 {
		limit = 10
	}

	contentService := c.MustGet("contentService").(ContentService)

	// Validate content type
	validTypes := map[string]bool{
		"post":        true,
		"hashtag":     true,
		"live_stream": true,
	}

	if !validTypes[contentType] {
		response.Error(c, http.StatusBadRequest, "Invalid content type", nil)
		return
	}

	content, err := contentService.GetTrendingContent(contentType, limit)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve trending content", err)
		return
	}

	response.Success(c, http.StatusOK, "Trending content retrieved successfully", content)
}

// GetContentHistory returns the edit history of a piece of content
func GetContentHistory(c *gin.Context) {
	contentType := c.Param("type")
	idStr := c.Param("id")

	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid content ID", err)
		return
	}

	contentService := c.MustGet("contentService").(ContentService)

	// Validate content type
	validTypes := map[string]bool{
		"post":    true,
		"comment": true,
	}

	if !validTypes[contentType] {
		response.Error(c, http.StatusBadRequest, "Invalid content type or history not available", nil)
		return
	}

	history, err := contentService.GetContentHistory(contentType, id)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve content history", err)
		return
	}

	response.Success(c, http.StatusOK, "Content history retrieved successfully", history)
}

// GetTagsList returns all hashtags/tags in the system
func GetTagsList(c *gin.Context) {
	contentService := c.MustGet("contentService").(ContentService)

	tags, err := contentService.GetTags()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve tags", err)
		return
	}

	response.Success(c, http.StatusOK, "Tags retrieved successfully", tags)
}

// ManageTag handles tag management (block, unblock, delete)
func ManageTag(c *gin.Context) {
	var req struct {
		Action string `json:"action" binding:"required"`
		Tag    string `json:"tag" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	contentService := c.MustGet("contentService").(ContentService)

	// Validate action
	validActions := map[string]bool{
		"block":   true,
		"unblock": true,
		"delete":  true,
	}

	if !validActions[req.Action] {
		response.Error(c, http.StatusBadRequest, "Invalid action", nil)
		return
	}

	if err := contentService.ManageTags(req.Action, req.Tag); err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to manage tag", err)
		return
	}

	response.Success(c, http.StatusOK, "Tag managed successfully", nil)
}
