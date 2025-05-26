package live

import (
	"context"
	"net/http"

	"github.com/Caqil/vyrall/internal/utils/response"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// LeaveStreamService defines the interface for leaving a live stream
type LeaveStreamService interface {
	LeaveStream(ctx context.Context, streamID, userID primitive.ObjectID) error
	DecrementViewerCount(ctx context.Context, streamID primitive.ObjectID) error
	GetStreamStats(ctx context.Context, streamID primitive.ObjectID) (map[string]interface{}, error)
}

// LeaveStream handles a user leaving a live stream
func LeaveStream(c *gin.Context) {
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
	_, err = liveStreamService.GetStreamByID(c.Request.Context(), streamID)
	if err != nil {
		response.Error(c, http.StatusNotFound, "Stream not found", err)
		return
	}

	// Get the leave stream service
	leaveStreamService := c.MustGet("leaveStreamService").(LeaveStreamService)

	// Leave the stream
	err = leaveStreamService.LeaveStream(c.Request.Context(), streamID, userID.(primitive.ObjectID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to leave stream", err)
		return
	}

	// Update viewer count
	err = leaveStreamService.DecrementViewerCount(c.Request.Context(), streamID)
	if err != nil {
		// Log the error but don't fail the request
		c.Error(err)
	}

	// Get stream stats for response
	stats, err := leaveStreamService.GetStreamStats(c.Request.Context(), streamID)
	if err != nil {
		// Log the error but don't fail the request
		c.Error(err)
		stats = map[string]interface{}{
			"stream_id": streamID.Hex(),
		}
	}

	response.Success(c, http.StatusOK, "Left stream successfully", stats)
}
