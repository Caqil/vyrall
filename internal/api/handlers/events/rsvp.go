package events

import (
	"context"
	"net/http"

	"github.com/Caqil/vyrall/internal/models"
	"github.com/Caqil/vyrall/internal/utils/response"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// EventService defines the interface for event operations
type EventService interface {
	GetEventByID(ctx context.Context, id primitive.ObjectID) (*models.Event, error)
	UpdateAttendeeStatus(ctx context.Context, eventID, userID primitive.ObjectID, rsvp string) error
	AddAttendee(ctx context.Context, eventID, userID primitive.ObjectID, rsvp string, guestCount int, note string) (primitive.ObjectID, error)
	GetAttendeeByID(ctx context.Context, eventID, userID primitive.ObjectID) (*models.EventAttendee, error)
	UpdateRSVPCounts(ctx context.Context, eventID primitive.ObjectID) error
}

// RSVPRequest represents the request payload for RSVPing to an event
type RSVPRequest struct {
	RSVP       string `json:"rsvp" binding:"required,oneof=going interested not_going"`
	GuestCount int    `json:"guest_count" binding:"min=0"`
	Note       string `json:"note"`
}

// RSVP handles a user's RSVP to an event
func RSVP(c *gin.Context) {
	// Get event ID from path parameter
	eventIDStr := c.Param("id")
	eventID, err := primitive.ObjectIDFromHex(eventIDStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid event ID", err)
		return
	}

	// Get the authenticated user's ID
	userID, exists := c.Get("userID")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	// Parse and validate the request
	var req RSVPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request format", err)
		return
	}

	// Get the event service from the context
	eventService := c.MustGet("eventService").(EventService)

	// Verify the event exists
	event, err := eventService.GetEventByID(c.Request.Context(), eventID)
	if err != nil {
		response.Error(c, http.StatusNotFound, "Event not found", err)
		return
	}

	// Check if the event has ended
	if event.EndTime.Before(c.Time()) {
		response.Error(c, http.StatusBadRequest, "Cannot RSVP to an event that has ended", nil)
		return
	}

	// Check if the event is at capacity
	if event.MaxAttendees > 0 && req.RSVP == "going" {
		currentAttendees := event.RSVPCount.Going
		if currentAttendees >= event.MaxAttendees {
			response.Error(c, http.StatusBadRequest, "Event has reached maximum capacity", nil)
			return
		}
	}

	// Check if the user has already RSVPed
	existingRSVP, err := eventService.GetAttendeeByID(c.Request.Context(), eventID, userID.(primitive.ObjectID))

	var attendeeID primitive.ObjectID
	if err == nil && existingRSVP != nil {
		// Update existing RSVP
		err = eventService.UpdateAttendeeStatus(c.Request.Context(), eventID, userID.(primitive.ObjectID), req.RSVP)
		if err != nil {
			response.Error(c, http.StatusInternalServerError, "Failed to update RSVP", err)
			return
		}
		attendeeID = existingRSVP.ID
	} else {
		// Create new RSVP
		attendeeID, err = eventService.AddAttendee(
			c.Request.Context(),
			eventID,
			userID.(primitive.ObjectID),
			req.RSVP,
			req.GuestCount,
			req.Note,
		)
		if err != nil {
			response.Error(c, http.StatusInternalServerError, "Failed to add attendee", err)
			return
		}
	}

	// Update RSVP counts for the event
	if err := eventService.UpdateRSVPCounts(c.Request.Context(), eventID); err != nil {
		// Log error but continue - counts will be updated eventually
		c.Error(err)
	}

	// Return success response
	response.Success(c, http.StatusOK, "RSVP successful", gin.H{
		"event_id":    eventID.Hex(),
		"attendee_id": attendeeID.Hex(),
		"status":      req.RSVP,
	})
}
