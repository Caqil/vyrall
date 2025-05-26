package posts

import (
	"net/http"

	"github.com/Caqil/vyrall/internal/models"
	"github.com/Caqil/vyrall/internal/services/post"
	"github.com/Caqil/vyrall/internal/utils/response"
	"github.com/Caqil/vyrall/internal/utils/validation"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// UpdateHandler handles post update operations
type UpdateHandler struct {
	postService *post.Service
}

// NewUpdateHandler creates a new update handler
func NewUpdateHandler(postService *post.Service) *UpdateHandler {
	return &UpdateHandler{
		postService: postService,
	}
}

// UpdatePost handles the request to update a post
func (h *UpdateHandler) UpdatePost(c *gin.Context) {
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
		Content        string   `json:"content,omitempty"`
		MediaIDs       []string `json:"media_ids,omitempty"`
		Tags           []string `json:"tags,omitempty"`
		MentionedUsers []string `json:"mentioned_users,omitempty"`
		Location       *struct {
			Name      string  `json:"name,omitempty"`
			Latitude  float64 `json:"latitude"`
			Longitude float64 `json:"longitude"`
		} `json:"location,omitempty"`
		Privacy       string `json:"privacy,omitempty"`
		AllowComments *bool  `json:"allow_comments,omitempty"`
		NSFW          *bool  `json:"nsfw,omitempty"`
		EnableLikes   *bool  `json:"enable_likes,omitempty"`
		EnableSharing *bool  `json:"enable_sharing,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "Invalid request body", err.Error())
		return
	}

	// Check if the post exists and belongs to the user
	post, err := h.postService.GetPost(c.Request.Context(), postID, userID.(primitive.ObjectID))
	if err != nil {
		response.NotFoundError(c, "Post not found")
		return
	}

	if post.UserID != userID.(primitive.ObjectID) {
		response.ForbiddenError(c, "You don't have permission to update this post")
		return
	}

	// Convert media IDs to ObjectIDs if provided
	var mediaIDs []primitive.ObjectID
	if req.MediaIDs != nil {
		mediaIDs = make([]primitive.ObjectID, 0, len(req.MediaIDs))
		for _, idStr := range req.MediaIDs {
			if !validation.IsValidObjectID(idStr) {
				continue // Skip invalid IDs
			}
			mediaID, _ := primitive.ObjectIDFromHex(idStr)
			mediaIDs = append(mediaIDs, mediaID)
		}
	}

	// Convert mentioned user IDs to ObjectIDs if provided
	var mentionedUserIDs []primitive.ObjectID
	if req.MentionedUsers != nil {
		mentionedUserIDs = make([]primitive.ObjectID, 0, len(req.MentionedUsers))
		for _, idStr := range req.MentionedUsers {
			if !validation.IsValidObjectID(idStr) {
				continue // Skip invalid IDs
			}
			mentionID, _ := primitive.ObjectIDFromHex(idStr)
			mentionedUserIDs = append(mentionedUserIDs, mentionID)
		}
	}

	// Create updates map
	updates := make(map[string]interface{})

	if req.Content != "" {
		updates["content"] = req.Content
	}
	if req.MediaIDs != nil {
		updates["media_ids"] = mediaIDs
	}
	if req.Tags != nil {
		updates["tags"] = req.Tags
	}
	if req.MentionedUsers != nil {
		updates["mentioned_users"] = mentionedUserIDs
	}
	if req.Location != nil {
		location := &models.Location{
			Name:      req.Location.Name,
			Latitude:  req.Location.Latitude,
			Longitude: req.Location.Longitude,
		}
		updates["location"] = location
	}
	if req.Privacy != "" {
		if req.Privacy != "public" && req.Privacy != "followers" && req.Privacy != "private" {
			response.ValidationError(c, "Invalid privacy setting. Must be 'public', 'followers', or 'private'", nil)
			return
		}
		updates["privacy"] = req.Privacy
	}
	if req.AllowComments != nil {
		updates["allow_comments"] = *req.AllowComments
	}
	if req.NSFW != nil {
		updates["nsfw"] = *req.NSFW
	}
	if req.EnableLikes != nil {
		updates["enable_likes"] = *req.EnableLikes
	}
	if req.EnableSharing != nil {
		updates["enable_sharing"] = *req.EnableSharing
	}

	if len(updates) == 0 {
		response.ValidationError(c, "No updates provided", nil)
		return
	}

	// Update the post
	updatedPost, err := h.postService.UpdatePost(c.Request.Context(), postID, updates)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to update post", err)
		return
	}

	// Return success response
	response.OK(c, "Post updated successfully", updatedPost)
}

// UpdatePollPost handles the request to update a poll post
func (h *UpdateHandler) UpdatePollPost(c *gin.Context) {
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
		Question        string   `json:"question,omitempty"`
		Options         []string `json:"options,omitempty"`
		ExpiresAt       string   `json:"expires_at,omitempty"`
		AllowMultiple   *bool    `json:"allow_multiple,omitempty"`
		AllowAddOptions *bool    `json:"allow_add_options,omitempty"`
		Privacy         string   `json:"privacy,omitempty"`
		Tags            []string `json:"tags,omitempty"`
		MentionedUsers  []string `json:"mentioned_users,omitempty"`
		Location        *struct {
			Name      string  `json:"name,omitempty"`
			Latitude  float64 `json:"latitude"`
			Longitude float64 `json:"longitude"`
		} `json:"location,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "Invalid request body", err.Error())
		return
	}

	// Check if the post exists, is a poll, and belongs to the user
	post, err := h.postService.GetPost(c.Request.Context(), postID, userID.(primitive.ObjectID))
	if err != nil {
		response.NotFoundError(c, "Post not found")
		return
	}

	if post.UserID != userID.(primitive.ObjectID) {
		response.ForbiddenError(c, "You don't have permission to update this post")
		return
	}

	if post.Type != "poll" {
		response.ValidationError(c, "This is not a poll post", nil)
		return
	}

	// Convert mentioned user IDs to ObjectIDs if provided
	var mentionedUserIDs []primitive.ObjectID
	if req.MentionedUsers != nil {
		mentionedUserIDs = make([]primitive.ObjectID, 0, len(req.MentionedUsers))
		for _, idStr := range req.MentionedUsers {
			if !validation.IsValidObjectID(idStr) {
				continue // Skip invalid IDs
			}
			mentionID, _ := primitive.ObjectIDFromHex(idStr)
			mentionedUserIDs = append(mentionedUserIDs, mentionID)
		}
	}

	// Create updates map
	updates := make(map[string]interface{})

	if req.Question != "" {
		updates["poll.question"] = req.Question
	}
	if req.Options != nil {
		if len(req.Options) < 2 {
			response.ValidationError(c, "Poll must have at least 2 options", nil)
			return
		}
		if len(req.Options) > 10 {
			response.ValidationError(c, "Poll can have at most 10 options", nil)
			return
		}
		updates["poll.options"] = req.Options
	}
	if req.ExpiresAt != "" {
		updates["poll.expires_at"] = req.ExpiresAt
	}
	if req.AllowMultiple != nil {
		updates["poll.allow_multiple"] = *req.AllowMultiple
	}
	if req.AllowAddOptions != nil {
		updates["poll.allow_add_options"] = *req.AllowAddOptions
	}
	if req.Privacy != "" {
		if req.Privacy != "public" && req.Privacy != "followers" && req.Privacy != "private" {
			response.ValidationError(c, "Invalid privacy setting. Must be 'public', 'followers', or 'private'", nil)
			return
		}
		updates["privacy"] = req.Privacy
	}
	if req.Tags != nil {
		updates["tags"] = req.Tags
	}
	if req.MentionedUsers != nil {
		updates["mentioned_users"] = mentionedUserIDs
	}
	if req.Location != nil {
		location := &models.Location{
			Name:      req.Location.Name,
			Latitude:  req.Location.Latitude,
			Longitude: req.Location.Longitude,
		}
		updates["location"] = location
	}

	if len(updates) == 0 {
		response.ValidationError(c, "No updates provided", nil)
		return
	}

	// Update the poll post
	updatedPost, err := h.postService.UpdatePost(c.Request.Context(), postID, updates)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to update poll post", err)
		return
	}

	// Return success response
	response.OK(c, "Poll post updated successfully", updatedPost)
}

// AddPollOption handles the request to add a new option to a poll
func (h *UpdateHandler) AddPollOption(c *gin.Context) {
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
		Option string `json:"option" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "Invalid request body", err.Error())
		return
	}

	// Add poll option
	updatedPoll, err := h.postService.AddPollOption(c.Request.Context(), postID, userID.(primitive.ObjectID), req.Option)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to add poll option", err)
		return
	}

	// Return success response
	response.OK(c, "Poll option added successfully", updatedPoll)
}

// PinPost handles the request to pin a post to user's profile
func (h *UpdateHandler) PinPost(c *gin.Context) {
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

	// Pin the post
	err := h.postService.PinPost(c.Request.Context(), postID, userID.(primitive.ObjectID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to pin post", err)
		return
	}

	// Return success response
	response.OK(c, "Post pinned successfully", nil)
}

// UnpinPost handles the request to unpin a post from user's profile
func (h *UpdateHandler) UnpinPost(c *gin.Context) {
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

	// Unpin the post
	err := h.postService.UnpinPost(c.Request.Context(), postID, userID.(primitive.ObjectID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to unpin post", err)
		return
	}

	// Return success response
	response.OK(c, "Post unpinned successfully", nil)
}
