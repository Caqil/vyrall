package business

import (
	"net/http"
	"time"

	"github.com/Caqil/vyrall/internal/utils/response"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// InsightService defines the interface for business insights operations
type InsightService interface {
	GetContentInsights(userID primitive.ObjectID, contentType string, startDate, endDate time.Time) (map[string]interface{}, error)
	GetAudienceInsights(userID primitive.ObjectID, startDate, endDate time.Time) (map[string]interface{}, error)
	GetCompetitorInsights(userID primitive.ObjectID, competitors []string, metrics []string) (map[string]interface{}, error)
	GetTrendInsights(userID primitive.ObjectID, category string, limit int) ([]map[string]interface{}, error)
	GetPublishingInsights(userID primitive.ObjectID, startDate, endDate time.Time) (map[string]interface{}, error)
	GetEngagementInsights(userID primitive.ObjectID, contentType string, startDate, endDate time.Time) (map[string]interface{}, error)
	GetCustomInsight(userID primitive.ObjectID, metric string, dimensions []string, filters map[string]interface{}, startDate, endDate time.Time) (map[string]interface{}, error)
}

// GetContentInsights returns insights about content performance
func GetContentInsights(c *gin.Context) {
	contentType := c.DefaultQuery("type", "post")

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

	insightService := c.MustGet("insightService").(InsightService)

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

	// Get content insights
	insights, err := insightService.GetContentInsights(userID.(primitive.ObjectID), contentType, startDate, endDate)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve content insights", err)
		return
	}

	response.Success(c, http.StatusOK, "Content insights retrieved successfully", insights)
}


// GetCompetitorInsights returns comparative insights with competitors
func GetCompetitorInsights(c *gin.Context) {
	var req struct {
		Competitors []string `json:"competitors" binding:"required"`
		Metrics     []string `json:"metrics"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Get user ID from authenticated user
	userID, exists := c.Get("userID")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "Authentication required", nil)
		return
	}

	insightService := c.MustGet("insightService").(InsightService)

	// Default metrics if not provided
	if len(req.Metrics) == 0 {
		req.Metrics = []string{"followers", "engagement_rate", "post_frequency", "growth_rate"}
	}

	// Get competitor insights
	insights, err := insightService.GetCompetitorInsights(userID.(primitive.ObjectID), req.Competitors, req.Metrics)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve competitor insights", err)
		return
	}

	response.Success(c, http.StatusOK, "Competitor insights retrieved successfully", insights)
}

// GetTrendInsights returns trend insights for a category
func GetTrendInsights(c *gin.Context) {
	category := c.DefaultQuery("category", "general")
	limitStr := c.DefaultQuery("limit", "10")

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

	insightService := c.MustGet("insightService").(InsightService)

	// Validate category
	validCategories := map[string]bool{
		"general":    true,
		"hashtags":   true,
		"content":    true,
		"engagement": true,
		"industry":   true,
	}

	if !validCategories[category] {
		response.Error(c, http.StatusBadRequest, "Invalid category", nil)
		return
	}

	// Get trend insights
	insights, err := insightService.GetTrendInsights(userID.(primitive.ObjectID), category, limit)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve trend insights", err)
		return
	}

	response.Success(c, http.StatusOK, "Trend insights retrieved successfully", insights)
}

// GetPublishingInsights returns insights about optimal publishing times
func GetPublishingInsights(c *gin.Context) {
	// Parse date parameters
	startDateStr := c.DefaultQuery("start_date", time.Now().AddDate(0, -3, 0).Format("2006-01-02"))
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

	insightService := c.MustGet("insightService").(InsightService)

	// Get publishing insights
	insights, err := insightService.GetPublishingInsights(userID.(primitive.ObjectID), startDate, endDate)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve publishing insights", err)
		return
	}

	response.Success(c, http.StatusOK, "Publishing insights retrieved successfully", insights)
}

// GetEngagementInsights returns detailed engagement insights
func GetEngagementInsights(c *gin.Context) {
	contentType := c.DefaultQuery("type", "post")

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

	insightService := c.MustGet("insightService").(InsightService)

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

	// Get engagement insights
	insights, err := insightService.GetEngagementInsights(userID.(primitive.ObjectID), contentType, startDate, endDate)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve engagement insights", err)
		return
	}

	response.Success(c, http.StatusOK, "Engagement insights retrieved successfully", insights)
}

// GetCustomInsight generates a custom insight based on specified parameters
func GetCustomInsight(c *gin.Context) {
	var req struct {
		Metric     string                 `json:"metric" binding:"required"`
		Dimensions []string               `json:"dimensions"`
		Filters    map[string]interface{} `json:"filters"`
		StartDate  string                 `json:"start_date"`
		EndDate    string                 `json:"end_date"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Parse date parameters
	startDateStr := req.StartDate
	if startDateStr == "" {
		startDateStr = time.Now().AddDate(0, -1, 0).Format("2006-01-02")
	}

	endDateStr := req.EndDate
	if endDateStr == "" {
		endDateStr = time.Now().Format("2006-01-02")
	}

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

	insightService := c.MustGet("insightService").(InsightService)

	// Get custom insight
	insight, err := insightService.GetCustomInsight(userID.(primitive.ObjectID), req.Metric, req.Dimensions, req.Filters, startDate, endDate)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to generate custom insight", err)
		return
	}

	response.Success(c, http.StatusOK, "Custom insight generated successfully", insight)
}
