package groups

import (
	"net/http"

	"github.com/Caqil/vyrall/internal/utils/response"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// KickRequest represents the request payload for kicking a member from a group
type KickRequest struct {
	Reason string `json:"reason"`
}

// KickMember handles removing a member from a group by an admin or moderator
func KickMember(c *gin.Context) {
	// Get group ID and user ID from URL parameters
	groupIDStr := c.Param("id")
	userIDStr := c.Param("user_id")

	groupID, err := primitive.ObjectIDFromHex(groupIDStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid group ID format", err)
		return
	}

	targetUserID, err := primitive.ObjectIDFromHex(userIDStr)
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

	// Parse request body (optional reason)
	var req KickRequest
	c.ShouldBindJSON(&req)

	// Get the group service
	groupService := c.MustGet("groupService").(GroupService)

	// Check if the group exists
	group, err := groupService.GetGroupByID(c.Request.Context(), groupID)
	if err != nil {
		response.Error(c, http.StatusNotFound, "Group not found", err)
		return
	}

	// Check if the current user has permission to kick members
	hasPermission, err := groupService.HasModeratorPermission(c.Request.Context(), groupID, currentUserID.(primitive.ObjectID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to check user permissions", err)
		return
	}

	if !hasPermission {
		response.Error(c, http.StatusForbidden, "Only group admins and moderators can remove members", nil)
		return
	}

	// Check if the target user is the group creator
	if group.CreatorID == targetUserID {
		response.Error(c, http.StatusForbidden, "Cannot remove the group creator", nil)
		return
	}

	// Check if the target user is an admin and current user is just a moderator
	isAdmin := false
	for _, adminID := range group.Admins {
		if adminID == targetUserID {
			isAdmin = true
			break
		}
	}

	isCurrentUserAdmin := false
	for _, adminID := range group.Admins {
		if adminID == currentUserID.(primitive.ObjectID) {
			isCurrentUserAdmin = true
			break
		}
	}

	if isAdmin && !isCurrentUserAdmin {
		response.Error(c, http.StatusForbidden, "Moderators cannot remove admins", nil)
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

	// Remove the member
	err = groupService.RemoveMember(c.Request.Context(), groupID, targetUserID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to remove member", err)
		return
	}

	// Send notification to the kicked user (asynchronously)
	go groupService.NotifyUserKicked(groupID, group.Name, targetUserID, req.Reason)

	response.Success(c, http.StatusOK, "Member removed successfully", nil)
}
