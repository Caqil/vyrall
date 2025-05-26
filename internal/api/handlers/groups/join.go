package groups

import (
	"net/http"

	"github.com/Caqil/vyrall/internal/utils/response"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// JoinRequest represents the request payload for joining a group
type JoinRequest struct {
	Message string `json:"message"`
}

// JoinGroup handles a user's request to join a group
func JoinGroup(c *gin.Context) {
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

	// Parse request body (optional message)
	var req JoinRequest
	c.ShouldBindJSON(&req)

	// Get the group service
	groupService := c.MustGet("groupService").(GroupService)

	// Check if the group exists
	group, err := groupService.GetGroupByID(c.Request.Context(), groupID)
	if err != nil {
		response.Error(c, http.StatusNotFound, "Group not found", err)
		return
	}

	// Check if the user is already a member
	isMember, err := groupService.IsGroupMember(c.Request.Context(), groupID, userID.(primitive.ObjectID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to check group membership", err)
		return
	}

	if isMember {
		response.Error(c, http.StatusBadRequest, "You are already a member of this group", nil)
		return
	}

	// Handle join request based on group settings
	if group.JoinApprovalRequired {
		// Submit join request
		err = groupService.SubmitJoinRequest(c.Request.Context(), groupID, userID.(primitive.ObjectID), req.Message)
		if err != nil {
			response.Error(c, http.StatusInternalServerError, "Failed to submit join request", err)
			return
		}

		response.Success(c, http.StatusOK, "Join request submitted successfully. Waiting for approval.", nil)
	} else {
		// Join directly
		err = groupService.AddMember(c.Request.Context(), groupID, userID.(primitive.ObjectID), "member")
		if err != nil {
			response.Error(c, http.StatusInternalServerError, "Failed to join group", err)
			return
		}

		response.Success(c, http.StatusOK, "Joined group successfully", nil)
	}
}

// JoinGroupByInviteLink handles a user joining a group via an invitation link
func JoinGroupByInviteLink(c *gin.Context) {
	// Get invite link from URL parameter
	inviteLink := c.Param("link")
	if inviteLink == "" {
		response.Error(c, http.StatusBadRequest, "Invalid invite link", nil)
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

	// Validate invite link and get group ID
	groupID, err := groupService.ValidateInviteLink(c.Request.Context(), inviteLink)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid or expired invite link", err)
		return
	}

	// Check if the user is already a member
	isMember, err := groupService.IsGroupMember(c.Request.Context(), groupID, userID.(primitive.ObjectID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to check group membership", err)
		return
	}

	if isMember {
		response.Error(c, http.StatusBadRequest, "You are already a member of this group", nil)
		return
	}

	// Join the group
	err = groupService.AddMember(c.Request.Context(), groupID, userID.(primitive.ObjectID), "member")
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to join group", err)
		return
	}

	// Get group details to return to user
	group, err := groupService.GetGroupByID(c.Request.Context(), groupID)
	if err != nil {
		// If this fails, we still joined the group successfully, so just log the error
		c.Error(err)
	}

	response.Success(c, http.StatusOK, "Joined group successfully", gin.H{
		"group_id":   groupID.Hex(),
		"group_name": group.Name,
	})
}

// ApproveJoinRequest handles approving a pending join request
func ApproveJoinRequest(c *gin.Context) {
	// Get group ID and user ID from URL parameters
	groupIDStr := c.Param("id")
	userIDStr := c.Param("user_id")

	groupID, err := primitive.ObjectIDFromHex(groupIDStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid group ID format", err)
		return
	}

	requestUserID, err := primitive.ObjectIDFromHex(userIDStr)
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

	// Check if the current user has permission to approve requests
	hasPermission, err := groupService.HasModeratorPermission(c.Request.Context(), groupID, currentUserID.(primitive.ObjectID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to check user permissions", err)
		return
	}

	if !hasPermission {
		response.Error(c, http.StatusForbidden, "Only group admins and moderators can approve join requests", nil)
		return
	}

	// Approve the join request
	err = groupService.ApproveJoinRequest(c.Request.Context(), groupID, requestUserID, currentUserID.(primitive.ObjectID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to approve join request", err)
		return
	}

	response.Success(c, http.StatusOK, "Join request approved successfully", nil)
}

// RejectJoinRequest handles rejecting a pending join request
func RejectJoinRequest(c *gin.Context) {
	// Get group ID and user ID from URL parameters
	groupIDStr := c.Param("id")
	userIDStr := c.Param("user_id")

	groupID, err := primitive.ObjectIDFromHex(groupIDStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid group ID format", err)
		return
	}

	requestUserID, err := primitive.ObjectIDFromHex(userIDStr)
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

	// Get rejection reason from request body
	var req struct {
		Reason string `json:"reason"`
	}
	c.ShouldBindJSON(&req)

	// Get the group service
	groupService := c.MustGet("groupService").(GroupService)

	// Check if the current user has permission to reject requests
	hasPermission, err := groupService.HasModeratorPermission(c.Request.Context(), groupID, currentUserID.(primitive.ObjectID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to check user permissions", err)
		return
	}

	if !hasPermission {
		response.Error(c, http.StatusForbidden, "Only group admins and moderators can reject join requests", nil)
		return
	}

	// Reject the join request
	err = groupService.RejectJoinRequest(c.Request.Context(), groupID, requestUserID, currentUserID.(primitive.ObjectID), req.Reason)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to reject join request", err)
		return
	}

	response.Success(c, http.StatusOK, "Join request rejected successfully", nil)
}
