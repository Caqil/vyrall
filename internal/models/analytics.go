package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// AnalyticsEvent represents a tracked user action
type AnalyticsEvent struct {
	ID         primitive.ObjectID     `bson:"_id,omitempty" json:"id"`
	UserID     *primitive.ObjectID    `bson:"user_id,omitempty" json:"user_id,omitempty"`
	SessionID  string                 `bson:"session_id" json:"session_id"`
	EventType  string                 `bson:"event_type" json:"event_type"`                       // page_view, post_like, profile_visit, etc.
	EntityType string                 `bson:"entity_type,omitempty" json:"entity_type,omitempty"` // post, user, comment, etc.
	EntityID   *primitive.ObjectID    `bson:"entity_id,omitempty" json:"entity_id,omitempty"`
	Properties map[string]interface{} `bson:"properties,omitempty" json:"properties,omitempty"`
	Device     DeviceInfo             `bson:"device" json:"device"`
	Location   *LocationInfo          `bson:"location,omitempty" json:"location,omitempty"`
	Referrer   string                 `bson:"referrer,omitempty" json:"referrer,omitempty"`
	UTMParams  map[string]string      `bson:"utm_params,omitempty" json:"utm_params,omitempty"`
	IPAddress  string                 `bson:"ip_address" json:"ip_address"`
	CreatedAt  time.Time              `bson:"created_at" json:"created_at"`
}

// LocationInfo contains geographical information about the user
type LocationInfo struct {
	Country     string    `bson:"country,omitempty" json:"country,omitempty"`
	Region      string    `bson:"region,omitempty" json:"region,omitempty"`
	City        string    `bson:"city,omitempty" json:"city,omitempty"`
	PostalCode  string    `bson:"postal_code,omitempty" json:"postal_code,omitempty"`
	Timezone    string    `bson:"timezone,omitempty" json:"timezone,omitempty"`
	Coordinates *GeoPoint `bson:"coordinates,omitempty" json:"coordinates,omitempty"`
}

// UserAnalytics contains metrics about user activity and engagement
type UserAnalytics struct {
	ID                  primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID              primitive.ObjectID `bson:"user_id" json:"user_id"`
	ProfileViews        int                `bson:"profile_views" json:"profile_views"`
	PostImpressions     int                `bson:"post_impressions" json:"post_impressions"`
	EngagementRate      float64            `bson:"engagement_rate" json:"engagement_rate"`
	TopPerformingPost   string             `bson:"top_performing_post" json:"top_performing_post"`
	LastProfileViewedAt time.Time          `bson:"last_profile_viewed_at" json:"last_profile_viewed_at"`
	ActiveDaysThisMonth int                `bson:"active_days_this_month" json:"active_days_this_month"`
	Period              string             `bson:"period" json:"period"` // daily, weekly, monthly
	StartDate           time.Time          `bson:"start_date" json:"start_date"`
	EndDate             time.Time          `bson:"end_date" json:"end_date"`
	UpdatedAt           time.Time          `bson:"updated_at" json:"updated_at"`
}

// PostAnalytics contains detailed metrics about a post's performance
type PostAnalytics struct {
	ID                   primitive.ObjectID     `bson:"_id,omitempty" json:"id"`
	PostID               primitive.ObjectID     `bson:"post_id" json:"post_id"`
	Impressions          int                    `bson:"impressions" json:"impressions"`
	Reach                int                    `bson:"reach" json:"reach"`
	EngagementRate       float64                `bson:"engagement_rate" json:"engagement_rate"`
	Clicks               int                    `bson:"clicks" json:"clicks"`
	Shares               int                    `bson:"shares" json:"shares"`
	Saves                int                    `bson:"saves" json:"saves"`
	ViewsBreakdown       map[string]int         `bson:"views_breakdown" json:"views_breakdown"` // By platform, device, etc.
	AudienceDemographics map[string]interface{} `bson:"audience_demographics,omitempty" json:"audience_demographics,omitempty"`
	TopReferrers         []string               `bson:"top_referrers,omitempty" json:"top_referrers,omitempty"`
	Period               string                 `bson:"period" json:"period"` // hourly, daily, weekly, all_time
	StartDate            time.Time              `bson:"start_date" json:"start_date"`
	EndDate              time.Time              `bson:"end_date" json:"end_date"`
	UpdatedAt            time.Time              `bson:"updated_at" json:"updated_at"`
}

// GroupAnalytics contains metrics about a group's activity
type GroupAnalytics struct {
	ID                 primitive.ObjectID   `bson:"_id,omitempty" json:"id"`
	GroupID            primitive.ObjectID   `bson:"group_id" json:"group_id"`
	DailyActiveUsers   int                  `bson:"daily_active_users" json:"daily_active_users"`
	MonthlyActiveUsers int                  `bson:"monthly_active_users" json:"monthly_active_users"`
	PostsThisWeek      int                  `bson:"posts_this_week" json:"posts_this_week"`
	CommentsThisWeek   int                  `bson:"comments_this_week" json:"comments_this_week"`
	GrowthRate         float64              `bson:"growth_rate" json:"growth_rate"` // % increase in members
	EngagementRate     float64              `bson:"engagement_rate" json:"engagement_rate"`
	TopContributors    []primitive.ObjectID `bson:"top_contributors,omitempty" json:"top_contributors,omitempty"`
	PopularTopics      []string             `bson:"popular_topics,omitempty" json:"popular_topics,omitempty"`
	PeakActivityTimes  []string             `bson:"peak_activity_times,omitempty" json:"peak_activity_times,omitempty"`
	Period             string               `bson:"period" json:"period"` // daily, weekly, monthly
	StartDate          time.Time            `bson:"start_date" json:"start_date"`
	EndDate            time.Time            `bson:"end_date" json:"end_date"`
	UpdatedAt          time.Time            `bson:"updated_at" json:"updated_at"`
}

// EventAnalytics contains metrics about an event
type EventAnalytics struct {
	ID               primitive.ObjectID     `bson:"_id,omitempty" json:"id"`
	EventID          primitive.ObjectID     `bson:"event_id" json:"event_id"`
	ViewCount        int                    `bson:"view_count" json:"view_count"`
	UniqueViewCount  int                    `bson:"unique_view_count" json:"unique_view_count"`
	ShareCount       int                    `bson:"share_count" json:"share_count"`
	ClickThroughRate float64                `bson:"click_through_rate" json:"click_through_rate"`
	ConversionRate   float64                `bson:"conversion_rate" json:"conversion_rate"`
	ReferralSources  map[string]int         `bson:"referral_sources,omitempty" json:"referral_sources,omitempty"`
	TicketRevenue    float64                `bson:"ticket_revenue,omitempty" json:"ticket_revenue,omitempty"`
	CheckInRate      float64                `bson:"check_in_rate,omitempty" json:"check_in_rate,omitempty"`
	Demographics     map[string]interface{} `bson:"demographics,omitempty" json:"demographics,omitempty"`
	UpdatedAt        time.Time              `bson:"updated_at" json:"updated_at"`
}
