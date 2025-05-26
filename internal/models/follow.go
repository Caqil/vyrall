package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Follow represents a follower relationship between users
type Follow struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	FollowerID  primitive.ObjectID `bson:"follower_id" json:"follower_id"`   // User who is following
	FollowingID primitive.ObjectID `bson:"following_id" json:"following_id"` // User being followed
	Status      string             `bson:"status" json:"status"`             // pending, accepted (for private accounts)
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at" json:"updated_at"`
	NotifyPosts bool               `bson:"notify_posts" json:"notify_posts"` // Get notifications for new posts
}
