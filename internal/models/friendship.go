package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Friendship represents a mutual connection between users
type Friendship struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID1     primitive.ObjectID `bson:"user_id_1" json:"user_id_1"`
	UserID2     primitive.ObjectID `bson:"user_id_2" json:"user_id_2"`
	Status      string             `bson:"status" json:"status"` // pending, accepted, rejected, blocked
	RequestedBy primitive.ObjectID `bson:"requested_by" json:"requested_by"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at" json:"updated_at"`
}
