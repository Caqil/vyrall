package search

import (
	"net/http"

	"github.com/Caqil/vyrall/internal/services/search"
	"github.com/Caqil/vyrall/internal/utils/response"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// TrendingHandler handles trending search operations
type TrendingHandler struct {
	searchService *search.Service
}

// NewTrendingHandler creates a new trending handler
func NewTrendingHandler(searchService *search.Service) *TrendingHandler {
	return &TrendingHandler{
		searchService: searchService,
	}
}

// GetTrendingSearches handles the request to get trending searches
func (h *TrendingHandler) GetTrendingSearches(c *gin.Context) {
	// Get user ID from context (may be nil for unauthenticated users)
	var userID primitive.ObjectID
	if id, exists := c.Get("userID"); exists {
		userID = id.(primitive.ObjectID)
	}

	// Get category parameter
	category := c.DefaultQuery("category", "all") // all, general, entertainment, news, etc.

	// Get limit for trending searches
	limit := 10
	if limitParam := c.Query("limit"); limitParam != "" {
		if parsedLimit, err := primitive.ParseInt(limitParam); err == nil && parsedLimit > 0 {
			limit = int(parsedLimit)
		}
	}

	// Get trending searches
	trending, err := h.searchService.GetTrendingSearches(
		c.Request.Context(),
		category,
		userID,
		limit,
	)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get trending searches", err)
		return
	}

	// Return success response
	response.OK(c, "Trending searches retrieved successfully", trending)
}

// GetTrendingTopics handles the request to get trending topics
func (h *TrendingHandler) GetTrendingTopics(c *gin.Context) {
	// Get user ID from context (may be nil for unauthenticated users)
	var userID primitive.ObjectID
	if id, exists := c.Get("userID"); exists {
		userID = id.(primitive.ObjectID)
	}

	// Get category and location parameters
	category := c.DefaultQuery("category", "all") // all, general, entertainment, news, etc.
	location := c.DefaultQuery("location", "")    // Global by default, or country/city

	// Get pagination parameters
	limit, offset := response.GetPaginationParams(c)

	// Get trending topics
	topics, total, err := h.searchService.GetTrendingTopics(
		c.Request.Context(),
		category,
		location,
		userID,
		limit,
		offset,
	)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get trending topics", err)
		return
	}

	// Return paginated response
	response.SuccessWithPagination(c, http.StatusOK, "Trending topics retrieved successfully", topics, limit, offset, total)
}
