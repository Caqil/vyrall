package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Ad represents an advertisement
type Ad struct {
	ID              primitive.ObjectID  `bson:"_id,omitempty" json:"id"`
	AdvertiserID    primitive.ObjectID  `bson:"advertiser_id" json:"advertiser_id"`
	CampaignID      primitive.ObjectID  `bson:"campaign_id" json:"campaign_id"`
	Title           string              `bson:"title" json:"title"`
	Description     string              `bson:"description" json:"description"`
	MediaFiles      []Media             `bson:"media_files" json:"media_files"`
	DestinationURL  string              `bson:"destination_url" json:"destination_url"`
	CallToAction    string              `bson:"call_to_action,omitempty" json:"call_to_action,omitempty"`
	Status          string              `bson:"status" json:"status"` // draft, pending_review, approved, rejected, paused, active, completed
	RejectionReason string              `bson:"rejection_reason,omitempty" json:"rejection_reason,omitempty"`
	Format          string              `bson:"format" json:"format"`       // feed, story, banner, sidebar, etc.
	Placement       string              `bson:"placement" json:"placement"` // feed, profile, story, search, etc.
	TargetAudience  AdTargeting         `bson:"target_audience" json:"target_audience"`
	Budget          AdBudget            `bson:"budget" json:"budget"`
	Metrics         AdMetrics           `bson:"metrics,omitempty" json:"metrics,omitempty"`
	StartDate       time.Time           `bson:"start_date" json:"start_date"`
	EndDate         time.Time           `bson:"end_date" json:"end_date"`
	Schedule        []AdSchedule        `bson:"schedule,omitempty" json:"schedule,omitempty"`
	IsAutomated     bool                `bson:"is_automated" json:"is_automated"`
	Keywords        []string            `bson:"keywords,omitempty" json:"keywords,omitempty"`
	CreatedAt       time.Time           `bson:"created_at" json:"created_at"`
	UpdatedAt       time.Time           `bson:"updated_at" json:"updated_at"`
	ReviewedAt      *time.Time          `bson:"reviewed_at,omitempty" json:"reviewed_at,omitempty"`
	ReviewedBy      *primitive.ObjectID `bson:"reviewed_by,omitempty" json:"reviewed_by,omitempty"`
}

// AdTargeting represents targeting parameters for an advertisement
type AdTargeting struct {
	AgeRange         []int                  `bson:"age_range,omitempty" json:"age_range,omitempty"` // [min, max]
	Genders          []string               `bson:"genders,omitempty" json:"genders,omitempty"`
	Locations        []string               `bson:"locations,omitempty" json:"locations,omitempty"` // Country, state, city
	Languages        []string               `bson:"languages,omitempty" json:"languages,omitempty"`
	Interests        []string               `bson:"interests,omitempty" json:"interests,omitempty"`
	IncludeKeywords  []string               `bson:"include_keywords,omitempty" json:"include_keywords,omitempty"`
	ExcludeKeywords  []string               `bson:"exclude_keywords,omitempty" json:"exclude_keywords,omitempty"`
	IncludeAudiences []primitive.ObjectID   `bson:"include_audiences,omitempty" json:"include_audiences,omitempty"`
	ExcludeAudiences []primitive.ObjectID   `bson:"exclude_audiences,omitempty" json:"exclude_audiences,omitempty"`
	Devices          []string               `bson:"devices,omitempty" json:"devices,omitempty"` // mobile, desktop, tablet
	PlatformVersions []string               `bson:"platform_versions,omitempty" json:"platform_versions,omitempty"`
	CustomAttributes map[string]interface{} `bson:"custom_attributes,omitempty" json:"custom_attributes,omitempty"`
}

