package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Media represents a media file attached to a post or message
type Media struct {
	ID               primitive.ObjectID     `bson:"_id,omitempty" json:"id"`
	UserID           primitive.ObjectID     `bson:"user_id" json:"user_id"`
	Type             string                 `bson:"type" json:"type"` // image, video, audio, document
	URL              string                 `bson:"url" json:"url"`
	ThumbnailURL     string                 `bson:"thumbnail_url,omitempty" json:"thumbnail_url,omitempty"`
	FileName         string                 `bson:"file_name" json:"file_name"`
	FileSize         int64                  `bson:"file_size" json:"file_size"`
	MimeType         string                 `bson:"mime_type" json:"mime_type"`
	Width            int                    `bson:"width,omitempty" json:"width,omitempty"`
	Height           int                    `bson:"height,omitempty" json:"height,omitempty"`
	Duration         float64                `bson:"duration,omitempty" json:"duration,omitempty"` // for video/audio
	ProcessingStatus string                 `bson:"processing_status" json:"processing_status"`   // pending, processing, completed, failed
	IsProcessed      bool                   `bson:"is_processed" json:"is_processed"`
	AltText          string                 `bson:"alt_text,omitempty" json:"alt_text,omitempty"`
	Caption          string                 `bson:"caption,omitempty" json:"caption,omitempty"`
	UploadedAt       time.Time              `bson:"uploaded_at" json:"uploaded_at"`
	ProcessedAt      *time.Time             `bson:"processed_at,omitempty" json:"processed_at,omitempty"`
	Metadata         map[string]interface{} `bson:"metadata,omitempty" json:"metadata,omitempty"`
	CreatedAt        time.Time              `bson:"created_at" json:"created_at"`
	UpdatedAt        time.Time              `bson:"updated_at" json:"updated_at"`
}

// Image represents an image with multiple resolutions
type Image struct {
	Original    string `bson:"original" json:"original"`
	Large       string `bson:"large,omitempty" json:"large,omitempty"`
	Medium      string `bson:"medium,omitempty" json:"medium,omitempty"`
	Small       string `bson:"small,omitempty" json:"small,omitempty"`
	Thumbnail   string `bson:"thumbnail,omitempty" json:"thumbnail,omitempty"`
	Width       int    `bson:"width,omitempty" json:"width,omitempty"`
	Height      int    `bson:"height,omitempty" json:"height,omitempty"`
	AltText     string `bson:"alt_text,omitempty" json:"alt_text,omitempty"`
	ContentType string `bson:"content_type,omitempty" json:"content_type,omitempty"`
}

// Video represents a video with multiple resolutions
type Video struct {
	Original    string  `bson:"original" json:"original"`
	HighRes     string  `bson:"high_res,omitempty" json:"high_res,omitempty"`
	MediumRes   string  `bson:"medium_res,omitempty" json:"medium_res,omitempty"`
	LowRes      string  `bson:"low_res,omitempty" json:"low_res,omitempty"`
	Thumbnail   string  `bson:"thumbnail,omitempty" json:"thumbnail,omitempty"`
	Duration    float64 `bson:"duration,omitempty" json:"duration,omitempty"` // in seconds
	Width       int     `bson:"width,omitempty" json:"width,omitempty"`
	Height      int     `bson:"height,omitempty" json:"height,omitempty"`
	ContentType string  `bson:"content_type,omitempty" json:"content_type,omitempty"`
	Caption     string  `bson:"caption,omitempty" json:"caption,omitempty"`
}

// Audio represents an audio file
type Audio struct {
	URL         string  `bson:"url" json:"url"`
	Duration    float64 `bson:"duration,omitempty" json:"duration,omitempty"` // in seconds
	ContentType string  `bson:"content_type,omitempty" json:"content_type,omitempty"`
	Title       string  `bson:"title,omitempty" json:"title,omitempty"`
	Artist      string  `bson:"artist,omitempty" json:"artist,omitempty"`
	Waveform    string  `bson:"waveform,omitempty" json:"waveform,omitempty"`
	Transcript  string  `bson:"transcript,omitempty" json:"transcript,omitempty"`
}

// File represents a generic file
type File struct {
	URL         string    `bson:"url" json:"url"`
	FileName    string    `bson:"file_name" json:"file_name"`
	Size        int64     `bson:"size" json:"size"` // in bytes
	ContentType string    `bson:"content_type" json:"content_type"`
	UploadedAt  time.Time `bson:"uploaded_at" json:"uploaded_at"`
}
