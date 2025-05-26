package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Notification represents a system notification to a user
type Notification struct {
	ID             primitive.ObjectID     `bson:"_id,omitempty" json:"id"`
	UserID         primitive.ObjectID     `bson:"user_id" json:"user_id"`
	Type           string                 `bson:"type" json:"type"`       // like, comment, follow, mention, etc.
	Actor          primitive.ObjectID     `bson:"actor" json:"actor"`     // User who triggered the notification
	Subject        string                 `bson:"subject" json:"subject"` // post, comment, profile, etc.
	SubjectID      primitive.ObjectID     `bson:"subject_id" json:"subject_id"`
	Message        string                 `bson:"message" json:"message"`
	SubjectPreview string                 `bson:"subject_preview,omitempty" json:"subject_preview,omitempty"`
	ImageURL       string                 `bson:"image_url,omitempty" json:"image_url,omitempty"`
	ActionURL      string                 `bson:"action_url" json:"action_url"` // Deep link to the relevant content
	IsRead         bool                   `bson:"is_read" json:"is_read"`
	ReadAt         *time.Time             `bson:"read_at,omitempty" json:"read_at,omitempty"`
	IsHidden       bool                   `bson:"is_hidden" json:"is_hidden"`
	IsSent         bool                   `bson:"is_sent" json:"is_sent"`
	SentAt         *time.Time             `bson:"sent_at,omitempty" json:"sent_at,omitempty"`
	DeliveryStatus map[string]string      `bson:"delivery_status,omitempty" json:"delivery_status,omitempty"` // Key is delivery method, value is status
	GroupKey       string                 `bson:"group_key,omitempty" json:"group_key,omitempty"`             // For grouping similar notifications
	Priority       string                 `bson:"priority" json:"priority"`                                   // high, normal, low
	ExpiresAt      *time.Time             `bson:"expires_at,omitempty" json:"expires_at,omitempty"`
	CreatedAt      time.Time              `bson:"created_at" json:"created_at"`
	UpdatedAt      time.Time              `bson:"updated_at" json:"updated_at"`
	Metadata       map[string]interface{} `bson:"metadata,omitempty" json:"metadata,omitempty"`
}
