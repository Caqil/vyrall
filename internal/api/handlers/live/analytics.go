package live

import (
	"net/http"
	"time"

	"github.com/Caqil/vyrall/internal/utils/response"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// AnalyticsService defines the interface for live stream analytics operations
type AnalyticsService interface {
	GetStreamAnalytics(ctx context.Context, streamID primitive.ObjectID) (map[string]interface{}, error)
	GetViewerStats(ctx context.Context, streamID primitive.ObjectID) (map[string]interface{}, error)
	GetEngagementMetrics(ctx context.Context, streamID primitive.ObjectID) (map[string]interface{}, error)
	GetStreamPerformance(ctx context.Context, streamID primitive.ObjectID, interval string) ([]map[string]interface{}, error)
	GetStreamsByUserAnalytics(ctx context.Context, userID primitive.ObjectID, startDate, endDate time.Time) ([]map[string]interface{}, error)
}

// LiveStreamService defines the interface for live stream operations
type LiveStreamService interface {
	GetStreamByID(ctx context.Context, id primitive.ObjectID) (*models.LiveStream, error)
	IsStreamHost(ctx context.Context, streamID, userID primitive.ObjectID) (bool, error)
	IsStreamModerator(ctx context.Context, streamID, userID primitive.ObjectID) (bool, error)
}

// GetStreamAnalytics returns analytics data for a specific live stream
func GetStreamAnalytics(c *gin.Context) {
	// Get stream ID from URL parameter
	streamIDStr := c.Param("id")
	streamID, err := primitive.ObjectIDFromHex(streamIDStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid stream ID format", err)
		return
	}

	// Get the authenticated user's ID
	userID, exists := c.Get("userID")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	// Get the live stream service
	liveStreamService := c.MustGet("liveStreamService").(LiveStreamService)

	// Check if the stream exists
	stream, err := liveStreamService.GetStreamByID(c.Request.Context(), streamID)
	if err != nil {
		response.Error(c, http.StatusNotFound, "Stream not found", err)
		return
	}

	// Check if user is authorized to view analytics (must be host or moderator)
	isHost, err := liveStreamService.IsStreamHost(c.Request.Context(), streamID, userID.(primitive.ObjectID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to check if user is host", err)
		return
	}

	isModerator, err := liveStreamService.IsStreamModerator(c.Request.Context(), streamID, userID.(primitive.ObjectID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to check if user is moderator", err)
		return
	}

	if !isHost && !isModerator {
		response.Error(c, http.StatusForbidden, "Only the stream host or moderators can view analytics", nil)
		return
	}

	// Get the analytics service
	analyticsService := c.MustGet("analyticsService").(AnalyticsService)

	// Get the stream analytics
	analytics, err := analyticsService.GetStreamAnalytics(c.Request.Context(), streamID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve stream analytics", err)
		return
	}

	response.Success(c, http.StatusOK, "Stream analytics retrieved successfully", analytics)
}

// GetViewerStats returns viewer statistics for a live stream
func GetViewerStats(c *gin.Context) {
	// Get stream ID from URL parameter
	streamIDStr := c.Param("id")
	streamID, err := primitive.ObjectIDFromHex(streamIDStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid stream ID format", err)
		return
	}

	// Get the authenticated user's ID
	userID, exists := c.Get("userID")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	// Get the live stream service
	liveStreamService := c.MustGet("liveStreamService").(LiveStreamService)

	// Check if the stream exists
	_, err = liveStreamService.GetStreamByID(c.Request.Context(), streamID)
	if err != nil {
		response.Error(c, http.StatusNotFound, "Stream not found", err)
		return
	}

	// Check if user is authorized to view viewer stats
	isHost, err := liveStreamService.IsStreamHost(c.Request.Context(), streamID, userID.(primitive.ObjectID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to check if user is host", err)
		return
	}

	isModerator, err := liveStreamService.IsStreamModerator(c.Request.Context(), streamID, userID.(primitive.ObjectID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to check if user is moderator", err)
		return
	}

	if !isHost && !isModerator {
		response.Error(c, http.StatusForbidden, "Only the stream host or moderators can view viewer statistics", nil)
		return
	}

	// Get the analytics service
	analyticsService := c.MustGet("analyticsService").(AnalyticsService)

	// Get viewer statistics
	viewerStats, err := analyticsService.GetViewerStats(c.Request.Context(), streamID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve viewer statistics", err)
		return
	}

	response.Success(c, http.StatusOK, "Viewer statistics retrieved successfully", viewerStats)
}

// GetEngagementMetrics returns engagement metrics for a live stream
func GetEngagementMetrics(c *gin.Context) {
	// Get stream ID from URL parameter
	streamIDStr := c.Param("id")
	streamID, err := primitive.ObjectIDFromHex(streamIDStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid stream ID format", err)
		return
	}

	// Get the authenticated user's ID
	userID, exists := c.Get("userID")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	// Get the live stream service
	liveStreamService := c.MustGet("liveStreamService").(LiveStreamService)

	// Check if the stream exists
	_, err = liveStreamService.GetStreamByID(c.Request.Context(), streamID)
	if err != nil {
		response.Error(c, http.StatusNotFound, "Stream not found", err)
		return
	}

	// Check if user is authorized to view engagement metrics
	isHost, err := liveStreamService.IsStreamHost(c.Request.Context(), streamID, userID.(primitive.ObjectID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to check if user is host", err)
		return
	}

	if !isHost {
		response.Error(c, http.StatusForbidden, "Only the stream host can view engagement metrics", nil)
		return
	}

	// Get the analytics service
	analyticsService := c.MustGet("analyticsService").(AnalyticsService)

	// Get engagement metrics
	metrics, err := analyticsService.GetEngagementMetrics(c.Request.Context(), streamID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve engagement metrics", err)
		return
	}

	response.Success(c, http.StatusOK, "Engagement metrics retrieved successfully", metrics)
}

// GetStreamPerformance returns performance data over time for a live stream
func GetStreamPerformance(c *gin.Context) {
	// Get stream ID from URL parameter
	streamIDStr := c.Param("id")
	streamID, err := primitive.ObjectIDFromHex(streamIDStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid stream ID format", err)
		return
	}

	// Get query parameters
	interval := c.DefaultQuery("interval", "minute") // minute, hour, day

	// Get the authenticated user's ID
	userID, exists := c.Get("userID")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	// Get the live stream service
	liveStreamService := c.MustGet("liveStreamService").(LiveStreamService)

	// Check if the stream exists
	_, err = liveStreamService.GetStreamByID(c.Request.Context(), streamID)
	if err != nil {
		response.Error(c, http.StatusNotFound, "Stream not found", err)
		return
	}

	// Check if user is authorized to view stream performance
	isHost, err := liveStreamService.IsStreamHost(c.Request.Context(), streamID, userID.(primitive.ObjectID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to check if user is host", err)
		return
	}

	if !isHost {
		response.Error(c, http.StatusForbidden, "Only the stream host can view stream performance data", nil)
		return
	}

	// Validate interval
	validIntervals := map[string]bool{"minute": true, "hour": true, "day": true}
	if !validIntervals[interval] {
		response.Error(c, http.StatusBadRequest, "Invalid interval. Must be 'minute', 'hour', or 'day'", nil)
		return
	}

	// Get the analytics service
	analyticsService := c.MustGet("analyticsService").(AnalyticsService)

	// Get performance data
	performance, err := analyticsService.GetStreamPerformance(c.Request.Context(), streamID, interval)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve stream performance data", err)
		return
	}

	response.Success(c, http.StatusOK, "Stream performance data retrieved successfully", performance)
}

// GetUserStreamAnalytics returns analytics for all streams by a user in a time period
func GetUserStreamAnalytics(c *gin.Context) {
	// Get user ID from URL parameter
	userIDStr := c.Param("user_id")
	var targetUserID primitive.ObjectID
	var err error

	if userIDStr != "" {
		targetUserID, err = primitive.ObjectIDFromHex(userIDStr)
		if err != nil {
			response.Error(c, http.StatusBadRequest, "Invalid user ID format", err)
			return
		}
	} else {
		// Use the authenticated user's ID if no user ID provided
		currentUserID, exists := c.Get("userID")
		if !exists {
			response.Error(c, http.StatusUnauthorized, "User not authenticated", nil)
			return
		}
		targetUserID = currentUserID.(primitive.ObjectID)
	}

	// Get query parameters
	startDateStr := c.DefaultQuery("start_date", "")
	endDateStr := c.DefaultQuery("end_date", "")

	// Parse dates
	var startDate, endDate time.Time
	if startDateStr != "" {
		startDate, err = time.Parse("2006-01-02", startDateStr)
		if err != nil {
			response.Error(c, http.StatusBadRequest, "Invalid start date format", err)
			return
		}
	} else {
		// Default to 30 days ago
		startDate = time.Now().AddDate(0, 0, -30)
	}

	if endDateStr != "" {
		endDate, err = time.Parse("2006-01-02", endDateStr)
		if err != nil {
			response.Error(c, http.StatusBadRequest, "Invalid end date format", err)
			return
		}
		// Include the entire end date
		endDate = endDate.Add(24 * time.Hour).Add(-time.Second)
	} else {
		// Default to now
		endDate = time.Now()
	}

	// Get the authenticated user's ID
	currentUserID, exists := c.Get("userID")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	// Check if user is authorized to view these analytics
	if targetUserID != currentUserID.(primitive.ObjectID) {
		// Check if current user is an admin
		userService := c.MustGet("userService").(UserService)
		isAdmin, err := userService.IsAdmin(c.Request.Context(), currentUserID.(primitive.ObjectID))
		if err != nil {
			response.Error(c, http.StatusInternalServerError, "Failed to check user permissions", err)
			return
		}

		if !isAdmin {
			response.Error(c, http.StatusForbidden, "You can only view your own stream analytics", nil)
			return
		}
	}

	// Get the analytics service
	analyticsService := c.MustGet("analyticsService").(AnalyticsService)

	// Get user stream analytics
	analytics, err := analyticsService.GetStreamsByUserAnalytics(c.Request.Context(), targetUserID, startDate, endDate)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve user stream analytics", err)
		return
	}

	response.Success(c, http.StatusOK, "User stream analytics retrieved successfully", analytics)
}
