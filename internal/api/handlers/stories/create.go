package stories

import (
	"net/http"
	"time"

	"github.com/Caqil/vyrall/internal/models"
	"github.com/Caqil/vyrall/internal/services/story"
	"github.com/Caqil/vyrall/internal/utils/response"
	"github.com/Caqil/vyrall/internal/utils/validation"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// CreateHandler handles story creation operations
type CreateHandler struct {
	storyService *story.Service
}

// NewCreateHandler creates a new create handler
func NewCreateHandler(storyService *story.Service) *CreateHandler {
	return &CreateHandler{
		storyService: storyService,
	}
}

// CreateStory handles the request to create a new story
func (h *CreateHandler) CreateStory(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Parse request body
	var req struct {
		MediaIDs       []string `json:"media_ids" binding:"required"`
		Caption        string   `json:"caption,omitempty"`
		Hashtags       []string `json:"hashtags,omitempty"`
		MentionedUsers []string `json:"mentioned_users,omitempty"`
		Location       *struct {
			Name      string  `json:"name,omitempty"`
			Latitude  float64 `json:"latitude"`
			Longitude float64 `json:"longitude"`
		} `json:"location,omitempty"`
		Duration       int      `json:"duration,omitempty"`     // in seconds, default 24 hours
		PrivacyType    string   `json:"privacy_type,omitempty"` // everyone, followers, close_friends
		HideFromUsers  []string `json:"hide_from_users,omitempty"`
		VisibleToUsers []string `json:"visible_to_users,omitempty"`
		Music          *struct {
			Title    string `json:"title,omitempty"`
			Artist   string `json:"artist,omitempty"`
			AudioURL string `json:"audio_url,omitempty"`
			CoverURL string `json:"cover_url,omitempty"`
		} `json:"music,omitempty"`
		Link *struct {
			URL          string `json:"url,omitempty"`
			Title        string `json:"title,omitempty"`
			Description  string `json:"description,omitempty"`
			ThumbnailURL string `json:"thumbnail_url,omitempty"`
		} `json:"link,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "Invalid request body", err.Error())
		return
	}

	if len(req.MediaIDs) == 0 {
		response.ValidationError(c, "At least one media file is required", nil)
		return
	}

	// Convert media IDs to ObjectIDs
	mediaIDs := make([]primitive.ObjectID, 0, len(req.MediaIDs))
	for _, idStr := range req.MediaIDs {
		if !validation.IsValidObjectID(idStr) {
			response.ValidationError(c, "Invalid media ID: "+idStr, nil)
			return
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

	// Convert hide from user IDs to ObjectIDs
	hideFromUserIDs := make([]primitive.ObjectID, 0, len(req.HideFromUsers))
	for _, idStr := range req.HideFromUsers {
		if !validation.IsValidObjectID(idStr) {
			continue // Skip invalid IDs
		}
		hideID, _ := primitive.ObjectIDFromHex(idStr)
		hideFromUserIDs = append(hideFromUserIDs, hideID)
	}

	// Convert visible to user IDs to ObjectIDs
	visibleToUserIDs := make([]primitive.ObjectID, 0, len(req.VisibleToUsers))
	for _, idStr := range req.VisibleToUsers {
		if !validation.IsValidObjectID(idStr) {
			continue // Skip invalid IDs
		}
		visibleID, _ := primitive.ObjectIDFromHex(idStr)
		visibleToUserIDs = append(visibleToUserIDs, visibleID)
	}

	// Create story object
	story := &models.Story{
		UserID:         userID.(primitive.ObjectID),
		Caption:        req.Caption,
		Hashtags:       req.Hashtags,
		MentionedUsers: mentionedUserIDs,
		CreatedAt:      time.Now(),
	}

	// Set duration (default to 24 hours if not provided)
	if req.Duration > 0 {
		story.Duration = req.Duration
	} else {
		story.Duration = 86400 // 24 hours in seconds
	}
	story.ExpiresAt = story.CreatedAt.Add(time.Duration(story.Duration) * time.Second)

	// Set location if provided
	if req.Location != nil {
		story.Location = &models.Location{
			Name: req.Location.Name,
			Coordinates: models.GeoPoint{
				Type:        "Point",
				Coordinates: []float64{req.Location.Longitude, req.Location.Latitude},
			},
		}
	}

	// Set privacy settings
	story.PrivacySettings = models.StoryPrivacy{
		Visibility:     "everyone", // Default
		HideFromUsers:  hideFromUserIDs,
		VisibleToUsers: visibleToUserIDs,
	}

	if req.PrivacyType != "" {
		if req.PrivacyType == "everyone" || req.PrivacyType == "followers" || req.PrivacyType == "close_friends" {
			story.PrivacySettings.Visibility = req.PrivacyType
		} else {
			response.ValidationError(c, "Invalid privacy type. Must be 'everyone', 'followers', or 'close_friends'", nil)
			return
		}
	}

	// Set music info if provided
	if req.Music != nil && req.Music.AudioURL != "" {
		story.MusicInfo = &models.StoryMusic{
			Title:    req.Music.Title,
			Artist:   req.Music.Artist,
			AudioURL: req.Music.AudioURL,
			CoverURL: req.Music.CoverURL,
		}
	}

	// Set link info if provided
	if req.Link != nil && req.Link.URL != "" {
		story.LinkInfo = &models.StoryLink{
			URL:          req.Link.URL,
			Title:        req.Link.Title,
			Description:  req.Link.Description,
			ThumbnailURL: req.Link.ThumbnailURL,
		}
	}

	// Create the story
	createdStory, err := h.storyService.CreateStory(c.Request.Context(), story, mediaIDs)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to create story", err)
		return
	}

	// Return success response
	response.Created(c, "Story created successfully", createdStory)
}

// CreatePollStory handles the request to create a story with a poll
func (h *CreateHandler) CreatePollStory(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Parse request body
	var req struct {
		MediaIDs       []string `json:"media_ids" binding:"required"`
		Caption        string   `json:"caption,omitempty"`
		PollQuestion   string   `json:"poll_question" binding:"required"`
		PollOptions    []string `json:"poll_options" binding:"required"`
		ResultsVisible bool     `json:"results_visible"`
		Duration       int      `json:"duration,omitempty"` // in seconds
		PrivacyType    string   `json:"privacy_type,omitempty"`
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

	if len(req.MediaIDs) == 0 {
		response.ValidationError(c, "At least one media file is required", nil)
		return
	}

	if len(req.PollOptions) < 2 {
		response.ValidationError(c, "At least two poll options are required", nil)
		return
	}

	if len(req.PollOptions) > 4 {
		response.ValidationError(c, "Maximum of four poll options allowed", nil)
		return
	}

	// Convert media IDs to ObjectIDs
	mediaIDs := make([]primitive.ObjectID, 0, len(req.MediaIDs))
	for _, idStr := range req.MediaIDs {
		if !validation.IsValidObjectID(idStr) {
			response.ValidationError(c, "Invalid media ID: "+idStr, nil)
			return
		}
		mediaID, _ := primitive.ObjectIDFromHex(idStr)
		mediaIDs = append(mediaIDs, mediaID)
	}

	// Create the poll story
	createdStory, err := h.storyService.CreatePollStory(
		c.Request.Context(),
		userID.(primitive.ObjectID),
		mediaIDs,
		req.Caption,
		req.PollQuestion,
		req.PollOptions,
		req.ResultsVisible,
		req.Duration,
		req.PrivacyType,
		req.Location,
	)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to create poll story", err)
		return
	}

	// Return success response
	response.Created(c, "Poll story created successfully", createdStory)
}

// CreateQuestionStory handles the request to create a story with a question
func (h *CreateHandler) CreateQuestionStory(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Parse request body
	var req struct {
		MediaIDs        []string `json:"media_ids" binding:"required"`
		Caption         string   `json:"caption,omitempty"`
		Question        string   `json:"question" binding:"required"`
		BackgroundColor string   `json:"background_color,omitempty"`
		ResponsesPublic bool     `json:"responses_public"`
		Duration        int      `json:"duration,omitempty"` // in seconds
		PrivacyType     string   `json:"privacy_type,omitempty"`
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

	if len(req.MediaIDs) == 0 {
		response.ValidationError(c, "At least one media file is required", nil)
		return
	}

	// Convert media IDs to ObjectIDs
	mediaIDs := make([]primitive.ObjectID, 0, len(req.MediaIDs))
	for _, idStr := range req.MediaIDs {
		if !validation.IsValidObjectID(idStr) {
			response.ValidationError(c, "Invalid media ID: "+idStr, nil)
			return
		}
		mediaID, _ := primitive.ObjectIDFromHex(idStr)
		mediaIDs = append(mediaIDs, mediaID)
	}

	// Create the question story
	createdStory, err := h.storyService.CreateQuestionStory(
		c.Request.Context(),
		userID.(primitive.ObjectID),
		mediaIDs,
		req.Caption,
		req.Question,
		req.BackgroundColor,
		req.ResponsesPublic,
		req.Duration,
		req.PrivacyType,
		req.Location,
	)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to create question story", err)
		return
	}

	// Return success response
	response.Created(c, "Question story created successfully", createdStory)
}
