package live

import (
	"net/http"
	"strconv"

	"github.com/Caqil/vyrall/internal/utils/response"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ViewerService defines the interface for live stream viewer operations
type ViewerService interface {
	GetViewers(ctx context.Context, streamID primitive.ObjectID, limit, offset int) ([]*models.LiveStreamViewer, int, error)
	GetActiveViewers(ctx context.Context, streamID primitive.ObjectID) ([]*models.LiveStreamViewer, error)
	GetViewerCount(ctx context.Context, streamID primitive.ObjectID) (int, error)
	GetViewerInfo(ctx context.Context, streamID, viewerID primitive.ObjectID) (*models.LiveStreamViewer, error)
	IsViewerActive(ctx context.Context, streamID, userID primitive.ObjectID) (bool, error)
}

// GetViewers returns the viewers of a live stream
func GetViewers(c *gin.Context) {
	// Get stream ID from URL parameter
	streamIDStr := c.Param("id")
	streamID, err := primitive.ObjectIDFromHex(streamIDStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid stream ID format", err)
		return
	}

	// Get query parameters
	activeOnly := c.DefaultQuery("active", "true") == "true"
	limitStr := c.DefaultQuery("limit", "50")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		limit = 50
	} else if limit > 100 {
		limit = 100 // Cap at 100 for performance
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	// Get the live stream service
	liveStreamService := c.MustGet("liveStreamService").(LiveStreamService)

	// Check if the stream exists
	stream, err := liveStreamService.GetStreamByID(c.Request.Context(), streamID)
	if err != nil {
		response.Error(c, http.StatusNotFound, "Stream not found", err)
		return
	}

	// Get the viewer service
	viewerService := c.MustGet("viewerService").(ViewerService)

	var viewers []*models.LiveStreamViewer
	var total int

	// Get viewers based on the active_only parameter
	if activeOnly {
		viewers, err = viewerService.GetActiveViewers(c.Request.Context(), streamID)
		if err != nil {
			response.Error(c, http.StatusInternalServerError, "Failed to retrieve active viewers", err)
			return
		}
		total = len(viewers)
	} else {
		viewers, total, err = viewerService.GetViewers(c.Request.Context(), streamID, limit, offset)
		if err != nil {
			response.Error(c, http.StatusInternalServerError, "Failed to retrieve viewers", err)
			return
		}
	}

	// If there are no viewers, return an empty array
	if len(viewers) == 0 {
		if activeOnly {
			response.Success(c, http.StatusOK, "No active viewers found", []interface{}{})
		} else {
			response.SuccessWithPagination(c, http.StatusOK, "No viewers found", []interface{}{}, limit, offset, total)
		}
		return
	}

	// Get the user service to include user details
	userService := c.MustGet("userService").(UserService)

	// Collect all user IDs
	userIDs := make([]primitive.ObjectID, 0, len(viewers))
	for _, viewer := range viewers {
		userIDs = append(userIDs, viewer.UserID)
	}

	// Get user details
	users, err := userService.GetUsersByIDs(c.Request.Context(), userIDs)
	if err != nil {
		// If user details retrieval fails, still return the viewers
		c.Error(err)
		if activeOnly {
			response.Success(c, http.StatusOK, "Active viewers retrieved successfully", viewers)
		} else {
			response.SuccessWithPagination(c, http.StatusOK, "Viewers retrieved successfully", viewers, limit, offset, total)
		}
		return
	}

	// Create a map of user ID to user for quick lookup
	userMap := make(map[string]*models.User)
	for _, user := range users {
		userMap[user.ID.Hex()] = user
	}

	// Combine viewers with user details
	viewerResponses := make([]map[string]interface{}, 0, len(viewers))
	for _, viewer := range viewers {
		user, exists := userMap[viewer.UserID.Hex()]

		viewerResponse := map[string]interface{}{
			"id":             viewer.ID.Hex(),
			"user_id":        viewer.UserID.Hex(),
			"joined_at":      viewer.JoinedAt,
			"device":         viewer.Device,
			"platform":       viewer.Platform,
			"is_active":      viewer.IsActive,
			"last_active_at": viewer.LastActiveAt,
		}

		if exists {
			viewerResponse["username"] = user.Username
			viewerResponse["display_name"] = user.DisplayName
			viewerResponse["profile_image"] = user.ProfileImage
		}

		viewerResponses = append(viewerResponses, viewerResponse)
	}

	if activeOnly {
		response.Success(c, http.StatusOK, "Active viewers retrieved successfully", viewerResponses)
	} else {
		response.SuccessWithPagination(c, http.StatusOK, "Viewers retrieved successfully", viewerResponses, limit, offset, total)
	}
}

// GetViewerCount returns the current viewer count for a live stream
func GetViewerCount(c *gin.Context) {
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
	_, err = liveStreamService.GetStreamByID(c.Request.Context(), streamID)
	if err != nil {
		response.Error(c, http.StatusNotFound, "Stream not found", err)
		return
	}

	// Get the viewer service
	viewerService := c.MustGet("viewerService").(ViewerService)

	// Get the viewer count
	count, err := viewerService.GetViewerCount(c.Request.Context(), streamID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve viewer count", err)
		return
	}

	response.Success(c, http.StatusOK, "Viewer count retrieved successfully", gin.H{
		"stream_id":    streamID.Hex(),
		"viewer_count": count,
	})
}

// GetViewerInfo returns details about a specific viewer in a live stream
func GetViewerInfo(c *gin.Context) {
	// Get stream ID and viewer ID from URL parameters
	streamIDStr := c.Param("id")
	viewerIDStr := c.Param("viewer_id")

	streamID, err := primitive.ObjectIDFromHex(streamIDStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid stream ID format", err)
		return
	}

	viewerID, err := primitive.ObjectIDFromHex(viewerIDStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid viewer ID format", err)
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

	// Check if the user is the stream host or a moderator
	isHost := stream.HostID == userID.(primitive.ObjectID)

	if !isHost {
		isModerator, err := liveStreamService.IsStreamModerator(c.Request.Context(), streamID, userID.(primitive.ObjectID))
		if err != nil {
			response.Error(c, http.StatusInternalServerError, "Failed to check if user is a moderator", err)
			return
		}

		if !isModerator {
			response.Error(c, http.StatusForbidden, "Only the stream host or moderators can view viewer details", nil)
			return
		}
	}

	// Get the viewer service
	viewerService := c.MustGet("viewerService").(ViewerService)

	// Get the viewer info
	viewer, err := viewerService.GetViewerInfo(c.Request.Context(), streamID, viewerID)
	if err != nil {
		response.Error(c, http.StatusNotFound, "Viewer not found", err)
		return
	}

	// Get the user service to include user details
	userService := c.MustGet("userService").(UserService)
	user, err := userService.GetUserByID(c.Request.Context(), viewer.UserID)

	// Create response with user details if available
	viewerResponse := map[string]interface{}{
		"id":             viewer.ID.Hex(),
		"user_id":        viewer.UserID.Hex(),
		"joined_at":      viewer.JoinedAt,
		"device":         viewer.Device,
		"platform":       viewer.Platform,
		"is_active":      viewer.IsActive,
		"last_active_at": viewer.LastActiveAt,
	}

	if err == nil {
		viewerResponse["username"] = user.Username
		viewerResponse["display_name"] = user.DisplayName
		viewerResponse["profile_image"] = user.ProfileImage
		viewerResponse["is_following_host"] = false // This would need to be determined by your service
		viewerResponse["account_created_at"] = user.CreatedAt
	}

	response.Success(c, http.StatusOK, "Viewer information retrieved successfully", viewerResponse)
}

// CheckViewerStatus checks if a user is currently viewing a live stream
func CheckViewerStatus(c *gin.Context) {
	// Get stream ID and user ID from URL parameters
	streamIDStr := c.Param("id")
	userIDStr := c.Param("user_id")

	streamID, err := primitive.ObjectIDFromHex(streamIDStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid stream ID format", err)
		return
	}

	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid user ID format", err)
		return
	}

	// Get the authenticated user's ID
	currentUserID, exists := c.Get("userID")
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

	// Check if the current user has permission to check viewer status
	// Only the stream host, moderators, or the user themselves can check their status
	isHost := stream.HostID == currentUserID.(primitive.ObjectID)
	isSelf := userID == currentUserID.(primitive.ObjectID)

	isModerator := false
	if !isHost && !isSelf {
		isModerator, err = liveStreamService.IsStreamModerator(c.Request.Context(), streamID, currentUserID.(primitive.ObjectID))
		if err != nil {
			response.Error(c, http.StatusInternalServerError, "Failed to check if user is a moderator", err)
			return
		}
	}

	if !isHost && !isModerator && !isSelf {
		response.Error(c, http.StatusForbidden, "You don't have permission to check this viewer's status", nil)
		return
	}

	// Get the viewer service
	viewerService := c.MustGet("viewerService").(ViewerService)

	// Check if the user is an active viewer
	isActive, err := viewerService.IsViewerActive(c.Request.Context(), streamID, userID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to check viewer status", err)
		return
	}

	response.Success(c, http.StatusOK, "Viewer status checked successfully", gin.H{
		"stream_id": streamID.Hex(),
		"user_id":   userID.Hex(),
		"is_active": isActive,
	})
}
