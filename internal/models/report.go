package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Report represents a user report on content
type Report struct {
	ID             primitive.ObjectID  `bson:"_id,omitempty" json:"id"`
	ReporterID     primitive.ObjectID  `bson:"reporter_id" json:"reporter_id"`
	ContentID      primitive.ObjectID  `bson:"content_id" json:"content_id"`
	ContentType    string              `bson:"content_type" json:"content_type"` // post, comment, user, message
	ReasonCode     string              `bson:"reason_code" json:"reason_code"`
	Description    string              `bson:"description,omitempty" json:"description,omitempty"`
	Status         string              `bson:"status" json:"status"` // pending, reviewed, actioned, dismissed
	ModeratorID    *primitive.ObjectID `bson:"moderator_id,omitempty" json:"moderator_id,omitempty"`
	ModeratorNotes string              `bson:"moderator_notes,omitempty" json:"moderator_notes,omitempty"`
	ActionTaken    string              `bson:"action_taken,omitempty" json:"action_taken,omitempty"`
	CreatedAt      time.Time           `bson:"created_at" json:"created_at"`
	UpdatedAt      time.Time           `bson:"updated_at" json:"updated_at"`
	ResolvedAt     *time.Time          `bson:"resolved_at,omitempty" json:"resolved_at,omitempty"`
}
