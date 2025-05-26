package events

import (
	"net/http"
	"strconv"

	"github.com/Caqil/vyrall/internal/utils/response"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// GetAttendees retrieves the list of attendees for an event
func GetAttendees(c *gin.Context) {
	// Get event ID from URL parameter
	eventIDStr := c.Param("id")
	eventID, err := primitive.ObjectIDFromHex(eventIDStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid event ID format", err)
		return
	}

	// Get query parameters
	rsvpStatus := c.DefaultQuery("status", "going") // Default to only showing confirmed attendees
	limitStr := c.DefaultQuery("limit", "20")
	offsetStr := c.DefaultQuery("offset", "0")

	// Parse pagination params
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		limit = 20
	}
	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	// Validate RSVP status
	validStatuses := map[string]bool{
		"going":      true,
		"interested": true,
		"not_going":  true,
		"all":        true,
	}
	if !validStatuses[rsvpStatus] {
		response.Error(c, http.StatusBadRequest, "Invalid RSVP status", nil)
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

	// Check if the user has permission to view attendees (for private events)
	if event.Privacy == "private" || event.Privacy == "invite-only" {
		userID, exists := c.Get("userID")
		if !exists {
			response.Error(c, http.StatusUnauthorized, "Authentication required for private events", nil)
			return
		}

		userObjectID := userID.(primitive.ObjectID)
		// Check if user is host, co-host or already an attendee
		isAuthorized := userObjectID == event.HostID

		if !isAuthorized {
			for _, coHostID := range event.CoHosts {
				if coHostID == userObjectID {
					isAuthorized = true
					break
				}
			}
		}

		if !isAuthorized {
			attendee, _ := eventService.GetAttendeeByID(c.Request.Context(), eventID, userObjectID)
			isAuthorized = attendee != nil
		}

		if !isAuthorized {
			response.Error(c, http.StatusForbidden, "Not authorized to view attendees for this event", nil)
			return
		}
	}

	// Get the attendees
	attendees, total, err := eventService.GetAttendees(c.Request.Context(), eventID, rsvpStatus, limit, offset)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve attendees", err)
		return
	}

	response.SuccessWithPagination(c, http.StatusOK, "Attendees retrieved successfully", attendees, limit, offset, total)
}

// CheckInAttendee marks an attendee as checked in to an event
func CheckInAttendee(c *gin.Context) {
	// Get event ID and user ID from URL parameters
	eventIDStr := c.Param("id")
	userIDStr := c.Param("user_id")

	eventID, err := primitive.ObjectIDFromHex(eventIDStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid event ID format", err)
		return
	}

	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid user ID format", err)
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

	// Verify the current user has permission to check in attendees (host or co-host)
	currentUserID, exists := c.Get("userID")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "Authentication required", nil)
		return
	}

	currentUserObjectID := currentUserID.(primitive.ObjectID)
	isAuthorized := currentUserObjectID == event.HostID

	if !isAuthorized {
		for _, coHostID := range event.CoHosts {
			if coHostID == currentUserObjectID {
				isAuthorized = true
				break
			}
		}
	}

	if !isAuthorized {
		response.Error(c, http.StatusForbidden, "Only hosts and co-hosts can check in attendees", nil)
		return
	}

	// Check in the attendee
	err = eventService.CheckInAttendee(c.Request.Context(), eventID, userID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to check in attendee", err)
		return
	}

	response.Success(c, http.StatusOK, "Attendee checked in successfully", nil)
}
