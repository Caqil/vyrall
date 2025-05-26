package notifications

import (
	"net/http"

	"github.com/Caqil/vyrall/internal/services/notification"
	"github.com/Caqil/vyrall/internal/utils/response"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ListHandler handles notification listing operations
type ListHandler struct {
	notificationService *notification.Service
}

// NewListHandler creates a new list handler
func NewListHandler(notificationService *notification.Service) *ListHandler {
	return &ListHandler{
		notificationService: notificationService,
	}
}

// GetNotifications handles the request to get user notifications
func (h *ListHandler) GetNotifications(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Get query parameters
	readStatus := c.DefaultQuery("read_status", "") // all, read, unread
	notificationType := c.DefaultQuery("type", "")  // Filter by notification type
	priority := c.DefaultQuery("priority", "")      // high, normal, low

	// Get pagination parameters
	limit, offset := response.GetPaginationParams(c)

	// Get notifications
	notifications, total, err := h.notificationService.GetNotifications(
		c.Request.Context(),
		userID.(primitive.ObjectID),
		readStatus,
		notificationType,
		priority,
		limit,
		offset,
	)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get notifications", err)
		return
	}

	// Return paginated response
	response.SuccessWithPagination(c, http.StatusOK, "Notifications retrieved successfully", notifications, limit, offset, total)
}

// GetGroupedNotifications handles the request to get notifications grouped by type
func (h *ListHandler) GetGroupedNotifications(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Get pagination parameters
	limit, offset := response.GetPaginationParams(c)

	// Get grouped notifications
	grouped, total, err := h.notificationService.GetGroupedNotifications(
		c.Request.Context(),
		userID.(primitive.ObjectID),
		limit,
		offset,
	)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get grouped notifications", err)
		return
	}

	// Return paginated response
	response.SuccessWithPagination(c, http.StatusOK, "Grouped notifications retrieved successfully", grouped, limit, offset, total)
}

// GetNotificationsByType handles the request to get notifications by type
func (h *ListHandler) GetNotificationsByType(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Get notification type from URL parameter
	notificationType := c.Param("type")
	if notificationType == "" {
		response.ValidationError(c, "Notification type is required", nil)
		return
	}

	// Get pagination parameters
	limit, offset := response.GetPaginationParams(c)

	// Get notifications by type
	notifications, total, err := h.notificationService.GetNotificationsByType(
		c.Request.Context(),
		userID.(primitive.ObjectID),
		notificationType,
		limit,
		offset,
	)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get notifications", err)
		return
	}

	// Return paginated response
	response.SuccessWithPagination(c, http.StatusOK, "Notifications retrieved successfully", notifications, limit, offset, total)
}

// SearchNotifications handles the request to search notifications
func (h *ListHandler) SearchNotifications(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Get search query
	query := c.Query("q")
	if query == "" {
		response.ValidationError(c, "Search query is required", nil)
		return
	}

	// Get pagination parameters
	limit, offset := response.GetPaginationParams(c)

	// Search notifications
	notifications, total, err := h.notificationService.SearchNotifications(
		c.Request.Context(),
		userID.(primitive.ObjectID),
		query,
		limit,
		offset,
	)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to search notifications", err)
		return
	}

	// Return paginated response
	response.SuccessWithPagination(c, http.StatusOK, "Notifications searched successfully", notifications, limit, offset, total)
}
