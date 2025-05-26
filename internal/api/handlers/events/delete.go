package events

import (
	"net/http"

	"github.com/Caqil/vyrall/internal/utils/response"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// DeleteEvent handles the deletion of an event
func DeleteEvent(c *gin.Context) {
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

	// Get the event service
	eventService := c.MustGet("eventService").(EventService)

	// Check if the event exists
	event, err := eventService.GetEventByID(c.Request.Context(), eventID)
	if err != nil {
		response.Error(c, http.StatusNotFound, "Event not found", err)
		return
	}

	// Check if the user has permission to delete the event (must be host)
	if event.HostID != userID.(primitive.ObjectID) {
		// Check if user is an admin
		isAdmin, err := eventService.IsUserAdmin(c.Request.Context(), userID.(primitive.ObjectID))
		if err != nil {
			response.Error(c, http.StatusInternalServerError, "Failed to check user permissions", err)
			return
		}

		if !isAdmin {
			response.Error(c, http.StatusForbidden, "Only the event host or an admin can delete this event", nil)
			return
		}
	}

	// If event is part of a group, check group permissions
	if event.GroupID != nil {
		hasPermission, err := eventService.CheckGroupEventPermission(c.Request.Context(), *event.GroupID, userID.(primitive.ObjectID))
		if err != nil {
			response.Error(c, http.StatusInternalServerError, "Failed to check group permissions", err)
			return
		}

		if !hasPermission {
			response.Error(c, http.StatusForbidden, "You don't have permission to delete events in this group", nil)
			return
		}
	}

	// Delete the event
	err = eventService.DeleteEvent(c.Request.Context(), eventID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to delete event", err)
		return
	}

	// Check if this is a recurring event instance
	if event.ParentEventID != nil {
		// Just delete this instance
		response.Success(c, http.StatusOK, "Event instance deleted successfully", nil)
		return
	}

	// If this is a recurring event parent, delete all instances
	if event.IsRecurring {
		err = eventService.DeleteRecurringEventInstances(c.Request.Context(), eventID)
		if err != nil {
			// Log error but don't fail the request - the main event was deleted
			c.Error(err)
		}
	}

	// Send notifications to attendees
	go eventService.NotifyEventCancellation(eventID, event.Title)

	response.Success(c, http.StatusOK, "Event deleted successfully", nil)
}

// CancelEvent changes an event's status to cancelled without deleting it
func CancelEvent(c *gin.Context) {
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

	// Get the event service
	eventService := c.MustGet("eventService").(EventService)

	// Check if the event exists
	event, err := eventService.GetEventByID(c.Request.Context(), eventID)
	if err != nil {
		response.Error(c, http.StatusNotFound, "Event not found", err)
		return
	}

	// Check if the user has permission to cancel the event (must be host or co-host)
	if event.HostID != userID.(primitive.ObjectID) {
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
				response.Error(c, http.StatusForbidden, "Only the event host, co-host, or an admin can cancel this event", nil)
				return
			}
		}
	}

	// Parse optional cancellation reason
	var req struct {
		Reason string `json:"reason"`
	}
	c.ShouldBindJSON(&req)

	// Cancel the event
	err = eventService.CancelEvent(c.Request.Context(), eventID, req.Reason)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to cancel event", err)
		return
	}

	// Send notifications to attendees
	go eventService.NotifyEventCancellation(eventID, event.Title)

	response.Success(c, http.StatusOK, "Event cancelled successfully", nil)
}
