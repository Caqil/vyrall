package search

import (
	"net/http"

	"github.com/Caqil/vyrall/internal/services/search"
	"github.com/Caqil/vyrall/internal/utils/response"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// UserSearchHandler handles user search operations
type UserSearchHandler struct {
	searchService *search.Service
}

// NewUserSearchHandler creates a new user search handler
func NewUserSearchHandler(searchService *search.Service) *UserSearchHandler {
	return &UserSearchHandler{
		searchService: searchService,
	}
}

// SearchUsers handles the request to search for users
func (h *UserSearchHandler) SearchUsers(c *gin.Context) {
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
	location := c.Query("location")
	interests := c.QueryArray("interest")
	verifiedOnly := c.DefaultQuery("verified", "false") == "true"
	sort := c.DefaultQuery("sort", "relevance") // relevance, followers, recent

	// Get pagination parameters
	limit, offset := response.GetPaginationParams(c)

	// Search users
	users, total, err := h.searchService.SearchUsers(
		c.Request.Context(),
		query,
		location,
		interests,
		verifiedOnly,
		sort,
		userID,
		limit,
		offset,
	)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to search users", err)
		return
	}

	// Return paginated response
	response.SuccessWithPagination(c, http.StatusOK, "User search results retrieved successfully", users, limit, offset, total)
}

// FindUsersByUsername handles the request to find users by username
func (h *UserSearchHandler) FindUsersByUsername(c *gin.Context) {
	// Get user ID from context (may be nil for unauthenticated users)
	var userID primitive.ObjectID
	if id, exists := c.Get("userID"); exists {
		userID = id.(primitive.ObjectID)
	}

	// Get username parameter
	username := c.Query("username")
	if username == "" {
		response.ValidationError(c, "Username is required", nil)
		return
	}

	// Get exact match parameter
	exactMatch := c.DefaultQuery("exact", "false") == "true"

	// Get pagination parameters
	limit, offset := response.GetPaginationParams(c)

	// Find users by username
	users, total, err := h.searchService.FindUsersByUsername(
		c.Request.Context(),
		username,
		exactMatch,
		userID,
		limit,
		offset,
	)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to find users by username", err)
		return
	}

	// Return paginated response
	response.SuccessWithPagination(c, http.StatusOK, "Users found successfully", users, limit, offset, total)
}

// SearchUsersByInterests handles the request to search users by interests
func (h *UserSearchHandler) SearchUsersByInterests(c *gin.Context) {
	// Get user ID from context (may be nil for unauthenticated users)
	var userID primitive.ObjectID
	if id, exists := c.Get("userID"); exists {
		userID = id.(primitive.ObjectID)
	}

	// Get interests parameter
	interests := c.QueryArray("interest")
	if len(interests) == 0 {
		response.ValidationError(c, "At least one interest is required", nil)
		return
	}

	// Get pagination parameters
	limit, offset := response.GetPaginationParams(c)

	// Search users by interests
	users, total, err := h.searchService.SearchUsersByInterests(
		c.Request.Context(),
		interests,
		userID,
		limit,
		offset,
	)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to search users by interests", err)
		return
	}

	// Return paginated response
	response.SuccessWithPagination(c, http.StatusOK, "Users found successfully", users, limit, offset, total)
}
