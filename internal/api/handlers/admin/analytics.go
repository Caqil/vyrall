package admin

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Caqil/vyrall/internal/utils/response"
	"github.com/gin-gonic/gin"
)

// AnalyticsService interface for analytics operations
type AnalyticsService interface {
	GetPlatformStats(startDate, endDate time.Time) (map[string]interface{}, error)
	GetUserGrowth(period string, startDate, endDate time.Time) (map[string]interface{}, error)
	GetContentStats(contentType string, startDate, endDate time.Time) (map[string]interface{}, error)
	GetEngagementMetrics(startDate, endDate time.Time) (map[string]interface{}, error)
	GetTopContent(contentType, metric string, limit int) ([]map[string]interface{}, error)
	GetUserDemographics() (map[string]interface{}, error)
	GetUserRetention(cohortDate time.Time, periods int) (map[string]interface{}, error)
	GetActiveUserStats(period string) (map[string]interface{}, error)
}

// GetDashboardAnalytics returns high-level platform analytics for the admin dashboard
func GetDashboardAnalytics(c *gin.Context) {
	analyticsService := c.MustGet("analyticsService").(AnalyticsService)

	// Parse date parameters with defaults
	startDateStr := c.DefaultQuery("start_date", time.Now().AddDate(0, -1, 0).Format("2006-01-02"))
	endDateStr := c.DefaultQuery("end_date", time.Now().Format("2006-01-02"))

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid start date format", err)
		return
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid end date format", err)
		return
	}

	// Get platform stats
	stats, err := analyticsService.GetPlatformStats(startDate, endDate)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get platform stats", err)
		return
	}

	response.Success(c, http.StatusOK, "Dashboard analytics retrieved successfully", stats)
}

// GetUserGrowthAnalytics returns user growth metrics over time
func GetUserGrowthAnalytics(c *gin.Context) {
	analyticsService := c.MustGet("analyticsService").(AnalyticsService)

	// Parse parameters
	period := c.DefaultQuery("period", "monthly")
	startDateStr := c.DefaultQuery("start_date", time.Now().AddDate(0, -6, 0).Format("2006-01-02"))
	endDateStr := c.DefaultQuery("end_date", time.Now().Format("2006-01-02"))

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid start date format", err)
		return
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid end date format", err)
		return
	}

	// Validate period
	if period != "daily" && period != "weekly" && period != "monthly" {
		response.Error(c, http.StatusBadRequest, "Invalid period. Must be daily, weekly, or monthly", nil)
		return
	}

	// Get user growth data
	data, err := analyticsService.GetUserGrowth(period, startDate, endDate)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get user growth data", err)
		return
	}

	response.Success(c, http.StatusOK, "User growth analytics retrieved successfully", data)
}

// GetContentAnalytics returns analytics for platform content
func GetContentAnalytics(c *gin.Context) {
	analyticsService := c.MustGet("analyticsService").(AnalyticsService)

	// Parse parameters
	contentType := c.DefaultQuery("type", "post")
	startDateStr := c.DefaultQuery("start_date", time.Now().AddDate(0, -1, 0).Format("2006-01-02"))
	endDateStr := c.DefaultQuery("end_date", time.Now().Format("2006-01-02"))

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid start date format", err)
		return
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid end date format", err)
		return
	}

	// Validate content type
	validTypes := map[string]bool{"post": true, "comment": true, "story": true, "live_stream": true}
	if !validTypes[contentType] {
		response.Error(c, http.StatusBadRequest, "Invalid content type", nil)
		return
	}

	// Get content stats
	stats, err := analyticsService.GetContentStats(contentType, startDate, endDate)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get content analytics", err)
		return
	}

	response.Success(c, http.StatusOK, "Content analytics retrieved successfully", stats)
}

