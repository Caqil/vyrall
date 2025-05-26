package groups

import (
	"net/http"

	"github.com/Caqil/vyrall/internal/utils/response"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// InviteRequest represents the request payload for inviting users to a group
type InviteRequest struct {
	UserIDs []string `json:"user_ids" binding:"required"`
	Message string   `json:"message"`
}

// InviteUsers handles inviting users to join a group
func InviteUsers(c *gin.Context) {
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

	// Parse and validate the request
	var req InviteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request format", err)
		return
	}

	// Check that at least one user ID is provided
	if len(req.UserIDs) == 0 {
		response.Error(c, http.StatusBadRequest, "At least one user ID must be provided", nil)
		return
	}

	// Convert string IDs to ObjectIDs
	var targetUserIDs []primitive.ObjectID
	for _, idStr := range req.UserIDs {
		id, err := primitive.ObjectIDFromHex(idStr)
		if err != nil {
			response.Error(c, http.StatusBadRequest, "Invalid user ID format: "+idStr, err)
			return
		}
		targetUserIDs = append(targetUserIDs, id)
	}

	// Get the group service
	groupService := c.MustGet("groupService").(GroupService)

	// Check if the group exists
	group, err := groupService.GetGroupByID(c.Request.Context(), groupID)
	if err != nil {
		response.Error(c, http.StatusNotFound, "Group not found", err)
		return
	}

	// Check if the current user is a member of the group
	isMember, err := groupService.IsGroupMember(c.Request.Context(), groupID, userID.(primitive.ObjectID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to check group membership", err)
		return
	}

	if !isMember {
		response.Error(c, http.StatusForbidden, "Only group members can invite others", nil)
		return
	}

	// Send invitations
	results, err := groupService.InviteUsers(c.Request.Context(), groupID, userID.(primitive.ObjectID), targetUserIDs, req.Message)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to send invitations", err)
		return
	}

	response.Success(c, http.StatusOK, "Invitations sent successfully", results)
}

// GenerateInviteLink creates or retrieves an invitation link for the group
func GenerateInviteLink(c *gin.Context) {
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

	// Check if the current user has permission to generate invite links
	hasPermission, err := groupService.HasModeratorPermission(c.Request.Context(), groupID, userID.(primitive.ObjectID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to check user permissions", err)
		return
	}

	if !hasPermission {
		response.Error(c, http.StatusForbidden, "Only group admins and moderators can generate invite links", nil)
		return
	}

	// Generate invite link
	link, err := groupService.GenerateInviteLink(c.Request.Context(), groupID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to generate invite link", err)
		return
	}

	response.Success(c, http.StatusOK, "Invite link generated successfully", gin.H{
		"invite_link": link,
	})
}
