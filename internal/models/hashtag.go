package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Hashtag represents a tracked hashtag in the system
type Hashtag struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name          string             `bson:"name" json:"name"`
	PostCount     int                `bson:"post_count" json:"post_count"`
	FollowerCount int                `bson:"follower_count" json:"follower_count"`
	IsTrending    bool               `bson:"is_trending" json:"is_trending"`
	TrendingRank  int                `bson:"trending_rank,omitempty" json:"trending_rank,omitempty"`
	IsRestricted  bool               `bson:"is_restricted" json:"is_restricted"`
	CreatedAt     time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt     time.Time          `bson:"updated_at" json:"updated_at"`
}

// HashtagFollow represents a user following a hashtag
type HashtagFollow struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID    primitive.ObjectID `bson:"user_id" json:"user_id"`
	HashtagID primitive.ObjectID `bson:"hashtag_id" json:"hashtag_id"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
}
