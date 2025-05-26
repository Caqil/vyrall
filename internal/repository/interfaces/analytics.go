package interfaces

import (
	"context"
	"time"

	"github.com/yourusername/social-media-platform/internal/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// AnalyticsRepository defines the interface for analytics data access
type AnalyticsRepository interface {
	// Analytics events
	CreateEvent(ctx context.Context, event *models.AnalyticsEvent) (primitive.ObjectID, error)
	GetEventsByType(ctx context.Context, eventType string, startDate, endDate time.Time, limit, offset int) ([]*models.AnalyticsEvent, int, error)
	GetEventsByUserID(ctx context.Context, userID primitive.ObjectID, eventTypes []string, startDate, endDate time.Time, limit, offset int) ([]*models.AnalyticsEvent, int, error)
	GetEventsByEntityID(ctx context.Context, entityType string, entityID primitive.ObjectID, startDate, endDate time.Time, limit, offset int) ([]*models.AnalyticsEvent, int, error)

	// User sessions
	CreateSession(ctx context.Context, session *models.UserSession) (primitive.ObjectID, error)
	UpdateSession(ctx context.Context, sessionID string, endTime time.Time, duration int, pageViews int, exitPage string) error
	GetSessionsByUserID(ctx context.Context, userID primitive.ObjectID, startDate, endDate time.Time, limit, offset int) ([]*models.UserSession, int, error)

	// User analytics
	CreateOrUpdateUserAnalytics(ctx context.Context, analytics *models.UserAnalytics) (primitive.ObjectID, error)
	GetUserAnalytics(ctx context.Context, userID primitive.ObjectID, period string, startDate, endDate time.Time) (*models.UserAnalytics, error)
	IncrementUserProfileViews(ctx context.Context, userID primitive.ObjectID) error
	IncrementUserPostImpressions(ctx context.Context, userID primitive.ObjectID, amount int) error
	UpdateUserEngagementRate(ctx context.Context, userID primitive.ObjectID, rate float64) error

	// Post analytics
	CreateOrUpdatePostAnalytics(ctx context.Context, analytics *models.PostAnalytics) (primitive.ObjectID, error)
	GetPostAnalytics(ctx context.Context, postID primitive.ObjectID, period string, startDate, endDate time.Time) (*models.PostAnalytics, error)
	IncrementPostImpressions(ctx context.Context, postID primitive.ObjectID, amount int) error
	IncrementPostReach(ctx context.Context, postID primitive.ObjectID, amount int) error
	UpdatePostEngagementRate(ctx context.Context, postID primitive.ObjectID, rate float64) error

	// Group analytics
	CreateOrUpdateGroupAnalytics(ctx context.Context, analytics *models.GroupAnalytics) (primitive.ObjectID, error)
	GetGroupAnalytics(ctx context.Context, groupID primitive.ObjectID, period string, startDate, endDate time.Time) (*models.GroupAnalytics, error)

	// Event analytics
	CreateOrUpdateEventAnalytics(ctx context.Context, analytics *models.EventAnalytics) (primitive.ObjectID, error)
	GetEventAnalytics(ctx context.Context, eventID primitive.ObjectID) (*models.EventAnalytics, error)

	// Aggregated analytics
	GetTopPerformingContent(ctx context.Context, contentType string, metric string, startDate, endDate time.Time, limit int) ([]map[string]interface{}, error)
	GetPlatformGrowth(ctx context.Context, period string, startDate, endDate time.Time) (map[string]interface{}, error)
	GetEngagementMetrics(ctx context.Context, period string, startDate, endDate time.Time) (map[string]interface{}, error)
	GetUserDemographics(ctx context.Context) (map[string]interface{}, error)
	GetRetentionMetrics(ctx context.Context, cohortDate time.Time, periods int) (map[string]interface{}, error)

	// Real-time analytics
	GetCurrentActiveUsers(ctx context.Context) (int, error)
	GetCurrentActiveSessions(ctx context.Context) (int, error)
	GetRealtimeEngagement(ctx context.Context, minutes int) (map[string]interface{}, error)

	// Location-based analytics
	GetUsersByLocation(ctx context.Context) (map[string]int, error)
	GetEventsByLocation(ctx context.Context, eventType string, startDate, endDate time.Time) (map[string]int, error)

	// Device analytics
	GetUsersByDevice(ctx context.Context, startDate, endDate time.Time) (map[string]int, error)
	GetSessionsByDevice(ctx context.Context, startDate, endDate time.Time) (map[string]int, error)

	// Time-based analytics
	GetActivityByTimeOfDay(ctx context.Context, eventType string, startDate, endDate time.Time) (map[int]int, error)
	GetActivityByDayOfWeek(ctx context.Context, eventType string, startDate, endDate time.Time) (map[int]int, error)

	// Data cleanup
	DeleteOldEvents(ctx context.Context, before time.Time) (int, error)
	DeleteOldSessions(ctx context.Context, before time.Time) (int, error)
}
