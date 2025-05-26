package notifications

import (
	"net/http"

	"github.com/Caqil/vyrall/internal/services/notification"
	"github.com/Caqil/vyrall/internal/utils/response"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// CountHandler handles notification count operations
type CountHandler struct {
	notificationService *notification.Service
}

// NewCountHandler creates a new count handler
func NewCountHandler(notificationService *notification.Service) *CountHandler {
	return &CountHandler{
		notificationService: notificationService,
	}
}

// GetUnreadCount handles the request to get the count of unread notifications
func (h *CountHandler) GetUnreadCount(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Get notification type filter (optional)
	notificationType := c.DefaultQuery("type", "") // Filter by notification type

	// Get unread count
	count, err := h.notificationService.GetUnreadCount(c.Request.Context(), userID.(primitive.ObjectID), notificationType)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get unread notification count", err)
		return
	}

	// Return success response
	response.OK(c, "Unread notification count retrieved successfully", gin.H{
		"unread_count": count,
	})
}

// GetNotificationCounts handles the request to get notification counts by type
func (h *CountHandler) GetNotificationCounts(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Get counts by type
	counts, err := h.notificationService.GetNotificationCountsByType(c.Request.Context(), userID.(primitive.ObjectID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get notification counts", err)
		return
	}

	// Return success response
	response.OK(c, "Notification counts retrieved successfully", counts)
}

// GetTotalCount handles the request to get the total count of notifications
func (h *CountHandler) GetTotalCount(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Get time range filter (optional)
	timeRange := c.DefaultQuery("time_range", "all") // all, today, week, month

	// Get total count
	count, err := h.notificationService.GetTotalCount(c.Request.Context(), userID.(primitive.ObjectID), timeRange)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get total notification count", err)
		return
	}

	// Return success response
	response.OK(c, "Total notification count retrieved successfully", gin.H{
		"total_count": count,
		"time_range":  timeRange,
	})
}
