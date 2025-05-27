package stories

import (
	"net/http"

	"github.com/Caqil/vyrall/internal/services/story"
	"github.com/Caqil/vyrall/internal/utils/response"
	"github.com/Caqil/vyrall/internal/utils/validation"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// AnalyticsHandler handles story analytics operations
type AnalyticsHandler struct {
	storyService *story.Service
}

// NewAnalyticsHandler creates a new analytics handler
func NewAnalyticsHandler(storyService *story.Service) *AnalyticsHandler {
	return &AnalyticsHandler{
		storyService: storyService,
	}
}

// GetStoryAnalytics handles the request to get analytics for a story
func (h *AnalyticsHandler) GetStoryAnalytics(c *gin.Context) {
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

	// Get story analytics
	analytics, err := h.storyService.GetStoryAnalytics(c.Request.Context(), storyID, userID.(primitive.ObjectID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get story analytics", err)
		return
	}

	// Return success response
	response.OK(c, "Story analytics retrieved successfully", analytics)
}

// GetStoriesAnalytics handles the request to get analytics for all user stories
func (h *AnalyticsHandler) GetStoriesAnalytics(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Get time range parameters
	timeRange := c.DefaultQuery("time_range", "week") // day, week, month, all

	// Get stories analytics
	analytics, err := h.storyService.GetStoriesAnalytics(c.Request.Context(), userID.(primitive.ObjectID), timeRange)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get stories analytics", err)
		return
	}

	// Return success response
	response.OK(c, "Stories analytics retrieved successfully", analytics)
}

// GetStoryReachMetrics handles the request to get reach metrics for a story
func (h *AnalyticsHandler) GetStoryReachMetrics(c *gin.Context) {
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

	// Get story reach metrics
	metrics, err := h.storyService.GetStoryReachMetrics(c.Request.Context(), storyID, userID.(primitive.ObjectID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get story reach metrics", err)
		return
	}

	// Return success response
	response.OK(c, "Story reach metrics retrieved successfully", metrics)
}

// GetStoryEngagementMetrics handles the request to get engagement metrics for a story
func (h *AnalyticsHandler) GetStoryEngagementMetrics(c *gin.Context) {
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

	// Get story engagement metrics
	metrics, err := h.storyService.GetStoryEngagementMetrics(c.Request.Context(), storyID, userID.(primitive.ObjectID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get story engagement metrics", err)
		return
	}

	// Return success response
	response.OK(c, "Story engagement metrics retrieved successfully", metrics)
}
