package live

import (
	"context"
	"net/http"

	"github.com/Caqil/vyrall/internal/models"
	"github.com/Caqil/vyrall/internal/utils/response"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ModerationService defines the interface for live stream moderation operations
type ModerationService interface {
	AddModerator(ctx context.Context, streamID, moderatorID, userID primitive.ObjectID) error
	RemoveModerator(ctx context.Context, streamID, moderatorID, userID primitive.ObjectID) error
	GetModerators(ctx context.Context, streamID primitive.ObjectID) ([]primitive.ObjectID, error)
	BanUser(ctx context.Context, streamID, bannedUserID, userID primitive.ObjectID, reason string, duration int) error
	UnbanUser(ctx context.Context, streamID, bannedUserID, userID primitive.ObjectID) error
	GetBannedUsers(ctx context.Context, streamID primitive.ObjectID) ([]models.BannedUser, error)
	UpdateChatSettings(ctx context.Context, streamID primitive.ObjectID, settings models.LiveStreamChatSettings, userID primitive.ObjectID) error
	GetChatSettings(ctx context.Context, streamID primitive.ObjectID) (models.LiveStreamChatSettings, error)
	ModerateComment(ctx context.Context, commentID primitive.ObjectID, isHidden bool, moderatorID primitive.ObjectID, reason string) error
}

// AddModerator handles adding a moderator to a live stream
func AddModerator(c *gin.Context) {
	// Get stream ID from URL parameter
	streamIDStr := c.Param("id")
	streamID, err := primitive.ObjectIDFromHex(streamIDStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid stream ID format", err)
		return
	}

	// Parse request body
	var req struct {
		UserID string `json:"user_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request format", err)
		return
	}

	moderatorID, err := primitive.ObjectIDFromHex(req.UserID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid user ID format", err)
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
		response.Error(c, http.StatusForbidden, "Only the stream host can add moderators", nil)
		return
	}

	// Check if the moderator is already the host
	if moderatorID == stream.HostID {
		response.Error(c, http.StatusBadRequest, "The stream host cannot be added as a moderator", nil)
		return
	}

	// Get the moderation service
	moderationService := c.MustGet("moderationService").(ModerationService)

	// Add the moderator
	err = moderationService.AddModerator(c.Request.Context(), streamID, moderatorID, userID.(primitive.ObjectID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to add moderator", err)
		return
	}

	// Get the user service to include moderator details
	userService := c.MustGet("userService").(UserService)
	moderator, err := userService.GetUserByID(c.Request.Context(), moderatorID)

	moderatorInfo := map[string]interface{}{
		"user_id": moderatorID.Hex(),
	}

	if err == nil {
		moderatorInfo["username"] = moderator.Username
		moderatorInfo["display_name"] = moderator.DisplayName
		moderatorInfo["profile_image"] = moderator.ProfileImage
	}

	response.Success(c, http.StatusOK, "Moderator added successfully", moderatorInfo)
}

// RemoveModerator handles removing a moderator from a live stream
func RemoveModerator(c *gin.Context) {
	// Get stream ID and moderator ID from URL parameters
	streamIDStr := c.Param("id")
	moderatorIDStr := c.Param("moderator_id")

	streamID, err := primitive.ObjectIDFromHex(streamIDStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid stream ID format", err)
		return
	}

	moderatorID, err := primitive.ObjectIDFromHex(moderatorIDStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid moderator ID format", err)
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
		response.Error(c, http.StatusForbidden, "Only the stream host can remove moderators", nil)
		return
	}

	// Get the moderation service
	moderationService := c.MustGet("moderationService").(ModerationService)

	// Remove the moderator
	err = moderationService.RemoveModerator(c.Request.Context(), streamID, moderatorID, userID.(primitive.ObjectID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to remove moderator", err)
		return
	}

	response.Success(c, http.StatusOK, "Moderator removed successfully", nil)
}

// GetModerators returns all moderators for a live stream
func GetModerators(c *gin.Context) {
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

	// Get the moderation service
	moderationService := c.MustGet("moderationService").(ModerationService)

	// Get the moderators
	moderatorIDs, err := moderationService.GetModerators(c.Request.Context(), streamID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve moderators", err)
		return
	}

	// If there are no moderators, return an empty array
	if len(moderatorIDs) == 0 {
		response.Success(c, http.StatusOK, "No moderators found", []interface{}{})
		return
	}

	// Get the user service to include moderator details
	userService := c.MustGet("userService").(UserService)
	moderators, err := userService.GetUsersByIDs(c.Request.Context(), moderatorIDs)
	if err != nil {
		// If user details retrieval fails, still return the moderator IDs
		c.Error(err)
		moderatorIDStrings := make([]string, len(moderatorIDs))
		for i, id := range moderatorIDs {
			moderatorIDStrings[i] = id.Hex()
		}
		response.Success(c, http.StatusOK, "Moderators retrieved successfully", gin.H{
			"moderator_ids": moderatorIDStrings,
		})
		return
	}

	// Create response with moderator details
	moderatorResponses := make([]map[string]interface{}, 0, len(moderators))
	for _, moderator := range moderators {
		moderatorResponses = append(moderatorResponses, map[string]interface{}{
			"user_id":       moderator.ID.Hex(),
			"username":      moderator.Username,
			"display_name":  moderator.DisplayName,
			"profile_image": moderator.ProfileImage,
		})
	}

	response.Success(c, http.StatusOK, "Moderators retrieved successfully", moderatorResponses)
}

// BanUser handles banning a user from a live stream
func BanUser(c *gin.Context) {
	// Get stream ID from URL parameter
	streamIDStr := c.Param("id")
	streamID, err := primitive.ObjectIDFromHex(streamIDStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid stream ID format", err)
		return
	}

	// Parse request body
	var req struct {
		UserID   string `json:"user_id" binding:"required"`
		Reason   string `json:"reason"`
		Duration int    `json:"duration"` // Ban duration in minutes, 0 for permanent
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request format", err)
		return
	}

	bannedUserID, err := primitive.ObjectIDFromHex(req.UserID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid user ID format", err)
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

	isModerator := false
	if !isHost {
		isModerator, err = liveStreamService.IsStreamModerator(c.Request.Context(), streamID, userID.(primitive.ObjectID))
		if err != nil {
			response.Error(c, http.StatusInternalServerError, "Failed to check if user is a moderator", err)
			return
		}
	}

	if !isHost && !isModerator {
		response.Error(c, http.StatusForbidden, "Only the stream host or moderators can ban users", nil)
		return
	}

	// Cannot ban the stream host
	if bannedUserID == stream.HostID {
		response.Error(c, http.StatusBadRequest, "Cannot ban the stream host", nil)
		return
	}

	// Moderators cannot ban other moderators
	if isModerator && !isHost {
		// Check if the user to ban is a moderator
		moderationService := c.MustGet("moderationService").(ModerationService)
		moderatorIDs, err := moderationService.GetModerators(c.Request.Context(), streamID)
		if err != nil {
			response.Error(c, http.StatusInternalServerError, "Failed to retrieve moderators", err)
			return
		}

		for _, modID := range moderatorIDs {
			if modID == bannedUserID {
				response.Error(c, http.StatusForbidden, "Moderators cannot ban other moderators", nil)
				return
			}
		}
	}

	// Get the moderation service
	moderationService := c.MustGet("moderationService").(ModerationService)

	// Ban the user
	err = moderationService.BanUser(c.Request.Context(), streamID, bannedUserID, userID.(primitive.ObjectID), req.Reason, req.Duration)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to ban user", err)
		return
	}

	response.Success(c, http.StatusOK, "User banned successfully", nil)
}

// UnbanUser handles unbanning a user from a live stream
func UnbanUser(c *gin.Context) {
	// Get stream ID and banned user ID from URL parameters
	streamIDStr := c.Param("id")
	bannedUserIDStr := c.Param("user_id")

	streamID, err := primitive.ObjectIDFromHex(streamIDStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid stream ID format", err)
		return
	}

	bannedUserID, err := primitive.ObjectIDFromHex(bannedUserIDStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid user ID format", err)
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
			response.Error(c, http.StatusForbidden, "Only the stream host or moderators can unban users", nil)
			return
		}
	}

	// Get the moderation service
	moderationService := c.MustGet("moderationService").(ModerationService)

	// Unban the user
	err = moderationService.UnbanUser(c.Request.Context(), streamID, bannedUserID, userID.(primitive.ObjectID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to unban user", err)
		return
	}

	response.Success(c, http.StatusOK, "User unbanned successfully", nil)
}

// GetBannedUsers returns all banned users for a live stream
func GetBannedUsers(c *gin.Context) {
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

	// Check if the user is the stream host or a moderator
	isHost := stream.HostID == userID.(primitive.ObjectID)

	if !isHost {
		isModerator, err := liveStreamService.IsStreamModerator(c.Request.Context(), streamID, userID.(primitive.ObjectID))
		if err != nil {
			response.Error(c, http.StatusInternalServerError, "Failed to check if user is a moderator", err)
			return
		}

		if !isModerator {
			response.Error(c, http.StatusForbidden, "Only the stream host or moderators can view banned users", nil)
			return
		}
	}

	// Get the moderation service
	moderationService := c.MustGet("moderationService").(ModerationService)

	// Get banned users
	bannedUsers, err := moderationService.GetBannedUsers(c.Request.Context(), streamID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve banned users", err)
		return
	}

	// If there are no banned users, return an empty array
	if len(bannedUsers) == 0 {
		response.Success(c, http.StatusOK, "No banned users found", []interface{}{})
		return
	}

	// Get the user service to include user details
	userService := c.MustGet("userService").(UserService)

	// Collect all user IDs
	userIDs := make([]primitive.ObjectID, len(bannedUsers))
	for i, banned := range bannedUsers {
		userIDs[i] = banned.UserID
	}

	// Get user details
	users, err := userService.GetUsersByIDs(c.Request.Context(), userIDs)
	if err != nil {
		// If user details retrieval fails, still return the banned user info
		c.Error(err)
		response.Success(c, http.StatusOK, "Banned users retrieved successfully", bannedUsers)
		return
	}

	// Create a map of user ID to user for quick lookup
	userMap := make(map[string]*models.User)
	for _, user := range users {
		userMap[user.ID.Hex()] = user
	}

	// Create response with user details
	bannedResponses := make([]map[string]interface{}, 0, len(bannedUsers))
	for _, banned := range bannedUsers {
		user, exists := userMap[banned.UserID.Hex()]

		bannedResponse := map[string]interface{}{
			"user_id":    banned.UserID.Hex(),
			"banned_by":  banned.BannedByUserID.Hex(),
			"banned_at":  banned.BannedAt,
			"reason":     banned.Reason,
			"duration":   banned.Duration,
			"expires_at": banned.ExpiresAt,
		}

		if exists {
			bannedResponse["username"] = user.Username
			bannedResponse["display_name"] = user.DisplayName
			bannedResponse["profile_image"] = user.ProfileImage
		}

		bannedResponses = append(bannedResponses, bannedResponse)
	}

	response.Success(c, http.StatusOK, "Banned users retrieved successfully", bannedResponses)
}

// UpdateChatSettings handles updating chat settings for a live stream
func UpdateChatSettings(c *gin.Context) {
	// Get stream ID from URL parameter
	streamIDStr := c.Param("id")
	streamID, err := primitive.ObjectIDFromHex(streamIDStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid stream ID format", err)
		return
	}

	// Parse request body
	var req models.LiveStreamChatSettings
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request format", err)
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
		response.Error(c, http.StatusForbidden, "Only the stream host can update chat settings", nil)
		return
	}

	// Validate settings
	if req.SlowMode && req.SlowModeInterval <= 0 {
		response.Error(c, http.StatusBadRequest, "Slow mode interval must be greater than 0", nil)
		return
	}

	// Get the moderation service
	moderationService := c.MustGet("moderationService").(ModerationService)

	// Update chat settings
	err = moderationService.UpdateChatSettings(c.Request.Context(), streamID, req, userID.(primitive.ObjectID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to update chat settings", err)
		return
	}

	response.Success(c, http.StatusOK, "Chat settings updated successfully", req)
}

// GetChatSettings returns the chat settings for a live stream
func GetChatSettings(c *gin.Context) {
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

	// Get the moderation service
	moderationService := c.MustGet("moderationService").(ModerationService)

	// Get chat settings
	settings, err := moderationService.GetChatSettings(c.Request.Context(), streamID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve chat settings", err)
		return
	}

	response.Success(c, http.StatusOK, "Chat settings retrieved successfully", settings)
}

// ModerateComment handles moderating a comment in a live stream
func ModerateComment(c *gin.Context) {
	// Get stream ID and comment ID from URL parameters
	streamIDStr := c.Param("id")
	commentIDStr := c.Param("comment_id")

	streamID, err := primitive.ObjectIDFromHex(streamIDStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid stream ID format", err)
		return
	}

	commentID, err := primitive.ObjectIDFromHex(commentIDStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid comment ID format", err)
		return
	}

	// Parse request body
	var req struct {
		IsHidden bool   `json:"is_hidden"`
		Reason   string `json:"reason"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request format", err)
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

	// Check if the user is the stream host or a moderator
	isHost, err := liveStreamService.IsStreamHost(c.Request.Context(), streamID, userID.(primitive.ObjectID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to check if user is host", err)
		return
	}

	isModerator, err := liveStreamService.IsStreamModerator(c.Request.Context(), streamID, userID.(primitive.ObjectID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to check if user is moderator", err)
		return
	}

	if !isHost && !isModerator {
		response.Error(c, http.StatusForbidden, "Only the stream host or moderators can moderate comments", nil)
		return
	}

	// Get the moderation service
	moderationService := c.MustGet("moderationService").(ModerationService)

	// Moderate the comment
	err = moderationService.ModerateComment(c.Request.Context(), commentID, req.IsHidden, userID.(primitive.ObjectID), req.Reason)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to moderate comment", err)
		return
	}

	if req.IsHidden {
		response.Success(c, http.StatusOK, "Comment hidden successfully", nil)
	} else {
		response.Success(c, http.StatusOK, "Comment unhidden successfully", nil)
	}
}
