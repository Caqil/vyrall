package notifications

import (
	"net/http"

	"github.com/Caqil/vyrall/internal/services/notification"
	"github.com/Caqil/vyrall/internal/utils/response"
	"github.com/Caqil/vyrall/internal/utils/validation"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// DeleteHandler handles notification deletion operations
type DeleteHandler struct {
	notificationService *notification.Service
}

// NewDeleteHandler creates a new delete handler
func NewDeleteHandler(notificationService *notification.Service) *DeleteHandler {
	return &DeleteHandler{
		notificationService: notificationService,
	}
}

// DeleteNotification handles the request to delete a notification
func (h *DeleteHandler) DeleteNotification(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Get notification ID from URL parameter
	notificationIDStr := c.Param("id")
	if !validation.IsValidObjectID(notificationIDStr) {
		response.ValidationError(c, "Invalid notification ID", nil)
		return
	}
	notificationID, _ := primitive.ObjectIDFromHex(notificationIDStr)

	// Delete the notification
	err := h.notificationService.DeleteNotification(c.Request.Context(), notificationID, userID.(primitive.ObjectID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to delete notification", err)
		return
	}

	// Return success response
	response.OK(c, "Notification deleted successfully", nil)
}

// DeleteAllNotifications handles the request to delete all notifications
func (h *DeleteHandler) DeleteAllNotifications(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Get notification type filter (optional)
	notificationType := c.DefaultQuery("type", "") // Filter by notification type

	// Delete all notifications
	count, err := h.notificationService.DeleteAllNotifications(c.Request.Context(), userID.(primitive.ObjectID), notificationType)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to delete notifications", err)
		return
	}

	// Return success response
	response.OK(c, "Notifications deleted successfully", gin.H{
		"deleted_count": count,
	})
}

// BulkDeleteNotifications handles the request to delete multiple notifications
func (h *DeleteHandler) BulkDeleteNotifications(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Parse request body
	var req struct {
		NotificationIDs []string `json:"notification_ids" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "Invalid request body", err.Error())
		return
	}

	if len(req.NotificationIDs) == 0 {
		response.ValidationError(c, "No notification IDs provided", nil)
		return
	}

	// Convert notification IDs to ObjectIDs
	notificationIDs := make([]primitive.ObjectID, 0, len(req.NotificationIDs))
	for _, idStr := range req.NotificationIDs {
		if !validation.IsValidObjectID(idStr) {
			continue // Skip invalid IDs
		}
		notificationID, _ := primitive.ObjectIDFromHex(idStr)
		notificationIDs = append(notificationIDs, notificationID)
	}

	if len(notificationIDs) == 0 {
		response.ValidationError(c, "No valid notification IDs provided", nil)
		return
	}

	// Delete the notifications
	count, err := h.notificationService.BulkDeleteNotifications(c.Request.Context(), notificationIDs, userID.(primitive.ObjectID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to delete notifications", err)
		return
	}

	// Return success response
	response.OK(c, "Notifications deleted successfully", gin.H{
		"deleted_count": count,
	})
}
