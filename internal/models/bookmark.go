package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Bookmark represents a saved post
type Bookmark struct {
	ID           primitive.ObjectID  `bson:"_id,omitempty" json:"id"`
	UserID       primitive.ObjectID  `bson:"user_id" json:"user_id"`
	PostID       primitive.ObjectID  `bson:"post_id" json:"post_id"`
	CollectionID *primitive.ObjectID `bson:"collection_id,omitempty" json:"collection_id,omitempty"`
	Notes        string              `bson:"notes,omitempty" json:"notes,omitempty"`
	CreatedAt    time.Time           `bson:"created_at" json:"created_at"`
}

// BookmarkCollection represents a collection of bookmarks
type BookmarkCollection struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID      primitive.ObjectID `bson:"user_id" json:"user_id"`
	Name        string             `bson:"name" json:"name"`
	Description string             `bson:"description,omitempty" json:"description,omitempty"`
	IsPrivate   bool               `bson:"is_private" json:"is_private"`
	ItemCount   int                `bson:"item_count" json:"item_count"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at" json:"updated_at"`
}
