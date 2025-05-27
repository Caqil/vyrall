package users

import (
	"net/http"
	"time"

	"github.com/Caqil/vyrall/internal/services/user"
	"github.com/Caqil/vyrall/internal/utils/response"
	"github.com/Caqil/vyrall/internal/utils/validation"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// MuteHandler handles user muting operations
type MuteHandler struct {
	userService *user.Service
}

// NewMuteHandler creates a new mute handler
func NewMuteHandler(userService *user.Service) *MuteHandler {
	return &MuteHandler{
		userService: userService,
	}
}

// MuteUser handles the request to mute a user
func (h *MuteHandler) MuteUser(c *gin.Context) {
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

	// Check if trying to mute self
	if userID.(primitive.ObjectID) == targetID {
		response.ValidationError(c, "You cannot mute yourself", nil)
		return
	}

	// Parse request body
	var req struct {
		Duration    int    `json:"duration,omitempty"` // in days, 0 means indefinite
		MuteStories bool   `json:"mute_stories,omitempty"`
		MutePosts   bool   `json:"mute_posts,omitempty"`
		Reason      string `json:"reason,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		// If there's an error parsing the request body, set defaults
		req.Duration = 0 // indefinite
		req.MuteStories = true
		req.MutePosts = true
	}

	// Calculate expiry time
	var expiresAt *time.Time
	if req.Duration > 0 {
		expiry := time.Now().AddDate(0, 0, req.Duration)
		expiresAt = &expiry
	}

	// Mute the user
	err := h.userService.MuteUser(
		c.Request.Context(),
		userID.(primitive.ObjectID),
		targetID,
		expiresAt,
		req.MuteStories,
		req.MutePosts,
		req.Reason,
	)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to mute user", err)
		return
	}

	// Return success response
	if expiresAt != nil {
		response.OK(c, "User muted successfully", gin.H{
			"expires_at": expiresAt,
		})
	} else {
		response.OK(c, "User muted successfully", gin.H{
			"expires_at": nil,
		})
	}
}

// GetMutedUsers handles the request to get a list of muted users
func (h *MuteHandler) GetMutedUsers(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Get pagination parameters
	limit, offset := response.GetPaginationParams(c)

	// Get muted users
	mutedUsers, total, err := h.userService.GetMutedUsers(c.Request.Context(), userID.(primitive.ObjectID), limit, offset)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve muted users", err)
		return
	}

	// Return paginated response
	response.SuccessWithPagination(c, http.StatusOK, "Muted users retrieved successfully", mutedUsers, limit, offset, total)
}

// UnmuteUser handles the request to unmute a user
func (h *MuteHandler) UnmuteUser(c *gin.Context) {
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

	// Unmute the user
	err := h.userService.UnmuteUser(c.Request.Context(), userID.(primitive.ObjectID), targetID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to unmute user", err)
		return
	}

	// Return success response
	response.OK(c, "User unmuted successfully", nil)
}

// CheckMuteStatus handles the request to check if a user is muted
func (h *MuteHandler) CheckMuteStatus(c *gin.Context) {
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

	// Check mute status
	status, err := h.userService.CheckMuteStatus(c.Request.Context(), userID.(primitive.ObjectID), targetID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to check mute status", err)
		return
	}

	// Return success response
	response.OK(c, "Mute status checked successfully", status)
}
