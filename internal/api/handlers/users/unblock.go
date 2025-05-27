package users

import (
	"net/http"

	"github.com/Caqil/vyrall/internal/services/user"
	"github.com/Caqil/vyrall/internal/utils/response"
	"github.com/Caqil/vyrall/internal/utils/validation"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// UnblockHandler handles user unblocking operations
type UnblockHandler struct {
	userService *user.Service
}

// NewUnblockHandler creates a new unblock handler
func NewUnblockHandler(userService *user.Service) *UnblockHandler {
	return &UnblockHandler{
		userService: userService,
	}
}

// UnblockUser handles the request to unblock a user
func (h *UnblockHandler) UnblockUser(c *gin.Context) {
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

	// Unblock the user
	err := h.userService.UnblockUser(c.Request.Context(), userID.(primitive.ObjectID), targetID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to unblock user", err)
		return
	}

	// Return success response
	response.OK(c, "User unblocked successfully", nil)
}

// BulkUnblockUsers handles the request to unblock multiple users
func (h *UnblockHandler) BulkUnblockUsers(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Parse request body
	var req struct {
		UserIDs []string `json:"user_ids" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "Invalid request body", err.Error())
		return
	}

	if len(req.UserIDs) == 0 {
		response.ValidationError(c, "No user IDs provided", nil)
		return
	}

	// Convert user IDs to ObjectIDs
	userIDs := make([]primitive.ObjectID, 0, len(req.UserIDs))
	for _, idStr := range req.UserIDs {
		if !validation.IsValidObjectID(idStr) {
			response.ValidationError(c, "Invalid user ID: "+idStr, nil)
			return
		}
		id, _ := primitive.ObjectIDFromHex(idStr)
		userIDs = append(userIDs, id)
	}

	// Unblock users
	count, err := h.userService.BulkUnblockUsers(c.Request.Context(), userID.(primitive.ObjectID), userIDs)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to unblock users", err)
		return
	}

	// Return success response
	response.OK(c, "Users unblocked successfully", gin.H{
		"unblocked_count": count,
	})
}
