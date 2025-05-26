package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Subscription represents a recurring subscription plan
type Subscription struct {
	ID                 primitive.ObjectID  `bson:"_id,omitempty" json:"id"`
	UserID             primitive.ObjectID  `bson:"user_id" json:"user_id"`
	BusinessID         *primitive.ObjectID `bson:"business_id,omitempty" json:"business_id,omitempty"`
	PlanID             primitive.ObjectID  `bson:"plan_id" json:"plan_id"`
	Status             string              `bson:"status" json:"status"` // active, canceled, paused, past_due
	CurrentPeriodStart time.Time           `bson:"current_period_start" json:"current_period_start"`
	CurrentPeriodEnd   time.Time           `bson:"current_period_end" json:"current_period_end"`
	CanceledAt         *time.Time          `bson:"canceled_at,omitempty" json:"canceled_at,omitempty"`
	CancelAtPeriodEnd  bool                `bson:"cancel_at_period_end" json:"cancel_at_period_end"`
	PaymentMethodID    primitive.ObjectID  `bson:"payment_method_id" json:"payment_method_id"`
	LastPaymentID      *primitive.ObjectID `bson:"last_payment_id,omitempty" json:"last_payment_id,omitempty"`
	NextBillingDate    time.Time           `bson:"next_billing_date" json:"next_billing_date"`
	TrialStart         *time.Time          `bson:"trial_start,omitempty" json:"trial_start,omitempty"`
	TrialEnd           *time.Time          `bson:"trial_end,omitempty" json:"trial_end,omitempty"`
	CreatedAt          time.Time           `bson:"created_at" json:"created_at"`
	UpdatedAt          time.Time           `bson:"updated_at" json:"updated_at"`
}

// SubscriptionPlan represents a subscription plan configuration
type SubscriptionPlan struct {
	ID              primitive.ObjectID     `bson:"_id,omitempty" json:"id"`
	Name            string                 `bson:"name" json:"name"`
	Description     string                 `bson:"description" json:"description"`
	Price           float64                `bson:"price" json:"price"`
	Currency        string                 `bson:"currency" json:"currency"`
	Interval        string                 `bson:"interval" json:"interval"` // month, year
	IntervalCount   int                    `bson:"interval_count" json:"interval_count"`
	TrialPeriodDays int                    `bson:"trial_period_days,omitempty" json:"trial_period_days,omitempty"`
	Features        map[string]interface{} `bson:"features" json:"features"`
	Limits          map[string]int         `bson:"limits" json:"limits"`
	IsActive        bool                   `bson:"is_active" json:"is_active"`
	DisplayOrder    int                    `bson:"display_order" json:"display_order"`
	CreatedAt       time.Time              `bson:"created_at" json:"created_at"`
	UpdatedAt       time.Time              `bson:"updated_at" json:"updated_at"`
}

// PaymentMethod represents a saved payment method
type PaymentMethod struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID      primitive.ObjectID `bson:"user_id" json:"user_id"`
	Type        string             `bson:"type" json:"type"`                       // credit_card, bank_account, paypal, etc.
	Provider    string             `bson:"provider" json:"provider"`               // stripe, paypal, etc.
	ExternalID  string             `bson:"external_id" json:"external_id"`         // ID in payment provider system
	Name        string             `bson:"name" json:"name"`                       // Nickname for this payment method
	Last4       string             `bson:"last4,omitempty" json:"last4,omitempty"` // Last 4 digits of card/account
	ExpiryMonth int                `bson:"expiry_month,omitempty" json:"expiry_month,omitempty"`
	ExpiryYear  int                `bson:"expiry_year,omitempty" json:"expiry_year,omitempty"`
	Brand       string             `bson:"brand,omitempty" json:"brand,omitempty"` // Visa, Mastercard, etc.
	Country     string             `bson:"country,omitempty" json:"country,omitempty"`
	IsDefault   bool               `bson:"is_default" json:"is_default"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at" json:"updated_at"`
}

// Payment represents a payment transaction
type Payment struct {
	ID                primitive.ObjectID  `bson:"_id,omitempty" json:"id"`
	UserID            primitive.ObjectID  `bson:"user_id" json:"user_id"`
	BusinessID        *primitive.ObjectID `bson:"business_id,omitempty" json:"business_id,omitempty"`
	PaymentMethodID   primitive.ObjectID  `bson:"payment_method_id" json:"payment_method_id"`
	Type              string              `bson:"type" json:"type"` // ad_payment, subscription, etc.
	Amount            float64             `bson:"amount" json:"amount"`
	Currency          string              `bson:"currency" json:"currency"`
	Status            string              `bson:"status" json:"status"` // pending, completed, failed, refunded
	Description       string              `bson:"description" json:"description"`
	ExternalID        string              `bson:"external_id,omitempty" json:"external_id,omitempty"` // Transaction ID in payment system
	InvoiceURL        string              `bson:"invoice_url,omitempty" json:"invoice_url,omitempty"`
	ReceiptURL        string              `bson:"receipt_url,omitempty" json:"receipt_url,omitempty"`
	RelatedEntityType string              `bson:"related_entity_type,omitempty" json:"related_entity_type,omitempty"` // campaign, ad, subscription
	RelatedEntityID   *primitive.ObjectID `bson:"related_entity_id,omitempty" json:"related_entity_id,omitempty"`
	CreatedAt         time.Time           `bson:"created_at" json:"created_at"`
	CompletedAt       *time.Time          `bson:"completed_at,omitempty" json:"completed_at,omitempty"`
}
