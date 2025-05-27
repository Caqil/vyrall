package users

import (
	"net/http"

	"github.com/Caqil/vyrall/internal/services/user"
	"github.com/Caqil/vyrall/internal/utils/response"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// AnalyticsHandler handles user analytics operations
type AnalyticsHandler struct {
	userService *user.Service
}

// NewAnalyticsHandler creates a new analytics handler
func NewAnalyticsHandler(userService *user.Service) *AnalyticsHandler {
	return &AnalyticsHandler{
		userService: userService,
	}
}

// GetUserAnalytics handles the request to get analytics for a user
func (h *AnalyticsHandler) GetUserAnalytics(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Get time range parameter
	timeRange := c.DefaultQuery("time_range", "month") // day, week, month, year, all

	// Get user analytics
	analytics, err := h.userService.GetUserAnalytics(c.Request.Context(), userID.(primitive.ObjectID), timeRange)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get user analytics", err)
		return
	}

	// Return success response
	response.OK(c, "User analytics retrieved successfully", analytics)
}

// GetProfileViewers handles the request to get profile viewers
func (h *AnalyticsHandler) GetProfileViewers(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Get time range parameter
	timeRange := c.DefaultQuery("time_range", "week") // day, week, month, all

	// Get pagination parameters
	limit, offset := response.GetPaginationParams(c)

	// Get profile viewers
	viewers, total, err := h.userService.GetProfileViewers(c.Request.Context(), userID.(primitive.ObjectID), timeRange, limit, offset)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get profile viewers", err)
		return
	}

	// Return paginated response
	response.SuccessWithPagination(c, http.StatusOK, "Profile viewers retrieved successfully", viewers, limit, offset, total)
}

// GetEngagementMetrics handles the request to get engagement metrics
func (h *AnalyticsHandler) GetEngagementMetrics(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Get time range parameter
	timeRange := c.DefaultQuery("time_range", "month") // day, week, month, year, all

	// Get engagement metrics
	metrics, err := h.userService.GetEngagementMetrics(c.Request.Context(), userID.(primitive.ObjectID), timeRange)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get engagement metrics", err)
		return
	}

	// Return success response
	response.OK(c, "Engagement metrics retrieved successfully", metrics)
}

// GetFollowerGrowth handles the request to get follower growth metrics
func (h *AnalyticsHandler) GetFollowerGrowth(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Get time range parameter
	timeRange := c.DefaultQuery("time_range", "month") // week, month, year, all

	// Get follower growth metrics
	growth, err := h.userService.GetFollowerGrowth(c.Request.Context(), userID.(primitive.ObjectID), timeRange)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get follower growth metrics", err)
		return
	}

	// Return success response
	response.OK(c, "Follower growth metrics retrieved successfully", growth)
}
