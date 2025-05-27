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

// UserAnalyticsService provides user analytics functionality
type UserAnalyticsService struct {
	db    *database.Database
	cache *database.RedisClient
	log   *logger.Logger
}

// NewUserAnalyticsService creates a new user analytics service
func NewUserAnalyticsService(db *database.Database, cache *database.RedisClient, log *logger.Logger) *UserAnalyticsService {
	return &UserAnalyticsService{
		db:    db,
		cache: cache,
		log:   log,
	}
}

// UserMetrics represents user metrics
type UserMetrics struct {
	Period                 string                 `json:"period"`
	TotalUsers             int                    `json:"total_users"`
	NewUsers               int                    `json:"new_users"`
	ActiveUsers            int                    `json:"active_users"`
	RetentionRate          float64                `json:"retention_rate"`
	ChurnRate              float64                `json:"churn_rate"`
	AverageSessionDuration float64                `json:"average_session_duration"`
	SessionsPerUser        float64                `json:"sessions_per_user"`
	UserGrowthRate         float64                `json:"user_growth_rate"`
	DemographicBreakdown   map[string]interface{} `json:"demographic_breakdown,omitempty"`
	GeographicDistribution map[string]interface{} `json:"geographic_distribution,omitempty"`
}

// UserEngagementMetrics represents user engagement metrics
type UserEngagementMetrics struct {
	UserID            primitive.ObjectID `json:"user_id"`
	Period            string             `json:"period"`
	PostCount         int                `json:"post_count"`
	TotalLikes        int                `json:"total_likes"`
	TotalComments     int                `json:"total_comments"`
	TotalShares       int                `json:"total_shares"`
	TotalViews        int                `json:"total_views"`
	TotalEngagements  int                `json:"total_engagements"`
	EngagementRate    float64            `json:"engagement_rate"`
	LikesGiven        int                `json:"likes_given"`
	CommentsGiven     int                `json:"comments_given"`
	SharesGiven       int                `json:"shares_given"`
	TotalInteractions int                `json:"total_interactions"`
}

// UserProfile represents a user's analytics profile
type UserProfile struct {
	UserID          primitive.ObjectID     `json:"user_id"`
	Username        string                 `json:"username"`
	JoinDate        time.Time              `json:"join_date"`
	LastActive      time.Time              `json:"last_active"`
	PostCount       int                    `json:"post_count"`
	FollowerCount   int                    `json:"follower_count"`
	FollowingCount  int                    `json:"following_count"`
	EngagementRate  float64                `json:"engagement_rate"`
	ActivityLevel   string                 `json:"activity_level"` // high, medium, low
	TopContentTypes []string               `json:"top_content_types,omitempty"`
	ActivityTrends  map[string]interface{} `json:"activity_trends,omitempty"`
	Demographics    map[string]interface{} `json:"demographics,omitempty"`
}

