package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Group represents a community group
type Group struct {
	ID                   primitive.ObjectID   `bson:"_id,omitempty" json:"id"`
	Name                 string               `bson:"name" json:"name"`
	Description          string               `bson:"description,omitempty" json:"description,omitempty"`
	Avatar               string               `bson:"avatar,omitempty" json:"avatar,omitempty"`
	CoverPhoto           string               `bson:"cover_photo,omitempty" json:"cover_photo,omitempty"`
	CreatorID            primitive.ObjectID   `bson:"creator_id" json:"creator_id"`
	Admins               []primitive.ObjectID `bson:"admins" json:"admins"`
	Moderators           []primitive.ObjectID `bson:"moderators,omitempty" json:"moderators,omitempty"`
	MemberCount          int                  `bson:"member_count" json:"member_count"`
	PostCount            int                  `bson:"post_count" json:"post_count"`
	EventCount           int                  `bson:"event_count" json:"event_count"`
	IsPublic             bool                 `bson:"is_public" json:"is_public"`
	IsVisible            bool                 `bson:"is_visible" json:"is_visible"` // Can be found in search
	JoinApprovalRequired bool                 `bson:"join_approval_required" json:"join_approval_required"`
	Location             *Location            `bson:"location,omitempty" json:"location,omitempty"`
	Categories           []string             `bson:"categories,omitempty" json:"categories,omitempty"`
	Tags                 []string             `bson:"tags,omitempty" json:"tags,omitempty"`
	Rules                []GroupRule          `bson:"rules,omitempty" json:"rules,omitempty"`
	Features             GroupFeatures        `bson:"features" json:"features"`
	CreatedAt            time.Time            `bson:"created_at" json:"created_at"`
	UpdatedAt            time.Time            `bson:"updated_at" json:"updated_at"`
	LastActivityAt       time.Time            `bson:"last_activity_at" json:"last_activity_at"`
	Status               string               `bson:"status" json:"status"` // active, archived, suspended
	JoinRequests         []GroupJoinRequest   `bson:"join_requests,omitempty" json:"join_requests,omitempty"`
	InviteLink           string               `bson:"invite_link,omitempty" json:"invite_link,omitempty"`
	IsVerified           bool                 `bson:"is_verified" json:"is_verified"`
}

// GroupRule defines a rule for a group
type GroupRule struct {
	Title       string    `bson:"title" json:"title"`
	Description string    `bson:"description" json:"description"`
	CreatedAt   time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time `bson:"updated_at" json:"updated_at"`
}

// GroupFeatures defines what features are enabled for a group
type GroupFeatures struct {
	Events          bool `bson:"events" json:"events"`
	Polls           bool `bson:"polls" json:"polls"`
	PhotoSharing    bool `bson:"photo_sharing" json:"photo_sharing"`
	VideoSharing    bool `bson:"video_sharing" json:"video_sharing"`
	FileSharing     bool `bson:"file_sharing" json:"file_sharing"`
	LiveStreaming   bool `bson:"live_streaming" json:"live_streaming"`
	MemberDirectory bool `bson:"member_directory" json:"member_directory"`
	Topics          bool `bson:"topics" json:"topics"`
	Announcements   bool `bson:"announcements" json:"announcements"`
	Chat            bool `bson:"chat" json:"chat"`
}

// GroupMember represents a user's membership in a group
type GroupMember struct {
	ID                   primitive.ObjectID  `bson:"_id,omitempty" json:"id"`
	GroupID              primitive.ObjectID  `bson:"group_id" json:"group_id"`
	UserID               primitive.ObjectID  `bson:"user_id" json:"user_id"`
	Role                 string              `bson:"role" json:"role"` // admin, moderator, member
	JoinedAt             time.Time           `bson:"joined_at" json:"joined_at"`
	InvitedBy            *primitive.ObjectID `bson:"invited_by,omitempty" json:"invited_by,omitempty"`
	NotificationSettings string              `bson:"notification_settings" json:"notification_settings"` // all, mentions, none
	IsActive             bool                `bson:"is_active" json:"is_active"`
	LastActiveAt         time.Time           `bson:"last_active_at" json:"last_active_at"`
	ContributionScore    int                 `bson:"contribution_score" json:"contribution_score"`
	Nickname             string              `bson:"nickname,omitempty" json:"nickname,omitempty"`
	BadgeIDs             []string            `bson:"badge_ids,omitempty" json:"badge_ids,omitempty"`
}

// GroupJoinRequest represents a request to join a group
type GroupJoinRequest struct {
	UserID      primitive.ObjectID  `bson:"user_id" json:"user_id"`
	Message     string              `bson:"message,omitempty" json:"message,omitempty"`
	RequestedAt time.Time           `bson:"requested_at" json:"requested_at"`
	Status      string              `bson:"status" json:"status"` // pending, approved, rejected
	ReviewedBy  *primitive.ObjectID `bson:"reviewed_by,omitempty" json:"reviewed_by,omitempty"`
	ReviewedAt  *time.Time          `bson:"reviewed_at,omitempty" json:"reviewed_at,omitempty"`
	Reason      string              `bson:"reason,omitempty" json:"reason,omitempty"`
}
