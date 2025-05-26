package admin

import (
	"net/http"

	"github.com/Caqil/vyrall/internal/utils/response"
	"github.com/gin-gonic/gin"
)

// DashboardService interface for dashboard operations
type DashboardService interface {
	GetDashboardSummary() (map[string]interface{}, error)
	GetRecentActivity(limit int) ([]map[string]interface{}, error)
	GetSystemStatus() (map[string]interface{}, error)
	GetPendingModeration() (map[string]int, error)
	GetUserGrowth(period string, duration int) ([]map[string]interface{}, error)
	GetContentGrowth(period string, duration int) ([]map[string]interface{}, error)
	GetEngagementStats(period string, duration int) ([]map[string]interface{}, error)
	GetTopPerformers(category string, limit int) ([]map[string]interface{}, error)
}

// GetDashboardSummary returns a summary of platform statistics for the admin dashboard
func GetDashboardSummary(c *gin.Context) {
	dashboardService := c.MustGet("dashboardService").(DashboardService)

	summary, err := dashboardService.GetDashboardSummary()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve dashboard summary", err)
		return
	}

	response.Success(c, http.StatusOK, "Dashboard summary retrieved successfully", summary)
}

// GetRecentActivity returns recent significant activity on the platform
func GetRecentActivity(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "10")

	limit, err := parseInt(limitStr)
	if err != nil || limit < 1 {
		limit = 10
	}

	dashboardService := c.MustGet("dashboardService").(DashboardService)

	activity, err := dashboardService.GetRecentActivity(limit)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve recent activity", err)
		return
	}

	response.Success(c, http.StatusOK, "Recent activity retrieved successfully", activity)
}

// GetSystemStatus returns the current status of system components and services
func GetSystemStatus(c *gin.Context) {
	dashboardService := c.MustGet("dashboardService").(DashboardService)

	status, err := dashboardService.GetSystemStatus()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve system status", err)
		return
	}

	response.Success(c, http.StatusOK, "System status retrieved successfully", status)
}

// GetPendingModeration returns counts of items pending moderation by category
func GetPendingModeration(c *gin.Context) {
	dashboardService := c.MustGet("dashboardService").(DashboardService)

	counts, err := dashboardService.GetPendingModeration()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve pending moderation counts", err)
		return
	}

	response.Success(c, http.StatusOK, "Pending moderation counts retrieved successfully", counts)
}

// GetUserGrowth returns user growth data for a given period
func GetUserGrowth(c *gin.Context) {
	period := c.DefaultQuery("period", "daily")
	durationStr := c.DefaultQuery("duration", "30")

	duration, err := parseInt(durationStr)
	if err != nil || duration < 1 {
		duration = 30
	}

	dashboardService := c.MustGet("dashboardService").(DashboardService)

	// Validate period
	validPeriods := map[string]bool{
		"daily":   true,
		"weekly":  true,
		"monthly": true,
	}

	if !validPeriods[period] {
		response.Error(c, http.StatusBadRequest, "Invalid period. Must be daily, weekly, or monthly", nil)
		return
	}

	data, err := dashboardService.GetUserGrowth(period, duration)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve user growth data", err)
		return
	}

	response.Success(c, http.StatusOK, "User growth data retrieved successfully", data)
}

// GetContentGrowth returns content growth data for a given period
func GetContentGrowth(c *gin.Context) {
	period := c.DefaultQuery("period", "daily")
	durationStr := c.DefaultQuery("duration", "30")

	duration, err := parseInt(durationStr)
	if err != nil || duration < 1 {
		duration = 30
	}

	dashboardService := c.MustGet("dashboardService").(DashboardService)

	// Validate period
	validPeriods := map[string]bool{
		"daily":   true,
		"weekly":  true,
		"monthly": true,
	}

	if !validPeriods[period] {
		response.Error(c, http.StatusBadRequest, "Invalid period. Must be daily, weekly, or monthly", nil)
		return
	}

	data, err := dashboardService.GetContentGrowth(period, duration)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve content growth data", err)
		return
	}

	response.Success(c, http.StatusOK, "Content growth data retrieved successfully", data)
}

// GetEngagementStats returns engagement statistics for a given period
func GetEngagementStats(c *gin.Context) {
	period := c.DefaultQuery("period", "daily")
	durationStr := c.DefaultQuery("duration", "30")

	duration, err := parseInt(durationStr)
	if err != nil || duration < 1 {
		duration = 30
	}

	dashboardService := c.MustGet("dashboardService").(DashboardService)

	// Validate period
	validPeriods := map[string]bool{
		"daily":   true,
		"weekly":  true,
		"monthly": true,
	}

	if !validPeriods[period] {
		response.Error(c, http.StatusBadRequest, "Invalid period. Must be daily, weekly, or monthly", nil)
		return
	}

	data, err := dashboardService.GetEngagementStats(period, duration)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve engagement statistics", err)
		return
	}

	response.Success(c, http.StatusOK, "Engagement statistics retrieved successfully", data)
}

// GetTopPerformers returns top performing content, users, or groups
func GetTopPerformers(c *gin.Context) {
	category := c.DefaultQuery("category", "posts")
	limitStr := c.DefaultQuery("limit", "10")

	limit, err := parseInt(limitStr)
	if err != nil || limit < 1 {
		limit = 10
	}

	dashboardService := c.MustGet("dashboardService").(DashboardService)

	// Validate category
	validCategories := map[string]bool{
		"posts":        true,
		"users":        true,
		"groups":       true,
		"hashtags":     true,
		"events":       true,
		"live_streams": true,
	}

	if !validCategories[category] {
		response.Error(c, http.StatusBadRequest, "Invalid category", nil)
		return
	}

	performers, err := dashboardService.GetTopPerformers(category, limit)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve top performers", err)
		return
	}

	response.Success(c, http.StatusOK, "Top performers retrieved successfully", performers)
}
