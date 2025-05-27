package users

import (
	"net/http"

	"github.com/Caqil/vyrall/internal/services/user"
	"github.com/Caqil/vyrall/internal/utils/response"
	"github.com/Caqil/vyrall/internal/utils/validation"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// BlockHandler handles user blocking operations
type BlockHandler struct {
	userService *user.Service
}

// NewBlockHandler creates a new block handler
func NewBlockHandler(userService *user.Service) *BlockHandler {
	return &BlockHandler{
		userService: userService,
	}
}

// BlockUser handles the request to block a user
func (h *BlockHandler) BlockUser(c *gin.Context) {
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

	// Check if trying to block self
	if userID.(primitive.ObjectID) == targetID {
		response.ValidationError(c, "You cannot block yourself", nil)
		return
	}

	// Block the user
	err := h.userService.BlockUser(c.Request.Context(), userID.(primitive.ObjectID), targetID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to block user", err)
		return
	}

	// Return success response
	response.OK(c, "User blocked successfully", nil)
}

// GetBlockedUsers handles the request to get a list of blocked users
func (h *BlockHandler) GetBlockedUsers(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Get pagination parameters
	limit, offset := response.GetPaginationParams(c)

	// Get blocked users
	blockedUsers, total, err := h.userService.GetBlockedUsers(c.Request.Context(), userID.(primitive.ObjectID), limit, offset)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve blocked users", err)
		return
	}

	// Return paginated response
	response.SuccessWithPagination(c, http.StatusOK, "Blocked users retrieved successfully", blockedUsers, limit, offset, total)
}

// CheckBlockStatus handles the request to check if a user is blocked
func (h *BlockHandler) CheckBlockStatus(c *gin.Context) {
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

	// Check block status
	isBlocked, err := h.userService.CheckBlockStatus(c.Request.Context(), userID.(primitive.ObjectID), targetID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to check block status", err)
		return
	}

	// Return success response
	response.OK(c, "Block status checked successfully", gin.H{
		"is_blocked": isBlocked,
	})
}
