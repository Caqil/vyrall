package groups

import (
	"net/http"

	"github.com/Caqil/vyrall/internal/utils/response"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// DeleteGroup handles the deletion of a group
func DeleteGroup(c *gin.Context) {
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

	// Check if the user is the creator of the group
	if group.CreatorID != userID.(primitive.ObjectID) {
		// Check if user is platform admin
		isAdmin, err := groupService.IsUserAdmin(c.Request.Context(), userID.(primitive.ObjectID))
		if err != nil {
			response.Error(c, http.StatusInternalServerError, "Failed to check user permissions", err)
			return
		}

		if !isAdmin {
			response.Error(c, http.StatusForbidden, "Only the group creator or a platform admin can delete this group", nil)
			return
		}
	}

	// Get deletion reason from request body
	var req struct {
		Reason string `json:"reason"`
	}
	c.ShouldBindJSON(&req)

	// Delete the group
	err = groupService.DeleteGroup(c.Request.Context(), groupID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to delete group", err)
		return
	}

	// Send notifications to members (asynchronously)
	go groupService.NotifyGroupDeletion(groupID, group.Name, req.Reason)

	response.Success(c, http.StatusOK, "Group deleted successfully", nil)
}

// ArchiveGroup changes a group's status to archived without deleting it
func ArchiveGroup(c *gin.Context) {
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

	// Check if the user is an admin of the group
	isAdmin := false
	for _, adminID := range group.Admins {
		if adminID == userID.(primitive.ObjectID) {
			isAdmin = true
			break
		}
	}

	if !isAdmin {
		// Check if user is platform admin
		isPlatformAdmin, err := groupService.IsUserAdmin(c.Request.Context(), userID.(primitive.ObjectID))
		if err != nil {
			response.Error(c, http.StatusInternalServerError, "Failed to check user permissions", err)
			return
		}

		if !isPlatformAdmin {
			response.Error(c, http.StatusForbidden, "Only group admins or platform admins can archive this group", nil)
			return
		}
	}

	// Archive the group
	updates := map[string]interface{}{
		"status": "archived",
	}
	err = groupService.UpdateGroup(c.Request.Context(), groupID, updates)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to archive group", err)
		return
	}

	response.Success(c, http.StatusOK, "Group archived successfully", nil)
}
