package notifications

import (
	"net/http"

	"github.com/Caqil/vyrall/internal/services/notification"
	"github.com/Caqil/vyrall/internal/utils/response"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// SettingsHandler handles notification settings operations
type SettingsHandler struct {
	notificationService *notification.Service
}

// NewSettingsHandler creates a new settings handler
func NewSettingsHandler(notificationService *notification.Service) *SettingsHandler {
	return &SettingsHandler{
		notificationService: notificationService,
	}
}

// GetSettings handles the request to get all notification settings
func (h *SettingsHandler) GetSettings(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Get notification settings
	settings, err := h.notificationService.GetAllNotificationSettings(c.Request.Context(), userID.(primitive.ObjectID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get notification settings", err)
		return
	}

	// Return success response
	response.OK(c, "Notification settings retrieved successfully", settings)
}

// UpdateDoNotDisturb handles the request to update "do not disturb" settings
func (h *SettingsHandler) UpdateDoNotDisturb(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Parse request body
	var req struct {
		Enabled    bool     `json:"enabled"`
		StartTime  string   `json:"start_time,omitempty"` // Format: HH:MM
		EndTime    string   `json:"end_time,omitempty"`   // Format: HH:MM
		Exceptions []string `json:"exceptions,omitempty"` // Types that can bypass DND
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "Invalid request body", err.Error())
		return
	}

	// Validate time formats if provided
	if req.Enabled {
		if req.StartTime == "" || req.EndTime == "" {
			response.ValidationError(c, "Start time and end time are required when enabling Do Not Disturb", nil)
			return
		}
	}

	// Update DND settings
	err := h.notificationService.UpdateDoNotDisturbSettings(
		c.Request.Context(),
		userID.(primitive.ObjectID),
		req.Enabled,
		req.StartTime,
		req.EndTime,
		req.Exceptions,
	)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to update Do Not Disturb settings", err)
		return
	}

	// Get updated settings
	settings, err := h.notificationService.GetDoNotDisturbSettings(c.Request.Context(), userID.(primitive.ObjectID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get updated settings", err)
		return
	}

	// Return success response
	response.OK(c, "Do Not Disturb settings updated successfully", settings)
}

// ToggleAllNotifications handles the request to enable or disable all notifications
func (h *SettingsHandler) ToggleAllNotifications(c *gin.Context) {
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

	// Toggle all notifications
	err := h.notificationService.ToggleAllNotifications(c.Request.Context(), userID.(primitive.ObjectID), req.Enabled)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to update notification settings", err)
		return
	}

	// Return success response
	status := "enabled"
	if !req.Enabled {
		status = "disabled"
	}

	response.OK(c, "All notifications "+status+" successfully", gin.H{
		"notifications_enabled": req.Enabled,
	})
}

// UpdateMessagePreviewSettings handles the request to update message preview settings
func (h *SettingsHandler) UpdateMessagePreviewSettings(c *gin.Context) {
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

	// Update message preview settings
	err := h.notificationService.UpdateMessagePreviewSettings(c.Request.Context(), userID.(primitive.ObjectID), req.Enabled)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to update message preview settings", err)
		return
	}

	// Return success response
	status := "enabled"
	if !req.Enabled {
		status = "disabled"
	}

	response.OK(c, "Message previews "+status+" successfully", gin.H{
		"message_previews_enabled": req.Enabled,
	})
}

// ExportSettings handles the request to export notification settings
func (h *SettingsHandler) ExportSettings(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Get format from query parameter
	format := c.DefaultQuery("format", "json")
	if format != "json" && format != "yaml" {
		response.ValidationError(c, "Invalid format. Supported formats: json, yaml", nil)
		return
	}

	// Export settings
	settings, contentType, err := h.notificationService.ExportNotificationSettings(c.Request.Context(), userID.(primitive.ObjectID), format)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to export notification settings", err)
		return
	}

	// Set headers for file download
	fileName := "notification_settings." + format
	c.Header("Content-Disposition", "attachment; filename="+fileName)
	c.Header("Content-Type", contentType)

	// Return the file content
	c.String(http.StatusOK, settings)
}

// ImportSettings handles the request to import notification settings
func (h *SettingsHandler) ImportSettings(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Get format from form field
	format := c.DefaultPostForm("format", "json")
	if format != "json" && format != "yaml" {
		response.ValidationError(c, "Invalid format. Supported formats: json, yaml", nil)
		return
	}

	// Get settings file
	file, err := c.FormFile("settings_file")
	if err != nil {
		response.ValidationError(c, "No settings file provided", err.Error())
		return
	}

	// Open file
	f, err := file.Open()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to open settings file", err)
		return
	}
	defer f.Close()

	// Import settings
	err = h.notificationService.ImportNotificationSettings(c.Request.Context(), userID.(primitive.ObjectID), f, format)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to import notification settings", err)
		return
	}

	// Get updated settings
	settings, err := h.notificationService.GetAllNotificationSettings(c.Request.Context(), userID.(primitive.ObjectID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get updated settings", err)
		return
	}

	// Return success response
	response.OK(c, "Notification settings imported successfully", settings)
}
