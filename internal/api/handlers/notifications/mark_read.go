package notifications

import (
	"net/http"
	"time"

	"github.com/Caqil/vyrall/internal/services/notification"
	"github.com/Caqil/vyrall/internal/utils/response"
	"github.com/Caqil/vyrall/internal/utils/validation"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// MarkReadHandler handles operations to mark individual notifications as read
type MarkReadHandler struct {
	notificationService *notification.Service
}

// NewMarkReadHandler creates a new mark read handler
func NewMarkReadHandler(notificationService *notification.Service) *MarkReadHandler {
	return &MarkReadHandler{
		notificationService: notificationService,
	}
}

// MarkAsRead handles the request to mark a notification as read
func (h *MarkReadHandler) MarkAsRead(c *gin.Context) {
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

	// Mark as read
	err := h.notificationService.MarkAsRead(c.Request.Context(), notificationID, userID.(primitive.ObjectID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to mark notification as read", err)
		return
	}

	// Return success response
	response.OK(c, "Notification marked as read", gin.H{
		"notification_id": notificationID.Hex(),
		"read_at":         time.Now(),
	})
}

// MarkAsUnread handles the request to mark a notification as unread
func (h *MarkReadHandler) MarkAsUnread(c *gin.Context) {
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

	// Mark as unread
	err := h.notificationService.MarkAsUnread(c.Request.Context(), notificationID, userID.(primitive.ObjectID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to mark notification as unread", err)
		return
	}

	// Return success response
	response.OK(c, "Notification marked as unread", gin.H{
		"notification_id": notificationID.Hex(),
	})
}

// BulkMarkAsRead handles the request to mark multiple notifications as read
func (h *MarkReadHandler) BulkMarkAsRead(c *gin.Context) {
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

	// Mark notifications as read
	count, err := h.notificationService.BulkMarkAsRead(c.Request.Context(), notificationIDs, userID.(primitive.ObjectID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to mark notifications as read", err)
		return
	}

	// Return success response
	response.OK(c, "Notifications marked as read", gin.H{
		"marked_count": count,
		"read_at":      time.Now(),
	})
}

// ToggleReadStatus handles the request to toggle a notification's read status
func (h *MarkReadHandler) ToggleReadStatus(c *gin.Context) {
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

	// Toggle read status
	isRead, err := h.notificationService.ToggleReadStatus(c.Request.Context(), notificationID, userID.(primitive.ObjectID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to toggle notification read status", err)
		return
	}

	// Determine status message
	status := "unread"
	if isRead {
		status = "read"
	}

	// Return success response
	response.OK(c, "Notification marked as "+status, gin.H{
		"notification_id": notificationID.Hex(),
		"is_read":         isRead,
		"updated_at":      time.Now(),
	})
}
