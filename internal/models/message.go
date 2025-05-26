package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Message represents a direct message between users
type Message struct {
	ID                 primitive.ObjectID   `bson:"_id,omitempty" json:"id"`
	ConversationID     primitive.ObjectID   `bson:"conversation_id" json:"conversation_id"`
	SenderID           primitive.ObjectID   `bson:"sender_id" json:"sender_id"`
	Content            string               `bson:"content,omitempty" json:"content,omitempty"`
	MediaFiles         []Media              `bson:"media_files,omitempty" json:"media_files,omitempty"`
	Reactions          []MessageReaction    `bson:"reactions,omitempty" json:"reactions,omitempty"`
	ReplyToID          *primitive.ObjectID  `bson:"reply_to_id,omitempty" json:"reply_to_id,omitempty"`
	ForwardedFrom      *primitive.ObjectID  `bson:"forwarded_from,omitempty" json:"forwarded_from,omitempty"`
	ForwardedMessageID *primitive.ObjectID  `bson:"forwarded_message_id,omitempty" json:"forwarded_message_id,omitempty"`
	IsEdited           bool                 `bson:"is_edited" json:"is_edited"`
	EditHistory        []MessageEdit        `bson:"edit_history,omitempty" json:"edit_history,omitempty"`
	IsEncrypted        bool                 `bson:"is_encrypted" json:"is_encrypted"`
	EncryptionDetails  *EncryptionDetails   `bson:"encryption_details,omitempty" json:"encryption_details,omitempty"`
	DeliveryStatus     map[string]string    `bson:"delivery_status" json:"delivery_status"` // Map of UserID to status (sent, delivered, read)
	ReadByUsers        []ReadReceipt        `bson:"read_by_users,omitempty" json:"read_by_users,omitempty"`
	IsDeleted          bool                 `bson:"is_deleted" json:"is_deleted"`
	DeletedFor         []primitive.ObjectID `bson:"deleted_for,omitempty" json:"deleted_for,omitempty"`
	ExpiresAt          *time.Time           `bson:"expires_at,omitempty" json:"expires_at,omitempty"` // For disappearing messages
	MessageType        string               `bson:"message_type" json:"message_type"`                 // text, media, voice, system, etc.
	SystemMessage      *SystemMessage       `bson:"system_message,omitempty" json:"system_message,omitempty"`
	MentionedUsers     []primitive.ObjectID `bson:"mentioned_users,omitempty" json:"mentioned_users,omitempty"`
	IsImportant        bool                 `bson:"is_important" json:"is_important"`
	CreatedAt          time.Time            `bson:"created_at" json:"created_at"`
	UpdatedAt          time.Time            `bson:"updated_at" json:"updated_at"`
	DeletedAt          *time.Time           `bson:"deleted_at,omitempty" json:"deleted_at,omitempty"`
}

// MessageEdit represents an edit to a message
type MessageEdit struct {
	PreviousContent string    `bson:"previous_content" json:"previous_content"`
	EditedAt        time.Time `bson:"edited_at" json:"edited_at"`
}

// MessageReaction represents a reaction to a message
type MessageReaction struct {
	UserID    primitive.ObjectID `bson:"user_id" json:"user_id"`
	Reaction  string             `bson:"reaction" json:"reaction"` // emoji code
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
}

// ReadReceipt tracks when a user read a message
type ReadReceipt struct {
	UserID primitive.ObjectID `bson:"user_id" json:"user_id"`
	ReadAt time.Time          `bson:"read_at" json:"read_at"`
	Device string             `bson:"device,omitempty" json:"device,omitempty"`
}

// EncryptionDetails contains information about encrypted messages
type EncryptionDetails struct {
	Algorithm      string            `bson:"algorithm" json:"algorithm"`
	KeyID          string            `bson:"key_id" json:"key_id"`
	IV             string            `bson:"iv" json:"iv"`
	RecipientKeys  map[string]string `bson:"recipient_keys" json:"recipient_keys"` // Map of UserID to encrypted key
	SignatureValid bool              `bson:"signature_valid" json:"signature_valid"`
}

// SystemMessage represents automated system messages in a conversation
type SystemMessage struct {
	Type       string                 `bson:"type" json:"type"` // user_added, user_left, name_changed, etc.
	Parameters map[string]interface{} `bson:"parameters" json:"parameters"`
}
