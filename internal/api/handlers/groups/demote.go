package groups

import (
	"net/http"

	"github.com/Caqil/vyrall/internal/utils/response"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// DemoteRequest represents the request payload for demoting a group member
type DemoteRequest struct {
	UserID string `json:"user_id" binding:"required"`
	Role   string `json:"role" binding:"required,oneof=member moderator"`
}

// DemoteMember handles demoting a group admin or moderator to a lower role
func DemoteMember(c *gin.Context) {
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
	var req DemoteRequest
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

	// Check if the current user is the group creator or an admin
	isCreator := group.CreatorID == currentUserID.(primitive.ObjectID)
	isAdmin := false
	for _, adminID := range group.Admins {
		if adminID == currentUserID.(primitive.ObjectID) {
			isAdmin = true
			break
		}
	}

	if !isCreator && !isAdmin {
		response.Error(c, http.StatusForbidden, "Only the group creator or admins can demote members", nil)
		return
	}

	// Check if the target user is the group creator
	if group.CreatorID == targetUserID {
		response.Error(c, http.StatusForbidden, "Cannot demote the group creator", nil)
		return
	}

	// Check if current user is trying to demote another admin and is not the creator
	isTargetAdmin := false
	for _, adminID := range group.Admins {
		if adminID == targetUserID {
			isTargetAdmin = true
			break
		}
	}

	if isTargetAdmin && !isCreator {
		response.Error(c, http.StatusForbidden, "Only the group creator can demote admins", nil)
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

	// Demote the member
	err = groupService.DemoteMember(c.Request.Context(), groupID, targetUserID, req.Role)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to demote member", err)
		return
	}

	response.Success(c, http.StatusOK, "Member demoted successfully", nil)
}
