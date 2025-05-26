package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Conversation represents a messaging conversation between users
type Conversation struct {
	ID                  primitive.ObjectID   `bson:"_id,omitempty" json:"id"`
	Type                string               `bson:"type" json:"type"` // direct, group
	Participants        []Participant        `bson:"participants" json:"participants"`
	Title               string               `bson:"title,omitempty" json:"title,omitempty"`   // For group conversations
	Avatar              string               `bson:"avatar,omitempty" json:"avatar,omitempty"` // For group conversations
	LastMessageID       *primitive.ObjectID  `bson:"last_message_id,omitempty" json:"last_message_id,omitempty"`
	LastMessagePreview  string               `bson:"last_message_preview" json:"last_message_preview"`
	LastMessageAt       *time.Time           `bson:"last_message_at,omitempty" json:"last_message_at,omitempty"`
	LastMessageSenderID *primitive.ObjectID  `bson:"last_message_sender_id,omitempty" json:"last_message_sender_id,omitempty"`
	MessageCount        int                  `bson:"message_count" json:"message_count"`
	IsEncrypted         bool                 `bson:"is_encrypted" json:"is_encrypted"`
	EncryptionEnabled   time.Time            `bson:"encryption_enabled,omitempty" json:"encryption_enabled,omitempty"`
	GroupInfo           *GroupChatInfo       `bson:"group_info,omitempty" json:"group_info,omitempty"`
	IsActive            bool                 `bson:"is_active" json:"is_active"`
	CreatedAt           time.Time            `bson:"created_at" json:"created_at"`
	UpdatedAt           time.Time            `bson:"updated_at" json:"updated_at"`
	ArchivedForUsers    []primitive.ObjectID `bson:"archived_for_users,omitempty" json:"archived_for_users,omitempty"`
	DeletedForUsers     []primitive.ObjectID `bson:"deleted_for_users,omitempty" json:"deleted_for_users,omitempty"`
	MutedForUsers       map[string]time.Time `bson:"muted_for_users,omitempty" json:"muted_for_users,omitempty"` // Map of UserID to mute expiration
}

// Participant represents a user in a conversation
type Participant struct {
	UserID              primitive.ObjectID  `bson:"user_id" json:"user_id"`
	JoinedAt            time.Time           `bson:"joined_at" json:"joined_at"`
	Role                string              `bson:"role,omitempty" json:"role,omitempty"` // admin, member
	LastReadMessageID   *primitive.ObjectID `bson:"last_read_message_id,omitempty" json:"last_read_message_id,omitempty"`
	LastReadAt          *time.Time          `bson:"last_read_at,omitempty" json:"last_read_at,omitempty"`
	UnreadCount         int                 `bson:"unread_count" json:"unread_count"`
	IsActive            bool                `bson:"is_active" json:"is_active"`
	TypingAt            *time.Time          `bson:"typing_at,omitempty" json:"typing_at,omitempty"`
	Nickname            string              `bson:"nickname,omitempty" json:"nickname,omitempty"`
	NotificationSetting string              `bson:"notification_setting" json:"notification_setting"` // all, mentions, none
}

// GroupChatInfo contains additional information for group conversations
type GroupChatInfo struct {
	Description  string               `bson:"description,omitempty" json:"description,omitempty"`
	CreatorID    primitive.ObjectID   `bson:"creator_id" json:"creator_id"`
	Admins       []primitive.ObjectID `bson:"admins" json:"admins"`
	IsPublic     bool                 `bson:"is_public" json:"is_public"`
	JoinRequests []JoinRequest        `bson:"join_requests,omitempty" json:"join_requests,omitempty"`
	Rules        []string             `bson:"rules,omitempty" json:"rules,omitempty"`
	LinkAccess   string               `bson:"link_access" json:"link_access"` // anyone, members, admins
	JoinLink     string               `bson:"join_link,omitempty" json:"join_link,omitempty"`
	MaxMembers   int                  `bson:"max_members" json:"max_members"`
}

// JoinRequest represents a request to join a group conversation
type JoinRequest struct {
	UserID      primitive.ObjectID  `bson:"user_id" json:"user_id"`
	RequestedAt time.Time           `bson:"requested_at" json:"requested_at"`
	Message     string              `bson:"message,omitempty" json:"message,omitempty"`
	Status      string              `bson:"status" json:"status"` // pending, approved, declined
	ReviewedBy  *primitive.ObjectID `bson:"reviewed_by,omitempty" json:"reviewed_by,omitempty"`
	ReviewedAt  *time.Time          `bson:"reviewed_at,omitempty" json:"reviewed_at,omitempty"`
}
