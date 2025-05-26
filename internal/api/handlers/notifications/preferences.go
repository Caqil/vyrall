package notifications

import (
	"net/http"

	"github.com/Caqil/vyrall/internal/models"
	"github.com/Caqil/vyrall/internal/services/notification"
	"github.com/Caqil/vyrall/internal/utils/response"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// PreferencesHandler handles notification preferences operations
type PreferencesHandler struct {
	notificationService *notification.Service
}

// NewPreferencesHandler creates a new preferences handler
func NewPreferencesHandler(notificationService *notification.Service) *PreferencesHandler {
	return &PreferencesHandler{
		notificationService: notificationService,
	}
}

// GetPreferences handles the request to get notification preferences
func (h *PreferencesHandler) GetPreferences(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Get notification preferences
	preferences, err := h.notificationService.GetNotificationPreferences(c.Request.Context(), userID.(primitive.ObjectID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get notification preferences", err)
		return
	}

	// Return success response
	response.OK(c, "Notification preferences retrieved successfully", preferences)
}

// UpdatePreferences handles the request to update notification preferences
func (h *PreferencesHandler) UpdatePreferences(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Parse request body
	var req struct {
		Push            *models.NotificationTypes `json:"push,omitempty"`
		Email           *models.NotificationTypes `json:"email,omitempty"`
		InApp           *models.NotificationTypes `json:"in_app,omitempty"`
		MessagePreviews *bool                     `json:"message_previews,omitempty"`
		DoNotDisturb    *struct {
			Enabled    bool     `json:"enabled"`
			StartTime  string   `json:"start_time"`
			EndTime    string   `json:"end_time"`
			Exceptions []string `json:"exceptions,omitempty"`
		} `json:"do_not_disturb,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "Invalid request body", err.Error())
		return
	}

	// Create updates map
	updates := make(map[string]interface{})

	// Add push preferences if provided
	if req.Push != nil {
		updates["notification_preferences.push"] = req.Push
	}

	// Add email preferences if provided
	if req.Email != nil {
		updates["notification_preferences.email"] = req.Email
	}

	// Add in-app preferences if provided
	if req.InApp != nil {
		updates["notification_preferences.in_app"] = req.InApp
	}

	// Add message previews setting if provided
	if req.MessagePreviews != nil {
		updates["notification_preferences.message_previews"] = *req.MessagePreviews
	}

	// Add do not disturb settings if provided
	if req.DoNotDisturb != nil {
		updates["notification_preferences.do_not_disturb"] = req.DoNotDisturb
	}

	if len(updates) == 0 {
		response.ValidationError(c, "No preferences provided", nil)
		return
	}

	// Update preferences
	err := h.notificationService.UpdateNotificationPreferences(c.Request.Context(), userID.(primitive.ObjectID), updates)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to update notification preferences", err)
		return
	}

	// Get updated preferences
	updatedPrefs, err := h.notificationService.GetNotificationPreferences(c.Request.Context(), userID.(primitive.ObjectID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get updated preferences", err)
		return
	}

	// Return success response
	response.OK(c, "Notification preferences updated successfully", updatedPrefs)
}

// ResetToDefault handles the request to reset notification preferences to default
func (h *PreferencesHandler) ResetToDefault(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Reset to default preferences
	err := h.notificationService.ResetNotificationPreferencesToDefault(c.Request.Context(), userID.(primitive.ObjectID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to reset notification preferences", err)
		return
	}

	// Get updated preferences
	updatedPrefs, err := h.notificationService.GetNotificationPreferences(c.Request.Context(), userID.(primitive.ObjectID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get updated preferences", err)
		return
	}

	// Return success response
	response.OK(c, "Notification preferences reset to default", updatedPrefs)
}

// UpdateChannelPreferences handles the request to update preferences for a specific channel
func (h *PreferencesHandler) UpdateChannelPreferences(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Get channel from URL parameter
	channel := c.Param("channel")
	if channel != "push" && channel != "email" && channel != "in_app" {
		response.ValidationError(c, "Invalid channel. Must be 'push', 'email', or 'in_app'", nil)
		return
	}

	// Parse request body
	var req models.NotificationTypes
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "Invalid request body", err.Error())
		return
	}

	// Update channel preferences
	err := h.notificationService.UpdateChannelPreferences(c.Request.Context(), userID.(primitive.ObjectID), channel, req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to update "+channel+" preferences", err)
		return
	}

	// Get updated preferences
	updatedPrefs, err := h.notificationService.GetNotificationPreferences(c.Request.Context(), userID.(primitive.ObjectID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get updated preferences", err)
		return
	}

	// Return success response
	response.OK(c, channel+" preferences updated successfully", updatedPrefs)
}
