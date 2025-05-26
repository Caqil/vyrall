package groups

import (
	"net/http"
	"time"

	"github.com/Caqil/vyrall/internal/models"
	"github.com/Caqil/vyrall/internal/utils/response"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// UpdateGroupRequest represents the request payload for updating a group
type UpdateGroupRequest struct {
	Name        *string          `json:"name"`
	Description *string          `json:"description"`
	Avatar      *string          `json:"avatar"`
	CoverPhoto  *string          `json:"cover_photo"`
	Location    *models.Location `json:"location"`
}

// UpdateGroup handles updating the basic information for a group
func UpdateGroup(c *gin.Context) {
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
	var req UpdateGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request format", err)
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

	// Check if the current user has permission to update the group
	isAdmin := false
	if group.CreatorID == userID.(primitive.ObjectID) {
		isAdmin = true
	} else {
		for _, adminID := range group.Admins {
			if adminID == userID.(primitive.ObjectID) {
				isAdmin = true
				break
			}
		}
	}

	if !isAdmin {
		response.Error(c, http.StatusForbidden, "Only group admins can update group information", nil)
		return
	}

	// Build updates map
	updates := make(map[string]interface{})

	if req.Name != nil {
		// Validate name length
		if len(*req.Name) < 3 || len(*req.Name) > 100 {
			response.Error(c, http.StatusBadRequest, "Group name must be between 3 and 100 characters", nil)
			return
		}
		updates["name"] = *req.Name
	}

	if req.Description != nil {
		// Validate description length
		if len(*req.Description) < 10 || len(*req.Description) > 5000 {
			response.Error(c, http.StatusBadRequest, "Group description must be between 10 and 5000 characters", nil)
			return
		}
		updates["description"] = *req.Description
	}

	if req.Avatar != nil {
		// You could add validation for the avatar URL here
		// For example, check if it's a valid URL or if it's from an allowed domain
		updates["avatar"] = *req.Avatar
	}

	if req.CoverPhoto != nil {
		// You could add validation for the cover photo URL here
		updates["cover_photo"] = *req.CoverPhoto
	}

	if req.Location != nil {
		// Validate location data
		if req.Location.Name == "" {
			response.Error(c, http.StatusBadRequest, "Location name is required", nil)
			return
		}
		updates["location"] = *req.Location
	}

	// Set updated timestamp
	updates["updated_at"] = time.Now()
	updates["last_activity_at"] = time.Now()

	// Update the group
	err = groupService.UpdateGroup(c.Request.Context(), groupID, updates)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to update group", err)
		return
	}

	// Get updated group to return in response
	updatedGroup, err := groupService.GetGroupByID(c.Request.Context(), groupID)
	if err != nil {
		// If this fails, we still updated the group successfully, so just log the error
		c.Error(err)
		response.Success(c, http.StatusOK, "Group updated successfully", gin.H{
			"group_id": groupID.Hex(),
		})
		return
	}

	response.Success(c, http.StatusOK, "Group updated successfully", updatedGroup)
}

// UpdateGroupStatus handles changing a group's status (active, archived, suspended)
func UpdateGroupStatus(c *gin.Context) {
	// Get group ID from URL parameter
	groupIDStr := c.Param("id")
	groupID, err := primitive.ObjectIDFromHex(groupIDStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid group ID format", err)
		return
	}

	// Parse request body
	var req struct {
		Status string `json:"status" binding:"required,oneof=active archived suspended"`
		Reason string `json:"reason"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request format", err)
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

	// Check if the current user is the creator or a platform admin
	isCreator := group.CreatorID == userID.(primitive.ObjectID)
	isPlatformAdmin, err := groupService.IsUserAdmin(c.Request.Context(), userID.(primitive.ObjectID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to check user permissions", err)
		return
	}

	if !isCreator && !isPlatformAdmin {
		response.Error(c, http.StatusForbidden, "Only the group creator or platform admin can change group status", nil)
		return
	}

	// Update the status
	updates := map[string]interface{}{
		"status":     req.Status,
		"updated_at": time.Now(),
	}

	err = groupService.UpdateGroup(c.Request.Context(), groupID, updates)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to update group status", err)
		return
	}

	// Notify members if the status changed significantly
	if req.Status == "suspended" || (group.Status != "suspended" && req.Status == "archived") {
		go groupService.NotifyGroupStatusChange(groupID, group.Name, req.Status, req.Reason)
	}

	response.Success(c, http.StatusOK, "Group status updated successfully", gin.H{
		"group_id": groupID.Hex(),
		"status":   req.Status,
	})
}
