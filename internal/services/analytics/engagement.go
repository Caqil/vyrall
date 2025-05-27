package analytics

import (
	"context"
	"encoding/json"
	"time"

	"github.com/Caqil/vyrall/internal/database"
	"github.com/Caqil/vyrall/internal/utils/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// EngagementService provides engagement analytics functionality
type EngagementService struct {
	db    *database.Database
	cache *database.RedisClient
	log   *logger.Logger
}

// NewEngagementService creates a new engagement service
func NewEngagementService(db *database.Database, cache *database.RedisClient, log *logger.Logger) *EngagementService {
	return &EngagementService{
		db:    db,
		cache: cache,
		log:   log,
	}
}

// EngagementMetrics represents engagement metrics
type EngagementMetrics struct {
	Period             string                 `json:"period"`
	TotalEngagements   int                    `json:"total_engagements"`
	TotalLikes         int                    `json:"total_likes"`
	TotalComments      int                    `json:"total_comments"`
	TotalShares        int                    `json:"total_shares"`
	AvgEngagementRate  float64                `json:"avg_engagement_rate"`
	TopEngagementHours []HourlyEngagement     `json:"top_engagement_hours"`
	TopEngagementDays  []DailyEngagement      `json:"top_engagement_days"`
	EngagementByType   map[string]int         `json:"engagement_by_type"`
	EngagementTrends   map[string]interface{} `json:"engagement_trends"`
}

// HourlyEngagement represents engagement metrics for a specific hour
type HourlyEngagement struct {
	Hour           int     `json:"hour"`
	Engagements    int     `json:"engagements"`
	EngagementRate float64 `json:"engagement_rate"`
}

// DailyEngagement represents engagement metrics for a specific day
type DailyEngagement struct {
	Day            string  `json:"day"`
	Engagements    int     `json:"engagements"`
	EngagementRate float64 `json:"engagement_rate"`
}

// GetAggregatedMetrics gets aggregated engagement metrics for a period
func (s *EngagementService) GetAggregatedMetrics(ctx context.Context, period string) (*EngagementMetrics, error) {
	// Determine time range based on period
	startTime, endTime, err := getTimeRangeForPeriod(period)
	if err != nil {
		return nil, err
	}

	// Try to get from cache first
	cacheKey := "engagement_metrics:" + period
	cachedMetrics, err := s.cache.Get(ctx, cacheKey)
	if err == nil && cachedMetrics != "" {
		var metrics EngagementMetrics
		if err := json.Unmarshal([]byte(cachedMetrics), &metrics); err == nil {
			return &metrics, nil
		}
	}

	// Fetch metrics from database
	metrics := &EngagementMetrics{
		Period: period,
	}

	// Aggregate engagements for posts
	postsPipeline := []bson.M{
		{
			"$match": bson.M{
				"created_at": bson.M{
					"$gte": startTime,
					"$lte": endTime,
				},
				"deleted_at": nil,
			},
		},
		{
			"$group": bson.M{
				"_id":            nil,
				"total_likes":    bson.M{"$sum": "$like_count"},
				"total_comments": bson.M{"$sum": "$comment_count"},
				"total_shares":   bson.M{"$sum": "$share_count"},
				"post_count":     bson.M{"$sum": 1},
			},
		},
	}

	postsResults, err := s.db.Aggregate(ctx, "posts", postsPipeline)
	if err != nil {
		s.log.Error("Failed to aggregate post engagements", "error", err)
		return nil, err
	}

	if len(postsResults) > 0 {
		metrics.TotalLikes = int(postsResults[0]["total_likes"].(int32))
		metrics.TotalComments = int(postsResults[0]["total_comments"].(int32))
		metrics.TotalShares = int(postsResults[0]["total_shares"].(int32))

		// Calculate total engagements
		metrics.TotalEngagements = metrics.TotalLikes + metrics.TotalComments + metrics.TotalShares

		// Calculate average engagement rate
		postCount := int(postsResults[0]["post_count"].(int32))
		if postCount > 0 {
			metrics.AvgEngagementRate = float64(metrics.TotalEngagements) / float64(postCount)
		}
	}

	// Get engagement by type
	metrics.EngagementByType = map[string]int{
		"likes":    metrics.TotalLikes,
		"comments": metrics.TotalComments,
		"shares":   metrics.TotalShares,
	}

	// Get top engagement hours
	metrics.TopEngagementHours, err = s.getTopEngagementHours(ctx, startTime, endTime)
	if err != nil {
		s.log.Error("Failed to get top engagement hours", "error", err)
		// Continue despite error
	}

	// Get top engagement days
	metrics.TopEngagementDays, err = s.getTopEngagementDays(ctx, startTime, endTime)
	if err != nil {
		s.log.Error("Failed to get top engagement days", "error", err)
		// Continue despite error
	}

	// Get engagement trends
	metrics.EngagementTrends, err = s.getEngagementTrends(ctx, period, startTime, endTime)
	if err != nil {
		s.log.Error("Failed to get engagement trends", "error", err)
		// Continue despite error
	}

	// Cache the results
	metricsJSON, err := json.Marshal(metrics)
	if err == nil {
		// Cache for an appropriate duration based on period
		var cacheDuration time.Duration
		switch period {
		case "day":
			cacheDuration = 1 * time.Hour
		case "week":
			cacheDuration = 6 * time.Hour
		case "month":
			cacheDuration = 12 * time.Hour
		default:
			cacheDuration = 24 * time.Hour
		}
		s.cache.SetWithExpiration(ctx, cacheKey, string(metricsJSON), cacheDuration)
	}

	return metrics, nil
}

// GetUserEngagementMetrics gets engagement metrics for a specific user
func (s *EngagementService) GetUserEngagementMetrics(ctx context.Context, userID primitive.ObjectID, period string) (*UserEngagementMetrics, error) {
	// Determine time range based on period
	startTime, endTime, err := getTimeRangeForPeriod(period)
	if err != nil {
		return nil, err
	}

	// Try to get from cache first
	cacheKey := "user_engagement:" + userID.Hex() + ":" + period
	cachedMetrics, err := s.cache.Get(ctx, cacheKey)
	if err == nil && cachedMetrics != "" {
		var metrics UserEngagementMetrics
		if err := json.Unmarshal([]byte(cachedMetrics), &metrics); err == nil {
			return &metrics, nil
		}
	}

	// Get user's posts within the time range
	userPostsPipeline := []bson.M{
		{
			"$match": bson.M{
				"user_id": userID,
				"created_at": bson.M{
					"$gte": startTime,
					"$lte": endTime,
				},
				"deleted_at": nil,
			},
		},
		{
			"$group": bson.M{
				"_id":            nil,
				"post_count":     bson.M{"$sum": 1},
				"total_likes":    bson.M{"$sum": "$like_count"},
				"total_comments": bson.M{"$sum": "$comment_count"},
				"total_shares":   bson.M{"$sum": "$share_count"},
				"total_views":    bson.M{"$sum": "$view_count"},
			},
		},
	}

	userPostsResults, err := s.db.Aggregate(ctx, "posts", userPostsPipeline)
	if err != nil {
		s.log.Error("Failed to aggregate user posts", "error", err, "user_id", userID.Hex())
		return nil, err
	}

	metrics := &UserEngagementMetrics{
		UserID: userID,
		Period: period,
	}

	if len(userPostsResults) > 0 {
		metrics.PostCount = int(userPostsResults[0]["post_count"].(int32))
		metrics.TotalLikes = int(userPostsResults[0]["total_likes"].(int32))
		metrics.TotalComments = int(userPostsResults[0]["total_comments"].(int32))
		metrics.TotalShares = int(userPostsResults[0]["total_shares"].(int32))
		metrics.TotalViews = int(userPostsResults[0]["total_views"].(int32))

		// Calculate total engagements
		metrics.TotalEngagements = metrics.TotalLikes + metrics.TotalComments + metrics.TotalShares

		// Calculate engagement rate
		if metrics.TotalViews > 0 {
			metrics.EngagementRate = float64(metrics.TotalEngagements) / float64(metrics.TotalViews)
		}
	}

	// Get user's engagement with other content
	userEngagementPipeline := []bson.M{
		{
			"$match": bson.M{
				"user_id": userID,
				"created_at": bson.M{
					"$gte": startTime,
					"$lte": endTime,
				},
			},
		},
		{
			"$group": bson.M{
				"_id":        nil,
				"like_count": bson.M{"$sum": 1},
			},
		},
	}

	userLikesResults, err := s.db.Aggregate(ctx, "likes", userEngagementPipeline)
	if err != nil {
		s.log.Error("Failed to aggregate user likes", "error", err, "user_id", userID.Hex())
		// Continue despite error
	} else if len(userLikesResults) > 0 {
		metrics.LikesGiven = int(userLikesResults[0]["like_count"].(int32))
	}

	// Get comments made by the user
	userCommentsPipeline := []bson.M{
		{
			"$match": bson.M{
				"user_id": userID,
				"created_at": bson.M{
					"$gte": startTime,
					"$lte": endTime,
				},
				"deleted_at": nil,
			},
		},
		{
			"$count": "comment_count",
		},
	}

	userCommentsResults, err := s.db.Aggregate(ctx, "comments", userCommentsPipeline)
	if err != nil {
		s.log.Error("Failed to aggregate user comments", "error", err, "user_id", userID.Hex())
		// Continue despite error
	} else if len(userCommentsResults) > 0 {
		metrics.CommentsGiven = int(userCommentsResults[0]["comment_count"].(int32))
	}

	// Get shares made by the user
	userSharesPipeline := []bson.M{
		{
			"$match": bson.M{
				"user_id": userID,
				"created_at": bson.M{
					"$gte": startTime,
					"$lte": endTime,
				},
			},
		},
		{
			"$count": "share_count",
		},
	}

	userSharesResults, err := s.db.Aggregate(ctx, "shares", userSharesPipeline)
	if err != nil {
		s.log.Error("Failed to aggregate user shares", "error", err, "user_id", userID.Hex())
		// Continue despite error
	} else if len(userSharesResults) > 0 {
		metrics.SharesGiven = int(userSharesResults[0]["share_count"].(int32))
	}

	// Calculate total interactions
	metrics.TotalInteractions = metrics.LikesGiven + metrics.CommentsGiven + metrics.SharesGiven

	// Cache the results
	metricsJSON, err := json.Marshal(metrics)
	if err == nil {
		// Cache for an appropriate duration based on period
		var cacheDuration time.Duration
		switch period {
		case "day":
			cacheDuration = 1 * time.Hour
		case "week":
			cacheDuration = 6 * time.Hour
		case "month":
			cacheDuration = 12 * time.Hour
		default:
			cacheDuration = 24 * time.Hour
		}
		s.cache.SetWithExpiration(ctx, cacheKey, string(metricsJSON), cacheDuration)
	}

	return metrics, nil
}

// Helper methods would be implemented here...

// getTopEngagementHours gets the top hours for engagement
func (s *EngagementService) getTopEngagementHours(ctx context.Context, startTime, endTime time.Time) ([]HourlyEngagement, error) {
	// Implementation for getting top engagement hours
	return nil, nil
}

// getTopEngagementDays gets the top days for engagement
func (s *EngagementService) getTopEngagementDays(ctx context.Context, startTime, endTime time.Time) ([]DailyEngagement, error) {
	// Implementation for getting top engagement days
	return nil, nil
}

// getEngagementTrends gets engagement trends over time
func (s *EngagementService) getEngagementTrends(ctx context.Context, period string, startTime, endTime time.Time) (map[string]interface{}, error) {
	// Implementation for getting engagement trends
	return nil, nil
}
