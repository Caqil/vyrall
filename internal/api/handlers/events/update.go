package events

import (
	"net/http"
	"time"

	"github.com/Caqil/vyrall/internal/models"
	"github.com/Caqil/vyrall/internal/utils/response"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// UpdateEventRequest represents the request payload for updating an event
type UpdateEventRequest struct {
	Title           *string                 `json:"title"`
	Description     *string                 `json:"description"`
	StartTime       *time.Time              `json:"start_time"`
	EndTime         *time.Time              `json:"end_time"`
	TimeZone        *string                 `json:"time_zone"`
	Location        *models.EventLocation   `json:"location"`
	CoverImage      *string                 `json:"cover_image"`
	Type            *string                 `json:"type"`
	Privacy         *string                 `json:"privacy"`
	Category        *string                 `json:"category"`
	Tags            []string                `json:"tags"`
	MaxAttendees    *int                    `json:"max_attendees"`
	URL             *string                 `json:"url"`
	TicketInfo      *models.EventTicketInfo `json:"ticket_info"`
	CoHosts         []string                `json:"co_hosts"`
	UpdateRecurring bool                    `json:"update_recurring"`
}

// UpdateEvent handles updating an existing event
func UpdateEvent(c *gin.Context) {
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
	var req UpdateEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request format", err)
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

	// Check if the user has permission to update the event
	if event.HostID != userID.(primitive.ObjectID) {
		// Check if user is a co-host
		isCoHost := false
		for _, coHostID := range event.CoHosts {
			if coHostID == userID.(primitive.ObjectID) {
				isCoHost = true
				break
			}
		}

		if !isCoHost {
			// Check if user is an admin
			isAdmin, err := eventService.IsUserAdmin(c.Request.Context(), userID.(primitive.ObjectID))
			if err != nil {
				response.Error(c, http.StatusInternalServerError, "Failed to check user permissions", err)
				return
			}

			if !isAdmin {
				response.Error(c, http.StatusForbidden, "Only the event host, co-host, or an admin can update this event", nil)
				return
			}
		}
	}

	// If the event is part of a group, check group permissions
	if event.GroupID != nil {
		hasPermission, err := eventService.CheckGroupEventPermission(c.Request.Context(), *event.GroupID, userID.(primitive.ObjectID))
		if err != nil {
			response.Error(c, http.StatusInternalServerError, "Failed to check group permissions", err)
			return
		}

		if !hasPermission {
			response.Error(c, http.StatusForbidden, "You don't have permission to update events in this group", nil)
			return
		}
	}

	// Create update map
	updates := make(map[string]interface{})
	updates["updated_at"] = time.Now()

	// Apply updates if provided
	if req.Title != nil {
		updates["title"] = *req.Title
	}

	if req.Description != nil {
		updates["description"] = *req.Description
	}

	if req.StartTime != nil {
		updates["start_time"] = *req.StartTime
	}

	if req.EndTime != nil {
		updates["end_time"] = *req.EndTime
	}

	// Additional validation if both start and end times are updated
	if req.StartTime != nil && req.EndTime != nil {
		if req.EndTime.Before(*req.StartTime) {
			response.Error(c, http.StatusBadRequest, "End time cannot be before start time", nil)
			return
		}
	} else if req.StartTime != nil && event.EndTime.Before(*req.StartTime) {
		response.Error(c, http.StatusBadRequest, "End time cannot be before start time", nil)
		return
	} else if req.EndTime != nil && req.EndTime.Before(event.StartTime) {
		response.Error(c, http.StatusBadRequest, "End time cannot be before start time", nil)
		return
	}

	if req.TimeZone != nil {
		updates["time_zone"] = *req.TimeZone
	}

	if req.Location != nil {
		updates["location"] = *req.Location
	}

	if req.CoverImage != nil {
		updates["cover_image"] = *req.CoverImage
	}

	if req.Type != nil {
		if *req.Type != "in-person" && *req.Type != "online" && *req.Type != "hybrid" {
			response.Error(c, http.StatusBadRequest, "Invalid event type", nil)
			return
		}
		updates["type"] = *req.Type
	}

	if req.Privacy != nil {
		if *req.Privacy != "public" && *req.Privacy != "private" && *req.Privacy != "invite-only" {
			response.Error(c, http.StatusBadRequest, "Invalid privacy setting", nil)
			return
		}
		updates["privacy"] = *req.Privacy
	}

	if req.Category != nil {
		updates["category"] = *req.Category
	}

	if req.Tags != nil {
		updates["tags"] = req.Tags
	}

	if req.MaxAttendees != nil {
		if *req.MaxAttendees < event.RSVPCount.Going && *req.MaxAttendees > 0 {
			response.Error(c, http.StatusBadRequest, "Cannot reduce maximum attendees below current count", nil)
			return
		}
		updates["max_attendees"] = *req.MaxAttendees
	}

	if req.URL != nil {
		updates["url"] = *req.URL
	}

	if req.TicketInfo != nil {
		updates["ticket_info"] = *req.TicketInfo
	}

	if req.CoHosts != nil {
		var coHostIDs []primitive.ObjectID
		for _, coHostStr := range req.CoHosts {
			coHostID, err := primitive.ObjectIDFromHex(coHostStr)
			if err != nil {
				response.Error(c, http.StatusBadRequest, "Invalid co-host ID: "+coHostStr, err)
				return
			}
			coHostIDs = append(coHostIDs, coHostID)
		}
		updates["co_hosts"] = coHostIDs
	}

	// Update the event
	err = eventService.UpdateEvent(c.Request.Context(), eventID, updates)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to update event", err)
		return
	}

	// Handle recurring event updates if needed
	if event.IsRecurring && req.UpdateRecurring {
		go eventService.UpdateRecurringEventInstances(eventID, updates)
	}

	response.Success(c, http.StatusOK, "Event updated successfully", nil)
}
