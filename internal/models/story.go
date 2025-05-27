package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Story represents an ephemeral content post
type Story struct {
	ID              primitive.ObjectID   `bson:"_id,omitempty" json:"id"`
	UserID          primitive.ObjectID   `bson:"user_id" json:"user_id"`
	MediaFiles      []Media              `bson:"media_files" json:"media_files"`
	Caption         string               `bson:"caption,omitempty" json:"caption,omitempty"`
	Hashtags        []string             `bson:"hashtags,omitempty" json:"tags,omitempty"`
	MentionedUsers  []primitive.ObjectID `bson:"mentioned_users,omitempty" json:"mentioned_users,omitempty"`
	Location        *Location            `bson:"location,omitempty" json:"location,omitempty"`
	ViewerCount     int                  `bson:"viewer_count" json:"viewer_count"`
	ReactionCount   int                  `bson:"reaction_count" json:"reaction_count"`
	Viewers         []StoryViewer        `bson:"viewers,omitempty" json:"viewers,omitempty"`
	Reactions       []StoryReaction      `bson:"reactions,omitempty" json:"reactions,omitempty"`
	Replies         []StoryReply         `bson:"replies,omitempty" json:"replies,omitempty"`
	ReplyCount      int                  `bson:"reply_count" json:"reply_count"`
	IsHighlight     bool                 `bson:"is_highlight" json:"is_highlight"`
	HighlightID     *primitive.ObjectID  `bson:"highlight_id,omitempty" json:"highlight_id,omitempty"`
	PrivacySettings StoryPrivacy         `bson:"privacy_settings" json:"privacy_settings"`
	Duration        int                  `bson:"duration" json:"duration"` // in seconds
	ExpiresAt       time.Time            `bson:"expires_at" json:"expires_at"`
	CreatedAt       time.Time            `bson:"created_at" json:"created_at"`
	Interactive     *StoryInteractive    `bson:"interactive,omitempty" json:"interactive,omitempty"`
	MusicInfo       *StoryMusic          `bson:"music_info,omitempty" json:"music_info,omitempty"`
	LinkInfo        *StoryLink           `bson:"link_info,omitempty" json:"link_info,omitempty"`
	IsArchived      bool                 `bson:"is_archived" json:"is_archived"`
}

// StoryViewer represents a user who has viewed a story
type StoryViewer struct {
	UserID    primitive.ObjectID `bson:"user_id" json:"user_id"`
	ViewedAt  time.Time          `bson:"viewed_at" json:"viewed_at"`
	ViewCount int                `bson:"view_count" json:"view_count"` // For replays
	Device    string             `bson:"device,omitempty" json:"device,omitempty"`
}

// StoryReaction represents a reaction to a story
type StoryReaction struct {
	UserID    primitive.ObjectID `bson:"user_id" json:"user_id"`
	Reaction  string             `bson:"reaction" json:"reaction"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
}

// StoryReply represents a direct reply to a story
type StoryReply struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID    primitive.ObjectID `bson:"user_id" json:"user_id"`
	Content   string             `bson:"content" json:"content"`
	MediaURL  string             `bson:"media_url,omitempty" json:"media_url,omitempty"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
	IsPrivate bool               `bson:"is_private" json:"is_private"` // Only visible to story creator
}

// StoryPrivacy defines who can see a story
type StoryPrivacy struct {
	Visibility     string               `bson:"visibility" json:"visibility"` // everyone, followers, close_friends
	HideFromUsers  []primitive.ObjectID `bson:"hide_from_users,omitempty" json:"hide_from_users,omitempty"`
	VisibleToUsers []primitive.ObjectID `bson:"visible_to_users,omitempty" json:"visible_to_users,omitempty"`
}

// StoryInteractive represents interactive elements in a story
type StoryInteractive struct {
	Type           string                 `bson:"type" json:"type"` // poll, question, quiz, slider
	Data           map[string]interface{} `bson:"data" json:"data"`
	Responses      []InteractiveResponse  `bson:"responses,omitempty" json:"responses,omitempty"`
	ResultsVisible bool                   `bson:"results_visible" json:"results_visible"`
}

// InteractiveResponse represents a user response to an interactive story element
type InteractiveResponse struct {
	UserID    primitive.ObjectID `bson:"user_id" json:"user_id"`
	Response  interface{}        `bson:"response" json:"response"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
}

// StoryMusic represents music attached to a story
type StoryMusic struct {
	Title    string `bson:"title" json:"title"`
	Artist   string `bson:"artist" json:"artist"`
	AudioURL string `bson:"audio_url" json:"audio_url"`
	CoverURL string `bson:"cover_url,omitempty" json:"cover_url,omitempty"`
	Duration int    `bson:"duration" json:"duration"` // in seconds
}

// StoryLink represents a link attached to a story
type StoryLink struct {
	URL          string `bson:"url" json:"url"`
	Title        string `bson:"title,omitempty" json:"title,omitempty"`
	Description  string `bson:"description,omitempty" json:"description,omitempty"`
	ThumbnailURL string `bson:"thumbnail_url,omitempty" json:"thumbnail_url,omitempty"`
	Clicks       int    `bson:"clicks" json:"clicks"`
}
