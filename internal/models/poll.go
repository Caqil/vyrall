package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Poll represents a poll attached to a post
type Poll struct {
	Question      string       `bson:"question" json:"question"`
	Options       []PollOption `bson:"options" json:"options"`
	AllowMultiple bool         `bson:"allow_multiple" json:"allow_multiple"`
	ExpiresAt     time.Time    `bson:"expires_at" json:"expires_at"`
	IsAnonymous   bool         `bson:"is_anonymous" json:"is_anonymous"`
	TotalVotes    int          `bson:"total_votes" json:"total_votes"`
}

// PollOption represents a single option in a poll
type PollOption struct {
	ID       primitive.ObjectID   `bson:"_id,omitempty" json:"id"`
	Text     string               `bson:"text" json:"text"`
	Count    int                  `bson:"count" json:"count"`
	Voters   []primitive.ObjectID `bson:"voters,omitempty" json:"voters,omitempty"`
	ImageURL string               `bson:"image_url,omitempty" json:"image_url,omitempty"`
}

// PollVote represents a user's vote on a poll
type PollVote struct {
	ID          primitive.ObjectID   `bson:"_id,omitempty" json:"id"`
	PollID      primitive.ObjectID   `bson:"poll_id" json:"poll_id"`
	UserID      primitive.ObjectID   `bson:"user_id" json:"user_id"`
	OptionIDs   []primitive.ObjectID `bson:"option_ids" json:"option_ids"`
	VotedAt     time.Time            `bson:"voted_at" json:"voted_at"`
	IsAnonymous bool                 `bson:"is_anonymous" json:"is_anonymous"`
}
