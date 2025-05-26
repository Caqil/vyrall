package groups

import (
	"context"
	"net/http"
	"time"

	"github.com/Caqil/vyrall/internal/models"
	"github.com/Caqil/vyrall/internal/utils/response"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// GroupService defines the interface for group operations
type GroupService interface {
	CreateGroup(ctx context.Context, group *models.Group) (primitive.ObjectID, error)
	GetGroupByID(ctx context.Context, id primitive.ObjectID) (*models.Group, error)
	UpdateGroup(ctx context.Context, id primitive.ObjectID, updates map[string]interface{}) error
	DeleteGroup(ctx context.Context, id primitive.ObjectID) error
	HasModeratorPermission(ctx context.Context, groupID, userID primitive.ObjectID) (bool, error)
	IsGroupMember(ctx context.Context, groupID, userID primitive.ObjectID) (bool, error)
	GetGroupMembers(ctx context.Context, groupID primitive.ObjectID, role string, limit, offset int) ([]*models.GroupMember, int, error)
	AddMember(ctx context.Context, groupID, userID primitive.ObjectID, role string) error
	RemoveMember(ctx context.Context, groupID, userID primitive.ObjectID) error
	GetPendingJoinRequests(ctx context.Context, groupID primitive.ObjectID, limit, offset int) ([]models.GroupJoinRequest, int, error)
	UpdateGroupRules(ctx context.Context, groupID primitive.ObjectID, rules []models.GroupRule) error
}

// CreateGroupRequest represents the request payload for creating a group
type CreateGroupRequest struct {
	Name                 string               `json:"name" binding:"required"`
	Description          string               `json:"description" binding:"required"`
	Avatar               string               `json:"avatar"`
	CoverPhoto           string               `json:"cover_photo"`
	IsPublic             bool                 `json:"is_public"`
	IsVisible            bool                 `json:"is_visible"`
	JoinApprovalRequired bool                 `json:"join_approval_required"`
	Location             *models.Location     `json:"location"`
	Categories           []string             `json:"categories"`
	Tags                 []string             `json:"tags"`
	Rules                []models.GroupRule   `json:"rules"`
	Features             models.GroupFeatures `json:"features"`
}

// CreateGroup handles the creation of a new group
func CreateGroup(c *gin.Context) {
	// Get the authenticated user's ID
	userID, exists := c.Get("userID")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	// Parse and validate the request
	var req CreateGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request format", err)
		return
	}

	// Validate group name length
	if len(req.Name) < 3 || len(req.Name) > 100 {
		response.Error(c, http.StatusBadRequest, "Group name must be between 3 and 100 characters", nil)
		return
	}

	// Validate group description length
	if len(req.Description) < 10 || len(req.Description) > 5000 {
		response.Error(c, http.StatusBadRequest, "Group description must be between 10 and 5000 characters", nil)
		return
	}

	// Get the group service
	groupService := c.MustGet("groupService").(GroupService)

	// Create new group object
	group := &models.Group{
		Name:                 req.Name,
		Description:          req.Description,
		Avatar:               req.Avatar,
		CoverPhoto:           req.CoverPhoto,
		CreatorID:            userID.(primitive.ObjectID),
		Admins:               []primitive.ObjectID{userID.(primitive.ObjectID)},
		MemberCount:          1, // Creator is the first member
		PostCount:            0,
		EventCount:           0,
		IsPublic:             req.IsPublic,
		IsVisible:            req.IsVisible,
		JoinApprovalRequired: req.JoinApprovalRequired,
		Location:             req.Location,
		Categories:           req.Categories,
		Tags:                 req.Tags,
		Rules:                req.Rules,
		Features:             req.Features,
		Status:               "active",
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
		LastActivityAt:       time.Now(),
	}

	// Create the group
	groupID, err := groupService.CreateGroup(c.Request.Context(), group)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to create group", err)
		return
	}

	// Add creator as admin member
	err = groupService.AddMember(c.Request.Context(), groupID, userID.(primitive.ObjectID), "admin")
	if err != nil {
		// If adding member fails, log the error but don't fail the request
		// The group was created successfully, and we can try to add the member again later
		c.Error(err)
	}

	response.Success(c, http.StatusCreated, "Group created successfully", gin.H{
		"group_id": groupID.Hex(),
	})
}
