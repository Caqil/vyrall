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

// DraftHandler handles post draft operations
type DraftHandler struct {
	postService *post.Service
}

// NewDraftHandler creates a new draft handler
func NewDraftHandler(postService *post.Service) *DraftHandler {
	return &DraftHandler{
		postService: postService,
	}
}

// CreateDraft handles the request to create a post draft
func (h *DraftHandler) CreateDraft(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Parse request body
	var req struct {
		Content        string   `json:"content,omitempty"`
		MediaIDs       []string `json:"media_ids,omitempty"`
		Hashtags       []string `json:"tags,omitempty"`
		MentionedUsers []string `json:"mentioned_users,omitempty"`
		Location       *struct {
			Name      string  `json:"name,omitempty"`
			Latitude  float64 `json:"latitude"`
			Longitude float64 `json:"longitude"`
		} `json:"location,omitempty"`
		Privacy       string   `json:"privacy,omitempty"` // public, followers, private
		AllowComments bool     `json:"allow_comments"`
		NSFW          bool     `json:"nsfw,omitempty"`
		PostAs        string   `json:"post_as,omitempty"` // user, page, group
		PageID        string   `json:"page_id,omitempty"`
		GroupID       string   `json:"group_id,omitempty"`
		DraftType     string   `json:"draft_type,omitempty"` // regular, poll, story
		PollOptions   []string `json:"poll_options,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "Invalid request body", err.Error())
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

	// Set default draft type if not provided
	draftType := req.DraftType
	if draftType == "" {
		draftType = "regular"
	} else if draftType != "regular" && draftType != "poll" && draftType != "story" {
		response.ValidationError(c, "Invalid draft type. Must be 'regular', 'poll', or 'story'", nil)
		return
	}

	// Validate poll options if draft type is poll
	if draftType == "poll" && len(req.PollOptions) < 2 {
		response.ValidationError(c, "Poll drafts must have at least 2 options", nil)
		return
	}

	// Create the draft
	draft, err := h.postService.CreateDraft(
		c.Request.Context(),
		userID.(primitive.ObjectID),
		req.Content,
		mediaIDs,
		req.Hashtags,
		mentionedUserIDs,
		location,
		req.Privacy,
		req.AllowComments,
		req.NSFW,
		pageID,
		groupID,
		draftType,
		req.PollOptions,
	)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to create draft", err)
		return
	}

	// Return success response
	response.Created(c, "Draft created successfully", draft)
}

// GetDrafts handles the request to get user's drafts
func (h *DraftHandler) GetDrafts(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Get query parameters
	draftType := c.DefaultQuery("type", "") // regular, poll, story, all

	// Get pagination parameters
	limit, offset := response.GetPaginationParams(c)

	// Get drafts
	drafts, total, err := h.postService.GetDrafts(c.Request.Context(), userID.(primitive.ObjectID), draftType, limit, offset)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve drafts", err)
		return
	}

	// Return paginated response
	response.SuccessWithPagination(c, http.StatusOK, "Drafts retrieved successfully", drafts, limit, offset, total)
}

// GetDraft handles the request to get a specific draft
func (h *DraftHandler) GetDraft(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Get draft ID from URL parameter
	draftIDStr := c.Param("id")
	if !validation.IsValidObjectID(draftIDStr) {
		response.ValidationError(c, "Invalid draft ID", nil)
		return
	}
	draftID, _ := primitive.ObjectIDFromHex(draftIDStr)

	// Get the draft
	draft, err := h.postService.GetDraft(c.Request.Context(), draftID, userID.(primitive.ObjectID))
	if err != nil {
		response.NotFoundError(c, "Draft not found")
		return
	}

	// Return success response
	response.OK(c, "Draft retrieved successfully", draft)
}

// UpdateDraft handles the request to update a draft
func (h *DraftHandler) UpdateDraft(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Get draft ID from URL parameter
	draftIDStr := c.Param("id")
	if !validation.IsValidObjectID(draftIDStr) {
		response.ValidationError(c, "Invalid draft ID", nil)
		return
	}
	draftID, _ := primitive.ObjectIDFromHex(draftIDStr)

	// Parse request body (similar to create draft)
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
		Privacy       string   `json:"privacy,omitempty"`
		AllowComments *bool    `json:"allow_comments,omitempty"`
		NSFW          *bool    `json:"nsfw,omitempty"`
		PollOptions   []string `json:"poll_options,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "Invalid request body", err.Error())
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

	if req.Tags != nil {
		updates["hashtags"] = req.Tags
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

	if req.PollOptions != nil {
		if len(req.PollOptions) < 2 {
			response.ValidationError(c, "Poll drafts must have at least 2 options", nil)
			return
		}
		updates["poll.options"] = req.PollOptions
	}

	if len(updates) == 0 {
		response.ValidationError(c, "No updates provided", nil)
		return
	}

	// Update the draft
	updatedDraft, err := h.postService.UpdateDraft(c.Request.Context(), draftID, userID.(primitive.ObjectID), updates)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to update draft", err)
		return
	}

	// Return success response
	response.OK(c, "Draft updated successfully", updatedDraft)
}

// DeleteDraft handles the request to delete a draft
func (h *DraftHandler) DeleteDraft(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Get draft ID from URL parameter
	draftIDStr := c.Param("id")
	if !validation.IsValidObjectID(draftIDStr) {
		response.ValidationError(c, "Invalid draft ID", nil)
		return
	}
	draftID, _ := primitive.ObjectIDFromHex(draftIDStr)

	// Delete the draft
	err := h.postService.DeleteDraft(c.Request.Context(), draftID, userID.(primitive.ObjectID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to delete draft", err)
		return
	}

	// Return success response
	response.OK(c, "Draft deleted successfully", nil)
}

// PublishDraft handles the request to publish a draft
func (h *DraftHandler) PublishDraft(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Get draft ID from URL parameter
	draftIDStr := c.Param("id")
	if !validation.IsValidObjectID(draftIDStr) {
		response.ValidationError(c, "Invalid draft ID", nil)
		return
	}
	draftID, _ := primitive.ObjectIDFromHex(draftIDStr)

	// Publish the draft
	publishedPost, err := h.postService.PublishDraft(c.Request.Context(), draftID, userID.(primitive.ObjectID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to publish draft", err)
		return
	}

	// Return success response
	response.OK(c, "Draft published successfully", publishedPost)
}
