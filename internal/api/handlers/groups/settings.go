package groups

import (
	"net/http"

	"github.com/Caqil/vyrall/internal/models"
	"github.com/Caqil/vyrall/internal/utils/response"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// UpdateSettingsRequest represents the request payload for updating group settings
type UpdateSettingsRequest struct {
	IsPublic             *bool                 `json:"is_public"`
	IsVisible            *bool                 `json:"is_visible"`
	JoinApprovalRequired *bool                 `json:"join_approval_required"`
	Features             *models.GroupFeatures `json:"features"`
	Categories           []string              `json:"categories"`
	Tags                 []string              `json:"tags"`
}

// UpdateSettings handles updating the settings for a group
func UpdateSettings(c *gin.Context) {
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
	var req UpdateSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request format", err)
		return
	}

	// Get the group service
	groupService := c.MustGet("groupService").(GroupService)

	// Check if the current user has permission to update settings
	hasPermission, err := groupService.HasAdminPermission(c.Request.Context(), groupID, userID.(primitive.ObjectID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to check user permissions", err)
		return
	}

	if !hasPermission {
		response.Error(c, http.StatusForbidden, "Only group admins can update settings", nil)
		return
	}

	// Build updates map
	updates := make(map[string]interface{})

	if req.IsPublic != nil {
		updates["is_public"] = *req.IsPublic
	}

	if req.IsVisible != nil {
		updates["is_visible"] = *req.IsVisible
	}

	if req.JoinApprovalRequired != nil {
		updates["join_approval_required"] = *req.JoinApprovalRequired
	}

	if req.Features != nil {
		updates["features"] = *req.Features
	}

	if req.Categories != nil {
		updates["categories"] = req.Categories
	}

	if req.Tags != nil {
		updates["tags"] = req.Tags
	}

	// Update the settings
	err = groupService.UpdateGroup(c.Request.Context(), groupID, updates)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to update group settings", err)
		return
	}

	// Get updated group
	group, err := groupService.GetGroupByID(c.Request.Context(), groupID)
	if err != nil {
		// If this fails, we still updated the settings successfully, so just log the error
		c.Error(err)
		response.Success(c, http.StatusOK, "Group settings updated successfully", nil)
		return
	}

	response.Success(c, http.StatusOK, "Group settings updated successfully", group)
}

// GetSettings retrieves the current settings for a group
func GetSettings(c *gin.Context) {
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

	// Check if the current user has permission to view settings
	hasPermission, err := groupService.HasModeratorPermission(c.Request.Context(), groupID, userID.(primitive.ObjectID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to check user permissions", err)
		return
	}

	if !hasPermission {
		response.Error(c, http.StatusForbidden, "Only group admins and moderators can view settings", nil)
		return
	}

	// Get the group
	group, err := groupService.GetGroupByID(c.Request.Context(), groupID)
	if err != nil {
		response.Error(c, http.StatusNotFound, "Group not found", err)
		return
	}

	// Extract settings
	settings := gin.H{
		"is_public":              group.IsPublic,
		"is_visible":             group.IsVisible,
		"join_approval_required": group.JoinApprovalRequired,
		"features":               group.Features,
		"categories":             group.Categories,
		"tags":                   group.Tags,
	}

	response.Success(c, http.StatusOK, "Group settings retrieved successfully", settings)
}
