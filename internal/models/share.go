package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Share represents a post share
type Share struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID    primitive.ObjectID `bson:"user_id" json:"user_id"`
	PostID    primitive.ObjectID `bson:"post_id" json:"post_id"`
	SharedTo  string             `bson:"shared_to" json:"shared_to"`                     // platform, profile, group
	TargetID  string             `bson:"target_id,omitempty" json:"target_id,omitempty"` // ID of target entity
	AddedText string             `bson:"added_text,omitempty" json:"added_text,omitempty"`
	IsPublic  bool               `bson:"is_public" json:"is_public"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
}