// GetEngagementAnalytics returns platform-wide engagement metrics
func GetEngagementAnalytics(c *gin.Context) {
	analyticsService := c.MustGet("analyticsService").(AnalyticsService)

	// Parse date parameters
	startDateStr := c.DefaultQuery("start_date", time.Now().AddDate(0, -1, 0).Format("2006-01-02"))
	endDateStr := c.DefaultQuery("end_date", time.Now().Format("2006-01-02"))

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid start date format", err)
		return
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid end date format", err)
		return
	}

	// Get engagement metrics
	metrics, err := analyticsService.GetEngagementMetrics(startDate, endDate)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get engagement metrics", err)
		return
	}

	response.Success(c, http.StatusOK, "Engagement analytics retrieved successfully", metrics)
}

// GetTopContentAnalytics returns the top performing content based on various metrics
func GetTopContentAnalytics(c *gin.Context) {
	analyticsService := c.MustGet("analyticsService").(AnalyticsService)

	// Parse parameters
	contentType := c.DefaultQuery("type", "post")
	metric := c.DefaultQuery("metric", "views")
	limitStr := c.DefaultQuery("limit", "10")

	// Parse limit
	limit := 10
	if parsedLimit, err := parseInt(limitStr); err == nil && parsedLimit > 0 {
		limit = parsedLimit
	}

	// Validate content type
	validTypes := map[string]bool{"post": true, "story": true, "live_stream": true}
	if !validTypes[contentType] {
		response.Error(c, http.StatusBadRequest, "Invalid content type", nil)
		return
	}

	// Validate metric
	validMetrics := map[string]bool{"views": true, "likes": true, "comments": true, "shares": true, "engagement_rate": true}
	if !validMetrics[metric] {
		response.Error(c, http.StatusBadRequest, "Invalid metric", nil)
		return
	}

	// Get top content
	topContent, err := analyticsService.GetTopContent(contentType, metric, limit)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get top content", err)
		return
	}

	response.Success(c, http.StatusOK, "Top content retrieved successfully", topContent)
}

// GetDemographicsAnalytics returns user demographic information
func GetDemographicsAnalytics(c *gin.Context) {
	analyticsService := c.MustGet("analyticsService").(AnalyticsService)

	// Get demographics data
	demographics, err := analyticsService.GetUserDemographics()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get demographics data", err)
		return
	}

	response.Success(c, http.StatusOK, "Demographics data retrieved successfully", demographics)
}

// GetRetentionAnalytics returns user retention cohort analysis
func GetRetentionAnalytics(c *gin.Context) {
	analyticsService := c.MustGet("analyticsService").(AnalyticsService)

	// Parse parameters
	cohortDateStr := c.DefaultQuery("cohort_date", time.Now().AddDate(0, -3, 0).Format("2006-01-02"))
	periodsStr := c.DefaultQuery("periods", "12")

	cohortDate, err := time.Parse("2006-01-02", cohortDateStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid cohort date format", err)
		return
	}

	periods := 12
	if parsedPeriods, err := parseInt(periodsStr); err == nil && parsedPeriods > 0 {
		periods = parsedPeriods
	}

	// Get retention data
	retentionData, err := analyticsService.GetUserRetention(cohortDate, periods)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get retention data", err)
		return
	}

	response.Success(c, http.StatusOK, "Retention data retrieved successfully", retentionData)
}

// GetActiveUsersAnalytics returns active user statistics
func GetActiveUsersAnalytics(c *gin.Context) {
	analyticsService := c.MustGet("analyticsService").(AnalyticsService)

	// Parse period parameter
	period := c.DefaultQuery("period", "daily")

	// Validate period
	if period != "daily" && period != "weekly" && period != "monthly" {
		response.Error(c, http.StatusBadRequest, "Invalid period. Must be daily, weekly, or monthly", nil)
		return
	}

	// Get active user stats
	stats, err := analyticsService.GetActiveUserStats(period)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get active user stats", err)
		return
	}

	response.Success(c, http.StatusOK, "Active user statistics retrieved successfully", stats)
}

// Helper function to parse int from string
func parseInt(str string) (int, error) {
	var value int
	_, err := fmt.Sscanf(str, "%d", &value)
	return value, err
}
