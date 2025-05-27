package websocket

import (
	"encoding/json"
	"log"
	"time"

	"github.com/Caqil/vyrall/internal/models"
	"github.com/Caqil/vyrall/internal/services/notification"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// NotificationsHandler manages notification websocket events
type NotificationsHandler struct {
	hub                 *Hub
	notificationService *notification.Service
}

// NewNotificationsHandler creates a new notifications handler
func NewNotificationsHandler(hub *Hub, notificationService *notification.Service) *NotificationsHandler {
	return &NotificationsHandler{
		hub:                 hub,
		notificationService: notificationService,
	}
}

// ProcessNotificationRead handles marking notifications as read
func (h *NotificationsHandler) ProcessNotificationRead(client *Client, data []byte) {
	var payload struct {
		NotificationID string `json:"notification_id"`
	}

	if err := json.Unmarshal(data, &payload); err != nil {
		log.Printf("Error unmarshaling notification read: %v", err)
		client.SendErrorMessage("Invalid notification read format")
		return
	}

	// Check if it's a specific notification or all notifications
	if payload.NotificationID == "all" {
		// Mark all notifications as read
		err := h.notificationService.MarkAllAsRead(client.ctx, client.UserID)
		if err != nil {
			log.Printf("Error marking all notifications as read: %v", err)
			client.SendErrorMessage("Failed to mark all notifications as read")
			return
		}
	} else {
		// Mark specific notification as read
		notificationID, err := primitive.ObjectIDFromHex(payload.NotificationID)
		if err != nil {
			client.SendErrorMessage("Invalid notification ID")
			return
		}

		err = h.notificationService.MarkAsRead(client.ctx, notificationID, client.UserID)
		if err != nil {
			log.Printf("Error marking notification as read: %v", err)
			client.SendErrorMessage("Failed to mark notification as read")
			return
		}
	}

	// Send updated unread count
	h.SendUnreadNotificationsCount(client)
}

// SendUnreadNotificationsCount sends the unread notifications count to a client
func (h *NotificationsHandler) SendUnreadNotificationsCount(client *Client) {
	// Get unread count
	count, err := h.notificationService.GetUnreadCount(client.ctx, client.UserID)
	if err != nil {
		log.Printf("Error getting unread notifications count: %v", err)
		return
	}

	// Send count to client
	payload, err := json.Marshal(map[string]interface{}{
		"type":      "unread_notifications_count",
		"count":     count,
		"timestamp": time.Now(),
	})
	if err != nil {
		log.Printf("Error marshaling unread count: %v", err)
		return
	}

	client.SendMessage(payload)
}

// SendNotification sends a notification to a user
func (h *NotificationsHandler) SendNotification(notification *models.Notification) {
	// Mark the notification as sent
	notification.IsSent = true
	notification.SentAt = new(time.Time)
	*notification.SentAt = time.Now()

	err := h.notificationService.UpdateNotification(nil, notification)
	if err != nil {
		log.Printf("Error updating notification: %v", err)
	}

	// Create notification payload
	payload, err := json.Marshal(map[string]interface{}{
		"type":            "notification",
		"notification_id": notification.ID.Hex(),
		"user_id":         notification.UserID.Hex(),
		"actor":           notification.Actor.Hex(),
		"type":            notification.Type,
		"subject":         notification.Subject,
		"subject_id":      notification.SubjectID.Hex(),
		"message":         notification.Message,
		"image_url":       notification.ImageURL,
		"action_url":      notification.ActionURL,
		"created_at":      notification.CreatedAt,
	})
	if err != nil {
		log.Printf("Error marshaling notification: %v", err)
		return
	}

	// Send to user
	h.hub.SendToUser(notification.UserID, payload)
}

// BroadcastNotification sends a notification to multiple users
func (h *NotificationsHandler) BroadcastNotification(userIDs []primitive.ObjectID, notificationType string, actor primitive.ObjectID, subject string, subjectID primitive.ObjectID, message string) {
	for _, userID := range userIDs {
		// Create notification object
		notification := &models.Notification{
			UserID:    userID,
			Type:      notificationType,
			Actor:     actor,
			Subject:   subject,
			SubjectID: subjectID,
			Message:   message,
			IsRead:    false,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		// Save notification to database
		createdNotification, err := h.notificationService.CreateNotification(nil, notification)
		if err != nil {
			log.Printf("Error creating notification: %v", err)
			continue
		}

		// Send notification
		h.SendNotification(createdNotification)
	}
}

// SendSystemNotification sends a system notification to a user
func (h *NotificationsHandler) SendSystemNotification(userID primitive.ObjectID, message string, actionURL string) {
	// Create notification object
	notification := &models.Notification{
		UserID:    userID,
		Type:      "system",
		Subject:   "system",
		Message:   message,
		ActionURL: actionURL,
		Priority:  "normal",
		IsRead:    false,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Save notification to database
	createdNotification, err := h.notificationService.CreateNotification(nil, notification)
	if err != nil {
		log.Printf("Error creating system notification: %v", err)
		return
	}

	// Send notification
	h.SendNotification(createdNotification)
}
