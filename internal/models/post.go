package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Post represents a social media post
type Post struct {
	ID             primitive.ObjectID   `bson:"_id,omitempty" json:"id"`
	UserID         primitive.ObjectID   `bson:"user_id" json:"user_id"`
	Content        string               `bson:"content" json:"content"`
	MediaFiles     []Media              `bson:"media_files,omitempty" json:"media_files,omitempty"`
	Hashtags       []string             `bson:"hashtags,omitempty" json:"tags,omitempty"`
	MentionedUsers []primitive.ObjectID `bson:"mentioned_users,omitempty" json:"mentioned_users,omitempty"`
	Location       *Location            `bson:"location,omitempty" json:"location,omitempty"`
	NSFW           bool                 `bson:"nsfw" json:"nsfw"`
	EnableLikes    bool                 `bson:"enable_likes" json:"enable_likes"`
	EnableSharing  bool                 `bson:"enable_sharing" json:"enable_sharing"`
	PageID         *primitive.ObjectID  `bson:"page_id,omitempty" json:"page_id,omitempty"`
	Privacy        string               `bson:"privacy" json:"privacy"` // public, friends, private
	LikeCount      int                  `bson:"like_count" json:"like_count"`
	CommentCount   int                  `bson:"comment_count" json:"comment_count"`
	ShareCount     int                  `bson:"share_count" json:"share_count"`
	ViewCount      int                  `bson:"view_count" json:"view_count"`
	IsEdited       bool                 `bson:"is_edited" json:"is_edited"`
	IsPinned       bool                 `bson:"is_pinned" json:"is_pinned"`
	IsArchived     bool                 `bson:"is_archived" json:"is_archived"`
	IsHidden       bool                 `bson:"is_hidden" json:"is_hidden"`
	IsFeatured     bool                 `bson:"is_featured" json:"is_featured"`
	IsSponsored    bool                 `bson:"is_sponsored" json:"is_sponsored"`
	AllowComments  bool                 `bson:"allow_comments" json:"allow_comments"`
	PublishedAt    time.Time            `bson:"published_at" json:"published_at"`
	ScheduledFor   *time.Time           `bson:"scheduled_for,omitempty" json:"scheduled_for,omitempty"`
	EditHistory    []EditRecord         `bson:"edit_history,omitempty" json:"edit_history,omitempty"`
	GroupID        *primitive.ObjectID  `bson:"group_id,omitempty" json:"group_id,omitempty"`
	EventID        *primitive.ObjectID  `bson:"event_id,omitempty" json:"event_id,omitempty"`
	RepostOf       *primitive.ObjectID  `bson:"repost_of,omitempty" json:"repost_of,omitempty"`
	Poll           *Poll                `bson:"poll,omitempty" json:"poll,omitempty"`
	ReactionCounts map[string]int       `bson:"reaction_counts,omitempty" json:"reaction_counts,omitempty"`
	CreatedAt      time.Time            `bson:"created_at" json:"created_at"`
	UpdatedAt      time.Time            `bson:"updated_at" json:"updated_at"`
	DeletedAt      *time.Time           `bson:"deleted_at,omitempty" json:"deleted_at,omitempty"`
}

// EditRecord tracks changes to a post
type EditRecord struct {
	Content  string             `bson:"content" json:"content"`
	EditedAt time.Time          `bson:"edited_at" json:"edited_at"`
	EditorID primitive.ObjectID `bson:"editor_id" json:"editor_id"`
	Reason   string             `bson:"reason,omitempty" json:"reason,omitempty"`
}
