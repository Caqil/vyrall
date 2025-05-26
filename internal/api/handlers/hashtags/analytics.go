package hashtags

import (
	"net/http"
	"strconv"
	"time"

	"github.com/Caqil/vyrall/internal/utils/response"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// AnalyticsService defines the interface for hashtag analytics operations
type AnalyticsService interface {
	GetHashtagAnalytics(ctx context.Context, hashtagID primitive.ObjectID) (map[string]interface{}, error)
	GetHashtagGrowth(ctx context.Context, hashtagID primitive.ObjectID, period string, startDate, endDate time.Time) (map[string]interface{}, error)
	GetHashtagEngagement(ctx context.Context, hashtagID primitive.ObjectID, startDate, endDate time.Time) (map[string]interface{}, error)
	GetTopHashtagUsers(ctx context.Context, hashtagID primitive.ObjectID, limit int) ([]map[string]interface{}, error)
	GetHashtagDemographics(ctx context.Context, hashtagID primitive.ObjectID) (map[string]interface{}, error)
}

// HashtagService defines the interface for hashtag operations
type HashtagService interface {
	GetHashtagByID(ctx context.Context, id primitive.ObjectID) (*models.Hashtag, error)
	GetHashtagByName(ctx context.Context, name string) (*models.Hashtag, error)
	IsAdminUser(ctx context.Context, userID primitive.ObjectID) (bool, error)
}

// GetHashtagAnalytics returns analytics data for a specific hashtag
func GetHashtagAnalytics(c *gin.Context) {
	// Get hashtag ID or name from URL parameter
	hashtagParam := c.Param("hashtag")

	// Get the hashtag service
	hashtagService := c.MustGet("hashtagService").(HashtagService)

	// Get the analytics service
	analyticsService := c.MustGet("analyticsService").(AnalyticsService)

	// Find the hashtag by ID or name
	var hashtagID primitive.ObjectID
	var err error
	var hashtag *models.Hashtag

	// Try to parse as ObjectID first
	if primitive.IsValidObjectID(hashtagParam) {
		hashtagID, _ = primitive.ObjectIDFromHex(hashtagParam)
		hashtag, err = hashtagService.GetHashtagByID(c.Request.Context(), hashtagID)
	} else {
		// If not a valid ObjectID, treat as hashtag name
		hashtag, err = hashtagService.GetHashtagByName(c.Request.Context(), hashtagParam)
		if err == nil {
			hashtagID = hashtag.ID
		}
	}

	if err != nil {
		response.Error(c, http.StatusNotFound, "Hashtag not found", err)
		return
	}

	// For sensitive analytics, we might want to restrict access to admins
	// Check if requesting user is an admin
	userID, exists := c.Get("userID")
	if exists {
		isAdmin, err := hashtagService.IsAdminUser(c.Request.Context(), userID.(primitive.ObjectID))
		if err != nil {
			response.Error(c, http.StatusInternalServerError, "Failed to check user permissions", err)
			return
		}

		// If not admin, provide limited analytics
		if !isAdmin {
			// Provide basic public stats
			response.Success(c, http.StatusOK, "Hashtag analytics retrieved successfully", gin.H{
				"hashtag":        hashtag.Name,
				"post_count":     hashtag.PostCount,
				"follower_count": hashtag.FollowerCount,
				"is_trending":    hashtag.IsTrending,
				"created_at":     hashtag.CreatedAt,
			})
			return
		}
	}

	// For admins or public analytics, get detailed stats
	analytics, err := analyticsService.GetHashtagAnalytics(c.Request.Context(), hashtagID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve hashtag analytics", err)
		return
	}

	response.Success(c, http.StatusOK, "Hashtag analytics retrieved successfully", analytics)
}

// GetHashtagGrowth returns growth data for a hashtag over time
func GetHashtagGrowth(c *gin.Context) {
	// Get hashtag ID or name from URL parameter
	hashtagParam := c.Param("hashtag")

	// Get query parameters
	period := c.DefaultQuery("period", "daily") // daily, weekly, monthly
	startDateStr := c.DefaultQuery("start_date", "")
	endDateStr := c.DefaultQuery("end_date", "")

	// Parse dates
	var startDate, endDate time.Time
	if startDateStr != "" {
		startDate, _ = time.Parse("2006-01-02", startDateStr)
	} else {
		// Default to 30 days ago
		startDate = time.Now().AddDate(0, 0, -30)
	}

	if endDateStr != "" {
		endDate, _ = time.Parse("2006-01-02", endDateStr)
	} else {
		// Default to now
		endDate = time.Now()
	}

	// Get the hashtag service
	hashtagService := c.MustGet("hashtagService").(HashtagService)

	// Get the analytics service
	analyticsService := c.MustGet("analyticsService").(AnalyticsService)

	// Find the hashtag by ID or name
	var hashtagID primitive.ObjectID
	var err error
	var hashtag *models.Hashtag

	// Try to parse as ObjectID first
	if primitive.IsValidObjectID(hashtagParam) {
		hashtagID, _ = primitive.ObjectIDFromHex(hashtagParam)
		hashtag, err = hashtagService.GetHashtagByID(c.Request.Context(), hashtagID)
	} else {
		// If not a valid ObjectID, treat as hashtag name
		hashtag, err = hashtagService.GetHashtagByName(c.Request.Context(), hashtagParam)
		if err == nil {
			hashtagID = hashtag.ID
		}
	}

	if err != nil {
		response.Error(c, http.StatusNotFound, "Hashtag not found", err)
		return
	}

	// Get growth data
	growthData, err := analyticsService.GetHashtagGrowth(c.Request.Context(), hashtagID, period, startDate, endDate)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve hashtag growth data", err)
		return
	}

	response.Success(c, http.StatusOK, "Hashtag growth data retrieved successfully", growthData)
}

// GetTopHashtagUsers returns users who most frequently use a hashtag
func GetTopHashtagUsers(c *gin.Context) {
	// Get hashtag ID or name from URL parameter
	hashtagParam := c.Param("hashtag")

	// Get limit parameter
	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		limit = 10
	}

	// Get the hashtag service
	hashtagService := c.MustGet("hashtagService").(HashtagService)

	// Get the analytics service
	analyticsService := c.MustGet("analyticsService").(AnalyticsService)

	// Find the hashtag by ID or name
	var hashtagID primitive.ObjectID
	var hashtag *models.Hashtag

	// Try to parse as ObjectID first
	if primitive.IsValidObjectID(hashtagParam) {
		hashtagID, _ = primitive.ObjectIDFromHex(hashtagParam)
		hashtag, err = hashtagService.GetHashtagByID(c.Request.Context(), hashtagID)
	} else {
		// If not a valid ObjectID, treat as hashtag name
		hashtag, err = hashtagService.GetHashtagByName(c.Request.Context(), hashtagParam)
		if err == nil {
			hashtagID = hashtag.ID
		}
	}

	if err != nil {
		response.Error(c, http.StatusNotFound, "Hashtag not found", err)
		return
	}

	// Get top users
	users, err := analyticsService.GetTopHashtagUsers(c.Request.Context(), hashtagID, limit)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve top hashtag users", err)
		return
	}

	response.Success(c, http.StatusOK, "Top hashtag users retrieved successfully", users)
}
