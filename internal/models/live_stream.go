package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// LiveStream represents a live video streaming session
type LiveStream struct {
	ID                 primitive.ObjectID     `bson:"_id,omitempty" json:"id"`
	UserID             primitive.ObjectID     `bson:"user_id" json:"user_id"`
	Title              string                 `bson:"title" json:"title"`
	Description        string                 `bson:"description,omitempty" json:"description,omitempty"`
	ThumbnailURL       string                 `bson:"thumbnail_url,omitempty" json:"thumbnail_url,omitempty"`
	Status             string                 `bson:"status" json:"status"`   // scheduled, live, ended, failed
	Privacy            string                 `bson:"privacy" json:"privacy"` // public, followers, private
	AllowedViewers     []primitive.ObjectID   `bson:"allowed_viewers,omitempty" json:"allowed_viewers,omitempty"`
	ScheduledStartTime *time.Time             `bson:"scheduled_start_time,omitempty" json:"scheduled_start_time,omitempty"`
	ActualStartTime    *time.Time             `bson:"actual_start_time,omitempty" json:"actual_start_time,omitempty"`
	EndTime            *time.Time             `bson:"end_time,omitempty" json:"end_time,omitempty"`
	Duration           int                    `bson:"duration,omitempty" json:"duration,omitempty"` // in seconds
	ViewerCount        int                    `bson:"viewer_count" json:"viewer_count"`
	PeakViewerCount    int                    `bson:"peak_viewer_count" json:"peak_viewer_count"`
	TotalViews         int                    `bson:"total_views" json:"total_views"` // Including replay views
	LikeCount          int                    `bson:"like_count" json:"like_count"`
	CommentCount       int                    `bson:"comment_count" json:"comment_count"`
	ShareCount         int                    `bson:"share_count" json:"share_count"`
	StreamURL          string                 `bson:"stream_url" json:"stream_url"`
	PlaybackURL        string                 `bson:"playback_url,omitempty" json:"playback_url,omitempty"`
	RTMPURL            string                 `bson:"rtmp_url,omitempty" json:"rtmp_url,omitempty"`
	StreamKey          string                 `bson:"stream_key,omitempty" json:"stream_key,omitempty"`
	RecordingEnabled   bool                   `bson:"recording_enabled" json:"recording_enabled"`
	RecordingURL       string                 `bson:"recording_url,omitempty" json:"recording_url,omitempty"`
	Tags               []string               `bson:"tags,omitempty" json:"tags,omitempty"`
	Categories         []string               `bson:"categories,omitempty" json:"categories,omitempty"`
	Location           *Location              `bson:"location,omitempty" json:"location,omitempty"`
	IsMonetized        bool                   `bson:"is_monetized" json:"is_monetized"`
	ChatEnabled        bool                   `bson:"chat_enabled" json:"chat_enabled"`
	ChatSettings       LiveStreamChatSettings `bson:"chat_settings" json:"chat_settings"`
	CreatedAt          time.Time              `bson:"created_at" json:"created_at"`
	UpdatedAt          time.Time              `bson:"updated_at" json:"updated_at"`
}

// LiveStreamChatSettings defines chat configuration for a live stream
type LiveStreamChatSettings struct {
	ModeratedMode       bool                 `bson:"moderated_mode" json:"moderated_mode"`
	SlowMode            bool                 `bson:"slow_mode" json:"slow_mode"`
	SlowModeInterval    int                  `bson:"slow_mode_interval,omitempty" json:"slow_mode_interval,omitempty"` // seconds between messages
	FollowersOnlyMode   bool                 `bson:"followers_only_mode" json:"followers_only_mode"`
	SubscribersOnlyMode bool                 `bson:"subscribers_only_mode" json:"subscribers_only_mode"`
	ModeratorIDs        []primitive.ObjectID `bson:"moderator_ids,omitempty" json:"moderator_ids,omitempty"`
	BannedUsers         []primitive.ObjectID `bson:"banned_users,omitempty" json:"banned_users,omitempty"`
	BannedKeywords      []string             `bson:"banned_keywords,omitempty" json:"banned_keywords,omitempty"`
	RequireApproval     bool                 `bson:"require_approval" json:"require_approval"`
}

// LiveStreamViewer represents a viewer of a live stream
type LiveStreamViewer struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	LiveStreamID  primitive.ObjectID `bson:"live_stream_id" json:"live_stream_id"`
	UserID        primitive.ObjectID `bson:"user_id" json:"user_id"`
	JoinedAt      time.Time          `bson:"joined_at" json:"joined_at"`
	LeftAt        *time.Time         `bson:"left_at,omitempty" json:"left_at,omitempty"`
	Duration      int                `bson:"duration,omitempty" json:"duration,omitempty"` // in seconds
	Device        string             `bson:"device,omitempty" json:"device,omitempty"`
	Platform      string             `bson:"platform,omitempty" json:"platform,omitempty"`
	IsActive      bool               `bson:"is_active" json:"is_active"`
	CommentCount  int                `bson:"comment_count" json:"comment_count"`
	ReactionCount int                `bson:"reaction_count" json:"reaction_count"`
}

// LiveStreamComment represents a comment on a live stream
type LiveStreamComment struct {
	ID              primitive.ObjectID   `bson:"_id,omitempty" json:"id"`
	LiveStreamID    primitive.ObjectID   `bson:"live_stream_id" json:"live_stream_id"`
	UserID          primitive.ObjectID   `bson:"user_id" json:"user_id"`
	Content         string               `bson:"content" json:"content"`
	TimestampSec    int                  `bson:"timestamp_sec" json:"timestamp_sec"` // Seconds from stream start
	IsPinned        bool                 `bson:"is_pinned" json:"is_pinned"`
	IsHidden        bool                 `bson:"is_hidden" json:"is_hidden"`
	IsModerated     bool                 `bson:"is_moderated" json:"is_moderated"`
	ModeratedBy     *primitive.ObjectID  `bson:"moderated_by,omitempty" json:"moderated_by,omitempty"`
	ModeratedAt     *time.Time           `bson:"moderated_at,omitempty" json:"moderated_at,omitempty"`
	ModeratedReason string               `bson:"moderated_reason,omitempty" json:"moderated_reason,omitempty"`
	ReactionCount   int                  `bson:"reaction_count" json:"reaction_count"`
	Reactions       []LiveStreamReaction `bson:"reactions,omitempty" json:"reactions,omitempty"`
	CreatedAt       time.Time            `bson:"created_at" json:"created_at"`
}

// LiveStreamReaction represents a reaction to a live stream or comment
type LiveStreamReaction struct {
	ID           primitive.ObjectID  `bson:"_id,omitempty" json:"id"`
	LiveStreamID primitive.ObjectID  `bson:"live_stream_id" json:"live_stream_id"`
	UserID       primitive.ObjectID  `bson:"user_id" json:"user_id"`
	Type         string              `bson:"type" json:"type"`                                   // like, heart, etc.
	TimestampSec int                 `bson:"timestamp_sec" json:"timestamp_sec"`                 // Seconds from stream start
	TargetType   string              `bson:"target_type,omitempty" json:"target_type,omitempty"` // stream or comment
	TargetID     *primitive.ObjectID `bson:"target_id,omitempty" json:"target_id,omitempty"`     // Comment ID if target is comment
	CreatedAt    time.Time           `bson:"created_at" json:"created_at"`
}
