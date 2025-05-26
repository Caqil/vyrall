package live

import (
	"context"
	"net/http"
	"time"

	"github.com/Caqil/vyrall/internal/models"
	"github.com/Caqil/vyrall/internal/utils/response"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// StartStreamService defines the interface for starting a live stream
type StartStreamService interface {
	CreateStream(ctx context.Context, stream *models.LiveStream) (primitive.ObjectID, error)
	StartStream(ctx context.Context, streamID, userID primitive.ObjectID) error
	GenerateStreamKey(ctx context.Context, userID primitive.ObjectID) (string, error)
	ValidateStreamKey(ctx context.Context, streamKey string, userID primitive.ObjectID) (bool, error)
	GetStreamStatus(ctx context.Context, streamID primitive.ObjectID) (string, error)
	UpdateStreamStatus(ctx context.Context, streamID primitive.ObjectID, status string) error
	NotifyFollowers(ctx context.Context, userID primitive.ObjectID, streamID primitive.ObjectID, title string) error
}

// StartStreamRequest represents the request body for starting a live stream
type StartStreamRequest struct {
	Title       string   `json:"title" binding:"required"`
	Description string   `json:"description"`
	Thumbnail   string   `json:"thumbnail"`
	Category    string   `json:"category"`
	Tags        []string `json:"tags"`
	IsPrivate   bool     `json:"is_private"`
	EnableChat  bool     `json:"enable_chat"`
	EnableVOD   bool     `json:"enable_vod"`
}

// CreateStream handles creating a new live stream
func CreateStream(c *gin.Context) {
	// Get the authenticated user's ID
	userID, exists := c.Get("userID")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	// Parse request body
	var req StartStreamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request format", err)
		return
	}

	// Validate request
	if len(req.Title) < 3 || len(req.Title) > 100 {
		response.Error(c, http.StatusBadRequest, "Title must be between 3 and 100 characters", nil)
		return
	}

	if len(req.Description) > 5000 {
		response.Error(c, http.StatusBadRequest, "Description cannot exceed 5000 characters", nil)
		return
	}

	if len(req.Tags) > 10 {
		response.Error(c, http.StatusBadRequest, "Maximum of 10 tags allowed", nil)
		return
	}

	// Get the start stream service
	startStreamService := c.MustGet("startStreamService").(StartStreamService)

	// Generate stream key
	streamKey, err := startStreamService.GenerateStreamKey(c.Request.Context(), userID.(primitive.ObjectID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to generate stream key", err)
		return
	}

	// Create new stream object
	stream := &models.LiveStream{
		HostID:      userID.(primitive.ObjectID),
		Title:       req.Title,
		Description: req.Description,
		Thumbnail:   req.Thumbnail,
		Category:    req.Category,
		Tags:        req.Tags,
		IsPrivate:   req.IsPrivate,
		Status:      "pending", // Stream is created but not yet started
		StreamKey:   streamKey,
		ChatSettings: models.LiveStreamChatSettings{
			Enabled:             req.EnableChat,
			SlowMode:            false,
			SlowModeInterval:    0,
			FollowersOnlyMode:   false,
			SubscribersOnlyMode: false,
			EmoteOnlyMode:       false,
		},
		EnableVOD:       req.EnableVOD,
		ViewerCount:     0,
		PeakViewerCount: 0,
		TotalViews:      0,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	// Create the stream
	streamID, err := startStreamService.CreateStream(c.Request.Context(), stream)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to create stream", err)
		return
	}

	// Return success response with stream details and instructions
	response.Success(c, http.StatusCreated, "Stream created successfully", gin.H{
		"stream_id":   streamID.Hex(),
		"stream_key":  streamKey,
		"title":       req.Title,
		"status":      "pending",
		"rtmp_url":    "rtmp://streaming.yourdomain.com/live",
		"ingest_url":  "rtmp://streaming.yourdomain.com/live/" + streamKey,
		"player_url":  "https://yourdomain.com/streams/" + streamID.Hex(),
		"created_at":  stream.CreatedAt,
		"is_private":  req.IsPrivate,
		"enable_chat": req.EnableChat,
		"enable_vod":  req.EnableVOD,
	})
}

// StartStream handles starting a live stream that was previously created
func StartStream(c *gin.Context) {
	// Get stream ID from URL parameter
	streamIDStr := c.Param("id")
	streamID, err := primitive.ObjectIDFromHex(streamIDStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid stream ID format", err)
		return
	}

	// Get the authenticated user's ID
	userID, exists := c.Get("userID")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	// Get the live stream service
	liveStreamService := c.MustGet("liveStreamService").(LiveStreamService)

	// Check if the stream exists
	stream, err := liveStreamService.GetStreamByID(c.Request.Context(), streamID)
	if err != nil {
		response.Error(c, http.StatusNotFound, "Stream not found", err)
		return
	}

	// Check if the user is the stream host
	if stream.HostID != userID.(primitive.ObjectID) {
		response.Error(c, http.StatusForbidden, "Only the stream host can start the stream", nil)
		return
	}

	// Check if the stream is already active
	if stream.Status == "active" {
		response.Error(c, http.StatusBadRequest, "Stream is already active", nil)
		return
	}

	// Get the start stream service
	startStreamService := c.MustGet("startStreamService").(StartStreamService)

	// Start the stream
	err = startStreamService.StartStream(c.Request.Context(), streamID, userID.(primitive.ObjectID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to start stream", err)
		return
	}

	// Update stream status
	err = startStreamService.UpdateStreamStatus(c.Request.Context(), streamID, "active")
	if err != nil {
		// Log the error but don't fail the request
		c.Error(err)
	}

	// Set start time
	startedAt := time.Now()

	// Notify followers (asynchronously)
	go startStreamService.NotifyFollowers(c.Request.Context(), userID.(primitive.ObjectID), streamID, stream.Title)

	// Return success response with stream details
	response.Success(c, http.StatusOK, "Stream started successfully", gin.H{
		"stream_id":  streamID.Hex(),
		"title":      stream.Title,
		"status":     "active",
		"started_at": startedAt,
		"player_url": "https://yourdomain.com/streams/" + streamID.Hex(),
	})
}

// ValidateStreamKey handles validating a stream key for external RTMP servers
func ValidateStreamKey(c *gin.Context) {
	// Parse request body
	var req struct {
		StreamKey string `json:"stream_key" binding:"required"`
		UserID    string `json:"user_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request format", err)
		return
	}

	userID, err := primitive.ObjectIDFromHex(req.UserID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid user ID format", err)
		return
	}

	// Get the start stream service
	startStreamService := c.MustGet("startStreamService").(StartStreamService)

	// Validate the stream key
	isValid, err := startStreamService.ValidateStreamKey(c.Request.Context(), req.StreamKey, userID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to validate stream key", err)
		return
	}

	if !isValid {
		response.Error(c, http.StatusUnauthorized, "Invalid stream key", nil)
		return
	}

	response.Success(c, http.StatusOK, "Stream key validated successfully", gin.H{
		"valid": true,
	})
}

// GetStreamStatus returns the current status of a live stream
func GetStreamStatus(c *gin.Context) {
	// Get stream ID from URL parameter
	streamIDStr := c.Param("id")
	streamID, err := primitive.ObjectIDFromHex(streamIDStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid stream ID format", err)
		return
	}

	// Get the live stream service
	liveStreamService := c.MustGet("liveStreamService").(LiveStreamService)

	// Check if the stream exists
	stream, err := liveStreamService.GetStreamByID(c.Request.Context(), streamID)
	if err != nil {
		response.Error(c, http.StatusNotFound, "Stream not found", err)
		return
	}

	// Get the start stream service
	startStreamService := c.MustGet("startStreamService").(StartStreamService)

	// Get the stream status
	status, err := startStreamService.GetStreamStatus(c.Request.Context(), streamID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve stream status", err)
		return
	}

	// Get user service to include host details
	userService := c.MustGet("userService").(UserService)
	host, err := userService.GetUserByID(c.Request.Context(), stream.HostID)

	hostInfo := map[string]interface{}{
		"user_id": stream.HostID.Hex(),
	}

	if err == nil {
		hostInfo["username"] = host.Username
		hostInfo["display_name"] = host.DisplayName
		hostInfo["profile_image"] = host.ProfileImage
	}

	response.Success(c, http.StatusOK, "Stream status retrieved successfully", gin.H{
		"stream_id":     streamID.Hex(),
		"title":         stream.Title,
		"description":   stream.Description,
		"thumbnail":     stream.Thumbnail,
		"status":        status,
		"host":          hostInfo,
		"viewer_count":  stream.ViewerCount,
		"total_views":   stream.TotalViews,
		"category":      stream.Category,
		"tags":          stream.Tags,
		"started_at":    stream.StartedAt,
		"is_private":    stream.IsPrivate,
		"has_recording": stream.RecordingURL != "",
		"recording_url": stream.RecordingURL,
		"player_url":    "https://yourdomain.com/streams/" + streamID.Hex(),
	})
}
