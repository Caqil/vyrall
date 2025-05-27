package analytics

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Caqil/vyrall/internal/database"
	"github.com/Caqil/vyrall/internal/utils/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ContentAnalyticsService provides analytics for content
type ContentAnalyticsService struct {
	db    *database.Database
	cache *database.RedisClient
	log   *logger.Logger
}

// NewContentAnalyticsService creates a new content analytics service
func NewContentAnalyticsService(db *database.Database, cache *database.RedisClient, log *logger.Logger) *ContentAnalyticsService {
	return &ContentAnalyticsService{
		db:    db,
		cache: cache,
		log:   log,
	}
}

// ContentMetrics represents metrics for content
type ContentMetrics struct {
	Period             string                 `json:"period"`
	TotalPosts         int                    `json:"total_posts"`
	TotalViews         int                    `json:"total_views"`
	TotalLikes         int                    `json:"total_likes"`
	TotalComments      int                    `json:"total_comments"`
	TotalShares        int                    `json:"total_shares"`
	AvgEngagementRate  float64                `json:"avg_engagement_rate"`
	TopContentTypes    []ContentTypeMetrics   `json:"top_content_types"`
	TopPerformingPosts []TopPerformingPost    `json:"top_performing_posts"`
	ContentTrends      map[string]interface{} `json:"content_trends"`
}

// ContentTypeMetrics represents metrics for a content type
type ContentTypeMetrics struct {
	Type              string  `json:"type"`
	Count             int     `json:"count"`
	AvgEngagementRate float64 `json:"avg_engagement_rate"`
	TrendPercentage   float64 `json:"trend_percentage"`
}

// TopPerformingPost represents a top performing post
type TopPerformingPost struct {
	ID             string  `json:"id"`
	UserID         string  `json:"user_id"`
	Type           string  `json:"type"`
	Views          int     `json:"views"`
	Likes          int     `json:"likes"`
	Comments       int     `json:"comments"`
	Shares         int     `json:"shares"`
	EngagementRate float64 `json:"engagement_rate"`
}

