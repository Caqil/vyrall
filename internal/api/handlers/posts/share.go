package posts

import (
	"net/http"

	"github.com/Caqil/vyrall/internal/services/post"
	"github.com/Caqil/vyrall/internal/utils/response"
	"github.com/Caqil/vyrall/internal/utils/validation"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ShareHandler handles post sharing operations
type ShareHandler struct {
	postService *post.Service
}

// NewShareHandler creates a new share handler
func NewShareHandler(postService *post.Service) *ShareHandler {
	return &ShareHandler{
		postService: postService,
	}
}

// SharePost handles the request to share a post
func (h *ShareHandler) SharePost(c *gin.Context) {
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
		Platform string `json:"platform,omitempty"` // internal, twitter, facebook, etc.
		Caption  string `json:"caption,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "Invalid request body", err.Error())
		return
	}

	// Default to internal platform if not specified
	if req.Platform == "" {
		req.Platform = "internal"
	}

	// Share the post
	shareResult, err := h.postService.SharePost(c.Request.Context(), postID, userID.(primitive.ObjectID), req.Platform, req.Caption)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to share post", err)
		return
	}

	// Return success response
	response.OK(c, "Post shared successfully", shareResult)
}

// GetShareLink handles the request to get a shareable link for a post
func (h *ShareHandler) GetShareLink(c *gin.Context) {
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

	// Get query parameters
	utm := c.DefaultQuery("utm", "")
	expiry := c.DefaultQuery("expiry", "")

	// Get share link
	shareLink, err := h.postService.GetShareLink(c.Request.Context(), postID, userID, utm, expiry)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to generate share link", err)
		return
	}

	// Return success response
	response.OK(c, "Share link generated successfully", gin.H{
		"share_link": shareLink,
	})
}

// ShareToDirectMessage handles the request to share a post via direct message
func (h *ShareHandler) ShareToDirectMessage(c *gin.Context) {
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
		RecipientIDs []string `json:"recipient_ids" binding:"required"`
		Message      string   `json:"message,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "Invalid request body", err.Error())
		return
	}

	if len(req.RecipientIDs) == 0 {
		response.ValidationError(c, "At least one recipient is required", nil)
		return
	}

	// Convert recipient IDs to ObjectIDs
	recipientIDs := make([]primitive.ObjectID, 0, len(req.RecipientIDs))
	for _, idStr := range req.RecipientIDs {
		if !validation.IsValidObjectID(idStr) {
			continue // Skip invalid IDs
		}
		recipientID, _ := primitive.ObjectIDFromHex(idStr)
		recipientIDs = append(recipientIDs, recipientID)
	}

	if len(recipientIDs) == 0 {
		response.ValidationError(c, "No valid recipient IDs provided", nil)
		return
	}

	// Share the post via direct message
	results, err := h.postService.ShareToDirectMessage(c.Request.Context(), postID, userID.(primitive.ObjectID), recipientIDs, req.Message)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to share post via direct message", err)
		return
	}

	// Return success response
	response.OK(c, "Post shared via direct message", results)
}

// GetSharedPosts handles the request to get posts that shared a specific post
func (h *ShareHandler) GetSharedPosts(c *gin.Context) {
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

	// Get shared posts
	posts, total, err := h.postService.GetSharedPosts(c.Request.Context(), postID, userID, limit, offset)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve shared posts", err)
		return
	}

	// Return paginated response
	response.SuccessWithPagination(c, http.StatusOK, "Shared posts retrieved successfully", posts, limit, offset, total)
}
