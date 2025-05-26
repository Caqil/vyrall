package events

import (
	"net/http"
	"strconv"
	"time"

	"github.com/Caqil/vyrall/internal/models"
	"github.com/Caqil/vyrall/internal/utils/response"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ReminderRequest represents the request payload for setting an event reminder
type ReminderRequest struct {
	ReminderTime time.Time `json:"reminder_time" binding:"required"`
	Type         string    `json:"type" binding:"required,oneof=email push sms"`
}

// SetReminder adds a reminder for an event
func SetReminder(c *gin.Context) {
	// Get event ID from URL parameter
	eventIDStr := c.Param("id")
	eventID, err := primitive.ObjectIDFromHex(eventIDStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid event ID format", err)
		return
	}

	// Get the authenticated user's ID
	userID, exists := c.Get("userID")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	// Parse and validate the request
	var req ReminderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request format", err)
		return
	}

	// Validate reminder time
	if req.ReminderTime.Before(time.Now()) {
		response.Error(c, http.StatusBadRequest, "Reminder time cannot be in the past", nil)
		return
	}

	// Get the event service
	eventService := c.MustGet("eventService").(EventService)

	// Check if the event exists
	event, err := eventService.GetEventByID(c.Request.Context(), eventID)
	if err != nil {
		response.Error(c, http.StatusNotFound, "Event not found", err)
		return
	}

	// Ensure reminder is before the event
	if req.ReminderTime.After(event.StartTime) {
		response.Error(c, http.StatusBadRequest, "Reminder time must be before event start time", nil)
		return
	}

	// Create the reminder
	reminder := &models.EventReminder{
		EventID:      eventID,
		UserID:       userID.(primitive.ObjectID),
		ReminderTime: req.ReminderTime,
		Type:         req.Type,
		Status:       "pending",
		CreatedAt:    time.Now(),
	}

	reminderID, err := eventService.CreateReminder(c.Request.Context(), reminder)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to set reminder", err)
		return
	}

	response.Success(c, http.StatusOK, "Reminder set successfully", gin.H{
		"reminder_id": reminderID.Hex(),
	})
}

// DeleteReminder removes a reminder for an event
func DeleteReminder(c *gin.Context) {
	// Get event ID and reminder ID from URL parameters
	eventIDStr := c.Param("id")
	reminderIDStr := c.Param("reminder_id")

	eventID, err := primitive.ObjectIDFromHex(eventIDStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid event ID format", err)
		return
	}

	reminderID, err := primitive.ObjectIDFromHex(reminderIDStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid reminder ID format", err)
		return
	}

	// Get the authenticated user's ID
	userID, exists := c.Get("userID")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	// Get the event service
	eventService := c.MustGet("eventService").(EventService)

	// Delete the reminder
	err = eventService.DeleteReminder(c.Request.Context(), reminderID, eventID, userID.(primitive.ObjectID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to delete reminder", err)
		return
	}

	response.Success(c, http.StatusOK, "Reminder deleted successfully", nil)
}

// GetReminders retrieves all reminders for an event set by the current user
func GetReminders(c *gin.Context) {
	// Get event ID from URL parameter
	eventIDStr := c.Param("id")
	eventID, err := primitive.ObjectIDFromHex(eventIDStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid event ID format", err)
		return
	}

	// Get the authenticated user's ID
	userID, exists := c.Get("userID")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	// Get pagination parameters
	limitStr := c.DefaultQuery("limit", "20")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		limit = 20
	}
	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	// Get the event service
	eventService := c.MustGet("eventService").(EventService)

	// Get the reminders
	reminders, total, err := eventService.GetReminders(c.Request.Context(), eventID, userID.(primitive.ObjectID), limit, offset)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve reminders", err)
		return
	}

	response.SuccessWithPagination(c, http.StatusOK, "Reminders retrieved successfully", reminders, limit, offset, total)
}