// GetAggregatedMetrics gets aggregated user metrics for a period
func (s *UserAnalyticsService) GetAggregatedMetrics(ctx context.Context, period string) (*UserMetrics, error) {
	// Determine time range based on period
	startTime, endTime, err := getTimeRangeForPeriod(period)
	if err != nil {
		return nil, err
	}

	// Try to get from cache first
	cacheKey := "user_metrics:" + period
	cachedMetrics, err := s.cache.Get(ctx, cacheKey)
	if err == nil && cachedMetrics != "" {
		var metrics UserMetrics
		if err := json.Unmarshal([]byte(cachedMetrics), &metrics); err == nil {
			return &metrics, nil
		}
	}

	// Fetch metrics from database
	metrics := &UserMetrics{
		Period: period,
	}

	// Get total users
	totalUsersFilter := bson.M{
		"created_at": bson.M{
			"$lte": endTime,
		},
		"deleted_at": nil,
	}
	totalUsers, err := s.db.CountDocuments(ctx, "users", totalUsersFilter)
	if err != nil {
		s.log.Error("Failed to count total users", "error", err)
		return nil, err
	}
	metrics.TotalUsers = int(totalUsers)

	// Get new users
	newUsersFilter := bson.M{
		"created_at": bson.M{
			"$gte": startTime,
			"$lte": endTime,
		},
		"deleted_at": nil,
	}
	newUsers, err := s.db.CountDocuments(ctx, "users", newUsersFilter)
	if err != nil {
		s.log.Error("Failed to count new users", "error", err)
		return nil, err
	}
	metrics.NewUsers = int(newUsers)

	// Get active users
	activeUsersFilter := bson.M{
		"last_active": bson.M{
			"$gte": startTime,
			"$lte": endTime,
		},
		"deleted_at": nil,
	}
	activeUsers, err := s.db.CountDocuments(ctx, "users", activeUsersFilter)
	if err != nil {
		s.log.Error("Failed to count active users", "error", err)
		return nil, err
	}
	metrics.ActiveUsers = int(activeUsers)

	// Calculate user growth rate
	previousPeriodStart, previousPeriodEnd, err := getPreviousPeriod(startTime, endTime)
	if err != nil {
		s.log.Error("Failed to get previous period", "error", err)
		// Continue despite error
	} else {
		previousPeriodFilter := bson.M{
			"created_at": bson.M{
				"$gte": previousPeriodStart,
				"$lte": previousPeriodEnd,
			},
			"deleted_at": nil,
		}
		previousPeriodUsers, err := s.db.CountDocuments(ctx, "users", previousPeriodFilter)
		if err != nil {
			s.log.Error("Failed to count previous period users", "error", err)
			// Continue despite error
		} else if previousPeriodUsers > 0 {
			metrics.UserGrowthRate = float64(metrics.NewUsers-int(previousPeriodUsers)) / float64(previousPeriodUsers)
		}
	}

	// Get session data for average duration
	sessionPipeline := []bson.M{
		{
			"$match": bson.M{
				"start_time": bson.M{
					"$gte": startTime,
					"$lte": endTime,
				},
				"end_time": bson.M{
					"$ne": nil,
				},
			},
		},
		{
			"$group": bson.M{
				"_id": nil,
				"total_duration": bson.M{
					"$sum": "$duration",
				},
				"total_sessions": bson.M{
					"$sum": 1,
				},
				"user_count": bson.M{
					"$addToSet": "$user_id",
				},
			},
		},
	}

	sessionResults, err := s.db.Aggregate(ctx, "user_sessions", sessionPipeline)
	if err != nil {
		s.log.Error("Failed to aggregate session data", "error", err)
		// Continue despite error
	} else if len(sessionResults) > 0 {
		totalDuration, _ := sessionResults[0]["total_duration"].(int32)
		totalSessions, _ := sessionResults[0]["total_sessions"].(int32)
		userCount := len(sessionResults[0]["user_count"].(primitive.A))

		if totalSessions > 0 {
			metrics.AverageSessionDuration = float64(totalDuration) / float64(totalSessions)
		}

		if userCount > 0 {
			metrics.SessionsPerUser = float64(totalSessions) / float64(userCount)
		}
	}

	// Calculate retention rate
	if err := s.calculateRetentionRate(ctx, metrics, startTime, endTime); err != nil {
		s.log.Error("Failed to calculate retention rate", "error", err)
		// Continue despite error
	}

	// Get demographic breakdown
	metrics.DemographicBreakdown, err = s.getDemographicBreakdown(ctx, startTime, endTime)
	if err != nil {
		s.log.Error("Failed to get demographic breakdown", "error", err)
		// Continue despite error
	}

	// Get geographic distribution
	metrics.GeographicDistribution, err = s.getGeographicDistribution(ctx, startTime, endTime)
	if err != nil {
		s.log.Error("Failed to get geographic distribution", "error", err)
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

// GetUserProfile gets a user's analytics profile
func (s *UserAnalyticsService) GetUserProfile(ctx context.Context, userID primitive.ObjectID) (*UserProfile, error) {
	// Try to get from cache first
	cacheKey := "user_profile:" + userID.Hex()
	cachedProfile, err := s.cache.Get(ctx, cacheKey)
	if err == nil && cachedProfile != "" {
		var profile UserProfile
		if err := json.Unmarshal([]byte(cachedProfile), &profile); err == nil {
			return &profile, nil
		}
	}

	// Get user data
	var user struct {
		ID         primitive.ObjectID `bson:"_id"`
		Username   string             `bson:"username"`
		CreatedAt  time.Time          `bson:"created_at"`
		LastActive time.Time          `bson:"last_active"`
	}

	err = s.db.FindOne(ctx, "users", bson.M{"_id": userID}, &user)
	if err != nil {
		s.log.Error("Failed to get user", "error", err, "user_id", userID.Hex())
		return nil, err
	}

	// Initialize user profile
	profile := &UserProfile{
		UserID:     userID,
		Username:   user.Username,
		JoinDate:   user.CreatedAt,
		LastActive: user.LastActive,
	}

	// Get post count
	postFilter := bson.M{
		"user_id":    userID,
		"deleted_at": nil,
	}
	postCount, err := s.db.CountDocuments(ctx, "posts", postFilter)
	if err != nil {
		s.log.Error("Failed to count posts", "error", err, "user_id", userID.Hex())
		// Continue despite error
	}
	profile.PostCount = int(postCount)

	// Get follower count
	followerFilter := bson.M{
		"following_id": userID,
		"status":       "accepted",
	}
	followerCount, err := s.db.CountDocuments(ctx, "follows", followerFilter)
	if err != nil {
		s.log.Error("Failed to count followers", "error", err, "user_id", userID.Hex())
		// Continue despite error
	}
	profile.FollowerCount = int(followerCount)

	// Get following count
	followingFilter := bson.M{
		"follower_id": userID,
		"status":      "accepted",
	}
	followingCount, err := s.db.CountDocuments(ctx, "follows", followingFilter)
	if err != nil {
		s.log.Error("Failed to count following", "error", err, "user_id", userID.Hex())
		// Continue despite error
	}
	profile.FollowingCount = int(followingCount)

	// Get engagement rate
	if err := s.calculateUserEngagementRate(ctx, profile); err != nil {
		s.log.Error("Failed to calculate engagement rate", "error", err, "user_id", userID.Hex())
		// Continue despite error
	}

	// Determine activity level
	profile.ActivityLevel = s.determineActivityLevel(profile)

	// Get top content types
	profile.TopContentTypes, err = s.getUserTopContentTypes(ctx, userID)
	if err != nil {
		s.log.Error("Failed to get top content types", "error", err, "user_id", userID.Hex())
		// Continue despite error
	}

	// Get activity trends
	profile.ActivityTrends, err = s.getUserActivityTrends(ctx, userID)
	if err != nil {
		s.log.Error("Failed to get activity trends", "error", err, "user_id", userID.Hex())
		// Continue despite error
	}

	// Get demographics
	profile.Demographics, err = s.getUserDemographics(ctx, userID)
	if err != nil {
		s.log.Error("Failed to get demographics", "error", err, "user_id", userID.Hex())
		// Continue despite error
	}

	// Cache the results
	profileJSON, err := json.Marshal(profile)
	if err == nil {
		// Cache for 6 hours
		s.cache.SetWithExpiration(ctx, cacheKey, string(profileJSON), 6*time.Hour)
	}

	return profile, nil
}

// GetUserGrowth gets user growth over a period
func (s *UserAnalyticsService) GetUserGrowth(ctx context.Context, period string, interval string) ([]UserGrowthPoint, error) {
	// Determine time range based on period
	startTime, endTime, err := getTimeRangeForPeriod(period)
	if err != nil {
		return nil, err
	}

	// Try to get from cache first
	cacheKey := "user_growth:" + period + ":" + interval
	cachedGrowth, err := s.cache.Get(ctx, cacheKey)
	if err == nil && cachedGrowth != "" {
		var growth []UserGrowthPoint
		if err := json.Unmarshal([]byte(cachedGrowth), &growth); err == nil {
			return growth, nil
		}
	}

	// Determine grouping based on interval
	var groupFormat string
	var timeFormat string
	switch interval {
	case "day":
		groupFormat = "%Y-%m-%d"
		timeFormat = "2006-01-02"
	case "week":
		groupFormat = "%Y-%U"
		timeFormat = "2006-W%V"
	case "month":
		groupFormat = "%Y-%m"
		timeFormat = "2006-01"
	default:
		return nil, fmt.Errorf("invalid interval: %s", interval)
	}

	// Aggregate user growth by interval
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
				"_id": bson.M{
					"$dateToString": bson.M{
						"format": groupFormat,
						"date":   "$created_at",
					},
				},
				"count": bson.M{"$sum": 1},
			},
		},
		{
			"$sort": bson.M{"_id": 1},
		},
	}

	results, err := s.db.Aggregate(ctx, "users", pipeline)
	if err != nil {
		s.log.Error("Failed to aggregate user growth", "error", err)
		return nil, err
	}

	// Format results
	growth := make([]UserGrowthPoint, 0, len(results))
	for _, result := range results {
		dateStr, _ := result["_id"].(string)
		count, _ := result["count"].(int32)

		var date time.Time
		if interval == "week" {
			// Special handling for week format
			year, week := 0, 0
			fmt.Sscanf(dateStr, "%d-%d", &year, &week)
			date = getDateOfISOWeek(year, week)
		} else {
			date, _ = time.Parse(timeFormat, dateStr)
		}

		growth = append(growth, UserGrowthPoint{
			Date:     date,
			NewUsers: int(count),
			Interval: interval,
		})
	}

	// Add cumulative total
	cumulativeTotal := 0
	for i := range growth {
		cumulativeTotal += growth[i].NewUsers
		growth[i].TotalUsers = cumulativeTotal
	}

	// Cache the results
	growthJSON, err := json.Marshal(growth)
	if err == nil {
		// Cache for 6 hours
		s.cache.SetWithExpiration(ctx, cacheKey, string(growthJSON), 6*time.Hour)
	}

	return growth, nil
}

