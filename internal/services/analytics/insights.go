package analytics

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Caqil/vyrall/internal/database"
	"github.com/Caqil/vyrall/internal/utils/logger"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// InsightsService provides analytics insights
type InsightsService struct {
	db    *database.Database
	cache *database.RedisClient
	log   *logger.Logger
}

// NewInsightsService creates a new insights service
func NewInsightsService(db *database.Database, cache *database.RedisClient, log *logger.Logger) *InsightsService {
	return &InsightsService{
		db:    db,
		cache: cache,
		log:   log,
	}
}

// EntityInsights represents insights for an entity
type EntityInsights struct {
	EntityType      string                 `json:"entity_type"`
	EntityID        string                 `json:"entity_id"`
	Period          string                 `json:"period"`
	MetricSummary   map[string]interface{} `json:"metric_summary"`
	Trends          map[string]interface{} `json:"trends"`
	Insights        []Insight              `json:"insights"`
	Recommendations []Recommendation       `json:"recommendations"`
}

// Insight represents a specific insight about an entity
type Insight struct {
	Type        string                 `json:"type"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Metrics     map[string]interface{} `json:"metrics"`
	Confidence  float64                `json:"confidence"`
	Priority    string                 `json:"priority"` // high, medium, low
}

// Recommendation represents a recommendation based on insights
type Recommendation struct {
	Type        string `json:"type"`
	Title       string `json:"title"`
	Description string `json:"description"`
	ActionText  string `json:"action_text,omitempty"`
	ActionURL   string `json:"action_url,omitempty"`
	Priority    string `json:"priority"` // high, medium, low
}

// GetEntityInsights gets insights for a specific entity
func (s *InsightsService) GetEntityInsights(ctx context.Context, entityType, entityID string, period string) (*EntityInsights, error) {
	// Validate entity type
	switch entityType {
	case "user", "post", "group", "event", "hashtag":
		// Valid entity types
	default:
		return nil, fmt.Errorf("invalid entity type: %s", entityType)
	}

	// Try to get from cache first
	cacheKey := fmt.Sprintf("insights:%s:%s:%s", entityType, entityID, period)
	cachedInsights, err := s.cache.Get(ctx, cacheKey)
	if err == nil && cachedInsights != "" {
		var insights EntityInsights
		if err := json.Unmarshal([]byte(cachedInsights), &insights); err == nil {
			return &insights, nil
		}
	}

	// Convert entity ID to ObjectID
	objectID, err := primitive.ObjectIDFromHex(entityID)
	if err != nil {
		return nil, fmt.Errorf("invalid entity ID: %s", entityID)
	}

	// Determine time range based on period
	startTime, endTime, err := getTimeRangeForPeriod(period)
	if err != nil {
		return nil, err
	}

	// Initialize entity insights
	insights := &EntityInsights{
		EntityType: entityType,
		EntityID:   entityID,
		Period:     period,
	}

	// Get metrics based on entity type
	switch entityType {
	case "user":
		insights, err = s.getUserInsights(ctx, objectID, period, startTime, endTime)
	case "post":
		insights, err = s.getPostInsights(ctx, objectID, period, startTime, endTime)
	case "group":
		insights, err = s.getGroupInsights(ctx, objectID, period, startTime, endTime)
	case "event":
		insights, err = s.getEventInsights(ctx, objectID, period, startTime, endTime)
	case "hashtag":
		insights, err = s.getHashtagInsights(ctx, objectID, period, startTime, endTime)
	}

	if err != nil {
		s.log.Error("Failed to get entity insights", "error", err, "entity_type", entityType, "entity_id", entityID)
		return nil, err
	}

	// Cache the results
	insightsJSON, err := json.Marshal(insights)
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
		s.cache.SetWithExpiration(ctx, cacheKey, string(insightsJSON), cacheDuration)
	}

	return insights, nil
}

// GetTrendingInsights gets trending insights across the platform
func (s *InsightsService) GetTrendingInsights(ctx context.Context, limit int) ([]Insight, error) {
	// Try to get from cache first
	cacheKey := fmt.Sprintf("trending_insights:%d", limit)
	cachedInsights, err := s.cache.Get(ctx, cacheKey)
	if err == nil && cachedInsights != "" {
		var insights []Insight
		if err := json.Unmarshal([]byte(cachedInsights), &insights); err == nil {
			return insights, nil
		}
	}

	// Get trending content
	trendingContent, err := s.getTrendingContent(ctx, limit)
	if err != nil {
		s.log.Error("Failed to get trending content", "error", err)
		return nil, err
	}

	// Get trending hashtags
	trendingHashtags, err := s.getTrendingHashtags(ctx, limit)
	if err != nil {
		s.log.Error("Failed to get trending hashtags", "error", err)
		// Continue despite error
	}

	// Get user growth trends
	userGrowthTrends, err := s.getUserGrowthTrends(ctx)
	if err != nil {
		s.log.Error("Failed to get user growth trends", "error", err)
		// Continue despite error
	}

	// Generate insights
	insights := []Insight{}

	// Add insights from trending content
	for _, content := range trendingContent {
		insight := Insight{
			Type:        "trending_content",
			Title:       "Trending Content",
			Description: fmt.Sprintf("Post ID %s is gaining significant traction", content["id"]),
			Metrics: map[string]interface{}{
				"views":       content["views"],
				"likes":       content["likes"],
				"comments":    content["comments"],
				"shares":      content["shares"],
				"growth_rate": content["growth_rate"],
			},
			Confidence: 0.9,
			Priority:   "high",
		}
		insights = append(insights, insight)
	}

	// Add insights from trending hashtags
	for _, hashtag := range trendingHashtags {
		insight := Insight{
			Type:        "trending_hashtag",
			Title:       "Trending Hashtag",
			Description: fmt.Sprintf("Hashtag #%s is gaining popularity", hashtag["name"]),
			Metrics: map[string]interface{}{
				"posts":       hashtag["post_count"],
				"users":       hashtag["user_count"],
				"growth_rate": hashtag["growth_rate"],
			},
			Confidence: 0.85,
			Priority:   "medium",
		}
		insights = append(insights, insight)
	}

	// Add user growth insights
	if userGrowthTrends != nil {
		growthRate, ok := userGrowthTrends["growth_rate"].(float64)
		if ok && growthRate > 0 {
			insight := Insight{
				Type:        "user_growth",
				Title:       "User Growth",
				Description: fmt.Sprintf("User growth rate is %.2f%% over the past week", growthRate*100),
				Metrics:     userGrowthTrends,
				Confidence:  0.95,
				Priority:    "high",
			}
			insights = append(insights, insight)
		}
	}

	// Limit insights to requested count
	if len(insights) > limit {
		insights = insights[:limit]
	}

	// Cache the results
	insightsJSON, err := json.Marshal(insights)
	if err == nil {
		// Cache for 1 hour
		s.cache.SetWithExpiration(ctx, cacheKey, string(insightsJSON), 1*time.Hour)
	}

	return insights, nil
}

// GetContentRecommendations gets content recommendations based on insights
func (s *InsightsService) GetContentRecommendations(ctx context.Context, userID primitive.ObjectID, limit int) ([]Recommendation, error) {
	// Try to get from cache first
	cacheKey := fmt.Sprintf("content_recommendations:%s:%d", userID.Hex(), limit)
	cachedRecommendations, err := s.cache.Get(ctx, cacheKey)
	if err == nil && cachedRecommendations != "" {
		var recommendations []Recommendation
		if err := json.Unmarshal([]byte(cachedRecommendations), &recommendations); err == nil {
			return recommendations, nil
		}
	}

	// Get user's content performance
	userContentPerformance, err := s.getUserContentPerformance(ctx, userID)
	if err != nil {
		s.log.Error("Failed to get user content performance", "error", err, "user_id", userID.Hex())
		return nil, err
	}

	// Get platform trends
	platformTrends, err := s.getPlatformTrends(ctx)
	if err != nil {
		s.log.Error("Failed to get platform trends", "error", err)
		// Continue despite error
	}

	// Get user's best performing content types
	bestContentTypes, err := s.getUserBestContentTypes(ctx, userID)
	if err != nil {
		s.log.Error("Failed to get best content types", "error", err, "user_id", userID.Hex())
		// Continue despite error
	}

	// Generate recommendations
	recommendations := []Recommendation{}

	// Add recommendations based on user's content performance
	if userContentPerformance != nil {
		if engagementRate, ok := userContentPerformance["engagement_rate"].(float64); ok {
			if engagementRate < 0.02 { // Less than 2% engagement rate
				recommendation := Recommendation{
					Type:        "engagement_improvement",
					Title:       "Improve Engagement Rate",
					Description: "Your content engagement rate is below average. Try posting more interactive content.",
					ActionText:  "Learn More",
					ActionURL:   "/insights/engagement-tips",
					Priority:    "high",
				}
				recommendations = append(recommendations, recommendation)
			}
		}

		if postFrequency, ok := userContentPerformance["post_frequency"].(float64); ok {
			if postFrequency < 2 { // Less than 2 posts per week
				recommendation := Recommendation{
					Type:        "posting_frequency",
					Title:       "Increase Posting Frequency",
					Description: "Posting more regularly can help increase your visibility and engagement.",
					ActionText:  "Create Post",
					ActionURL:   "/posts/create",
					Priority:    "medium",
				}
				recommendations = append(recommendations, recommendation)
			}
		}
	}

	// Add recommendations based on platform trends
	if platformTrends != nil {
		if trendingTopics, ok := platformTrends["trending_topics"].([]interface{}); ok && len(trendingTopics) > 0 {
			topicNames := make([]string, 0, len(trendingTopics))
			for _, topic := range trendingTopics {
				if topicMap, ok := topic.(map[string]interface{}); ok {
					if name, ok := topicMap["name"].(string); ok {
						topicNames = append(topicNames, name)
					}
				}
			}

			if len(topicNames) > 0 {
				recommendation := Recommendation{
					Type:        "trending_topics",
					Title:       "Leverage Trending Topics",
					Description: fmt.Sprintf("Consider creating content about trending topics like %s", joinStrings(topicNames, 3)),
					ActionText:  "Explore Trends",
					ActionURL:   "/trends",
					Priority:    "medium",
				}
				recommendations = append(recommendations, recommendation)
			}
		}

		if bestTime, ok := platformTrends["best_posting_time"].(string); ok {
			recommendation := Recommendation{
				Type:        "posting_time",
				Title:       "Optimize Posting Time",
				Description: fmt.Sprintf("Try posting at %s when engagement is highest", bestTime),
				Priority:    "low",
			}
			recommendations = append(recommendations, recommendation)
		}
	}

	// Add recommendations based on best content types
	if bestContentTypes != nil && len(bestContentTypes) > 0 {
		bestType := bestContentTypes[0]
		recommendation := Recommendation{
			Type:        "content_type",
			Title:       "Focus on Your Strengths",
			Description: fmt.Sprintf("Your %s content performs best. Consider creating more of this type.", bestType),
			ActionText:  "Create Content",
			ActionURL:   "/posts/create",
			Priority:    "medium",
		}
		recommendations = append(recommendations, recommendation)
	}

	// Limit recommendations to requested count
	if len(recommendations) > limit {
		recommendations = recommendations[:limit]
	}

	// Cache the results
	recommendationsJSON, err := json.Marshal(recommendations)
	if err == nil {
		// Cache for 24 hours
		s.cache.SetWithExpiration(ctx, cacheKey, string(recommendationsJSON), 24*time.Hour)
	}

	return recommendations, nil
}

// Helper methods would be implemented here...

// getUserInsights gets insights for a user
func (s *InsightsService) getUserInsights(ctx context.Context, userID primitive.ObjectID, period string, startTime, endTime time.Time) (*EntityInsights, error) {
	// Implementation for getting user insights
	return nil, nil
}

// getPostInsights gets insights for a post
func (s *InsightsService) getPostInsights(ctx context.Context, postID primitive.ObjectID, period string, startTime, endTime time.Time) (*EntityInsights, error) {
	// Implementation for getting post insights
	return nil, nil
}

// getGroupInsights gets insights for a group
func (s *InsightsService) getGroupInsights(ctx context.Context, groupID primitive.ObjectID, period string, startTime, endTime time.Time) (*EntityInsights, error) {
	// Implementation for getting group insights
	return nil, nil
}

// getEventInsights gets insights for an event
func (s *InsightsService) getEventInsights(ctx context.Context, eventID primitive.ObjectID, period string, startTime, endTime time.Time) (*EntityInsights, error) {
	// Implementation for getting event insights
	return nil, nil
}

// getHashtagInsights gets insights for a hashtag
func (s *InsightsService) getHashtagInsights(ctx context.Context, hashtagID primitive.ObjectID, period string, startTime, endTime time.Time) (*EntityInsights, error) {
	// Implementation for getting hashtag insights
	return nil, nil
}

// getTrendingContent gets trending content
func (s *InsightsService) getTrendingContent(ctx context.Context, limit int) ([]map[string]interface{}, error) {
	// Implementation for getting trending content
	return nil, nil
}

// getTrendingHashtags gets trending hashtags
func (s *InsightsService) getTrendingHashtags(ctx context.Context, limit int) ([]map[string]interface{}, error) {
	// Implementation for getting trending hashtags
	return nil, nil
}

// getUserGrowthTrends gets user growth trends
func (s *InsightsService) getUserGrowthTrends(ctx context.Context) (map[string]interface{}, error) {
	// Implementation for getting user growth trends
	return nil, nil
}

// getUserContentPerformance gets a user's content performance
func (s *InsightsService) getUserContentPerformance(ctx context.Context, userID primitive.ObjectID) (map[string]interface{}, error) {
	// Implementation for getting user content performance
	return nil, nil
}

// getPlatformTrends gets platform trends
func (s *InsightsService) getPlatformTrends(ctx context.Context) (map[string]interface{}, error) {
	// Implementation for getting platform trends
	return nil, nil
}

// getUserBestContentTypes gets a user's best performing content types
func (s *InsightsService) getUserBestContentTypes(ctx context.Context, userID primitive.ObjectID) ([]string, error) {
	// Implementation for getting user's best content types
	return nil, nil
}

// joinStrings joins a slice of strings with commas
func joinStrings(strs []string, limit int) string {
	if len(strs) == 0 {
		return ""
	}

	if len(strs) > limit {
		strs = strs[:limit]
	}

	result := strs[0]
	for i := 1; i < len(strs); i++ {
		if i == len(strs)-1 {
			result += " and " + strs[i]
		} else {
			result += ", " + strs[i]
		}
	}

	return result
}
