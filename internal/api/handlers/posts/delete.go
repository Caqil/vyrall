package posts

import (
	"net/http"

	"github.com/Caqil/vyrall/internal/services/post"
	"github.com/Caqil/vyrall/internal/utils/response"
	"github.com/Caqil/vyrall/internal/utils/validation"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// DeleteHandler handles post deletion operations
type DeleteHandler struct {
	postService *post.Service
}

// NewDeleteHandler creates a new delete handler
func NewDeleteHandler(postService *post.Service) *DeleteHandler {
	return &DeleteHandler{
		postService: postService,
	}
}

// DeletePost handles the request to delete a post
func (h *DeleteHandler) DeletePost(c *gin.Context) {
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

	// Check if the post exists and belongs to the user
	post, err := h.postService.GetPost(c.Request.Context(), postID, userID.(primitive.ObjectID))
	if err != nil {
		response.NotFoundError(c, "Post not found")
		return
	}

	// Check if user is post owner or has admin permissions
	isAdmin, _ := c.Get("isAdmin")
	if post.UserID != userID.(primitive.ObjectID) && !(isAdmin != nil && isAdmin.(bool)) {
		response.ForbiddenError(c, "You don't have permission to delete this post")
		return
	}

	// Delete the post
	err = h.postService.DeletePost(c.Request.Context(), postID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to delete post", err)
		return
	}

	// Return success response
	response.OK(c, "Post deleted successfully", nil)
}

// BulkDeletePosts handles the request to delete multiple posts
func (h *DeleteHandler) BulkDeletePosts(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Parse request body
	var req struct {
		PostIDs []string `json:"post_ids" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "Invalid request body", err.Error())
		return
	}

	if len(req.PostIDs) == 0 {
		response.ValidationError(c, "No post IDs provided", nil)
		return
	}

	// Convert post IDs to ObjectIDs
	postIDs := make([]primitive.ObjectID, 0, len(req.PostIDs))
	for _, idStr := range req.PostIDs {
		if !validation.IsValidObjectID(idStr) {
			continue // Skip invalid IDs
		}
		postID, _ := primitive.ObjectIDFromHex(idStr)
		postIDs = append(postIDs, postID)
	}

	if len(postIDs) == 0 {
		response.ValidationError(c, "No valid post IDs provided", nil)
		return
	}

	// Check if user is admin
	isAdmin, _ := c.Get("isAdmin")
	isAdminBool := isAdmin != nil && isAdmin.(bool)

	// Delete the posts (service will verify ownership for non-admins)
	deletedCount, err := h.postService.BulkDeletePosts(c.Request.Context(), postIDs, userID.(primitive.ObjectID), isAdminBool)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to delete posts", err)
		return
	}

	// Return success response
	response.OK(c, "Posts deleted successfully", gin.H{
		"deleted_count": deletedCount,
	})
}

// DeleteComment handles the request to delete a comment
func (h *DeleteHandler) DeleteComment(c *gin.Context) {
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

	// Delete the comment (service will verify ownership)
	err := h.postService.DeleteComment(c.Request.Context(), commentID, userID.(primitive.ObjectID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to delete comment", err)
		return
	}

	// Return success response
	response.OK(c, "Comment deleted successfully", nil)
}