// GetAggregatedMetrics gets aggregated content metrics for a period
func (s *ContentAnalyticsService) GetAggregatedMetrics(ctx context.Context, period string) (*ContentMetrics, error) {
	// Determine time range based on period
	startTime, endTime, err := getTimeRangeForPeriod(period)
	if err != nil {
		return nil, err
	}

	// Try to get from cache first
	cacheKey := "content_metrics:" + period
	cachedMetrics, err := s.cache.Get(ctx, cacheKey)
	if err == nil && cachedMetrics != "" {
		var metrics ContentMetrics
		if err := json.Unmarshal([]byte(cachedMetrics), &metrics); err == nil {
			return &metrics, nil
		}
	}

	// Fetch metrics from database
	metrics := &ContentMetrics{
		Period: period,
	}

	// Aggregate total posts and metrics
	pipeline := []bson.M{
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
				"total_posts":    bson.M{"$sum": 1},
				"total_views":    bson.M{"$sum": "$view_count"},
				"total_likes":    bson.M{"$sum": "$like_count"},
				"total_comments": bson.M{"$sum": "$comment_count"},
				"total_shares":   bson.M{"$sum": "$share_count"},
			},
		},
	}

	results, err := s.db.Aggregate(ctx, "posts", pipeline)
	if err != nil {
		s.log.Error("Failed to aggregate content metrics", "error", err)
		return nil, err
	}

	if len(results) > 0 {
		metrics.TotalPosts = int(results[0]["total_posts"].(int32))
		metrics.TotalViews = int(results[0]["total_views"].(int32))
		metrics.TotalLikes = int(results[0]["total_likes"].(int32))
		metrics.TotalComments = int(results[0]["total_comments"].(int32))
		metrics.TotalShares = int(results[0]["total_shares"].(int32))

		// Calculate average engagement rate
		totalEngagements := metrics.TotalLikes + metrics.TotalComments + metrics.TotalShares
		if metrics.TotalPosts > 0 {
			metrics.AvgEngagementRate = float64(totalEngagements) / float64(metrics.TotalPosts)
		}
	}

	// Get top content types
	metrics.TopContentTypes, err = s.getTopContentTypes(ctx, startTime, endTime)
	if err != nil {
		s.log.Error("Failed to get top content types", "error", err)
		// Continue despite error
	}

	// Get top performing posts
	metrics.TopPerformingPosts, err = s.getTopPerformingPosts(ctx, startTime, endTime)
	if err != nil {
		s.log.Error("Failed to get top performing posts", "error", err)
		// Continue despite error
	}

	// Get content trends
	metrics.ContentTrends, err = s.getContentTrends(ctx, period, startTime, endTime)
	if err != nil {
		s.log.Error("Failed to get content trends", "error", err)
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

// GetPostAnalytics gets analytics for a specific post
func (s *ContentAnalyticsService) GetPostAnalytics(ctx context.Context, postID primitive.ObjectID) (*PostAnalytics, error) {
	// Try to get from cache first
	cacheKey := "post_analytics:" + postID.Hex()
	cachedAnalytics, err := s.cache.Get(ctx, cacheKey)
	if err == nil && cachedAnalytics != "" {
		var analytics PostAnalytics
		if err := json.Unmarshal([]byte(cachedAnalytics), &analytics); err == nil {
			return &analytics, nil
		}
	}

	// Get the post details
	var post struct {
		ID           primitive.ObjectID `bson:"_id"`
		UserID       primitive.ObjectID `bson:"user_id"`
		Content      string             `bson:"content"`
		LikeCount    int                `bson:"like_count"`
		CommentCount int                `bson:"comment_count"`
		ShareCount   int                `bson:"share_count"`
		ViewCount    int                `bson:"view_count"`
		CreatedAt    time.Time          `bson:"created_at"`
	}

	err = s.db.FindOne(ctx, "posts", bson.M{"_id": postID}, &post)
	if err != nil {
		return nil, err
	}

	// Get view analytics
	viewsBreakdown, err := s.getViewsBreakdown(ctx, postID)
	if err != nil {
		s.log.Error("Failed to get views breakdown", "error", err, "post_id", postID.Hex())
		// Continue despite error
	}

	// Get audience demographics
	audienceDemographics, err := s.getAudienceDemographics(ctx, postID)
	if err != nil {
		s.log.Error("Failed to get audience demographics", "error", err, "post_id", postID.Hex())
		// Continue despite error
	}

	// Get referrers
	topReferrers, err := s.getTopReferrers(ctx, postID)
	if err != nil {
		s.log.Error("Failed to get top referrers", "error", err, "post_id", postID.Hex())
		// Continue despite error
	}

	// Calculate engagement rate
	var engagementRate float64
	totalEngagements := post.LikeCount + post.CommentCount + post.ShareCount
	if post.ViewCount > 0 {
		engagementRate = float64(totalEngagements) / float64(post.ViewCount)
	}

	// Create analytics result
	analytics := &PostAnalytics{
		ID:                   post.ID,
		Impressions:          post.ViewCount,
		Reach:                0, // To be calculated separately
		EngagementRate:       engagementRate,
		Likes:                post.LikeCount,
		Comments:             post.CommentCount,
		Shares:               post.ShareCount,
		ViewsBreakdown:       viewsBreakdown,
		AudienceDemographics: audienceDemographics,
		TopReferrers:         topReferrers,
		Period:               "all_time",
		StartDate:            post.CreatedAt,
		EndDate:              time.Now(),
		UpdatedAt:            time.Now(),
	}

	// Cache the results
	analyticsJSON, err := json.Marshal(analytics)
	if err == nil {
		// Cache for 1 hour
		s.cache.SetWithExpiration(ctx, cacheKey, string(analyticsJSON), 1*time.Hour)
	}

	return analytics, nil
}

// GetContentTypeAnalytics gets analytics for a specific content type
func (s *ContentAnalyticsService) GetContentTypeAnalytics(ctx context.Context, contentType, period string) (*ContentTypeAnalytics, error) {
	// Determine time range based on period
	startTime, endTime, err := getTimeRangeForPeriod(period)
	if err != nil {
		return nil, err
	}

	// Try to get from cache first
	cacheKey := "content_type_analytics:" + contentType + ":" + period
	cachedAnalytics, err := s.cache.Get(ctx, cacheKey)
	if err == nil && cachedAnalytics != "" {
		var analytics ContentTypeAnalytics
		if err := json.Unmarshal([]byte(cachedAnalytics), &analytics); err == nil {
			return &analytics, nil
		}
	}

	// Determine how to identify the content type in the database
	var contentTypeQuery bson.M
	switch contentType {
	case "text":
		contentTypeQuery = bson.M{
			"media_files": bson.M{"$size": 0},
		}
	case "image":
		contentTypeQuery = bson.M{
			"media_files.type": "image",
		}
	case "video":
		contentTypeQuery = bson.M{
			"media_files.type": "video",
		}
	case "poll":
		contentTypeQuery = bson.M{
			"poll": bson.M{"$ne": nil},
		}
	default:
		return nil, fmt.Errorf("unknown content type: %s", contentType)
	}

	// Combine with time range
	query := bson.M{
		"$and": []bson.M{
			contentTypeQuery,
			{
				"created_at": bson.M{
					"$gte": startTime,
					"$lte": endTime,
				},
				"deleted_at": nil,
			},
		},
	}

	// Aggregate metrics for this content type
	pipeline := []bson.M{
		{
			"$match": query,
		},
		{
			"$group": bson.M{
				"_id":               nil,
				"count":             bson.M{"$sum": 1},
				"total_views":       bson.M{"$sum": "$view_count"},
				"total_likes":       bson.M{"$sum": "$like_count"},
				"total_comments":    bson.M{"$sum": "$comment_count"},
				"total_shares":      bson.M{"$sum": "$share_count"},
				"avg_view_count":    bson.M{"$avg": "$view_count"},
				"avg_like_count":    bson.M{"$avg": "$like_count"},
				"avg_comment_count": bson.M{"$avg": "$comment_count"},
				"avg_share_count":   bson.M{"$avg": "$share_count"},
			},
		},
	}

	results, err := s.db.Aggregate(ctx, "posts", pipeline)
	if err != nil {
		s.log.Error("Failed to aggregate content type metrics", "error", err)
		return nil, err
	}

	// Create analytics result
	analytics := &ContentTypeAnalytics{
		Type:   contentType,
		Period: period,
	}

	if len(results) > 0 {
		analytics.Count = int(results[0]["count"].(int32))
		analytics.TotalViews = int(results[0]["total_views"].(int32))
		analytics.TotalLikes = int(results[0]["total_likes"].(int32))
		analytics.TotalComments = int(results[0]["total_comments"].(int32))
		analytics.TotalShares = int(results[0]["total_shares"].(int32))
		analytics.AvgViewCount = results[0]["avg_view_count"].(float64)
		analytics.AvgLikeCount = results[0]["avg_like_count"].(float64)
		analytics.AvgCommentCount = results[0]["avg_comment_count"].(float64)
		analytics.AvgShareCount = results[0]["avg_share_count"].(float64)

		// Calculate average engagement rate
		totalEngagements := analytics.TotalLikes + analytics.TotalComments + analytics.TotalShares
		if analytics.Count > 0 {
			analytics.AvgEngagementRate = float64(totalEngagements) / float64(analytics.Count)
		}
	}

	// Get performance trend
	analytics.PerformanceTrend, err = s.getContentTypeTrend(ctx, contentType, period)
	if err != nil {
		s.log.Error("Failed to get content type trend", "error", err)
		// Continue despite error
	}

	// Cache the results
	analyticsJSON, err := json.Marshal(analytics)
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
		s.cache.SetWithExpiration(ctx, cacheKey, string(analyticsJSON), cacheDuration)
	}

	return analytics, nil
}

