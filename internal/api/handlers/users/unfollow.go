package users

import (
	"net/http"

	"github.com/Caqil/vyrall/internal/services/user"
	"github.com/Caqil/vyrall/internal/utils/response"
	"github.com/Caqil/vyrall/internal/utils/validation"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// UnfollowHandler handles user unfollow operations
type UnfollowHandler struct {
	userService *user.Service
}

// NewUnfollowHandler creates a new unfollow handler
func NewUnfollowHandler(userService *user.Service) *UnfollowHandler {
	return &UnfollowHandler{
		userService: userService,
	}
}

// UnfollowUser handles the request to unfollow a user
func (h *UnfollowHandler) UnfollowUser(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Get target user ID from URL parameter
	targetIDStr := c.Param("id")
	if !validation.IsValidObjectID(targetIDStr) {
		response.ValidationError(c, "Invalid user ID", nil)
		return
	}
	targetID, _ := primitive.ObjectIDFromHex(targetIDStr)

	// Unfollow the user
	err := h.userService.UnfollowUser(c.Request.Context(), userID.(primitive.ObjectID), targetID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to unfollow user", err)
		return
	}

	// Return success response
	response.OK(c, "User unfollowed successfully", nil)
}

// CancelFollowRequest handles the request to cancel a pending follow request
func (h *UnfollowHandler) CancelFollowRequest(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Get target user ID from URL parameter
	targetIDStr := c.Param("id")
	if !validation.IsValidObjectID(targetIDStr) {
		response.ValidationError(c, "Invalid user ID", nil)
		return
	}
	targetID, _ := primitive.ObjectIDFromHex(targetIDStr)

	// Cancel follow request
	err := h.userService.CancelFollowRequest(c.Request.Context(), userID.(primitive.ObjectID), targetID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to cancel follow request", err)
		return
	}

	// Return success response
	response.OK(c, "Follow request cancelled successfully", nil)
}

// RemoveFollower handles the request to remove a follower
func (h *UnfollowHandler) RemoveFollower(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Get follower ID from URL parameter
	followerIDStr := c.Param("id")
	if !validation.IsValidObjectID(followerIDStr) {
		response.ValidationError(c, "Invalid user ID", nil)
		return
	}
	followerID, _ := primitive.ObjectIDFromHex(followerIDStr)

	// Remove follower
	err := h.userService.RemoveFollower(c.Request.Context(), userID.(primitive.ObjectID), followerID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to remove follower", err)
		return
	}

	// Return success response
	response.OK(c, "Follower removed successfully", nil)
}
