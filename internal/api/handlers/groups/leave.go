package groups

import (
	"net/http"

	"github.com/Caqil/vyrall/internal/utils/response"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// LeaveGroup handles a user leaving a group
func LeaveGroup(c *gin.Context) {
	// Get group ID from URL parameter
	groupIDStr := c.Param("id")
	groupID, err := primitive.ObjectIDFromHex(groupIDStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid group ID format", err)
		return
	}

	// Get the authenticated user's ID
	userID, exists := c.Get("userID")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "User not authenticated", nil)
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

	// Check if the user is the group creator
	if group.CreatorID == userID.(primitive.ObjectID) {
		response.Error(c, http.StatusForbidden, "Group creators cannot leave their own groups. Transfer ownership first or delete the group.", nil)
		return
	}

	// Check if the user is a member of the group
	isMember, err := groupService.IsGroupMember(c.Request.Context(), groupID, userID.(primitive.ObjectID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to check group membership", err)
		return
	}

	if !isMember {
		response.Error(c, http.StatusBadRequest, "You are not a member of this group", nil)
		return
	}

	// Leave the group
	err = groupService.RemoveMember(c.Request.Context(), groupID, userID.(primitive.ObjectID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to leave group", err)
		return
	}

	response.Success(c, http.StatusOK, "Left group successfully", nil)
}

// TransferOwnership handles transferring group ownership to another member
func TransferOwnership(c *gin.Context) {
	// Get group ID from URL parameter
	groupIDStr := c.Param("id")
	groupID, err := primitive.ObjectIDFromHex(groupIDStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid group ID format", err)
		return
	}

	// Parse request body
	var req struct {
		NewOwnerID string `json:"new_owner_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request format", err)
		return
	}

	newOwnerID, err := primitive.ObjectIDFromHex(req.NewOwnerID)
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

	// Get the group service
	groupService := c.MustGet("groupService").(GroupService)

	// Check if the group exists
	group, err := groupService.GetGroupByID(c.Request.Context(), groupID)
	if err != nil {
		response.Error(c, http.StatusNotFound, "Group not found", err)
		return
	}

	// Check if the current user is the group creator
	if group.CreatorID != currentUserID.(primitive.ObjectID) {
		response.Error(c, http.StatusForbidden, "Only the group creator can transfer ownership", nil)
		return
	}

	// Check if the new owner is a member of the group
	isMember, err := groupService.IsGroupMember(c.Request.Context(), groupID, newOwnerID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to check if user is a member", err)
		return
	}

	if !isMember {
		response.Error(c, http.StatusBadRequest, "The new owner must be a member of the group", nil)
		return
	}

	// Transfer ownership
	err = groupService.TransferOwnership(c.Request.Context(), groupID, currentUserID.(primitive.ObjectID), newOwnerID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to transfer ownership", err)
		return
	}

	response.Success(c, http.StatusOK, "Group ownership transferred successfully", nil)
}
