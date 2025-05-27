package users

import (
	"net/http"

	"github.com/Caqil/vyrall/internal/services/user"
	"github.com/Caqil/vyrall/internal/utils/response"
	"github.com/Caqil/vyrall/internal/utils/validation"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ProfileHandler handles user profile operations
type ProfileHandler struct {
	userService *user.Service
}

// NewProfileHandler creates a new profile handler
func NewProfileHandler(userService *user.Service) *ProfileHandler {
	return &ProfileHandler{
		userService: userService,
	}
}

// GetProfile handles the request to get a user's profile
func (h *ProfileHandler) GetProfile(c *gin.Context) {
	// Get target user ID or username from URL parameter
	userIdentifier := c.Param("identifier")
	if userIdentifier == "" {
		response.ValidationError(c, "User identifier is required", nil)
		return
	}

	// Get authenticated user ID (may be nil for unauthenticated users)
	var userID primitive.ObjectID
	if id, exists := c.Get("userID"); exists {
		userID = id.(primitive.ObjectID)
	}

	var profile interface{}
	var err error

	// Check if identifier is an ObjectID or username
	if validation.IsValidObjectID(userIdentifier) {
		targetID, _ := primitive.ObjectIDFromHex(userIdentifier)
		profile, err = h.userService.GetProfileByID(c.Request.Context(), targetID, userID)
	} else {
		profile, err = h.userService.GetProfileByUsername(c.Request.Context(), userIdentifier, userID)
	}

	if err != nil {
		response.NotFoundError(c, "User not found")
		return
	}

	// Return success response
	response.OK(c, "Profile retrieved successfully", profile)
}

// GetMyProfile handles the request to get the authenticated user's profile
func (h *ProfileHandler) GetMyProfile(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Get own profile
	profile, err := h.userService.GetMyProfile(c.Request.Context(), userID.(primitive.ObjectID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve profile", err)
		return
	}

	// Return success response
	response.OK(c, "Profile retrieved successfully", profile)
}

// UpdateProfile handles the request to update a user's profile
func (h *ProfileHandler) UpdateProfile(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Parse request body
	var req struct {
		DisplayName    string `json:"display_name,omitempty"`
		Bio            string `json:"bio,omitempty"`
		Website        string `json:"website,omitempty"`
		Location       string `json:"location,omitempty"`
		ProfilePicture string `json:"profile_picture,omitempty"`
		CoverPhoto     string `json:"cover_photo,omitempty"`
		IsPrivate      *bool  `json:"is_private,omitempty"`
		ShowActivity   *bool  `json:"show_activity,omitempty"`
		AllowTags      *bool  `json:"allow_tags,omitempty"`
		Gender         string `json:"gender,omitempty"`
		BirthDate      string `json:"birth_date,omitempty"`
		Phone          string `json:"phone,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "Invalid request body", err.Error())
		return
	}

	// Create updates map
	updates := make(map[string]interface{})

	if req.DisplayName != "" {
		updates["display_name"] = req.DisplayName
	}
	if req.Bio != "" {
		updates["bio"] = req.Bio
	}
	if req.Website != "" {
		updates["website"] = req.Website
	}
	if req.Location != "" {
		updates["location"] = req.Location
	}
	if req.ProfilePicture != "" {
		updates["profile_picture"] = req.ProfilePicture
	}
	if req.CoverPhoto != "" {
		updates["cover_photo"] = req.CoverPhoto
	}
	if req.IsPrivate != nil {
		updates["is_private"] = *req.IsPrivate
	}
	if req.ShowActivity != nil {
		updates["settings.show_activity"] = *req.ShowActivity
	}
	if req.AllowTags != nil {
		updates["settings.allow_tags"] = *req.AllowTags
	}
	if req.Gender != "" {
		updates["gender"] = req.Gender
	}
	if req.BirthDate != "" {
		updates["date_of_birth"] = req.BirthDate
	}
	if req.Phone != "" {
		updates["phone_number"] = req.Phone
	}

	if len(updates) == 0 {
		response.ValidationError(c, "No updates provided", nil)
		return
	}

	// Update the profile
	updatedProfile, err := h.userService.UpdateProfile(c.Request.Context(), userID.(primitive.ObjectID), updates)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to update profile", err)
		return
	}

	// Return success response
	response.OK(c, "Profile updated successfully", updatedProfile)
}

// ChangeUsername handles the request to change a user's username
func (h *ProfileHandler) ChangeUsername(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Parse request body
	var req struct {
		Username string `json:"username" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "Invalid request body", err.Error())
		return
	}

	if req.Username == "" {
		response.ValidationError(c, "Username cannot be empty", nil)
		return
	}

	// Change username
	err := h.userService.ChangeUsername(c.Request.Context(), userID.(primitive.ObjectID), req.Username)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to change username", err)
		return
	}

	// Return success response
	response.OK(c, "Username changed successfully", gin.H{
		"username": req.Username,
	})
}

// ViewProfile handles the request to record a profile view
func (h *ProfileHandler) ViewProfile(c *gin.Context) {
	// Get user ID from context (may be nil for unauthenticated users)
	var viewerID primitive.ObjectID
	if id, exists := c.Get("userID"); exists {
		viewerID = id.(primitive.ObjectID)
	}

	// Get target user ID from URL parameter
	targetIDStr := c.Param("id")
	if !validation.IsValidObjectID(targetIDStr) {
		response.ValidationError(c, "Invalid user ID", nil)
		return
	}
	targetID, _ := primitive.ObjectIDFromHex(targetIDStr)

	// Skip if viewing own profile or not authenticated
	if viewerID == targetID || viewerID.IsZero() {
		response.OK(c, "Profile view recorded", nil)
		return
	}

	// Record profile view
	err := h.userService.ViewProfile(c.Request.Context(), viewerID, targetID)
	if err != nil {
		// Just log the error and return success anyway to not disrupt the user experience
		// The client doesn't need to know if recording the view failed
	}

	// Return success response
	response.OK(c, "Profile view recorded", nil)
}
