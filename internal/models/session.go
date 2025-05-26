package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Session represents a user login session
type Session struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID       primitive.ObjectID `bson:"user_id" json:"user_id"`
	Token        string             `bson:"token" json:"token"`
	RefreshToken string             `bson:"refresh_token" json:"refresh_token"`
	UserAgent    string             `bson:"user_agent" json:"user_agent"`
	IPAddress    string             `bson:"ip_address" json:"ip_address"`
	Device       string             `bson:"device" json:"device"`
	Location     string             `bson:"location" json:"location"`
	ExpiresAt    time.Time          `bson:"expires_at" json:"expires_at"`
	CreatedAt    time.Time          `bson:"created_at" json:"created_at"`
	LastUsedAt   time.Time          `bson:"last_used_at" json:"last_used_at"`
	IsActive     bool               `bson:"is_active" json:"is_active"`
}

// UserSession represents a user browsing session for analytics
type UserSession struct {
	ID              primitive.ObjectID  `bson:"_id,omitempty" json:"id"`
	UserID          *primitive.ObjectID `bson:"user_id,omitempty" json:"user_id,omitempty"` // May be anonymous
	SessionID       string              `bson:"session_id" json:"session_id"`
	StartTime       time.Time           `bson:"start_time" json:"start_time"`
	EndTime         *time.Time          `bson:"end_time,omitempty" json:"end_time,omitempty"`
	Duration        int                 `bson:"duration,omitempty" json:"duration,omitempty"` // in seconds
	PageViews       int                 `bson:"page_views" json:"page_views"`
	Device          DeviceInfo          `bson:"device" json:"device"`
	IPAddress       string              `bson:"ip_address" json:"ip_address"`
	EntryPage       string              `bson:"entry_page" json:"entry_page"`
	ExitPage        string              `bson:"exit_page,omitempty" json:"exit_page,omitempty"`
	Referrer        string              `bson:"referrer,omitempty" json:"referrer,omitempty"`
	UTMParams       map[string]string   `bson:"utm_params,omitempty" json:"utm_params,omitempty"`
	Location        *Location           `bson:"location,omitempty" json:"location,omitempty"`
	IsAuthenticated bool                `bson:"is_authenticated" json:"is_authenticated"`
	Events          []string            `bson:"events,omitempty" json:"events,omitempty"` // IDs of related analytics events
}

// DeviceInfo contains information about the user's device
type DeviceInfo struct {
	Type           string `bson:"type" json:"type"` // mobile, tablet, desktop
	Brand          string `bson:"brand,omitempty" json:"brand,omitempty"`
	Model          string `bson:"model,omitempty" json:"model,omitempty"`
	OS             string `bson:"os" json:"os"`
	OSVersion      string `bson:"os_version" json:"os_version"`
	Browser        string `bson:"browser,omitempty" json:"browser,omitempty"`
	BrowserVersion string `bson:"browser_version,omitempty" json:"browser_version,omitempty"`
	AppVersion     string `bson:"app_version,omitempty" json:"app_version,omitempty"`
	ScreenSize     string `bson:"screen_size,omitempty" json:"screen_size,omitempty"`
}
