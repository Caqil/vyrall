package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Comment represents a comment on a post or another comment
type Comment struct {
	ID             primitive.ObjectID   `bson:"_id,omitempty" json:"id"`
	PostID         primitive.ObjectID   `bson:"post_id" json:"post_id"`
	UserID         primitive.ObjectID   `bson:"user_id" json:"user_id"`
	Content        string               `bson:"content" json:"content"`
	MediaFiles     []Media              `bson:"media_files,omitempty" json:"media_files,omitempty"`
	MentionedUsers []primitive.ObjectID `bson:"mentioned_users,omitempty" json:"mentioned_users,omitempty"`
	ParentID       *primitive.ObjectID  `bson:"parent_id,omitempty" json:"parent_id,omitempty"` // For threaded comments
	LikeCount      int                  `bson:"like_count" json:"like_count"`
	ReplyCount     int                  `bson:"reply_count" json:"reply_count"`
	IsEdited       bool                 `bson:"is_edited" json:"is_edited"`
	IsPinned       bool                 `bson:"is_pinned" json:"is_pinned"`
	IsHidden       bool                 `bson:"is_hidden" json:"is_hidden"`
	EditHistory    []EditRecord         `bson:"edit_history,omitempty" json:"edit_history,omitempty"`
	ReactionCounts map[string]int       `bson:"reaction_counts,omitempty" json:"reaction_counts,omitempty"`
	CreatedAt      time.Time            `bson:"created_at" json:"created_at"`
	UpdatedAt      time.Time            `bson:"updated_at" json:"updated_at"`
	DeletedAt      *time.Time           `bson:"deleted_at,omitempty" json:"deleted_at,omitempty"`
}