// GetTopUsers gets the top users by various metrics
func (s *UserAnalyticsService) GetTopUsers(ctx context.Context, metric string, limit int) ([]TopUser, error) {
	// Try to get from cache first
	cacheKey := fmt.Sprintf("top_users:%s:%d", metric, limit)
	cachedUsers, err := s.cache.Get(ctx, cacheKey)
	if err == nil && cachedUsers != "" {
		var users []TopUser
		if err := json.Unmarshal([]byte(cachedUsers), &users); err == nil {
			return users, nil
		}
	}

	// Determine sort field based on metric
	var sortField string
	var collection string
	var filterField string
	var pipeline []bson.M

	switch metric {
	case "followers":
		// Count followers for each user
		pipeline = []bson.M{
			{
				"$match": bson.M{
					"status": "accepted",
				},
			},
			{
				"$group": bson.M{
					"_id":   "$following_id",
					"count": bson.M{"$sum": 1},
				},
			},
			{
				"$sort": bson.M{"count": -1},
			},
			{
				"$limit": limit,
			},
		}
		collection = "follows"
		filterField = "following_id"
	case "posts":
		// Count posts for each user
		pipeline = []bson.M{
			{
				"$match": bson.M{
					"deleted_at": nil,
				},
			},
			{
				"$group": bson.M{
					"_id":   "$user_id",
					"count": bson.M{"$sum": 1},
				},
			},
			{
				"$sort": bson.M{"count": -1},
			},
			{
				"$limit": limit,
			},
		}
		collection = "posts"
		filterField = "user_id"
	case "engagement":
		// Aggregate engagement for each user
		pipeline = []bson.M{
			{
				"$match": bson.M{
					"deleted_at": nil,
				},
			},
			{
				"$group": bson.M{
					"_id": "$user_id",
					"engagement": bson.M{
						"$sum": bson.M{
							"$add": []string{"$like_count", "$comment_count", "$share_count"},
						},
					},
					"posts": bson.M{"$sum": 1},
				},
			},
			{
				"$project": bson.M{
					"_id":        1,
					"engagement": 1,
					"posts":      1,
					"avg_engagement": bson.M{
						"$cond": bson.M{
							"if":   bson.M{"$eq": []interface{}{"$posts", 0}},
							"then": 0,
							"else": bson.M{"$divide": []string{"$engagement", "$posts"}},
						},
					},
				},
			},
			{
				"$sort": bson.M{"avg_engagement": -1},
			},
			{
				"$limit": limit,
			},
		}
		collection = "posts"
		filterField = "user_id"
	default:
		return nil, fmt.Errorf("invalid metric: %s", metric)
	}

	// Execute aggregation
	results, err := s.db.Aggregate(ctx, collection, pipeline)
	if err != nil {
		s.log.Error("Failed to aggregate top users", "error", err, "metric", metric)
		return nil, err
	}

	// Extract user IDs
	userIDs := make([]primitive.ObjectID, 0, len(results))
	userMetrics := make(map[string]int)
	userEngagement := make(map[string]float64)

	for _, result := range results {
		userID, _ := result["_id"].(primitive.ObjectID)
		userIDs = append(userIDs, userID)

		// Store metrics
		if metric == "engagement" {
			if avgEngagement, ok := result["avg_engagement"].(float64); ok {
				userEngagement[userID.Hex()] = avgEngagement
			}
		} else {
			if count, ok := result["count"].(int32); ok {
				userMetrics[userID.Hex()] = int(count)
			}
		}
	}

	// Get user details
	userMap, err := s.getUserDetails(ctx, userIDs)
	if err != nil {
		s.log.Error("Failed to get user details", "error", err)
		return nil, err
	}

	// Create top users list
	topUsers := make([]TopUser, 0, len(results))
	for _, userID := range userIDs {
		user, exists := userMap[userID.Hex()]
		if !exists {
			continue
		}

		topUser := TopUser{
			UserID:       userID,
			Username:     user.Username,
			DisplayName:  user.DisplayName,
			ProfileImage: user.ProfileImage,
		}

		if metric == "engagement" {
			topUser.Value = userEngagement[userID.Hex()]
		} else {
			topUser.Value = float64(userMetrics[userID.Hex()])
		}

		topUsers = append(topUsers, topUser)
	}

	// Cache the results
	topUsersJSON, err := json.Marshal(topUsers)
	if err == nil {
		// Cache for 6 hours
		s.cache.SetWithExpiration(ctx, cacheKey, string(topUsersJSON), 6*time.Hour)
	}

	return topUsers, nil
}

