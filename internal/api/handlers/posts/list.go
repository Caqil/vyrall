package posts

import (
	"net/http"

	"github.com/Caqil/vyrall/internal/services/post"
	"github.com/Caqil/vyrall/internal/utils/response"
	"github.com/Caqil/vyrall/internal/utils/validation"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ListHandler handles post listing operations
type ListHandler struct {
	postService *post.Service
}

// NewListHandler creates a new list handler
func NewListHandler(postService *post.Service) *ListHandler {
	return &ListHandler{
		postService: postService,
	}
}

// ListPosts handles the request to list posts with various filters
func (h *ListHandler) ListPosts(c *gin.Context) {
	// Get user ID from context (may be nil for unauthenticated users)
	var userID primitive.ObjectID
	if id, exists := c.Get("userID"); exists {
		userID = id.(primitive.ObjectID)
	}

	// Get query parameters
	postType := c.DefaultQuery("type", "")           // text, media, poll, shared
	privacy := c.DefaultQuery("privacy", "")         // public, followers, private
	sortBy := c.DefaultQuery("sort_by", "recent")    // recent, popular
	timeRange := c.DefaultQuery("time_range", "all") // all, today, week, month, year
	tags := c.QueryArray("tag")                      // Array of tags to filter by
	nsfw := c.DefaultQuery("nsfw", "")               // true, false, all

	// Get pagination parameters
	limit, offset := response.GetPaginationParams(c)

	// Build filter options
	filters := map[string]interface{}{
		"type":       postType,
		"privacy":    privacy,
		"time_range": timeRange,
		"tags":       tags,
		"nsfw":       nsfw,
	}

	// List posts
	posts, total, err := h.postService.ListPosts(c.Request.Context(), filters, sortBy, userID, limit, offset)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to list posts", err)
		return
	}

	// Return paginated response
	response.SuccessWithPagination(c, http.StatusOK, "Posts retrieved successfully", posts, limit, offset, total)
}

// ListTrendingPosts handles the request to list trending posts
func (h *ListHandler) ListTrendingPosts(c *gin.Context) {
	// Get user ID from context (may be nil for unauthenticated users)
	var userID primitive.ObjectID
	if id, exists := c.Get("userID"); exists {
		userID = id.(primitive.ObjectID)
	}

	// Get query parameters
	category := c.DefaultQuery("category", "")       // Optional category filter
	timeRange := c.DefaultQuery("time_range", "day") // day, week, month

	// Get pagination parameters
	limit, offset := response.GetPaginationParams(c)

	// List trending posts
	posts, total, err := h.postService.ListTrendingPosts(c.Request.Context(), category, timeRange, userID, limit, offset)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to list trending posts", err)
		return
	}

	// Return paginated response
	response.SuccessWithPagination(c, http.StatusOK, "Trending posts retrieved successfully", posts, limit, offset, total)
}

// ListRecentPosts handles the request to list recent posts
func (h *ListHandler) ListRecentPosts(c *gin.Context) {
	// Get user ID from context (may be nil for unauthenticated users)
	var userID primitive.ObjectID
	if id, exists := c.Get("userID"); exists {
		userID = id.(primitive.ObjectID)
	}

	// Get query parameters
	category := c.DefaultQuery("category", "") // Optional category filter

	// Get pagination parameters
	limit, offset := response.GetPaginationParams(c)

	// List recent posts
	posts, total, err := h.postService.ListRecentPosts(c.Request.Context(), category, userID, limit, offset)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to list recent posts", err)
		return
	}

	// Return paginated response
	response.SuccessWithPagination(c, http.StatusOK, "Recent posts retrieved successfully", posts, limit, offset, total)
}

// ListPopularPosts handles the request to list popular posts
func (h *ListHandler) ListPopularPosts(c *gin.Context) {
	// Get user ID from context (may be nil for unauthenticated users)
	var userID primitive.ObjectID
	if id, exists := c.Get("userID"); exists {
		userID = id.(primitive.ObjectID)
	}

	// Get query parameters
	timeRange := c.DefaultQuery("time_range", "week") // day, week, month, year
	metric := c.DefaultQuery("metric", "engagement")  // engagement, likes, comments, shares, views

	// Get pagination parameters
	limit, offset := response.GetPaginationParams(c)

	// List popular posts
	posts, total, err := h.postService.ListPopularPosts(c.Request.Context(), timeRange, metric, userID, limit, offset)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to list popular posts", err)
		return
	}

	// Return paginated response
	response.SuccessWithPagination(c, http.StatusOK, "Popular posts retrieved successfully", posts, limit, offset, total)
}

// ListRelatedPosts handles the request to list posts related to a specific post
func (h *ListHandler) ListRelatedPosts(c *gin.Context) {
	// Get user ID from context (may be nil for unauthenticated users)
	var userID primitive.ObjectID
	if id, exists := c.Get("userID"); exists {
		userID = id.(primitive.ObjectID)
	}

	// Get post ID from URL parameter
	postIDStr := c.Param("id")
	if !validation.IsValidObjectID(postIDStr) {
		response.ValidationError(c, "Invalid post ID", nil)
		return
	}
	postID, _ := primitive.ObjectIDFromHex(postIDStr)

	// Get pagination parameters
	limit, offset := response.GetPaginationParams(c)

	// List related posts
	posts, total, err := h.postService.ListRelatedPosts(c.Request.Context(), postID, userID, limit, offset)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to list related posts", err)
		return
	}

	// Return paginated response
	response.SuccessWithPagination(c, http.StatusOK, "Related posts retrieved successfully", posts, limit, offset, total)
}
