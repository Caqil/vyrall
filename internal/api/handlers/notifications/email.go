package notifications

import (
	"net/http"

	"github.com/Caqil/vyrall/internal/services/notification"
	"github.com/Caqil/vyrall/internal/utils/response"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// EmailHandler handles email notification operations
type EmailHandler struct {
	notificationService *notification.Service
}

// NewEmailHandler creates a new email handler
func NewEmailHandler(notificationService *notification.Service) *EmailHandler {
	return &EmailHandler{
		notificationService: notificationService,
	}
}

// UpdateEmailPreferences handles the request to update email notification preferences
func (h *EmailHandler) UpdateEmailPreferences(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Parse request body
	var req struct {
		Likes          *bool `json:"likes,omitempty"`
		Comments       *bool `json:"comments,omitempty"`
		Follows        *bool `json:"follows,omitempty"`
		Messages       *bool `json:"messages,omitempty"`
		TaggedPosts    *bool `json:"tagged_posts,omitempty"`
		Stories        *bool `json:"stories,omitempty"`
		LiveStreams    *bool `json:"live_streams,omitempty"`
		GroupActivity  *bool `json:"group_activity,omitempty"`
		EventReminders *bool `json:"event_reminders,omitempty"`
		SystemUpdates  *bool `json:"system_updates,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "Invalid request body", err.Error())
		return
	}

	// Create preferences map
	preferences := make(map[string]bool)

	if req.Likes != nil {
		preferences["likes"] = *req.Likes
	}
	if req.Comments != nil {
		preferences["comments"] = *req.Comments
	}
	if req.Follows != nil {
		preferences["follows"] = *req.Follows
	}
	if req.Messages != nil {
		preferences["messages"] = *req.Messages
	}
	if req.TaggedPosts != nil {
		preferences["tagged_posts"] = *req.TaggedPosts
	}
	if req.Stories != nil {
		preferences["stories"] = *req.Stories
	}
	if req.LiveStreams != nil {
		preferences["live_streams"] = *req.LiveStreams
	}
	if req.GroupActivity != nil {
		preferences["group_activity"] = *req.GroupActivity
	}
	if req.EventReminders != nil {
		preferences["event_reminders"] = *req.EventReminders
	}
	if req.SystemUpdates != nil {
		preferences["system_updates"] = *req.SystemUpdates
	}

	if len(preferences) == 0 {
		response.ValidationError(c, "No preferences provided", nil)
		return
	}

	// Update email preferences
	err := h.notificationService.UpdateEmailPreferences(c.Request.Context(), userID.(primitive.ObjectID), preferences)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to update email preferences", err)
		return
	}

	// Get updated preferences
	updatedPrefs, err := h.notificationService.GetNotificationPreferences(c.Request.Context(), userID.(primitive.ObjectID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get updated preferences", err)
		return
	}

	// Return success response
	response.OK(c, "Email preferences updated successfully", updatedPrefs)
}

// GetEmailPreferences handles the request to get email notification preferences
func (h *EmailHandler) GetEmailPreferences(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Get email preferences
	preferences, err := h.notificationService.GetEmailPreferences(c.Request.Context(), userID.(primitive.ObjectID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get email preferences", err)
		return
	}

	// Return success response
	response.OK(c, "Email preferences retrieved successfully", preferences)
}

// ToggleAllEmailNotifications handles the request to enable or disable all email notifications
func (h *EmailHandler) ToggleAllEmailNotifications(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Parse request body
	var req struct {
		Enabled bool `json:"enabled" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "Invalid request body", err.Error())
		return
	}

	// Toggle all email notifications
	err := h.notificationService.ToggleAllEmailNotifications(c.Request.Context(), userID.(primitive.ObjectID), req.Enabled)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to update email notification settings", err)
		return
	}

	// Return success response
	status := "enabled"
	if !req.Enabled {
		status = "disabled"
	}

	response.OK(c, "Email notifications "+status+" successfully", gin.H{
		"email_notifications_enabled": req.Enabled,
	})
}

// SendTestEmail handles the request to send a test email notification
func (h *EmailHandler) SendTestEmail(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Send test email
	err := h.notificationService.SendTestEmail(c.Request.Context(), userID.(primitive.ObjectID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to send test email", err)
		return
	}

	// Return success response
	response.OK(c, "Test email sent successfully", nil)
}
