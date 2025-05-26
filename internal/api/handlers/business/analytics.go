package business

import (
	"net/http"
	"strings"
	"time"

	"github.com/Caqil/vyrall/internal/utils/response"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// AnalyticsService defines the interface for business analytics operations
type AnalyticsService interface {
	GetBusinessOverview(userID primitive.ObjectID, startDate, endDate time.Time) (map[string]interface{}, error)
	GetAudienceInsights(userID primitive.ObjectID, startDate, endDate time.Time) (map[string]interface{}, error)
	GetContentPerformance(userID primitive.ObjectID, contentType string, limit int, startDate, endDate time.Time) ([]map[string]interface{}, error)
	GetFollowerGrowth(userID primitive.ObjectID, period string, startDate, endDate time.Time) ([]map[string]interface{}, error)
	GetEngagementMetrics(userID primitive.ObjectID, startDate, endDate time.Time) (map[string]interface{}, error)
	GetReachMetrics(userID primitive.ObjectID, startDate, endDate time.Time) (map[string]interface{}, error)
	GetReferralSources(userID primitive.ObjectID, startDate, endDate time.Time) ([]map[string]interface{}, error)
	ExportAnalytics(userID primitive.ObjectID, metrics []string, format string, startDate, endDate time.Time) (string, error)
	GetAdPerformanceAnalytics(userID primitive.ObjectID, adID *primitive.ObjectID, startDate, endDate time.Time) (map[string]interface{}, error)
	GetMonetizationAnalytics(userID primitive.ObjectID, startDate, endDate time.Time) (map[string]interface{}, error)
}

// GetBusinessOverview returns an overview of business metrics
func GetBusinessOverview(c *gin.Context) {
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

	// Get user ID from authenticated user
	userID, exists := c.Get("userID")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "Authentication required", nil)
		return
	}

	analyticsService := c.MustGet("analyticsService").(AnalyticsService)

	// Get business overview
	overview, err := analyticsService.GetBusinessOverview(userID.(primitive.ObjectID), startDate, endDate)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve business overview", err)
		return
	}

	response.Success(c, http.StatusOK, "Business overview retrieved successfully", overview)
}

// GetAudienceInsights returns insights about the audience
func GetAudienceInsights(c *gin.Context) {
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

	// Get user ID from authenticated user
	userID, exists := c.Get("userID")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "Authentication required", nil)
		return
	}

	analyticsService := c.MustGet("analyticsService").(AnalyticsService)

	// Get audience insights
	insights, err := analyticsService.GetAudienceInsights(userID.(primitive.ObjectID), startDate, endDate)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve audience insights", err)
		return
	}

	response.Success(c, http.StatusOK, "Audience insights retrieved successfully", insights)
}

// GetContentPerformance returns performance metrics for content
func GetContentPerformance(c *gin.Context) {
	contentType := c.DefaultQuery("type", "post")
	limitStr := c.DefaultQuery("limit", "10")

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

	limit := 10
	if parsedLimit, err := parseInt(limitStr); err == nil && parsedLimit > 0 {
		limit = parsedLimit
	}

	// Get user ID from authenticated user
	userID, exists := c.Get("userID")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "Authentication required", nil)
		return
	}

	analyticsService := c.MustGet("analyticsService").(AnalyticsService)

	// Validate content type
	validTypes := map[string]bool{
		"post":        true,
		"story":       true,
		"live_stream": true,
		"all":         true,
	}

	if !validTypes[contentType] {
		response.Error(c, http.StatusBadRequest, "Invalid content type", nil)
		return
	}

	// Get content performance
	performance, err := analyticsService.GetContentPerformance(userID.(primitive.ObjectID), contentType, limit, startDate, endDate)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve content performance", err)
		return
	}

	response.Success(c, http.StatusOK, "Content performance retrieved successfully", performance)
}

// GetFollowerGrowth returns follower growth over time
func GetFollowerGrowth(c *gin.Context) {
	period := c.DefaultQuery("period", "daily")

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

	// Get user ID from authenticated user
	userID, exists := c.Get("userID")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "Authentication required", nil)
		return
	}

	analyticsService := c.MustGet("analyticsService").(AnalyticsService)

	// Validate period
	validPeriods := map[string]bool{
		"daily":   true,
		"weekly":  true,
		"monthly": true,
	}

	if !validPeriods[period] {
		response.Error(c, http.StatusBadRequest, "Invalid period", nil)
		return
	}

	// Get follower growth
	growth, err := analyticsService.GetFollowerGrowth(userID.(primitive.ObjectID), period, startDate, endDate)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve follower growth", err)
		return
	}

	response.Success(c, http.StatusOK, "Follower growth retrieved successfully", growth)
}

