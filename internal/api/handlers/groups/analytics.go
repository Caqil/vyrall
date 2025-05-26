package groups

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/Caqil/vyrall/internal/models"
	"github.com/Caqil/vyrall/internal/utils/response"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// AnalyticsService defines the interface for group analytics operations
type AnalyticsService interface {
	GetGroupAnalytics(ctx context.Context, groupID primitive.ObjectID) (*models.GroupAnalytics, error)
	GetGroupMemberStats(ctx context.Context, groupID primitive.ObjectID) (map[string]interface{}, error)
	GetGroupContentStats(ctx context.Context, groupID primitive.ObjectID) (map[string]interface{}, error)
	GetGroupActivityOverTime(ctx context.Context, groupID primitive.ObjectID, startDate, endDate time.Time) ([]map[string]interface{}, error)
	GetTopContributors(ctx context.Context, groupID primitive.ObjectID, limit int) ([]map[string]interface{}, error)
	GetGroupGrowthRate(ctx context.Context, groupID primitive.ObjectID) (float64, error)
	GetPopularTopics(ctx context.Context, groupID primitive.ObjectID, limit int) ([]map[string]interface{}, error)
}

// GetGroupAnalytics retrieves analytics data for a specific group
func GetGroupAnalytics(c *gin.Context) {
	// Get group ID from URL parameter
	groupIDStr := c.Param("id")
	groupID, err := primitive.ObjectIDFromHex(groupIDStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid group ID format", err)
		return
	}

	// Get the authenticated user's ID
	userID, exists := c.Get("userID")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	// Get the group service
	groupService := c.MustGet("groupService").(GroupService)

	// Check if the group exists
	group, err := groupService.GetGroupByID(c.Request.Context(), groupID)
	if err != nil {
		response.Error(c, http.StatusNotFound, "Group not found", err)
		return
	}

	// Check if the user has permission to view analytics (admin or moderator)
	hasPermission, err := groupService.HasModeratorPermission(c.Request.Context(), groupID, userID.(primitive.ObjectID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to check user permissions", err)
		return
	}

	if !hasPermission {
		response.Error(c, http.StatusForbidden, "Only group admins and moderators can view analytics", nil)
		return
	}

	// Get the analytics service
	analyticsService := c.MustGet("analyticsService").(AnalyticsService)

	// Get group analytics
	analytics, err := analyticsService.GetGroupAnalytics(c.Request.Context(), groupID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve group analytics", err)
		return
	}

	response.Success(c, http.StatusOK, "Group analytics retrieved successfully", analytics)
}

// GetGroupMemberStats retrieves statistics about group members
func GetGroupMemberStats(c *gin.Context) {
	// Get group ID from URL parameter
	groupIDStr := c.Param("id")
	groupID, err := primitive.ObjectIDFromHex(groupIDStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid group ID format", err)
		return
	}

	// Get the authenticated user's ID
	userID, exists := c.Get("userID")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	// Get the group service
	groupService := c.MustGet("groupService").(GroupService)

	// Check if the group exists
	_, err = groupService.GetGroupByID(c.Request.Context(), groupID)
	if err != nil {
		response.Error(c, http.StatusNotFound, "Group not found", err)
		return
	}

	// Check if the user has permission to view analytics
	hasPermission, err := groupService.HasModeratorPermission(c.Request.Context(), groupID, userID.(primitive.ObjectID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to check user permissions", err)
		return
	}

	if !hasPermission {
		response.Error(c, http.StatusForbidden, "Only group admins and moderators can view analytics", nil)
		return
	}

	// Get the analytics service
	analyticsService := c.MustGet("analyticsService").(AnalyticsService)

	// Get member statistics
	stats, err := analyticsService.GetGroupMemberStats(c.Request.Context(), groupID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve member statistics", err)
		return
	}

	response.Success(c, http.StatusOK, "Member statistics retrieved successfully", stats)
}

// GetGroupContentStats retrieves statistics about group content
func GetGroupContentStats(c *gin.Context) {
	// Get group ID from URL parameter
	groupIDStr := c.Param("id")
	groupID, err := primitive.ObjectIDFromHex(groupIDStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid group ID format", err)
		return
	}

	// Get the authenticated user's ID
	userID, exists := c.Get("userID")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	// Get the group service
	groupService := c.MustGet("groupService").(GroupService)

	// Check if the group exists
	_, err = groupService.GetGroupByID(c.Request.Context(), groupID)
	if err != nil {
		response.Error(c, http.StatusNotFound, "Group not found", err)
		return
	}

	// Check if the user has permission to view analytics
	hasPermission, err := groupService.HasModeratorPermission(c.Request.Context(), groupID, userID.(primitive.ObjectID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to check user permissions", err)
		return
	}

	if !hasPermission {
		response.Error(c, http.StatusForbidden, "Only group admins and moderators can view analytics", nil)
		return
	}

	// Get the analytics service
	analyticsService := c.MustGet("analyticsService").(AnalyticsService)

	// Get content statistics
	stats, err := analyticsService.GetGroupContentStats(c.Request.Context(), groupID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve content statistics", err)
		return
	}

	response.Success(c, http.StatusOK, "Content statistics retrieved successfully", stats)
}

// GetGroupActivityOverTime retrieves group activity data over a specific time period
func GetGroupActivityOverTime(c *gin.Context) {
	// Get group ID from URL parameter
	groupIDStr := c.Param("id")
	groupID, err := primitive.ObjectIDFromHex(groupIDStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid group ID format", err)
		return
	}

	// Get the authenticated user's ID
	userID, exists := c.Get("userID")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "User not authenticated", nil)
		return
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
		endDate = endDate.Add(24 * time.Hour)
	} else {
		// Default to now
		endDate = time.Now()
	}

	// Get the group service
	groupService := c.MustGet("groupService").(GroupService)

	// Check if the group exists
	_, err = groupService.GetGroupByID(c.Request.Context(), groupID)
	if err != nil {
		response.Error(c, http.StatusNotFound, "Group not found", err)
		return
	}

	// Check if the user has permission to view analytics
	hasPermission, err := groupService.HasModeratorPermission(c.Request.Context(), groupID, userID.(primitive.ObjectID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to check user permissions", err)
		return
	}

	if !hasPermission {
		response.Error(c, http.StatusForbidden, "Only group admins and moderators can view analytics", nil)
		return
	}

	// Get the analytics service
	analyticsService := c.MustGet("analyticsService").(AnalyticsService)

	// Get activity data by time range
	activityData, err := analyticsService.GetGroupActivityOverTime(c.Request.Context(), groupID, startDate, endDate)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve activity data", err)
		return
	}

	response.Success(c, http.StatusOK, "Group activity data retrieved successfully", activityData)
}

// GetTopContributors retrieves the top contributors to a group
func GetTopContributors(c *gin.Context) {
	// Get group ID from URL parameter
	groupIDStr := c.Param("id")
	groupID, err := primitive.ObjectIDFromHex(groupIDStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid group ID format", err)
		return
	}

	// Get the authenticated user's ID
	userID, exists := c.Get("userID")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	// Get limit parameter
	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		limit = 10
	}

	// Get the group service
	groupService := c.MustGet("groupService").(GroupService)

	// Check if the group exists
	_, err = groupService.GetGroupByID(c.Request.Context(), groupID)
	if err != nil {
		response.Error(c, http.StatusNotFound, "Group not found", err)
		return
	}

	// Check if the user has permission to view analytics
	hasPermission, err := groupService.HasModeratorPermission(c.Request.Context(), groupID, userID.(primitive.ObjectID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to check user permissions", err)
		return
	}

	if !hasPermission {
		response.Error(c, http.StatusForbidden, "Only group admins and moderators can view analytics", nil)
		return
	}

	// Get the analytics service
	analyticsService := c.MustGet("analyticsService").(AnalyticsService)

	// Get top contributors
	contributors, err := analyticsService.GetTopContributors(c.Request.Context(), groupID, limit)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve top contributors", err)
		return
	}

	response.Success(c, http.StatusOK, "Top contributors retrieved successfully", contributors)
}
