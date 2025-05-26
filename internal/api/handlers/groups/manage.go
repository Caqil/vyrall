package groups

import (
	"net/http"
	"strconv"

	"github.com/Caqil/vyrall/internal/utils/response"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// GetMembers retrieves the members of a group
func GetMembers(c *gin.Context) {
	// Get group ID from URL parameter
	groupIDStr := c.Param("id")
	groupID, err := primitive.ObjectIDFromHex(groupIDStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid group ID format", err)
		return
	}

	// Get query parameters
	role := c.DefaultQuery("role", "all") // all, admin, moderator, member
	limitStr := c.DefaultQuery("limit", "20")
	offsetStr := c.DefaultQuery("offset", "0")

	// Parse pagination params
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		limit = 20
	}
	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	// Get the group service
	groupService := c.MustGet("groupService").(GroupService)

	// Check if the group exists
	group, err := groupService.GetGroupByID(c.Request.Context(), groupID)
	if err != nil {
		response.Error(c, http.StatusNotFound, "Group not found", err)
		return
	}

	// If the group is not public, check if the user is a member
	if !group.IsPublic {
		// Get the authenticated user's ID
		userID, exists := c.Get("userID")
		if !exists {
			response.Error(c, http.StatusUnauthorized, "Authentication required for private groups", nil)
			return
		}

		// Check if the user is a member of the group
		isMember, err := groupService.IsGroupMember(c.Request.Context(), groupID, userID.(primitive.ObjectID))
		if err != nil {
			response.Error(c, http.StatusInternalServerError, "Failed to check group membership", err)
			return
		}

		if !isMember {
			response.Error(c, http.StatusForbidden, "You must be a member to view group members", nil)
			return
		}
	}

	// Validate role
	validRoles := map[string]bool{
		"all":       true,
		"admin":     true,
		"moderator": true,
		"member":    true,
	}
	if !validRoles[role] {
		response.Error(c, http.StatusBadRequest, "Invalid role filter", nil)
		return
	}

	// Get members
	members, total, err := groupService.GetGroupMembers(c.Request.Context(), groupID, role, limit, offset)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve group members", err)
		return
	}

	response.SuccessWithPagination(c, http.StatusOK, "Group members retrieved successfully", members, limit, offset, total)
}

// GetPendingRequests retrieves pending join requests for a group
func GetPendingRequests(c *gin.Context) {
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

	// Get pagination parameters
	limitStr := c.DefaultQuery("limit", "20")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		limit = 20
	}
	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	// Get the group service
	groupService := c.MustGet("groupService").(GroupService)

	// Check if the user has permission to view pending requests
	hasPermission, err := groupService.HasModeratorPermission(c.Request.Context(), groupID, userID.(primitive.ObjectID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to check user permissions", err)
		return
	}

	if !hasPermission {
		response.Error(c, http.StatusForbidden, "Only group admins and moderators can view pending join requests", nil)
		return
	}

	// Get pending join requests
	requests, total, err := groupService.GetPendingJoinRequests(c.Request.Context(), groupID, limit, offset)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve pending join requests", err)
		return
	}

	response.SuccessWithPagination(c, http.StatusOK, "Pending join requests retrieved successfully", requests, limit, offset, total)
}

// BanMember handles banning a user from a group
func BanMember(c *gin.Context) {
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

	// Parse request body
	var req struct {
		Reason   string `json:"reason"`
		Duration int    `json:"duration"` // Ban duration in days, 0 for permanent
	}
	c.ShouldBindJSON(&req)

	// Get the group service
	groupService := c.MustGet("groupService").(GroupService)

	// Check if the current user has permission to ban members
	hasPermission, err := groupService.HasModeratorPermission(c.Request.Context(), groupID, currentUserID.(primitive.ObjectID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to check user permissions", err)
		return
	}

	if !hasPermission {
		response.Error(c, http.StatusForbidden, "Only group admins and moderators can ban members", nil)
		return
	}

	// Cannot ban the group creator
	group, err := groupService.GetGroupByID(c.Request.Context(), groupID)
	if err != nil {
		response.Error(c, http.StatusNotFound, "Group not found", err)
		return
	}

	if group.CreatorID == targetUserID {
		response.Error(c, http.StatusForbidden, "Cannot ban the group creator", nil)
		return
	}

	// Ban the member
	err = groupService.BanMember(c.Request.Context(), groupID, targetUserID, currentUserID.(primitive.ObjectID), req.Reason, req.Duration)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to ban member", err)
		return
	}

	response.Success(c, http.StatusOK, "Member banned successfully", nil)
}

// UnbanMember handles unbanning a user from a group
func UnbanMember(c *gin.Context) {
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

	// Get the group service
	groupService := c.MustGet("groupService").(GroupService)

	// Check if the current user has permission to unban members
	hasPermission, err := groupService.HasModeratorPermission(c.Request.Context(), groupID, currentUserID.(primitive.ObjectID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to check user permissions", err)
		return
	}

	if !hasPermission {
		response.Error(c, http.StatusForbidden, "Only group admins and moderators can unban members", nil)
		return
	}

	// Unban the member
	err = groupService.UnbanMember(c.Request.Context(), groupID, targetUserID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to unban member", err)
		return
	}

	response.Success(c, http.StatusOK, "Member unbanned successfully", nil)
}
