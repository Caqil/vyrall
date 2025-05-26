package posts

import (
	"net/http"

	"github.com/Caqil/vyrall/internal/services/post"
	"github.com/Caqil/vyrall/internal/utils/response"
	"github.com/Caqil/vyrall/internal/utils/validation"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// UnlikeHandler handles post unlike operations
type UnlikeHandler struct {
	postService *post.Service
}

// NewUnlikeHandler creates a new unlike handler
func NewUnlikeHandler(postService *post.Service) *UnlikeHandler {
	return &UnlikeHandler{
		postService: postService,
	}
}

// UnlikePost handles the request to unlike a post
func (h *UnlikeHandler) UnlikePost(c *gin.Context) {
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

	// Unlike the post
	err := h.postService.UnlikePost(c.Request.Context(), postID, userID.(primitive.ObjectID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to unlike post", err)
		return
	}

	// Return success response
	response.OK(c, "Post unliked successfully", nil)
}

// UnlikeComment handles the request to unlike a comment
func (h *UnlikeHandler) UnlikeComment(c *gin.Context) {
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

	// Unlike the comment
	err := h.postService.UnlikeComment(c.Request.Context(), commentID, userID.(primitive.ObjectID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to unlike comment", err)
		return
	}

	// Return success response
	response.OK(c, "Comment unliked successfully", nil)
}