// Helper methods would be implemented here...

// UserGrowthPoint represents a point in the user growth chart
type UserGrowthPoint struct {
	Date       time.Time `json:"date"`
	NewUsers   int       `json:"new_users"`
	TotalUsers int       `json:"total_users"`
	Interval   string    `json:"interval"`
}

// TopUser represents a top user by some metric
type TopUser struct {
	UserID       primitive.ObjectID `json:"user_id"`
	Username     string             `json:"username"`
	DisplayName  string             `json:"display_name"`
	ProfileImage string             `json:"profile_image"`
	Value        float64            `json:"value"`
}

// calculateRetentionRate calculates retention rate for the given period
func (s *UserAnalyticsService) calculateRetentionRate(ctx context.Context, metrics *UserMetrics, startTime, endTime time.Time) error {
	// Implementation for calculating retention rate
	return nil
}

// getDemographicBreakdown gets demographic breakdown of users
func (s *UserAnalyticsService) getDemographicBreakdown(ctx context.Context, startTime, endTime time.Time) (map[string]interface{}, error) {
	// Implementation for getting demographic breakdown
	return nil, nil
}

// getGeographicDistribution gets geographic distribution of users
func (s *UserAnalyticsService) getGeographicDistribution(ctx context.Context, startTime, endTime time.Time) (map[string]interface{}, error) {
	// Implementation for getting geographic distribution
	return nil, nil
}

