package notifications

import (
	"net/http"

	"github.com/Caqil/vyrall/internal/services/notification"
	"github.com/Caqil/vyrall/internal/utils/response"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// PushHandler handles push notification operations
type PushHandler struct {
	notificationService *notification.Service
}

// NewPushHandler creates a new push handler
func NewPushHandler(notificationService *notification.Service) *PushHandler {
	return &PushHandler{
		notificationService: notificationService,
	}
}

// RegisterDevice handles the request to register a device for push notifications
func (h *PushHandler) RegisterDevice(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Parse request body
	var req struct {
		DeviceToken string `json:"device_token" binding:"required"`
		DeviceType  string `json:"device_type" binding:"required"` // ios, android, web
		DeviceName  string `json:"device_name,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "Invalid request body", err.Error())
		return
	}

	// Validate device type
	if req.DeviceType != "ios" && req.DeviceType != "android" && req.DeviceType != "web" {
		response.ValidationError(c, "Invalid device type. Must be 'ios', 'android', or 'web'", nil)
		return
	}

	// Register device
	deviceID, err := h.notificationService.RegisterDevice(
		c.Request.Context(),
		userID.(primitive.ObjectID),
		req.DeviceToken,
		req.DeviceType,
		req.DeviceName,
	)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to register device", err)
		return
	}

	// Return success response
	response.Created(c, "Device registered successfully", gin.H{
		"device_id": deviceID,
	})
}

// UnregisterDevice handles the request to unregister a device from push notifications
func (h *PushHandler) UnregisterDevice(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Parse request body
	var req struct {
		DeviceToken string `json:"device_token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "Invalid request body", err.Error())
		return
	}

	// Unregister device
	err := h.notificationService.UnregisterDevice(c.Request.Context(), userID.(primitive.ObjectID), req.DeviceToken)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to unregister device", err)
		return
	}

	// Return success response
	response.OK(c, "Device unregistered successfully", nil)
}

// UpdatePushPreferences handles the request to update push notification preferences
func (h *PushHandler) UpdatePushPreferences(c *gin.Context) {
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

	// Update push preferences
	err := h.notificationService.UpdatePushPreferences(c.Request.Context(), userID.(primitive.ObjectID), preferences)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to update push preferences", err)
		return
	}

	// Get updated preferences
	updatedPrefs, err := h.notificationService.GetPushPreferences(c.Request.Context(), userID.(primitive.ObjectID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get updated preferences", err)
		return
	}

	// Return success response
	response.OK(c, "Push preferences updated successfully", updatedPrefs)
}

// SendTestPush handles the request to send a test push notification
func (h *PushHandler) SendTestPush(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Send test push notification
	err := h.notificationService.SendTestPushNotification(c.Request.Context(), userID.(primitive.ObjectID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to send test push notification", err)
		return
	}

	// Return success response
	response.OK(c, "Test push notification sent successfully", nil)
}

// GetDevices handles the request to get registered devices for a user
func (h *PushHandler) GetDevices(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Get registered devices
	devices, err := h.notificationService.GetRegisteredDevices(c.Request.Context(), userID.(primitive.ObjectID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get registered devices", err)
		return
	}

	// Return success response
	response.OK(c, "Registered devices retrieved successfully", devices)
}
