package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Like represents a like or reaction on a post or comment
type Like struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID       primitive.ObjectID `bson:"user_id" json:"user_id"`
	ContentID    primitive.ObjectID `bson:"content_id" json:"content_id"`       // Post or Comment ID
	ContentType  string             `bson:"content_type" json:"content_type"`   // "post" or "comment"
	ReactionType string             `bson:"reaction_type" json:"reaction_type"` // "like", "love", "haha", etc.
	CreatedAt    time.Time          `bson:"created_at" json:"created_at"`
}
