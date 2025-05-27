package users

import (
	"net/http"

	"github.com/Caqil/vyrall/internal/services/user"
	"github.com/Caqil/vyrall/internal/utils/response"
	"github.com/Caqil/vyrall/internal/utils/validation"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// FollowHandler handles user follow operations
type FollowHandler struct {
	userService *user.Service
}

// NewFollowHandler creates a new follow handler
func NewFollowHandler(userService *user.Service) *FollowHandler {
	return &FollowHandler{
		userService: userService,
	}
}

// FollowUser handles the request to follow a user
func (h *FollowHandler) FollowUser(c *gin.Context) {
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

	// Check if trying to follow self
	if userID.(primitive.ObjectID) == targetID {
		response.ValidationError(c, "You cannot follow yourself", nil)
		return
	}

	// Parse request body (optional)
	var req struct {
		NotifyPosts bool `json:"notify_posts,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		// If there's an error parsing the request body, just ignore it
		req.NotifyPosts = false
	}

	// Follow the user
	status, err := h.userService.FollowUser(c.Request.Context(), userID.(primitive.ObjectID), targetID, req.NotifyPosts)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to follow user", err)
		return
	}

	// Return success response
	if status == "pending" {
		response.OK(c, "Follow request sent", gin.H{
			"status": status,
		})
	} else {
		response.OK(c, "User followed successfully", gin.H{
			"status": status,
		})
	}
}

// GetFollowers handles the request to get a user's followers
func (h *FollowHandler) GetFollowers(c *gin.Context) {
	// Get target user ID from URL parameter
	targetIDStr := c.Param("id")
	if !validation.IsValidObjectID(targetIDStr) {
		response.ValidationError(c, "Invalid user ID", nil)
		return
	}
	targetID, _ := primitive.ObjectIDFromHex(targetIDStr)

	// Get authenticated user ID (may be nil for unauthenticated users)
	var userID primitive.ObjectID
	if id, exists := c.Get("userID"); exists {
		userID = id.(primitive.ObjectID)
	}

	// Get pagination parameters
	limit, offset := response.GetPaginationParams(c)

	// Get followers
	followers, total, err := h.userService.GetFollowers(c.Request.Context(), targetID, userID, limit, offset)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve followers", err)
		return
	}

	// Return paginated response
	response.SuccessWithPagination(c, http.StatusOK, "Followers retrieved successfully", followers, limit, offset, total)
}

// GetFollowing handles the request to get users a user is following
func (h *FollowHandler) GetFollowing(c *gin.Context) {
	// Get target user ID from URL parameter
	targetIDStr := c.Param("id")
	if !validation.IsValidObjectID(targetIDStr) {
		response.ValidationError(c, "Invalid user ID", nil)
		return
	}
	targetID, _ := primitive.ObjectIDFromHex(targetIDStr)

	// Get authenticated user ID (may be nil for unauthenticated users)
	var userID primitive.ObjectID
	if id, exists := c.Get("userID"); exists {
		userID = id.(primitive.ObjectID)
	}

	// Get pagination parameters
	limit, offset := response.GetPaginationParams(c)

	// Get following
	following, total, err := h.userService.GetFollowing(c.Request.Context(), targetID, userID, limit, offset)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve following", err)
		return
	}

	// Return paginated response
	response.SuccessWithPagination(c, http.StatusOK, "Following retrieved successfully", following, limit, offset, total)
}

// GetFollowRequests handles the request to get pending follow requests
func (h *FollowHandler) GetFollowRequests(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Get pagination parameters
	limit, offset := response.GetPaginationParams(c)

	// Get follow requests
	requests, total, err := h.userService.GetFollowRequests(c.Request.Context(), userID.(primitive.ObjectID), limit, offset)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve follow requests", err)
		return
	}

	// Return paginated response
	response.SuccessWithPagination(c, http.StatusOK, "Follow requests retrieved successfully", requests, limit, offset, total)
}

// ApproveFollowRequest handles the request to approve a follow request
func (h *FollowHandler) ApproveFollowRequest(c *gin.Context) {
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

	// Approve the follow request
	err := h.userService.ApproveFollowRequest(c.Request.Context(), userID.(primitive.ObjectID), followerID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to approve follow request", err)
		return
	}

	// Return success response
	response.OK(c, "Follow request approved", nil)
}

// RejectFollowRequest handles the request to reject a follow request
func (h *FollowHandler) RejectFollowRequest(c *gin.Context) {
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

	// Reject the follow request
	err := h.userService.RejectFollowRequest(c.Request.Context(), userID.(primitive.ObjectID), followerID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to reject follow request", err)
		return
	}

	// Return success response
	response.OK(c, "Follow request rejected", nil)
}
