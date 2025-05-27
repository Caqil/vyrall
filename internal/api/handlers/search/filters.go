package search

import (
	"net/http"

	"github.com/Caqil/vyrall/internal/services/search"
	"github.com/Caqil/vyrall/internal/utils/response"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// GroupSearchHandler handles group search operations
type GroupSearchHandler struct {
	searchService *search.Service
}

// NewGroupSearchHandler creates a new group search handler
func NewGroupSearchHandler(searchService *search.Service) *GroupSearchHandler {
	return &GroupSearchHandler{
		searchService: searchService,
	}
}

// SearchGroups handles the request to search for groups
func (h *GroupSearchHandler) SearchGroups(c *gin.Context) {
	// Get user ID from context (may be nil for unauthenticated users)
	var userID primitive.ObjectID
	if id, exists := c.Get("userID"); exists {
		userID = id.(primitive.ObjectID)
	}

	// Get query parameters
	query := c.Query("q")
	if query == "" {
		response.ValidationError(c, "Search query is required", nil)
		return
	}

	// Get additional filters
	categories := c.QueryArray("category")
	memberCount := c.DefaultQuery("member_count", "") // min:max format
	publicOnly := c.DefaultQuery("public_only", "false") == "true"
	sort := c.DefaultQuery("sort", "relevance") // relevance, recent, popular, member_count

	// Get pagination parameters
	limit, offset := response.GetPaginationParams(c)

	// Search groups
	groups, total, err := h.searchService.SearchGroups(
		c.Request.Context(),
		query,
		categories,
		memberCount,
		publicOnly,
		sort,
		userID,
		limit,
		offset,
	)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to search groups", err)
		return
	}

	// Return paginated response
	response.SuccessWithPagination(c, http.StatusOK, "Group search results retrieved successfully", groups, limit, offset, total)
}

// GetRecommendedGroups handles the request to get recommended groups
func (h *GroupSearchHandler) GetRecommendedGroups(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Get category filter
	category := c.DefaultQuery("category", "")

	// Get pagination parameters
	limit, offset := response.GetPaginationParams(c)

	// Get recommended groups
	groups, total, err := h.searchService.GetRecommendedGroups(
		c.Request.Context(),
		userID.(primitive.ObjectID),
		category,
		limit,
		offset,
	)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get recommended groups", err)
		return
	}

	// Return paginated response
	response.SuccessWithPagination(c, http.StatusOK, "Recommended groups retrieved successfully", groups, limit, offset, total)
}