// GetEngagementMetrics returns engagement metrics
func GetEngagementMetrics(c *gin.Context) {
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

	// Get user ID from authenticated user
	userID, exists := c.Get("userID")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "Authentication required", nil)
		return
	}

	analyticsService := c.MustGet("analyticsService").(AnalyticsService)

	// Get engagement metrics
	metrics, err := analyticsService.GetEngagementMetrics(userID.(primitive.ObjectID), startDate, endDate)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve engagement metrics", err)
		return
	}

	response.Success(c, http.StatusOK, "Engagement metrics retrieved successfully", metrics)
}

// GetReachMetrics returns reach and impression metrics
func GetReachMetrics(c *gin.Context) {
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

	// Get user ID from authenticated user
	userID, exists := c.Get("userID")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "Authentication required", nil)
		return
	}

	analyticsService := c.MustGet("analyticsService").(AnalyticsService)

	// Get reach metrics
	metrics, err := analyticsService.GetReachMetrics(userID.(primitive.ObjectID), startDate, endDate)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve reach metrics", err)
		return
	}

	response.Success(c, http.StatusOK, "Reach metrics retrieved successfully", metrics)
}

// GetReferralSources returns referral sources
func GetReferralSources(c *gin.Context) {
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

	// Get user ID from authenticated user
	userID, exists := c.Get("userID")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "Authentication required", nil)
		return
	}

	analyticsService := c.MustGet("analyticsService").(AnalyticsService)

	// Get referral sources
	sources, err := analyticsService.GetReferralSources(userID.(primitive.ObjectID), startDate, endDate)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve referral sources", err)
		return
	}

	response.Success(c, http.StatusOK, "Referral sources retrieved successfully", sources)
}

// ExportAnalytics exports analytics data
func ExportAnalytics(c *gin.Context) {
	format := c.DefaultQuery("format", "csv")
	var metrics []string
	if metricsParam := c.Query("metrics"); metricsParam != "" {
		metrics = strings.Split(metricsParam, ",")
	} else {
		metrics = []string{"followers", "engagement", "reach", "views"}
	}

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

	// Get user ID from authenticated user
	userID, exists := c.Get("userID")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "Authentication required", nil)
		return
	}

	analyticsService := c.MustGet("analyticsService").(AnalyticsService)

	// Validate format
	validFormats := map[string]bool{
		"csv":  true,
		"xlsx": true,
		"json": true,
	}

	if !validFormats[format] {
		response.Error(c, http.StatusBadRequest, "Invalid export format", nil)
		return
	}

	// Export analytics data
	fileURL, err := analyticsService.ExportAnalytics(userID.(primitive.ObjectID), metrics, format, startDate, endDate)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to export analytics", err)
		return
	}

	response.Success(c, http.StatusOK, "Analytics exported successfully", gin.H{
		"file_url": fileURL,
	})
}

// GetAdPerformanceAnalytics returns ad performance analytics
func GetAdPerformanceAnalytics(c *gin.Context) {
	// Optional ad ID parameter
	var adID *primitive.ObjectID
	if adIDStr := c.Query("ad_id"); adIDStr != "" {
		parsedID, err := primitive.ObjectIDFromHex(adIDStr)
		if err != nil {
			response.Error(c, http.StatusBadRequest, "Invalid ad ID", err)
			return
		}
		adID = &parsedID
	}

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

	// Get user ID from authenticated user
	userID, exists := c.Get("userID")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "Authentication required", nil)
		return
	}

	analyticsService := c.MustGet("analyticsService").(AnalyticsService)

	// Get ad performance analytics
	performance, err := analyticsService.GetAdPerformanceAnalytics(userID.(primitive.ObjectID), adID, startDate, endDate)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve ad performance analytics", err)
		return
	}

	response.Success(c, http.StatusOK, "Ad performance analytics retrieved successfully", performance)
}

// GetMonetizationAnalytics returns monetization analytics
func GetMonetizationAnalytics(c *gin.Context) {
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

	// Get user ID from authenticated user
	userID, exists := c.Get("userID")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "Authentication required", nil)
		return
	}

	analyticsService := c.MustGet("analyticsService").(AnalyticsService)

	// Get monetization analytics
	analytics, err := analyticsService.GetMonetizationAnalytics(userID.(primitive.ObjectID), startDate, endDate)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve monetization analytics", err)
		return
	}

	response.Success(c, http.StatusOK, "Monetization analytics retrieved successfully", analytics)
}
