package posts

import (
	"net/http"
	"time"

	"github.com/Caqil/vyrall/internal/models"
	"github.com/Caqil/vyrall/internal/services/post"
	"github.com/Caqil/vyrall/internal/utils/response"
	"github.com/Caqil/vyrall/internal/utils/validation"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ScheduleHandler handles scheduled post operations
type ScheduleHandler struct {
	postService *post.Service
}

// NewScheduleHandler creates a new schedule handler
func NewScheduleHandler(postService *post.Service) *ScheduleHandler {
	return &ScheduleHandler{
		postService: postService,
	}
}

// SchedulePost handles the request to schedule a post for future publishing
func (h *ScheduleHandler) SchedulePost(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Parse request body
	var req struct {
		Content        string   `json:"content"`
		MediaIDs       []string `json:"media_ids,omitempty"`
		ScheduledFor   string   `json:"scheduled_for" binding:"required"` // ISO 8601 format
		Tags           []string `json:"tags,omitempty"`
		MentionedUsers []string `json:"mentioned_users,omitempty"`
		Location       *struct {
			Name      string  `json:"name,omitempty"`
			Latitude  float64 `json:"latitude"`
			Longitude float64 `json:"longitude"`
		} `json:"location,omitempty"`
		Privacy       string `json:"privacy,omitempty"` // public, followers, private
		AllowComments bool   `json:"allow_comments"`
		NSFW          bool   `json:"nsfw,omitempty"`
		PostAs        string `json:"post_as,omitempty"` // user, page, group
		PageID        string `json:"page_id,omitempty"`
		GroupID       string `json:"group_id,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "Invalid request body", err.Error())
		return
	}

	// Validate content or media
	if req.Content == "" && len(req.MediaIDs) == 0 {
		response.ValidationError(c, "Post must contain either text content or media", nil)
		return
	}

	// Parse scheduled time
	scheduledTime, err := time.Parse(time.RFC3339, req.ScheduledFor)
	if err != nil {
		response.ValidationError(c, "Invalid scheduled time format. Use ISO 8601 format (YYYY-MM-DDTHH:MM:SSZ)", err.Error())
		return
	}

	// Ensure scheduled time is in the future
	if scheduledTime.Before(time.Now()) {
		response.ValidationError(c, "Scheduled time must be in the future", nil)
		return
	}

	// Convert media IDs to ObjectIDs
	mediaIDs := make([]primitive.ObjectID, 0, len(req.MediaIDs))
	for _, idStr := range req.MediaIDs {
		if !validation.IsValidObjectID(idStr) {
			continue // Skip invalid IDs
		}
		mediaID, _ := primitive.ObjectIDFromHex(idStr)
		mediaIDs = append(mediaIDs, mediaID)
	}

	// Convert mentioned user IDs to ObjectIDs
	mentionedUserIDs := make([]primitive.ObjectID, 0, len(req.MentionedUsers))
	for _, idStr := range req.MentionedUsers {
		if !validation.IsValidObjectID(idStr) {
			continue // Skip invalid IDs
		}
		mentionID, _ := primitive.ObjectIDFromHex(idStr)
		mentionedUserIDs = append(mentionedUserIDs, mentionID)
	}

	// Set default privacy if not provided
	if req.Privacy == "" {
		req.Privacy = "public"
	} else if req.Privacy != "public" && req.Privacy != "followers" && req.Privacy != "private" {
		response.ValidationError(c, "Invalid privacy setting. Must be 'public', 'followers', or 'private'", nil)
		return
	}

	// Create location if provided
	var location *models.Location
	if req.Location != nil {
		location = &models.Location{
			Name: req.Location.Name,
			Coordinates: models.GeoPoint{
				Type: "Point",
				// Note: GeoJSON uses [longitude, latitude] order
				Coordinates: []float64{req.Location.Longitude, req.Location.Latitude},
			},
		}
	}

	// Determine post owner (user, page, or group)
	var pageID, groupID *primitive.ObjectID
	if req.PostAs == "page" && req.PageID != "" {
		if !validation.IsValidObjectID(req.PageID) {
			response.ValidationError(c, "Invalid page ID", nil)
			return
		}
		id, _ := primitive.ObjectIDFromHex(req.PageID)
		pageID = &id
	} else if req.PostAs == "group" && req.GroupID != "" {
		if !validation.IsValidObjectID(req.GroupID) {
			response.ValidationError(c, "Invalid group ID", nil)
			return
		}
		id, _ := primitive.ObjectIDFromHex(req.GroupID)
		groupID = &id
	}

	// Schedule the post
	scheduledPost, err := h.postService.SchedulePost(
		c.Request.Context(),
		userID.(primitive.ObjectID),
		req.Content,
		mediaIDs,
		scheduledTime,
		req.Tags,
		mentionedUserIDs,
		location,
		req.Privacy,
		req.AllowComments,
		req.NSFW,
		pageID,
		groupID,
	)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to schedule post", err)
		return
	}

	// Return success response
	response.Created(c, "Post scheduled successfully", scheduledPost)
}

// GetScheduledPosts handles the request to get scheduled posts
func (h *ScheduleHandler) GetScheduledPosts(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Get pagination parameters
	limit, offset := response.GetPaginationParams(c)

	// Get scheduled posts
	posts, total, err := h.postService.GetScheduledPosts(c.Request.Context(), userID.(primitive.ObjectID), limit, offset)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve scheduled posts", err)
		return
	}

	// Return paginated response
	response.SuccessWithPagination(c, http.StatusOK, "Scheduled posts retrieved successfully", posts, limit, offset, total)
}

// UpdateScheduledPost handles the request to update a scheduled post
func (h *ScheduleHandler) UpdateScheduledPost(c *gin.Context) {
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
		ScheduledFor   string   `json:"scheduled_for,omitempty"` // ISO 8601 format
		Hashtags       []string `json:"tags,omitempty"`
		MentionedUsers []string `json:"mentioned_users,omitempty"`
		Location       *struct {
			Name      string  `json:"name,omitempty"`
			Latitude  float64 `json:"latitude"`
			Longitude float64 `json:"longitude"`
		} `json:"location,omitempty"`
		Privacy       string `json:"privacy,omitempty"` // public, followers, private
		AllowComments *bool  `json:"allow_comments,omitempty"`
		NSFW          *bool  `json:"nsfw,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "Invalid request body", err.Error())
		return
	}

	// Check if the post exists, is scheduled, and belongs to the user
	_, err := h.postService.GetScheduledPost(c.Request.Context(), postID, userID.(primitive.ObjectID))
	if err != nil {
		response.NotFoundError(c, "Scheduled post not found")
		return
	}

	// Create updates map
	updates := make(map[string]interface{})

	if req.Content != "" {
		updates["content"] = req.Content
	}

	if req.MediaIDs != nil {
		mediaIDs := make([]primitive.ObjectID, 0, len(req.MediaIDs))
		for _, idStr := range req.MediaIDs {
			if !validation.IsValidObjectID(idStr) {
				continue // Skip invalid IDs
			}
			mediaID, _ := primitive.ObjectIDFromHex(idStr)
			mediaIDs = append(mediaIDs, mediaID)
		}
		updates["media_files"] = mediaIDs
	}

	if req.ScheduledFor != "" {
		scheduledTime, err := time.Parse(time.RFC3339, req.ScheduledFor)
		if err != nil {
			response.ValidationError(c, "Invalid scheduled time format. Use ISO 8601 format (YYYY-MM-DDTHH:MM:SSZ)", err.Error())
			return
		}

		// Ensure scheduled time is in the future
		if scheduledTime.Before(time.Now()) {
			response.ValidationError(c, "Scheduled time must be in the future", nil)
			return
		}

		updates["scheduled_for"] = scheduledTime
	}

	if req.Hashtags != nil {
		updates["hashtags"] = req.Hashtags
	}

	if req.MentionedUsers != nil {
		mentionedUserIDs := make([]primitive.ObjectID, 0, len(req.MentionedUsers))
		for _, idStr := range req.MentionedUsers {
			if !validation.IsValidObjectID(idStr) {
				continue // Skip invalid IDs
			}
			mentionID, _ := primitive.ObjectIDFromHex(idStr)
			mentionedUserIDs = append(mentionedUserIDs, mentionID)
		}
		updates["mentioned_users"] = mentionedUserIDs
	}

	if req.Location != nil {
		location := &models.Location{
			Name: req.Location.Name,
			Coordinates: models.GeoPoint{
				Type: "Point",
				// Note: GeoJSON uses [longitude, latitude] order
				Coordinates: []float64{req.Location.Longitude, req.Location.Latitude},
			},
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

	if len(updates) == 0 {
		response.ValidationError(c, "No updates provided", nil)
		return
	}

	// Update the scheduled post
	updatedPost, err := h.postService.UpdateScheduledPost(c.Request.Context(), postID, userID.(primitive.ObjectID), updates)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to update scheduled post", err)
		return
	}

	// Return success response
	response.OK(c, "Scheduled post updated successfully", updatedPost)
}

// DeleteScheduledPost handles the request to delete a scheduled post
func (h *ScheduleHandler) DeleteScheduledPost(c *gin.Context) {
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

	// Delete the scheduled post
	err := h.postService.DeleteScheduledPost(c.Request.Context(), postID, userID.(primitive.ObjectID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to delete scheduled post", err)
		return
	}

	// Return success response
	response.OK(c, "Scheduled post deleted successfully", nil)
}

// PublishScheduledPost handles the request to publish a scheduled post immediately
func (h *ScheduleHandler) PublishScheduledPost(c *gin.Context) {
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

	// Publish the scheduled post
	publishedPost, err := h.postService.PublishScheduledPost(c.Request.Context(), postID, userID.(primitive.ObjectID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to publish scheduled post", err)
		return
	}

	// Return success response
	response.OK(c, "Scheduled post published successfully", publishedPost)
}
