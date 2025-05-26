package posts

import (
	"net/http"

	"github.com/Caqil/vyrall/internal/services/post"
	"github.com/Caqil/vyrall/internal/utils/response"
	"github.com/Caqil/vyrall/internal/utils/validation"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// UnbookmarkHandler handles post unbookmark operations
type UnbookmarkHandler struct {
	postService *post.Service
}

// NewUnbookmarkHandler creates a new unbookmark handler
func NewUnbookmarkHandler(postService *post.Service) *UnbookmarkHandler {
	return &UnbookmarkHandler{
		postService: postService,
	}
}

// UnbookmarkPost handles the request to remove a post from bookmarks
func (h *UnbookmarkHandler) UnbookmarkPost(c *gin.Context) {
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

	// Unbookmark the post
	err := h.postService.UnbookmarkPost(c.Request.Context(), postID, userID.(primitive.ObjectID), collection)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to remove post from bookmarks", err)
		return
	}

	// Return success response
	response.OK(c, "Post removed from bookmarks successfully", nil)
}

// MoveBookmarkedPost handles the request to move a bookmarked post to another collection
func (h *UnbookmarkHandler) MoveBookmarkedPost(c *gin.Context) {
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

	// Parse request body
	var req struct {
		SourceCollection      string `json:"source_collection,omitempty"`
		DestinationCollection string `json:"destination_collection" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "Invalid request body", err.Error())
		return
	}

	// Move the bookmarked post
	err := h.postService.MoveBookmarkedPost(
		c.Request.Context(),
		postID,
		userID.(primitive.ObjectID),
		req.SourceCollection,
		req.DestinationCollection,
	)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to move bookmarked post", err)
		return
	}

	// Return success response
	response.OK(c, "Bookmarked post moved successfully", nil)
}
