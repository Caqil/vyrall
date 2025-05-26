package posts

import (
	"net/http"

	"github.com/Caqil/vyrall/internal/services/post"
	"github.com/Caqil/vyrall/internal/utils/response"
	"github.com/Caqil/vyrall/internal/utils/validation"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// LikeHandler handles post like operations
type LikeHandler struct {
	postService *post.Service
}

// NewLikeHandler creates a new like handler
func NewLikeHandler(postService *post.Service) *LikeHandler {
	return &LikeHandler{
		postService: postService,
	}
}

// LikePost handles the request to like a post
func (h *LikeHandler) LikePost(c *gin.Context) {
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

	// Like the post
	err := h.postService.LikePost(c.Request.Context(), postID, userID.(primitive.ObjectID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to like post", err)
		return
	}

	// Return success response
	response.OK(c, "Post liked successfully", nil)
}

// LikeComment handles the request to like a comment
func (h *LikeHandler) LikeComment(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Get comment ID from URL parameter
	commentIDStr := c.Param("id")
	if !validation.IsValidObjectID(commentIDStr) {
		response.ValidationError(c, "Invalid comment ID", nil)
		return
	}
	commentID, _ := primitive.ObjectIDFromHex(commentIDStr)

	// Like the comment
	err := h.postService.LikeComment(c.Request.Context(), commentID, userID.(primitive.ObjectID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to like comment", err)
		return
	}

	// Return success response
	response.OK(c, "Comment liked successfully", nil)
}

// GetPostLikes handles the request to get users who liked a post
func (h *LikeHandler) GetPostLikes(c *gin.Context) {
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

	// Get post likes
	users, total, err := h.postService.GetPostLikes(c.Request.Context(), postID, userID, limit, offset)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve post likes", err)
		return
	}

	// Return paginated response
	response.SuccessWithPagination(c, http.StatusOK, "Post likes retrieved successfully", users, limit, offset, total)
}

// GetCommentLikes handles the request to get users who liked a comment
func (h *LikeHandler) GetCommentLikes(c *gin.Context) {
	// Get user ID from context (may be nil for unauthenticated users)
	var userID primitive.ObjectID
	if id, exists := c.Get("userID"); exists {
		userID = id.(primitive.ObjectID)
	}

	// Get comment ID from URL parameter
	commentIDStr := c.Param("id")
	if !validation.IsValidObjectID(commentIDStr) {
		response.ValidationError(c, "Invalid comment ID", nil)
		return
	}
	commentID, _ := primitive.ObjectIDFromHex(commentIDStr)

	// Get pagination parameters
	limit, offset := response.GetPaginationParams(c)

	// Get comment likes
	users, total, err := h.postService.GetCommentLikes(c.Request.Context(), commentID, userID, limit, offset)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve comment likes", err)
		return
	}

	// Return paginated response
	response.SuccessWithPagination(c, http.StatusOK, "Comment likes retrieved successfully", users, limit, offset, total)
}
