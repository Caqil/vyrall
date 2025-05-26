package groups

import (
	"net/http"
	"time"

	"github.com/Caqil/vyrall/internal/models"
	"github.com/Caqil/vyrall/internal/utils/response"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// UpdateRulesRequest represents the request payload for updating group rules
type UpdateRulesRequest struct {
	Rules []models.GroupRule `json:"rules" binding:"required"`
}

// UpdateRules handles updating the rules for a group
func UpdateRules(c *gin.Context) {
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
	var req UpdateRulesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request format", err)
		return
	}

	// Validate rules
	if len(req.Rules) > 20 {
		response.Error(c, http.StatusBadRequest, "Maximum of 20 rules allowed", nil)
		return
	}

	for i, rule := range req.Rules {
		if rule.Title == "" {
			response.Error(c, http.StatusBadRequest, "Rule title is required", nil)
			return
		}
		if len(rule.Title) > 100 {
			response.Error(c, http.StatusBadRequest, "Rule title must be less than 100 characters", nil)
			return
		}
		if len(rule.Description) > 1000 {
			response.Error(c, http.StatusBadRequest, "Rule description must be less than 1000 characters", nil)
			return
		}

		// Set created time for new rules if not provided
		if rule.CreatedAt.IsZero() {
			req.Rules[i].CreatedAt = time.Now()
		}
		// Always update the updated time
		req.Rules[i].UpdatedAt = time.Now()
	}

	// Get the group service
	groupService := c.MustGet("groupService").(GroupService)

	// Check if the current user has permission to update rules
	hasPermission, err := groupService.HasModeratorPermission(c.Request.Context(), groupID, userID.(primitive.ObjectID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to check user permissions", err)
		return
	}

	if !hasPermission {
		response.Error(c, http.StatusForbidden, "Only group admins and moderators can update rules", nil)
		return
	}

	// Update the rules
	err = groupService.UpdateGroupRules(c.Request.Context(), groupID, req.Rules)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to update group rules", err)
		return
	}

	response.Success(c, http.StatusOK, "Group rules updated successfully", req.Rules)
}

// GetRules retrieves the rules for a group
func GetRules(c *gin.Context) {
	// Get group ID from URL parameter
	groupIDStr := c.Param("id")
	groupID, err := primitive.ObjectIDFromHex(groupIDStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid group ID format", err)
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
			response.Error(c, http.StatusForbidden, "You must be a member to view group rules", nil)
			return
		}
	}

	response.Success(c, http.StatusOK, "Group rules retrieved successfully", group.Rules)
}
