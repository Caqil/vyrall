package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Verification represents user verification requests
type Verification struct {
	ID               primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID           primitive.ObjectID `bson:"user_id" json:"user_id"`
	Type             string             `bson:"type" json:"type"` // email, phone, identity
	VerificationCode string             `bson:"verification_code" json:"verification_code,omitempty"`
	Status           string             `bson:"status" json:"status"`                 // pending, approved, rejected
	Documents        []string           `bson:"documents" json:"documents,omitempty"` // for identity verification
	CreatedAt        time.Time          `bson:"created_at" json:"created_at"`
	ExpiresAt        time.Time          `bson:"expires_at" json:"expires_at"`
	VerifiedAt       *time.Time         `bson:"verified_at,omitempty" json:"verified_at,omitempty"`
	RejectionReason  string             `bson:"rejection_reason,omitempty" json:"rejection_reason,omitempty"`
}

// VerificationRequest represents a user request for account verification (blue check)
type VerificationRequest struct {
	ID              primitive.ObjectID  `bson:"_id,omitempty" json:"id"`
	UserID          primitive.ObjectID  `bson:"user_id" json:"user_id"`
	Category        string              `bson:"category" json:"category"` // celebrity, brand, journalist, etc.
	FullName        string              `bson:"full_name" json:"full_name"`
	DocumentType    string              `bson:"document_type" json:"document_type"` // passport, license, etc.
	DocumentImages  []string            `bson:"document_images" json:"document_images"`
	AdditionalInfo  string              `bson:"additional_info,omitempty" json:"additional_info,omitempty"`
	Status          string              `bson:"status" json:"status"` // pending, approved, rejected
	ReviewerID      *primitive.ObjectID `bson:"reviewer_id,omitempty" json:"reviewer_id,omitempty"`
	ReviewedAt      *time.Time          `bson:"reviewed_at,omitempty" json:"reviewed_at,omitempty"`
	RejectionReason string              `bson:"rejection_reason,omitempty" json:"rejection_reason,omitempty"`
	CreatedAt       time.Time           `bson:"created_at" json:"created_at"`
	UpdatedAt       time.Time           `bson:"updated_at" json:"updated_at"`
}
