package posts

import (
	"net/http"

	"github.com/Caqil/vyrall/internal/services/post"
	"github.com/Caqil/vyrall/internal/utils/response"
	"github.com/Caqil/vyrall/internal/utils/validation"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// BookmarkHandler handles post bookmark operations
type BookmarkHandler struct {
	postService *post.Service
}

// NewBookmarkHandler creates a new bookmark handler
func NewBookmarkHandler(postService *post.Service) *BookmarkHandler {
	return &BookmarkHandler{
		postService: postService,
	}
}

// BookmarkPost handles the request to bookmark a post
func (h *BookmarkHandler) BookmarkPost(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Get post ID from URL parameter
	postIDStr := c.Param("id")
	if !validation.IsValidObjectID(postIDStr) {
		response.ValidationError(c, "Invalid post ID", nil)
		return
	}
	postID, _ := primitive.ObjectIDFromHex(postIDStr)

	// Get collection name from query parameter (optional)
	collection := c.DefaultQuery("collection", "")

	// Bookmark the post
	err := h.postService.BookmarkPost(c.Request.Context(), postID, userID.(primitive.ObjectID), collection)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to bookmark post", err)
		return
	}

	// Return success response
	response.OK(c, "Post bookmarked successfully", nil)
}

// GetBookmarkedPosts handles the request to get bookmarked posts
func (h *BookmarkHandler) GetBookmarkedPosts(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Get collection name from query parameter (optional)
	collection := c.DefaultQuery("collection", "")

	// Get sort options
	sortBy := c.DefaultQuery("sort_by", "recent") // recent, popular

	// Get pagination parameters
	limit, offset := response.GetPaginationParams(c)

	// Get bookmarked posts
	posts, total, err := h.postService.GetBookmarkedPosts(c.Request.Context(), userID.(primitive.ObjectID), collection, sortBy, limit, offset)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve bookmarked posts", err)
		return
	}

	// Return paginated response
	response.SuccessWithPagination(c, http.StatusOK, "Bookmarked posts retrieved successfully", posts, limit, offset, total)
}

// GetBookmarkCollections handles the request to get bookmark collections
func (h *BookmarkHandler) GetBookmarkCollections(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Get bookmark collections
	collections, err := h.postService.GetBookmarkCollections(c.Request.Context(), userID.(primitive.ObjectID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve bookmark collections", err)
		return
	}

	// Return success response
	response.OK(c, "Bookmark collections retrieved successfully", collections)
}

// CreateBookmarkCollection handles the request to create a bookmark collection
func (h *BookmarkHandler) CreateBookmarkCollection(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Parse request body
	var req struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description,omitempty"`
		IsPrivate   bool   `json:"is_private,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "Invalid request body", err.Error())
		return
	}

	// Create bookmark collection
	collection, err := h.postService.CreateBookmarkCollection(c.Request.Context(), userID.(primitive.ObjectID), req.Name, req.Description, req.IsPrivate)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to create bookmark collection", err)
		return
	}

	// Return success response
	response.Created(c, "Bookmark collection created successfully", collection)
}

// UpdateBookmarkCollection handles the request to update a bookmark collection
func (h *BookmarkHandler) UpdateBookmarkCollection(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Get collection ID from URL parameter
	collectionID := c.Param("id")

	// Parse request body
	var req struct {
		Name        string `json:"name,omitempty"`
		Description string `json:"description,omitempty"`
		IsPrivate   *bool  `json:"is_private,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "Invalid request body", err.Error())
		return
	}

	// Create updates map
	updates := make(map[string]interface{})

	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if req.IsPrivate != nil {
		updates["is_private"] = *req.IsPrivate
	}

	if len(updates) == 0 {
		response.ValidationError(c, "No updates provided", nil)
		return
	}

	// Update bookmark collection
	updatedCollection, err := h.postService.UpdateBookmarkCollection(c.Request.Context(), userID.(primitive.ObjectID), collectionID, updates)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to update bookmark collection", err)
		return
	}

	// Return success response
	response.OK(c, "Bookmark collection updated successfully", updatedCollection)
}

// DeleteBookmarkCollection handles the request to delete a bookmark collection
func (h *BookmarkHandler) DeleteBookmarkCollection(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Get collection ID from URL parameter
	collectionID := c.Param("id")

	// Delete bookmark collection
	err := h.postService.DeleteBookmarkCollection(c.Request.Context(), userID.(primitive.ObjectID), collectionID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to delete bookmark collection", err)
		return
	}

	// Return success response
	response.OK(c, "Bookmark collection deleted successfully", nil)
}