// calculateUserEngagementRate calculates a user's engagement rate
func (s *UserAnalyticsService) calculateUserEngagementRate(ctx context.Context, profile *UserProfile) error {
	// Implementation for calculating user engagement rate
	return nil
}

// determineActivityLevel determines a user's activity level
func (s *UserAnalyticsService) determineActivityLevel(profile *UserProfile) string {
	// Implementation for determining activity level
	return ""
}

// getUserTopContentTypes gets a user's top content types
func (s *UserAnalyticsService) getUserTopContentTypes(ctx context.Context, userID primitive.ObjectID) ([]string, error) {
	// Implementation for getting user's top content types
	return nil, nil
}

// getUserActivityTrends gets a user's activity trends
func (s *UserAnalyticsService) getUserActivityTrends(ctx context.Context, userID primitive.ObjectID) (map[string]interface{}, error) {
	// Implementation for getting user's activity trends
	return nil, nil
}

// getUserDemographics gets a user's demographics
func (s *UserAnalyticsService) getUserDemographics(ctx context.Context, userID primitive.ObjectID) (map[string]interface{}, error) {
	// Implementation for getting user's demographics
	return nil, nil
}

// getUserDetails gets details for multiple users
func (s *UserAnalyticsService) getUserDetails(ctx context.Context, userIDs []primitive.ObjectID) (map[string]struct{ Username, DisplayName, ProfileImage string }, error) {
	// Implementation for getting user details
	return nil, nil
}

// getDateOfISOWeek gets the date of a specific ISO week
func getDateOfISOWeek(year, week int) time.Time {
	// Implementation for getting date of ISO week
	return time.Now()
}
