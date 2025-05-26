package live

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/Caqil/vyrall/internal/models"
	"github.com/Caqil/vyrall/internal/utils/response"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ReactionService defines the interface for live stream reaction operations
type ReactionService interface {
	AddReaction(ctx context.Context, reaction *models.LiveStreamReaction) (primitive.ObjectID, error)
	GetReactions(ctx context.Context, streamID primitive.ObjectID, limit, offset int) ([]*models.LiveStreamReaction, int, error)
	GetReactionCounts(ctx context.Context, streamID primitive.ObjectID) (map[string]int, error)
	IncrementReactionCount(ctx context.Context, streamID primitive.ObjectID, reactionType string) error
}

// AddReaction handles adding a reaction to a live stream
func AddReaction(c *gin.Context) {
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
		Type string `json:"type" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request format", err)
		return
	}

	// Validate reaction type
	validReactions := map[string]bool{
		"like":     true,
		"love":     true,
		"haha":     true,
		"wow":      true,
		"sad":      true,
		"angry":    true,
		"fire":     true,
		"clap":     true,
		"star":     true,
		"question": true,
	}

	if !validReactions[req.Type] {
		response.Error(c, http.StatusBadRequest, "Invalid reaction type", nil)
		return
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
		response.Error(c, http.StatusBadRequest, "Cannot react to an inactive stream", nil)
		return
	}

	// Check if the user is banned from the stream
	isBanned, err := liveStreamService.IsUserBanned(c.Request.Context(), streamID, userID.(primitive.ObjectID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to check if user is banned", err)
		return
	}

	if isBanned {
		response.Error(c, http.StatusForbidden, "You are banned from interacting with this stream", nil)
		return
	}

	// Get the reaction service
	reactionService := c.MustGet("reactionService").(ReactionService)

	// Create new reaction
	reaction := &models.LiveStreamReaction{
		StreamID:  streamID,
		UserID:    userID.(primitive.ObjectID),
		Type:      req.Type,
		CreatedAt: time.Now(),
	}

	// Add the reaction
	reactionID, err := reactionService.AddReaction(c.Request.Context(), reaction)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to add reaction", err)
		return
	}

	// Update reaction count
	err = reactionService.IncrementReactionCount(c.Request.Context(), streamID, req.Type)
	if err != nil {
		// Log the error but don't fail the request
		c.Error(err)
	}

	response.Success(c, http.StatusCreated, "Reaction added successfully", gin.H{
		"reaction_id": reactionID.Hex(),
		"stream_id":   streamID.Hex(),
		"type":        req.Type,
		"created_at":  reaction.CreatedAt,
	})
}

// GetReactions returns reactions for a live stream
func GetReactions(c *gin.Context) {
	// Get stream ID from URL parameter
	streamIDStr := c.Param("id")
	streamID, err := primitive.ObjectIDFromHex(streamIDStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid stream ID format", err)
		return
	}

	// Get pagination parameters
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
	_, err = liveStreamService.GetStreamByID(c.Request.Context(), streamID)
	if err != nil {
		response.Error(c, http.StatusNotFound, "Stream not found", err)
		return
	}

	// Get the reaction service
	reactionService := c.MustGet("reactionService").(ReactionService)

	// Get the reactions
	reactions, total, err := reactionService.GetReactions(c.Request.Context(), streamID, limit, offset)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve reactions", err)
		return
	}

	// If there are no reactions, return an empty array
	if len(reactions) == 0 {
		response.SuccessWithPagination(c, http.StatusOK, "No reactions found", []interface{}{}, limit, offset, total)
		return
	}

	// Get the user service to include user details
	userService := c.MustGet("userService").(UserService)

	// Collect all user IDs
	userIDs := make([]primitive.ObjectID, 0, len(reactions))
	for _, reaction := range reactions {
		userIDs = append(userIDs, reaction.UserID)
	}

	// Get user details
	users, err := userService.GetUsersByIDs(c.Request.Context(), userIDs)
	if err != nil {
		// If user details retrieval fails, still return the reactions
		c.Error(err)
		response.SuccessWithPagination(c, http.StatusOK, "Reactions retrieved successfully", reactions, limit, offset, total)
		return
	}

	// Create a map of user ID to user for quick lookup
	userMap := make(map[string]*models.User)
	for _, user := range users {
		userMap[user.ID.Hex()] = user
	}

	// Combine reactions with user details
	reactionResponses := make([]map[string]interface{}, 0, len(reactions))
	for _, reaction := range reactions {
		user, exists := userMap[reaction.UserID.Hex()]

		reactionResponse := map[string]interface{}{
			"id":         reaction.ID.Hex(),
			"stream_id":  reaction.StreamID.Hex(),
			"user_id":    reaction.UserID.Hex(),
			"type":       reaction.Type,
			"created_at": reaction.CreatedAt,
		}

		if exists {
			reactionResponse["username"] = user.Username
			reactionResponse["display_name"] = user.DisplayName
			reactionResponse["profile_image"] = user.ProfileImage
		}

		reactionResponses = append(reactionResponses, reactionResponse)
	}

	response.SuccessWithPagination(c, http.StatusOK, "Reactions retrieved successfully", reactionResponses, limit, offset, total)
}

// GetReactionCounts returns a summary of reaction counts by type for a live stream
func GetReactionCounts(c *gin.Context) {
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

	// Get the reaction service
	reactionService := c.MustGet("reactionService").(ReactionService)

	// Get reaction counts
	counts, err := reactionService.GetReactionCounts(c.Request.Context(), streamID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve reaction counts", err)
		return
	}

	// Calculate total reactions
	total := 0
	for _, count := range counts {
		total += count
	}

	response.Success(c, http.StatusOK, "Reaction counts retrieved successfully", gin.H{
		"counts": counts,
		"total":  total,
	})
}
