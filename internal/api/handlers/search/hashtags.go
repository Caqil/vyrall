package search

import (
	"net/http"

	"github.com/Caqil/vyrall/internal/services/search"
	"github.com/Caqil/vyrall/internal/utils/response"
	"github.com/gin-gonic/gin"
)

// HashtagSearchHandler handles hashtag search operations
type HashtagSearchHandler struct {
	searchService *search.Service
}

// NewHashtagSearchHandler creates a new hashtag search handler
func NewHashtagSearchHandler(searchService *search.Service) *HashtagSearchHandler {
	return &HashtagSearchHandler{
		searchService: searchService,
	}
}

// SearchHashtags handles the request to search for hashtags
func (h *HashtagSearchHandler) SearchHashtags(c *gin.Context) {
	// Get query parameters
	query := c.Query("q")
	if query == "" {
		response.ValidationError(c, "Search query is required", nil)
		return
	}

	// Get additional filters
	trendingOnly := c.DefaultQuery("trending_only", "false") == "true"
	sort := c.DefaultQuery("sort", "relevance") // relevance, popular, recent

	// Get pagination parameters
	limit, offset := response.GetPaginationParams(c)

	// Search hashtags
	hashtags, total, err := h.searchService.SearchHashtags(
		c.Request.Context(),
		query,
		trendingOnly,
		sort,
		limit,
		offset,
	)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to search hashtags", err)
		return
	}

	// Return paginated response
	response.SuccessWithPagination(c, http.StatusOK, "Hashtag search results retrieved successfully", hashtags, limit, offset, total)
}

// GetRelatedHashtags handles the request to get related hashtags
func (h *HashtagSearchHandler) GetRelatedHashtags(c *gin.Context) {
	// Get hashtag parameter
	hashtag := c.Param("tag")
	if hashtag == "" {
		response.ValidationError(c, "Hashtag is required", nil)
		return
	}

	// Get limit parameter
	limit := 10
	if limitParam := c.Query("limit"); limitParam != "" {
		if parsedLimit, err := primitive.ParseInt(limitParam); err == nil && parsedLimit > 0 {
			limit = int(parsedLimit)
		}
	}

	// Get related hashtags
	relatedTags, err := h.searchService.GetRelatedHashtags(
		c.Request.Context(),
		hashtag,
		limit,
	)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get related hashtags", err)
		return
	}

	// Return success response
	response.OK(c, "Related hashtags retrieved successfully", relatedTags)
}
