package live

import (
	"context"
	"net/http"
	"time"

	"github.com/Caqil/vyrall/internal/utils/response"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// EndStreamService defines the interface for ending a live stream
type EndStreamService interface {
	EndStream(ctx context.Context, streamID, userID primitive.ObjectID) error
	UpdateStreamStatus(ctx context.Context, streamID primitive.ObjectID, status string, endedAt time.Time) error
	SaveStreamRecording(ctx context.Context, streamID primitive.ObjectID, recordingURL string) error
}

// EndStream handles ending a live stream
func EndStream(c *gin.Context) {
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
	isHost, err := liveStreamService.IsStreamHost(c.Request.Context(), streamID, userID.(primitive.ObjectID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to check if user is host", err)
		return
	}

	if !isHost {
		response.Error(c, http.StatusForbidden, "Only the stream host can end the stream", nil)
		return
	}

	// Check if the stream is already ended
	if stream.Status != "active" {
		response.Error(c, http.StatusBadRequest, "Stream is not active", nil)
		return
	}

	// Parse request body (optional)
	var req struct {
		SaveRecording bool   `json:"save_recording"`
		RecordingURL  string `json:"recording_url"`
	}
	c.ShouldBindJSON(&req)

	// Get the end stream service
	endStreamService := c.MustGet("endStreamService").(EndStreamService)

	// End the stream
	err = endStreamService.EndStream(c.Request.Context(), streamID, userID.(primitive.ObjectID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to end stream", err)
		return
	}

	// Update stream status
	endedAt := time.Now()
	err = endStreamService.UpdateStreamStatus(c.Request.Context(), streamID, "ended", endedAt)
	if err != nil {
		// Log the error but don't fail the request
		c.Error(err)
	}

	// Save recording if requested
	if req.SaveRecording && req.RecordingURL != "" {
		err = endStreamService.SaveStreamRecording(c.Request.Context(), streamID, req.RecordingURL)
		if err != nil {
			// Log the error but don't fail the request
			c.Error(err)
		}
	}

	// Calculate stream duration
	duration := endedAt.Sub(stream.StartedAt)

	// Return success response with stream summary
	response.Success(c, http.StatusOK, "Stream ended successfully", gin.H{
		"stream_id":     streamID.Hex(),
		"title":         stream.Title,
		"duration":      duration.Seconds(),
		"started_at":    stream.StartedAt,
		"ended_at":      endedAt,
		"viewer_count":  stream.ViewerCount,
		"peak_viewers":  stream.PeakViewerCount,
		"total_viewers": stream.TotalViews,
		"status":        "ended",
	})
}
