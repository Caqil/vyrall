package posts

import (
	"net/http"
	"time"

	"github.com/Caqil/vyrall/internal/services/post"
	"github.com/Caqil/vyrall/internal/utils/response"
	"github.com/Caqil/vyrall/internal/utils/validation"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// AnalyticsHandler handles post analytics operations
type AnalyticsHandler struct {
	postService *post.Service
}

// NewAnalyticsHandler creates a new analytics handler
func NewAnalyticsHandler(postService *post.Service) *AnalyticsHandler {
	return &AnalyticsHandler{
		postService: postService,
	}
}

// GetPostAnalytics handles the request to get analytics for a post
func (h *AnalyticsHandler) GetPostAnalytics(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Get post ID from URL parameter
	postIDStr := c.Param("id")
	if !validation.IsValidObjectID(postIDStr) {
		response.ValidationError(c, "Invalid post ID", nil)
		return
	}
	postID, _ := primitive.ObjectIDFromHex(postIDStr)

	// Get time range parameters
	timeRange := c.DefaultQuery("time_range", "all") // all, day, week, month, year

	// Get the post analytics
	analytics, err := h.postService.GetPostAnalytics(c.Request.Context(), postID, userID.(primitive.ObjectID), timeRange)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get post analytics", err)
		return
	}

	// Return success response
	response.OK(c, "Post analytics retrieved successfully", analytics)
}

// GetPostEngagement handles the request to get engagement metrics for a post
func (h *AnalyticsHandler) GetPostEngagement(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Get post ID from URL parameter
	postIDStr := c.Param("id")
	if !validation.IsValidObjectID(postIDStr) {
		response.ValidationError(c, "Invalid post ID", nil)
		return
	}
	postID, _ := primitive.ObjectIDFromHex(postIDStr)

	// Get the post engagement metrics
	engagement, err := h.postService.GetPostEngagement(c.Request.Context(), postID, userID.(primitive.ObjectID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get post engagement metrics", err)
		return
	}

	// Return success response
	response.OK(c, "Post engagement metrics retrieved successfully", engagement)
}

// GetPostReach handles the request to get reach metrics for a post
func (h *AnalyticsHandler) GetPostReach(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Get post ID from URL parameter
	postIDStr := c.Param("id")
	if !validation.IsValidObjectID(postIDStr) {
		response.ValidationError(c, "Invalid post ID", nil)
		return
	}
	postID, _ := primitive.ObjectIDFromHex(postIDStr)

	// Get time range parameters
	timeRange := c.DefaultQuery("time_range", "all") // all, day, week, month, year

	// Get the post reach metrics
	reach, err := h.postService.GetPostReach(c.Request.Context(), postID, userID.(primitive.ObjectID), timeRange)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get post reach metrics", err)
		return
	}

	// Return success response
	response.OK(c, "Post reach metrics retrieved successfully", reach)
}

// GetPostDemographics handles the request to get audience demographics for a post
func (h *AnalyticsHandler) GetPostDemographics(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Get post ID from URL parameter
	postIDStr := c.Param("id")
	if !validation.IsValidObjectID(postIDStr) {
		response.ValidationError(c, "Invalid post ID", nil)
		return
	}
	postID, _ := primitive.ObjectIDFromHex(postIDStr)

	// Get the post demographics
	demographics, err := h.postService.GetPostDemographics(c.Request.Context(), postID, userID.(primitive.ObjectID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get post demographics", err)
		return
	}

	// Return success response
	response.OK(c, "Post demographics retrieved successfully", demographics)
}

// GetUserPostsAnalytics handles the request to get analytics for a user's posts
func (h *AnalyticsHandler) GetUserPostsAnalytics(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Get query parameters
	timeRange := c.DefaultQuery("time_range", "month") // day, week, month, year, all
	metric := c.DefaultQuery("metric", "all")          // all, engagement, reach, views, likes, comments, shares

	// Parse start and end dates if provided
	var startDate, endDate time.Time
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	if startDateStr != "" {
		if parsedDate, err := time.Parse(time.RFC3339, startDateStr); err == nil {
			startDate = parsedDate
		}
	}

	if endDateStr != "" {
		if parsedDate, err := time.Parse(time.RFC3339, endDateStr); err == nil {
			endDate = parsedDate
		} else {
			endDate = time.Now()
		}
	}

	// Get analytics for user's posts
	analytics, err := h.postService.GetUserPostsAnalytics(
		c.Request.Context(),
		userID.(primitive.ObjectID),
		timeRange,
		metric,
		startDate,
		endDate,
	)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get user posts analytics", err)
		return
	}

	// Return success response
	response.OK(c, "User posts analytics retrieved successfully", analytics)
}

// GetTopPerformingPosts handles the request to get a user's top performing posts
func (h *AnalyticsHandler) GetTopPerformingPosts(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Get query parameters
	timeRange := c.DefaultQuery("time_range", "month") // day, week, month, year, all
	metric := c.DefaultQuery("metric", "engagement")   // engagement, reach, views, likes, comments, shares

	// Get pagination parameters
	limit, offset := response.GetPaginationParams(c)

	// Get top performing posts
	posts, total, err := h.postService.GetTopPerformingPosts(
		c.Request.Context(),
		userID.(primitive.ObjectID),
		timeRange,
		metric,
		limit,
		offset,
	)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get top performing posts", err)
		return
	}

	// Return paginated response
	response.SuccessWithPagination(c, http.StatusOK, "Top performing posts retrieved successfully", posts, limit, offset, total)
}

// ExportPostAnalytics handles the request to export post analytics
func (h *AnalyticsHandler) ExportPostAnalytics(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Get post ID from URL parameter
	postIDStr := c.Param("id")
	if !validation.IsValidObjectID(postIDStr) {
		response.ValidationError(c, "Invalid post ID", nil)
		return
	}
	postID, _ := primitive.ObjectIDFromHex(postIDStr)

	// Get query parameters
	format := c.DefaultQuery("format", "csv")        // csv, json, pdf
	timeRange := c.DefaultQuery("time_range", "all") // all, day, week, month, year

	// Export post analytics
	exportData, contentType, filename, err := h.postService.ExportPostAnalytics(
		c.Request.Context(),
		postID,
		userID.(primitive.ObjectID),
		format,
		timeRange,
	)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to export post analytics", err)
		return
	}

	// Set headers for file download
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Header("Content-Type", contentType)

	// Send the file data
	c.Data(http.StatusOK, contentType, exportData)
}
