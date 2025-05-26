package notifications

import (
	"net/http"
	"time"

	"github.com/Caqil/vyrall/internal/services/notification"
	"github.com/Caqil/vyrall/internal/utils/response"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// MarkAllReadHandler handles operations to mark all notifications as read
type MarkAllReadHandler struct {
	notificationService *notification.Service
}

// NewMarkAllReadHandler creates a new mark all read handler
func NewMarkAllReadHandler(notificationService *notification.Service) *MarkAllReadHandler {
	return &MarkAllReadHandler{
		notificationService: notificationService,
	}
}

// MarkAllAsRead handles the request to mark all notifications as read
func (h *MarkAllReadHandler) MarkAllAsRead(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Get notification type filter (optional)
	notificationType := c.DefaultQuery("type", "") // Filter by notification type

	// Mark all as read
	count, err := h.notificationService.MarkAllAsRead(c.Request.Context(), userID.(primitive.ObjectID), notificationType)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to mark notifications as read", err)
		return
	}

	// Return success response
	response.OK(c, "All notifications marked as read", gin.H{
		"marked_count": count,
		"marked_at":    time.Now(),
	})
}

// MarkAllAsReadByDate handles the request to mark all notifications as read before a certain date
func (h *MarkAllReadHandler) MarkAllAsReadByDate(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Parse request body
	var req struct {
		BeforeDate string `json:"before_date" binding:"required"` // ISO format date
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "Invalid request body", err.Error())
		return
	}

	// Parse the date
	beforeDate, err := time.Parse(time.RFC3339, req.BeforeDate)
	if err != nil {
		response.ValidationError(c, "Invalid date format. Use ISO 8601 format (YYYY-MM-DDTHH:MM:SSZ)", err.Error())
		return
	}

	// Mark all as read before date
	count, err := h.notificationService.MarkAllAsReadBeforeDate(c.Request.Context(), userID.(primitive.ObjectID), beforeDate)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to mark notifications as read", err)
		return
	}

	// Return success response
	response.OK(c, "Notifications before "+req.BeforeDate+" marked as read", gin.H{
		"marked_count": count,
		"before_date":  req.BeforeDate,
		"marked_at":    time.Now(),
	})
}

// MarkAllAsReadBySubject handles the request to mark all notifications as read for a specific subject
func (h *MarkAllReadHandler) MarkAllAsReadBySubject(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Get subject parameters
	subject := c.Query("subject")
	subjectID := c.Query("subject_id")

	if subject == "" || subjectID == "" {
		response.ValidationError(c, "Both subject and subject_id are required", nil)
		return
	}

	// Validate subject ID
	if !primitive.IsValidObjectID(subjectID) {
		response.ValidationError(c, "Invalid subject ID", nil)
		return
	}
	subjectObjID, _ := primitive.ObjectIDFromHex(subjectID)

	// Mark all as read for subject
	count, err := h.notificationService.MarkAllAsReadBySubject(c.Request.Context(), userID.(primitive.ObjectID), subject, subjectObjID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to mark notifications as read", err)
		return
	}

	// Return success response
	response.OK(c, "All notifications for "+subject+" marked as read", gin.H{
		"marked_count": count,
		"subject":      subject,
		"subject_id":   subjectID,
		"marked_at":    time.Now(),
	})
}
