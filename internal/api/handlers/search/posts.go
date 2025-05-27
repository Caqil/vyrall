package search

import (
	"net/http"

	"github.com/Caqil/vyrall/internal/services/search"
	"github.com/Caqil/vyrall/internal/utils/response"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// PostSearchHandler handles post search operations
type PostSearchHandler struct {
	searchService *search.Service
}

// NewPostSearchHandler creates a new post search handler
func NewPostSearchHandler(searchService *search.Service) *PostSearchHandler {
	return &PostSearchHandler{
		searchService: searchService,
	}
}

// SearchPosts handles the request to search for posts
func (h *PostSearchHandler) SearchPosts(c *gin.Context) {
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
	contentType := c.DefaultQuery("type", "all")     // all, text, media, poll
	timeRange := c.DefaultQuery("time_range", "all") // all, day, week, month, year
	sort := c.DefaultQuery("sort", "relevance")      // relevance, recent, popular
	tags := c.QueryArray("tag")

	// Get pagination parameters
	limit, offset := response.GetPaginationParams(c)

	// Search posts
	posts, total, err := h.searchService.SearchPosts(
		c.Request.Context(),
		query,
		contentType,
		timeRange,
		sort,
		tags,
		userID,
		limit,
		offset,
	)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to search posts", err)
		return
	}

	// Return paginated response
	response.SuccessWithPagination(c, http.StatusOK, "Post search results retrieved successfully", posts, limit, offset, total)
}

// AdvancedPostSearch handles the request for advanced post search
func (h *PostSearchHandler) AdvancedPostSearch(c *gin.Context) {
	// Get user ID from context (may be nil for unauthenticated users)
	var userID primitive.ObjectID
	if id, exists := c.Get("userID"); exists {
		userID = id.(primitive.ObjectID)
	}

	// Parse request body
	var req struct {
		Query          string   `json:"query" binding:"required"`
		ContentType    string   `json:"content_type,omitempty"` // text, media, poll, all
		TimeRange      string   `json:"time_range,omitempty"`   // day, week, month, year, all
		Tags           []string `json:"tags,omitempty"`
		AuthorIDs      []string `json:"author_ids,omitempty"`
		LocationName   string   `json:"location_name,omitempty"`
		LocationRadius string   `json:"location_radius,omitempty"`
		Sort           string   `json:"sort,omitempty"` // relevance, recent, popular
		ExcludeWords   []string `json:"exclude_words,omitempty"`
		ExactPhrase    bool     `json:"exact_phrase,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "Invalid request body", err.Error())
		return
	}

	// Convert author IDs to ObjectIDs if provided
	authorIDs := make([]primitive.ObjectID, 0, len(req.AuthorIDs))
	for _, idStr := range req.AuthorIDs {
		if id, err := primitive.ObjectIDFromHex(idStr); err == nil {
			authorIDs = append(authorIDs, id)
		}
	}

	// Get pagination parameters
	limit, offset := response.GetPaginationParams(c)

	// Perform advanced post search
	posts, total, err := h.searchService.AdvancedPostSearch(
		c.Request.Context(),
		req.Query,
		req.ContentType,
		req.TimeRange,
		req.Tags,
		authorIDs,
		req.LocationName,
		req.LocationRadius,
		req.Sort,
		req.ExcludeWords,
		req.ExactPhrase,
		userID,
		limit,
		offset,
	)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to perform advanced post search", err)
		return
	}

	// Return paginated response
	response.SuccessWithPagination(c, http.StatusOK, "Advanced post search results retrieved successfully", posts, limit, offset, total)
}
