package posts

import (
	"net/http"

	"github.com/Caqil/vyrall/internal/services/post"
	"github.com/Caqil/vyrall/internal/utils/response"
	"github.com/Caqil/vyrall/internal/utils/validation"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// GetHandler handles post retrieval operations
type GetHandler struct {
	postService *post.Service
}

// NewGetHandler creates a new get handler
func NewGetHandler(postService *post.Service) *GetHandler {
	return &GetHandler{
		postService: postService,
	}
}

// GetPost handles the request to get a specific post
func (h *GetHandler) GetPost(c *gin.Context) {
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

	// Get the post
	post, err := h.postService.GetPost(c.Request.Context(), postID, userID)
	if err != nil {
		response.NotFoundError(c, "Post not found")
		return
	}

	// Return success response
	response.OK(c, "Post retrieved successfully", post)
}

// GetPostWithComments handles the request to get a post with its comments
func (h *GetHandler) GetPostWithComments(c *gin.Context) {
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

	// Get sort order for comments
	sortBy := c.DefaultQuery("sort_by", "recent") // recent, popular

	// Get pagination parameters for comments
	limit, offset := response.GetPaginationParams(c)

	// Get the post with comments
	post, comments, totalComments, err := h.postService.GetPostWithComments(c.Request.Context(), postID, userID, sortBy, limit, offset)
	if err != nil {
		response.NotFoundError(c, "Post not found")
		return
	}

	// Return success response
	response.OK(c, "Post retrieved successfully", gin.H{
		"post":     post,
		"comments": comments,
		"pagination": gin.H{
			"total":  totalComments,
			"limit":  limit,
			"offset": offset,
		},
	})
}

// GetUserPosts handles the request to get posts from a specific user
func (h *GetHandler) GetUserPosts(c *gin.Context) {
	// Get authenticated user ID (may be nil for unauthenticated users)
	var authUserID primitive.ObjectID
	if id, exists := c.Get("userID"); exists {
		authUserID = id.(primitive.ObjectID)
	}

	// Get target user ID from URL parameter
	userIDStr := c.Param("userId")
	if !validation.IsValidObjectID(userIDStr) {
		response.ValidationError(c, "Invalid user ID", nil)
		return
	}
	userID, _ := primitive.ObjectIDFromHex(userIDStr)

	// Get filters and sort options
	postType := c.DefaultQuery("type", "")        // all, text, media, poll, shared
	sortBy := c.DefaultQuery("sort_by", "recent") // recent, popular

	// Get pagination parameters
	limit, offset := response.GetPaginationParams(c)

	// Get user posts
	posts, total, err := h.postService.GetUserPosts(c.Request.Context(), userID, authUserID, postType, sortBy, limit, offset)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve user posts", err)
		return
	}

	// Return paginated response
	response.SuccessWithPagination(c, http.StatusOK, "User posts retrieved successfully", posts, limit, offset, total)
}

// GetPostsByTag handles the request to get posts with a specific tag
func (h *GetHandler) GetPostsByTag(c *gin.Context) {
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
	sortBy := c.DefaultQuery("sort_by", "recent") // recent, popular

	// Get pagination parameters
	limit, offset := response.GetPaginationParams(c)

	// Get posts by tag
	posts, total, err := h.postService.GetPostsByTag(c.Request.Context(), tag, userID, sortBy, limit, offset)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve posts", err)
		return
	}

	// Return paginated response
	response.SuccessWithPagination(c, http.StatusOK, "Posts retrieved successfully", posts, limit, offset, total)
}

// GetPostsByLocation handles the request to get posts from a specific location
func (h *GetHandler) GetPostsByLocation(c *gin.Context) {
	// Get user ID from context (may be nil for unauthenticated users)
	var userID primitive.ObjectID
	if id, exists := c.Get("userID"); exists {
		userID = id.(primitive.ObjectID)
	}

	// Parse request query parameters
	lat, _ := primitive.ParseFloat(c.Query("lat"))
	lng, _ := primitive.ParseFloat(c.Query("lng"))
	radius, _ := primitive.ParseFloat(c.DefaultQuery("radius", "5")) // Default 5km radius
	locationName := c.Query("name")

	// Validate coordinates
	if lat == 0 && lng == 0 && locationName == "" {
		response.ValidationError(c, "Either coordinates (lat/lng) or location name is required", nil)
		return
	}

	// Get sort options
	sortBy := c.DefaultQuery("sort_by", "recent") // recent, popular

	// Get pagination parameters
	limit, offset := response.GetPaginationParams(c)

	// Get posts by location
	posts, total, err := h.postService.GetPostsByLocation(
		c.Request.Context(),
		lat,
		lng,
		radius,
		locationName,
		userID,
		sortBy,
		limit,
		offset,
	)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve posts", err)
		return
	}

	// Return paginated response
	response.SuccessWithPagination(c, http.StatusOK, "Posts retrieved successfully", posts, limit, offset, total)
}

// GetLikedPosts handles the request to get posts liked by a user
func (h *GetHandler) GetLikedPosts(c *gin.Context) {
	// Get authenticated user ID
	authUserID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Get target user ID from URL parameter (optional, defaults to authenticated user)
	userIDStr := c.Param("userId")
	userID := authUserID.(primitive.ObjectID)

	if userIDStr != "" && userIDStr != "me" {
		if !validation.IsValidObjectID(userIDStr) {
			response.ValidationError(c, "Invalid user ID", nil)
			return
		}
		targetID, _ := primitive.ObjectIDFromHex(userIDStr)
		userID = targetID
	}

	// Get sort options
	sortBy := c.DefaultQuery("sort_by", "recent") // recent, popular

	// Get pagination parameters
	limit, offset := response.GetPaginationParams(c)

	// Get liked posts
	posts, total, err := h.postService.GetLikedPosts(c.Request.Context(), userID, authUserID.(primitive.ObjectID), sortBy, limit, offset)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve liked posts", err)
		return
	}

	// Return paginated response
	response.SuccessWithPagination(c, http.StatusOK, "Liked posts retrieved successfully", posts, limit, offset, total)
}
