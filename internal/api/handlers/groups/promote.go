package groups

import (
	"net/http"

	"github.com/Caqil/vyrall/internal/utils/response"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// PromoteRequest represents the request payload for promoting a group member
type PromoteRequest struct {
	UserID string `json:"user_id" binding:"required"`
	Role   string `json:"role" binding:"required,oneof=moderator admin"`
}

// PromoteMember handles promoting a group member to a higher role
func PromoteMember(c *gin.Context) {
	// Get group ID from URL parameter
	groupIDStr := c.Param("id")
	groupID, err := primitive.ObjectIDFromHex(groupIDStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid group ID format", err)
		return
	}

	// Get the authenticated user's ID
	currentUserID, exists := c.Get("userID")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	// Parse and validate the request
	var req PromoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request format", err)
		return
	}

	// Parse target user ID
	targetUserID, err := primitive.ObjectIDFromHex(req.UserID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid user ID format", err)
		return
	}

	// Get the group service
	groupService := c.MustGet("groupService").(GroupService)

	// Check if the group exists
	group, err := groupService.GetGroupByID(c.Request.Context(), groupID)
	if err != nil {
		response.Error(c, http.StatusNotFound, "Group not found", err)
		return
	}

	// Check if the current user has permission to promote members
	// Only admins can promote others, and only the creator can promote to admin
	isCreator := group.CreatorID == currentUserID.(primitive.ObjectID)
	isAdmin := false
	for _, adminID := range group.Admins {
		if adminID == currentUserID.(primitive.ObjectID) {
			isAdmin = true
			break
		}
	}

	if !isAdmin {
		response.Error(c, http.StatusForbidden, "Only group admins can promote members", nil)
		return
	}

	// If promoting to admin, check if current user is the creator
	if req.Role == "admin" && !isCreator {
		response.Error(c, http.StatusForbidden, "Only the group creator can promote members to admin", nil)
		return
	}

	// Check if the target user is a member of the group
	isMember, err := groupService.IsGroupMember(c.Request.Context(), groupID, targetUserID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to check if user is a member", err)
		return
	}

	if !isMember {
		response.Error(c, http.StatusBadRequest, "User is not a member of this group", nil)
		return
	}

	// Promote the member
	err = groupService.PromoteMember(c.Request.Context(), groupID, targetUserID, req.Role)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to promote member", err)
		return
	}

	response.Success(c, http.StatusOK, "Member promoted successfully", nil)
}
