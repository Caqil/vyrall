package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Common contains shared fields used across multiple models
type Common struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at" json:"updated_at"`
	DeletedAt *time.Time         `bson:"deleted_at,omitempty" json:"deleted_at,omitempty"`
}

// Pagination represents pagination metadata for API responses
type Pagination struct {
	Page       int  `json:"page"`
	PageSize   int  `json:"page_size"`
	TotalRows  int  `json:"total_rows"`
	TotalPages int  `json:"total_pages"`
	HasNext    bool `json:"has_next"`
	HasPrev    bool `json:"has_prev"`
}

// SortOption represents sorting options for queries
type SortOption struct {
	Field     string `json:"field"`
	Direction int    `json:"direction"` // 1 for ascending, -1 for descending
}

// FilterOption represents filtering options for queries
type FilterOption struct {
	Field    string      `json:"field"`
	Operator string      `json:"operator"` // eq, ne, gt, gte, lt, lte, in, nin, etc.
	Value    interface{} `json:"value"`
}

// SearchOption represents search options for queries
type SearchOption struct {
	Query      string             `json:"query"`
	Fields     []string           `json:"fields"`
	Fuzzy      bool               `json:"fuzzy"`
	Exact      bool               `json:"exact"`
	Weightings map[string]float64 `json:"weightings,omitempty"`
}

// GeoPoint represents a geographical point with latitude and longitude
type GeoPoint struct {
	Type        string    `bson:"type" json:"type"`               // "Point"
	Coordinates []float64 `bson:"coordinates" json:"coordinates"` // [longitude, latitude]
}

// Location represents geographical coordinates and information
type Location struct {
	Name        string   `bson:"name" json:"name"`
	Address     string   `bson:"address,omitempty" json:"address,omitempty"`
	City        string   `bson:"city,omitempty" json:"city,omitempty"`
	Country     string   `bson:"country,omitempty" json:"country,omitempty"`
	Coordinates GeoPoint `bson:"coordinates" json:"coordinates"`
}

// DateRange represents a time period between two dates
type DateRange struct {
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
}
