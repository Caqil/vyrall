package hashtags

import (
	"net/http"
	"strconv"
	"time"

	"github.com/Caqil/vyrall/internal/utils/response"
	"github.com/gin-gonic/gin"
)

// TrendingService defines the interface for trending hashtag operations
type TrendingService interface {
	GetTrendingHashtags(ctx context.Context, limit int) ([]*models.Hashtag, error)
	GetTrendingHashtagsByCategory(ctx context.Context, category string, limit int) ([]*models.Hashtag, error)
	GetHashtagTrends(ctx context.Context, startDate, endDate time.Time, limit int) ([]map[string]interface{}, error)
}

// GetTrendingHashtags retrieves the currently trending hashtags
func GetTrendingHashtags(c *gin.Context) {
	// Get query parameters
	limitStr := c.DefaultQuery("limit", "20")
	category := c.DefaultQuery("category", "")

	// Parse limit
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		limit = 20
	}

	// Get the trending service
	trendingService := c.MustGet("trendingService").(TrendingService)

	// Get trending hashtags, filtered by category if specified
	var hashtags []*models.Hashtag
	var err error

	if category != "" {
		hashtags, err = trendingService.GetTrendingHashtagsByCategory(c.Request.Context(), category, limit)
	} else {
		hashtags, err = trendingService.GetTrendingHashtags(c.Request.Context(), limit)
	}

	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve trending hashtags", err)
		return
	}

	response.Success(c, http.StatusOK, "Trending hashtags retrieved successfully", hashtags)
}

// GetHashtagTrends retrieves trending data over time
func GetHashtagTrends(c *gin.Context) {
	// Get query parameters
	startDateStr := c.DefaultQuery("start_date", "")
	endDateStr := c.DefaultQuery("end_date", "")
	limitStr := c.DefaultQuery("limit", "10")

	// Parse dates
	var startDate, endDate time.Time
	if startDateStr != "" {
		startDate, _ = time.Parse("2006-01-02", startDateStr)
	} else {
		// Default to 7 days ago
		startDate = time.Now().AddDate(0, 0, -7)
	}

	if endDateStr != "" {
		endDate, _ = time.Parse("2006-01-02", endDateStr)
	} else {
		// Default to now
		endDate = time.Now()
	}

	// Parse limit
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		limit = 10
	}

	// Get the trending service
	trendingService := c.MustGet("trendingService").(TrendingService)

	// Get hashtag trends
	trends, err := trendingService.GetHashtagTrends(c.Request.Context(), startDate, endDate, limit)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve hashtag trends", err)
		return
	}

	response.Success(c, http.StatusOK, "Hashtag trends retrieved successfully", trends)
}
