package events

import (
	"net/http"
	"time"

	"github.com/Caqil/vyrall/internal/models"
	"github.com/Caqil/vyrall/internal/utils/response"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// CreateEventRequest represents the request payload for creating an event
type CreateEventRequest struct {
	Title          string                  `json:"title" binding:"required"`
	Description    string                  `json:"description" binding:"required"`
	StartTime      time.Time               `json:"start_time" binding:"required"`
	EndTime        time.Time               `json:"end_time" binding:"required"`
	TimeZone       string                  `json:"time_zone" binding:"required"`
	Location       models.EventLocation    `json:"location" binding:"required"`
	CoverImage     string                  `json:"cover_image"`
	Type           string                  `json:"type" binding:"required,oneof=in-person online hybrid"`
	Privacy        string                  `json:"privacy" binding:"required,oneof=public private invite-only"`
	GroupID        *string                 `json:"group_id"`
	Category       string                  `json:"category"`
	Tags           []string                `json:"tags"`
	MaxAttendees   int                     `json:"max_attendees"`
	IsRecurring    bool                    `json:"is_recurring"`
	RecurrenceRule string                  `json:"recurrence_rule"`
	URL            string                  `json:"url"`
	TicketInfo     *models.EventTicketInfo `json:"ticket_info"`
	CoHosts        []string                `json:"co_hosts"`
}

// CreateEvent handles the creation of a new event
func CreateEvent(c *gin.Context) {
	// Get the authenticated user's ID
	userID, exists := c.Get("userID")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	// Parse and validate the request
	var req CreateEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request format", err)
		return
	}

	// Additional validation
	if req.EndTime.Before(req.StartTime) {
		response.Error(c, http.StatusBadRequest, "End time cannot be before start time", nil)
		return
	}

	if req.StartTime.Before(time.Now()) {
		response.Error(c, http.StatusBadRequest, "Start time cannot be in the past", nil)
		return
	}

	// Get the event service
	eventService := c.MustGet("eventService").(EventService)

	// Create new event object
	event := &models.Event{
		Title:        req.Title,
		Description:  req.Description,
		HostID:       userID.(primitive.ObjectID),
		StartTime:    req.StartTime,
		EndTime:      req.EndTime,
		TimeZone:     req.TimeZone,
		Location:     req.Location,
		CoverImage:   req.CoverImage,
		Type:         req.Type,
		Privacy:      req.Privacy,
		Category:     req.Category,
		Tags:         req.Tags,
		MaxAttendees: req.MaxAttendees,
		IsRecurring:  req.IsRecurring,
		Status:       "scheduled",
		URL:          req.URL,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		RSVPCount: models.EventRSVPCounts{
			Going:      0,
			Interested: 0,
			NotGoing:   0,
			NoReply:    0,
			Waitlist:   0,
		},
	}

	// Handle group events
	if req.GroupID != nil {
		groupID, err := primitive.ObjectIDFromHex(*req.GroupID)
		if err != nil {
			response.Error(c, http.StatusBadRequest, "Invalid group ID", err)
			return
		}

		// Verify user has permission to create events in this group
		hasPermission, err := eventService.CheckGroupEventPermission(c.Request.Context(), groupID, userID.(primitive.ObjectID))
		if err != nil {
			response.Error(c, http.StatusInternalServerError, "Failed to check group permissions", err)
			return
		}

		if !hasPermission {
			response.Error(c, http.StatusForbidden, "You don't have permission to create events in this group", nil)
			return
		}

		event.GroupID = &groupID
	}

	// Handle co-hosts
	if len(req.CoHosts) > 0 {
		var coHostIDs []primitive.ObjectID
		for _, coHostStr := range req.CoHosts {
			coHostID, err := primitive.ObjectIDFromHex(coHostStr)
			if err != nil {
				response.Error(c, http.StatusBadRequest, "Invalid co-host ID: "+coHostStr, err)
				return
			}
			coHostIDs = append(coHostIDs, coHostID)
		}
		event.CoHosts = coHostIDs
	}

	// Handle ticket info
	if req.TicketInfo != nil {
		event.TicketInfo = req.TicketInfo
	}

	// Handle recurring events
	if req.IsRecurring && req.RecurrenceRule != "" {
		event.RecurrenceRule = req.RecurrenceRule
	}

	// Create the event
	eventID, err := eventService.CreateEvent(c.Request.Context(), event)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to create event", err)
		return
	}

	// Handle recurring event instances if needed
	if req.IsRecurring && req.RecurrenceRule != "" {
		go eventService.CreateRecurringInstances(eventID, req.RecurrenceRule, req.EndTime.AddDate(1, 0, 0)) // Create instances for up to a year
	}

	response.Success(c, http.StatusCreated, "Event created successfully", gin.H{
		"event_id": eventID.Hex(),
	})
}