// Helper methods would be implemented here...

// getTopContentTypes gets the top performing content types
func (s *ContentAnalyticsService) getTopContentTypes(ctx context.Context, startTime, endTime time.Time) ([]ContentTypeMetrics, error) {
	// Implementation for getting top content types
	return nil, nil
}

// getTopPerformingPosts gets the top performing posts
func (s *ContentAnalyticsService) getTopPerformingPosts(ctx context.Context, startTime, endTime time.Time) ([]TopPerformingPost, error) {
	// Implementation for getting top performing posts
	return nil, nil
}

// getContentTrends gets content trends
func (s *ContentAnalyticsService) getContentTrends(ctx context.Context, period string, startTime, endTime time.Time) (map[string]interface{}, error) {
	// Implementation for getting content trends
	return nil, nil
}

// getViewsBreakdown gets views breakdown for a post
func (s *ContentAnalyticsService) getViewsBreakdown(ctx context.Context, postID primitive.ObjectID) (map[string]int, error) {
	// Implementation for getting views breakdown
	return nil, nil
}

// getAudienceDemographics gets audience demographics for a post
func (s *ContentAnalyticsService) getAudienceDemographics(ctx context.Context, postID primitive.ObjectID) (map[string]interface{}, error) {
	// Implementation for getting audience demographics
	return nil, nil
}

// getTopReferrers gets top referrers for a post
func (s *ContentAnalyticsService) getTopReferrers(ctx context.Context, postID primitive.ObjectID) ([]string, error) {
	// Implementation for getting top referrers
	return nil, nil
}

// getContentTypeTrend gets performance trend for a content type
func (s *ContentAnalyticsService) getContentTypeTrend(ctx context.Context, contentType, period string) (map[string]interface{}, error) {
	// Implementation for getting content type trend
	return nil, nil
}