// AdBudget represents budget information for an advertisement
type AdBudget struct {
	TotalAmount     float64 `bson:"total_amount" json:"total_amount"`
	Currency        string  `bson:"currency" json:"currency"`
	DailyLimit      float64 `bson:"daily_limit,omitempty" json:"daily_limit,omitempty"`
	RemainingBudget float64 `bson:"remaining_budget" json:"remaining_budget"`
	SpentAmount     float64 `bson:"spent_amount" json:"spent_amount"`
	BidType         string  `bson:"bid_type" json:"bid_type"` // cpc, cpm, cpa, etc.
	BidAmount       float64 `bson:"bid_amount" json:"bid_amount"`
	IsAutomatic     bool    `bson:"is_automatic" json:"is_automatic"` // Auto bid optimization
}

// AdMetrics represents performance metrics for an advertisement
type AdMetrics struct {
	Impressions           int     `bson:"impressions" json:"impressions"`
	Clicks                int     `bson:"clicks" json:"clicks"`
	CTR                   float64 `bson:"ctr" json:"ctr"` // Click-through rate
	UniqueReach           int     `bson:"unique_reach" json:"unique_reach"`
	Conversions           int     `bson:"conversions,omitempty" json:"conversions,omitempty"`
	ConversionRate        float64 `bson:"conversion_rate,omitempty" json:"conversion_rate,omitempty"`
	CostPerClick          float64 `bson:"cost_per_click" json:"cost_per_click"`
	CostPerImpression     float64 `bson:"cost_per_impression" json:"cost_per_impression"`
	CostPerConversion     float64 `bson:"cost_per_conversion,omitempty" json:"cost_per_conversion,omitempty"`
	EngagementRate        float64 `bson:"engagement_rate,omitempty" json:"engagement_rate,omitempty"`
	VideoViews            int     `bson:"video_views,omitempty" json:"video_views,omitempty"`
	VideoCompletionRate   float64 `bson:"video_completion_rate,omitempty" json:"video_completion_rate,omitempty"`
	AverageWatchTime      float64 `bson:"average_watch_time,omitempty" json:"average_watch_time,omitempty"`
	TargetAudienceReached float64 `bson:"target_audience_reached" json:"target_audience_reached"` // percentage
	ROI                   float64 `bson:"roi,omitempty" json:"roi,omitempty"`                     // Return on investment
}

// AdSchedule represents a time schedule for ad display
type AdSchedule struct {
	DaysOfWeek []int  `bson:"days_of_week" json:"days_of_week"` // 0-6, Sunday to Saturday
	StartTime  string `bson:"start_time" json:"start_time"`     // HH:MM format in 24 hour time
	EndTime    string `bson:"end_time" json:"end_time"`         // HH:MM format in 24 hour time
	TimeZone   string `bson:"time_zone,omitempty" json:"time_zone,omitempty"`
}

// Campaign represents an advertising campaign
type Campaign struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	AdvertiserID    primitive.ObjectID `bson:"advertiser_id" json:"advertiser_id"`
	Name            string             `bson:"name" json:"name"`
	Description     string             `bson:"description,omitempty" json:"description,omitempty"`
	Objective       string             `bson:"objective" json:"objective"` // awareness, consideration, conversion
	Status          string             `bson:"status" json:"status"`       // draft, active, paused, completed
	StartDate       time.Time          `bson:"start_date" json:"start_date"`
	EndDate         time.Time          `bson:"end_date" json:"end_date"`
	TotalBudget     float64            `bson:"total_budget" json:"total_budget"`
	Currency        string             `bson:"currency" json:"currency"`
	SpentBudget     float64            `bson:"spent_budget" json:"spent_budget"`
	RemainingBudget float64            `bson:"remaining_budget" json:"remaining_budget"`
	DailyBudget     float64            `bson:"daily_budget,omitempty" json:"daily_budget,omitempty"`
	AdCount         int                `bson:"ad_count" json:"ad_count"`
	Tags            []string           `bson:"tags,omitempty" json:"tags,omitempty"`
	UTMParameters   map[string]string  `bson:"utm_parameters,omitempty" json:"utm_parameters,omitempty"`
	CreatedAt       time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt       time.Time          `bson:"updated_at" json:"updated_at"`
}
