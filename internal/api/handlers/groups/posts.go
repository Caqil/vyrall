package groups

import (
	"net/http"
	"strconv"

	"github.com/Caqil/vyrall/internal/utils/response"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// PostService defines the interface for group post operations
type PostService interface {
	GetPostsByGroupID(ctx context.Context, groupID primitive.ObjectID, filter map[string]interface{}, sort map[string]int, limit, offset int) ([]*models.Post, int, error)
	GetPinnedPostsByGroupID(ctx context.Context, groupID primitive.ObjectID) ([]*models.Post, error)
	CanCreateGroupPost(ctx context.Context, groupID, userID primitive.ObjectID) (bool, error)
}

// GetGroupPosts retrieves posts for a specific group
func GetGroupPosts(c *gin.Context) {
	// Get group ID from URL parameter
	groupIDStr := c.Param("id")
	groupID, err := primitive.ObjectIDFromHex(groupIDStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid group ID format", err)
		return
	}

	// Get query parameters
	sortBy := c.DefaultQuery("sort_by", "created_at")
	sortOrder := c.DefaultQuery("sort_order", "desc")
	limitStr := c.DefaultQuery("limit", "20")
	offsetStr := c.DefaultQuery("offset", "0")
	filterType := c.DefaultQuery("filter", "all") // all, pinned

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
			response.Error(c, http.StatusForbidden, "You must be a member to view group posts", nil)
			return
		}
	}

	// Get the post service
	postService := c.MustGet("postService").(PostService)

	// Special case for pinned posts
	if filterType == "pinned" {
		pinnedPosts, err := postService.GetPinnedPostsByGroupID(c.Request.Context(), groupID)
		if err != nil {
			response.Error(c, http.StatusInternalServerError, "Failed to retrieve pinned posts", err)
			return
		}

		response.Success(c, http.StatusOK, "Pinned posts retrieved successfully", pinnedPosts)
		return
	}

	// Build filter
	filter := map[string]interface{}{
		"group_id":   groupID,
		"is_hidden":  false,
		"deleted_at": nil,
	}

	// Build sort
	sort := map[string]int{}
	if sortOrder == "desc" {
		sort[sortBy] = -1
	} else {
		sort[sortBy] = 1
	}

	// Get posts
	posts, total, err := postService.GetPostsByGroupID(c.Request.Context(), groupID, filter, sort, limit, offset)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve group posts", err)
		return
	}

	response.SuccessWithPagination(c, http.StatusOK, "Group posts retrieved successfully", posts, limit, offset, total)
}

// CanCreatePost checks if the current user can create posts in this group
func CanCreatePost(c *gin.Context) {
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
	_, err = groupService.GetGroupByID(c.Request.Context(), groupID)
	if err != nil {
		response.Error(c, http.StatusNotFound, "Group not found", err)
		return
	}

	// Get the post service
	postService := c.MustGet("postService").(PostService)

	// Check if the user can create posts
	canCreate, err := postService.CanCreateGroupPost(c.Request.Context(), groupID, userID.(primitive.ObjectID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to check permissions", err)
		return
	}

	response.Success(c, http.StatusOK, "Permission check completed", gin.H{
		"can_create_post": canCreate,
	})
}
