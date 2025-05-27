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

// CreateHandler handles post creation operations
type CreateHandler struct {
	postService *post.Service
}

// NewCreateHandler creates a new create handler
func NewCreateHandler(postService *post.Service) *CreateHandler {
	return &CreateHandler{
		postService: postService,
	}
}

// CreatePost handles the request to create a new post
func (h *CreateHandler) CreatePost(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Parse request body
	var req struct {
		Content  string   `json:"content"`
		MediaIDs []string `json:"media_ids,omitempty"`
		Location *struct {
			Name      string  `json:"name,omitempty"`
			Latitude  float64 `json:"latitude"`
			Longitude float64 `json:"longitude"`
		} `json:"location,omitempty"`
		Hashtags       []string `json:"tags,omitempty"`
		MentionedUsers []string `json:"mentioned_users,omitempty"`
		Privacy        string   `json:"privacy,omitempty"` // public, followers, private
		AllowComments  bool     `json:"allow_comments"`
		NSFW           bool     `json:"nsfw,omitempty"`
		EnableLikes    bool     `json:"enable_likes"`
		EnableSharing  bool     `json:"enable_sharing"`
		PostAs         string   `json:"post_as,omitempty"` // user, page, group
		PageID         string   `json:"page_id,omitempty"`
		GroupID        string   `json:"group_id,omitempty"`
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
				Type:        "Point",
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

	// Create post
	post := &models.Post{
		UserID:         userID.(primitive.ObjectID),
		Content:        req.Content,
		Hashtags:       req.Hashtags,
		MentionedUsers: mentionedUserIDs,
		Location:       location,
		Privacy:        req.Privacy,
		AllowComments:  req.AllowComments,
		NSFW:           req.NSFW,
		EnableLikes:    req.EnableLikes,
		EnableSharing:  req.EnableSharing,
		PageID:         pageID,
		GroupID:        groupID,
	}

	// Create the post
	createdPost, err := h.postService.CreatePost(c.Request.Context(), post, mediaIDs)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to create post", err)
		return
	}

	// Return success response
	response.Created(c, "Post created successfully", createdPost)
}

// CreatePoll handles the request to create a poll post
func (h *CreateHandler) CreatePoll(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Parse request body
	var req struct {
		Question        string   `json:"question" binding:"required"`
		Options         []string `json:"options" binding:"required"`
		ExpiresAt       string   `json:"expires_at,omitempty"` // ISO 8601 format
		AllowMultiple   bool     `json:"allow_multiple,omitempty"`
		AllowAddOptions bool     `json:"allow_add_options,omitempty"`
		Privacy         string   `json:"privacy,omitempty"` // public, followers, private
		MediaIDs        []string `json:"media_ids,omitempty"`
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

	// Validate poll options
	if len(req.Options) < 2 {
		response.ValidationError(c, "Poll must have at least 2 options", nil)
		return
	}
	if len(req.Options) > 10 {
		response.ValidationError(c, "Poll can have at most 10 options", nil)
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

	// Create poll
	poll := &models.Poll{
		Question:        req.Question,
		Options:         req.Options,
		ExpiresAt:       req.ExpiresAt,
		AllowMultiple:   req.AllowMultiple,
		AllowAddOptions: req.AllowAddOptions,
	}

	// Create the poll post
	createdPost, err := h.postService.CreatePollPost(
		c.Request.Context(),
		userID.(primitive.ObjectID),
		poll,
		req.Privacy,
		mediaIDs,
		req.Tags,
		mentionedUserIDs,
		location,
	)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to create poll", err)
		return
	}

	// Return success response
	response.Created(c, "Poll created successfully", createdPost)
}

// CreateSharedPost handles the request to share an existing post
func (h *CreateHandler) CreateSharedPost(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Parse request body
	var req struct {
		OriginalPostID string   `json:"original_post_id" binding:"required"`
		Caption        string   `json:"caption,omitempty"`
		Privacy        string   `json:"privacy,omitempty"` // public, followers, private
		Tags           []string `json:"tags,omitempty"`
		MentionedUsers []string `json:"mentioned_users,omitempty"`
		Location       *struct {
			Name      string  `json:"name,omitempty"`
			Latitude  float64 `json:"latitude"`
			Longitude float64 `json:"longitude"`
		} `json:"location,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "Invalid request body", err.Error())
		return
	}

	// Validate original post ID
	if !validation.IsValidObjectID(req.OriginalPostID) {
		response.ValidationError(c, "Invalid original post ID", nil)
		return
	}
	originalPostID, _ := primitive.ObjectIDFromHex(req.OriginalPostID)

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

	// Create the shared post
	sharedPost, err := h.postService.CreateSharedPost(
		c.Request.Context(),
		userID.(primitive.ObjectID),
		originalPostID,
		req.Caption,
		req.Privacy,
		req.Tags,
		mentionedUserIDs,
		location,
	)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to create shared post", err)
		return
	}

	// Return success response
	response.Created(c, "Post shared successfully", sharedPost)
}
