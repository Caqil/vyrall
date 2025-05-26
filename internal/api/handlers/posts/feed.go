package posts

import (
	"net/http"

	"github.com/Caqil/vyrall/internal/services/post"
	"github.com/Caqil/vyrall/internal/utils/response"
	"github.com/Caqil/vyrall/internal/utils/validation"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// FeedHandler handles post feed operations
type FeedHandler struct {
	postService *post.Service
}

// NewFeedHandler creates a new feed handler
func NewFeedHandler(postService *post.Service) *FeedHandler {
	return &FeedHandler{
		postService: postService,
	}
}

// GetFeed handles the request to get the user's personalized feed
func (h *FeedHandler) GetFeed(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Get query parameters
	feedType := c.DefaultQuery("type", "mixed") // following, for_you, mixed

	// Get pagination parameters
	limit, offset := response.GetPaginationParams(c)

	// Get feed
	posts, total, err := h.postService.GetFeed(c.Request.Context(), userID.(primitive.ObjectID), feedType, limit, offset)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve feed", err)
		return
	}

	// Return paginated response
	response.SuccessWithPagination(c, http.StatusOK, "Feed retrieved successfully", posts, limit, offset, total)
}

// GetDiscoverFeed handles the request to get the discovery feed
func (h *FeedHandler) GetDiscoverFeed(c *gin.Context) {
	// Get user ID from context (may be nil for unauthenticated users)
	var userID primitive.ObjectID
	if id, exists := c.Get("userID"); exists {
		userID = id.(primitive.ObjectID)
	}

	// Get query parameters
	category := c.DefaultQuery("category", "") // Optional category filter

	// Get pagination parameters
	limit, offset := response.GetPaginationParams(c)

	// Get discover feed
	posts, total, err := h.postService.GetDiscoverFeed(c.Request.Context(), userID, category, limit, offset)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve discover feed", err)
		return
	}

	// Return paginated response
	response.SuccessWithPagination(c, http.StatusOK, "Discover feed retrieved successfully", posts, limit, offset, total)
}

// GetTagFeed handles the request to get a feed for a specific tag
func (h *FeedHandler) GetTagFeed(c *gin.Context) {
	// Get user ID from context (may be nil for unauthenticated users)
	var userID primitive.ObjectID
	if id, exists := c.Get("userID"); exists {
		userID = id.(primitive.ObjectID)
	}

	// Get tag from URL parameter
	tag := c.Param("tag")
	if tag == "" {
		response.ValidationError(c, "Tag is required", nil)
		return
	}

	// Get sort options
	sortBy := c.DefaultQuery("sort_by", "recent") // recent, top

	// Get pagination parameters
	limit, offset := response.GetPaginationParams(c)

	// Get tag feed
	posts, total, err := h.postService.GetTagFeed(c.Request.Context(), tag, userID, sortBy, limit, offset)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve tag feed", err)
		return
	}

	// Return paginated response
	response.SuccessWithPagination(c, http.StatusOK, "Tag feed retrieved successfully", posts, limit, offset, total)
}

// RefreshFeed handles the request to refresh the user's feed
func (h *FeedHandler) RefreshFeed(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Get query parameters
	feedType := c.DefaultQuery("type", "mixed") // following, for_you, mixed

	// Get pagination parameters
	limit, _ := response.GetPaginationParams(c)

	// Refresh feed
	posts, total, err := h.postService.RefreshFeed(c.Request.Context(), userID.(primitive.ObjectID), feedType, limit)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to refresh feed", err)
		return
	}

	// Return success response
	response.OK(c, "Feed refreshed successfully", gin.H{
		"posts": posts,
		"total": total,
	})
}

// GetCustomFeed handles the request to get a custom feed based on user preferences
func (h *FeedHandler) GetCustomFeed(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Parse request body
	var req struct {
		Tags       []string `json:"tags,omitempty"`
		Categories []string `json:"categories,omitempty"`
		UserIDs    []string `json:"user_ids,omitempty"`
		SortBy     string   `json:"sort_by,omitempty"` // recent, popular
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "Invalid request body", err.Error())
		return
	}

	// Convert user IDs to ObjectIDs if provided
	userIDs := make([]primitive.ObjectID, 0, len(req.UserIDs))
	for _, idStr := range req.UserIDs {
		if !validation.IsValidObjectID(idStr) {
			continue // Skip invalid IDs
		}
		id, _ := primitive.ObjectIDFromHex(idStr)
		userIDs = append(userIDs, id)
	}

	// Set default sort option if not provided
	sortBy := req.SortBy
	if sortBy == "" {
		sortBy = "recent"
	}

	// Get pagination parameters
	limit, offset := response.GetPaginationParams(c)

	// Get custom feed
	posts, total, err := h.postService.GetCustomFeed(
		c.Request.Context(),
		userID.(primitive.ObjectID),
		req.Tags,
		req.Categories,
		userIDs,
		sortBy,
		limit,
		offset,
	)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve custom feed", err)
		return
	}

	// Return paginated response
	response.SuccessWithPagination(c, http.StatusOK, "Custom feed retrieved successfully", posts, limit, offset, total)
}
