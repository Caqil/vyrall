package users

import (
	"net/http"

	"github.com/Caqil/vyrall/internal/services/user"
	"github.com/Caqil/vyrall/internal/utils/response"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// SearchHandler handles user search operations
type SearchHandler struct {
	userService *user.Service
}

// NewSearchHandler creates a new search handler
func NewSearchHandler(userService *user.Service) *SearchHandler {
	return &SearchHandler{
		userService: userService,
	}
}

// SearchUsers handles the request to search for users
func (h *SearchHandler) SearchUsers(c *gin.Context) {
	// Get authenticated user ID (may be nil for unauthenticated users)
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
	verifiedOnly := c.DefaultQuery("verified", "false") == "true"
	nearbyOnly := c.DefaultQuery("nearby", "false") == "true"

	// Get pagination parameters
	limit, offset := response.GetPaginationParams(c)

	// Search users
	users, total, err := h.userService.SearchUsers(c.Request.Context(), query, verifiedOnly, nearbyOnly, userID, limit, offset)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to search users", err)
		return
	}

	// Return paginated response
	response.SuccessWithPagination(c, http.StatusOK, "Users found", users, limit, offset, total)
}

// SearchUsersByInterests handles the request to search users by interests
func (h *SearchHandler) SearchUsersByInterests(c *gin.Context) {
	// Get authenticated user ID (may be nil for unauthenticated users)
	var userID primitive.ObjectID
	if id, exists := c.Get("userID"); exists {
		userID = id.(primitive.ObjectID)
	}

	// Get query parameters
	interests := c.QueryArray("interest")
	if len(interests) == 0 {
		response.ValidationError(c, "At least one interest is required", nil)
		return
	}

	// Get pagination parameters
	limit, offset := response.GetPaginationParams(c)

	// Search users by interests
	users, total, err := h.userService.SearchUsersByInterests(c.Request.Context(), interests, userID, limit, offset)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to search users by interests", err)
		return
	}

	// Return paginated response
	response.SuccessWithPagination(c, http.StatusOK, "Users found", users, limit, offset, total)
}

// SearchUsersByLocation handles the request to search users by location
func (h *SearchHandler) SearchUsersByLocation(c *gin.Context) {
	// Get authenticated user ID (may be nil for unauthenticated users)
	var userID primitive.ObjectID
	if id, exists := c.Get("userID"); exists {
		userID = id.(primitive.ObjectID)
	}

	// Get query parameters
	location := c.Query("location")
	lat := c.Query("lat")
	lng := c.Query("lng")
	radius := c.DefaultQuery("radius", "50") // Default 50km radius

	if location == "" && (lat == "" || lng == "") {
		response.ValidationError(c, "Either location or coordinates (lat/lng) are required", nil)
		return
	}

	// Get pagination parameters
	limit, offset := response.GetPaginationParams(c)

	// Search users by location
	users, total, err := h.userService.SearchUsersByLocation(c.Request.Context(), location, lat, lng, radius, userID, limit, offset)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to search users by location", err)
		return
	}

	// Return paginated response
	response.SuccessWithPagination(c, http.StatusOK, "Users found", users, limit, offset, total)
}
