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

// JoinStreamService defines the interface for joining a live stream
type JoinStreamService interface {
	JoinStream(ctx context.Context, streamID, userID primitive.ObjectID, device, platform string) (primitive.ObjectID, error)
	IncrementViewerCount(ctx context.Context, streamID primitive.ObjectID) error
	IncrementTotalViews(ctx context.Context, streamID primitive.ObjectID) error
	UpdatePeakViewerCount(ctx context.Context, streamID primitive.ObjectID) error
	GetStreamChatSettings(ctx context.Context, streamID primitive.ObjectID) (*models.LiveStreamChatSettings, error)
}

// UserService defines the interface for user operations
type UserService interface {
	GetUserByID(ctx context.Context, id primitive.ObjectID) (*models.User, error)
	GetUsersByIDs(ctx context.Context, ids []primitive.ObjectID) ([]*models.User, error)
}

// JoinStream handles a user joining a live stream
func JoinStream(c *gin.Context) {
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

	// Parse request body
	var req struct {
		Device   string `json:"device"`
		Platform string `json:"platform"`
	}
	c.ShouldBindJSON(&req)

	// Set defaults if not provided
	if req.Device == "" {
		req.Device = "unknown"
	}
	if req.Platform == "" {
		req.Platform = "web"
	}

	// Get the live stream service
	liveStreamService := c.MustGet("liveStreamService").(LiveStreamService)

	// Check if the stream exists and is active
	stream, err := liveStreamService.GetStreamByID(c.Request.Context(), streamID)
	if err != nil {
		response.Error(c, http.StatusNotFound, "Stream not found", err)
		return
	}

	if stream.Status != "active" {
		response.Error(c, http.StatusBadRequest, "Stream is not active", nil)
		return
	}

	// Check if the user is banned from the stream
	isBanned, err := liveStreamService.IsUserBanned(c.Request.Context(), streamID, userID.(primitive.ObjectID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to check if user is banned", err)
		return
	}

	if isBanned {
		response.Error(c, http.StatusForbidden, "You are banned from this stream", nil)
		return
	}

	// Get the join stream service
	joinStreamService := c.MustGet("joinStreamService").(JoinStreamService)

	// Join the stream
	viewerID, err := joinStreamService.JoinStream(c.Request.Context(), streamID, userID.(primitive.ObjectID), req.Device, req.Platform)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to join stream", err)
		return
	}

	// Update viewer counts
	err = joinStreamService.IncrementViewerCount(c.Request.Context(), streamID)
	if err != nil {
		// Log the error but don't fail the request
		c.Error(err)
	}

	err = joinStreamService.IncrementTotalViews(c.Request.Context(), streamID)
	if err != nil {
		// Log the error but don't fail the request
		c.Error(err)
	}

	err = joinStreamService.UpdatePeakViewerCount(c.Request.Context(), streamID)
	if err != nil {
		// Log the error but don't fail the request
		c.Error(err)
	}

	// Get chat settings
	chatSettings, err := joinStreamService.GetStreamChatSettings(c.Request.Context(), streamID)
	if err != nil {
		// Use default settings if retrieval fails
		chatSettings = &models.LiveStreamChatSettings{
			Enabled:             true,
			SlowMode:            false,
			SlowModeInterval:    0,
			FollowersOnlyMode:   false,
			SubscribersOnlyMode: false,
			EmoteOnlyMode:       false,
		}
	}

	// Get the host user's information
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

	// Return success response with stream info
	response.Success(c, http.StatusOK, "Joined stream successfully", gin.H{
		"viewer_id":     viewerID.Hex(),
		"stream_id":     streamID.Hex(),
		"title":         stream.Title,
		"description":   stream.Description,
		"host":          hostInfo,
		"status":        stream.Status,
		"started_at":    stream.StartedAt,
		"viewer_count":  stream.ViewerCount + 1, // Include the newly joined viewer
		"category":      stream.Category,
		"tags":          stream.Tags,
		"chat_settings": chatSettings,
		"joined_at":     time.Now(),
	})
}
