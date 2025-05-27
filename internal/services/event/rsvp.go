package event

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/Caqil/vyrall/internal/internal/models"
	"github.com/Caqil/vyrall/internal/pkg/errors"
	"github.com/Caqil/vyrall/internal/pkg/logging"
)

// RSVPService handles event RSVPs
type RSVPService struct {
	eventRepo       EventRepository
	attendeeRepo    AttendeeRepository
	userRepo        UserRepository
	notificationSvc *NotificationsService
	logger          logging.Logger
}

// NewRSVPService creates a new RSVP service
func NewRSVPService(
	eventRepo EventRepository,
	attendeeRepo AttendeeRepository,
	userRepo UserRepository,
	notificationSvc *NotificationsService,
	logger logging.Logger,
) *RSVPService {
	return &RSVPService{
		eventRepo:       eventRepo,
		attendeeRepo:    attendeeRepo,
		userRepo:        userRepo,
		notificationSvc: notificationSvc,
		logger:          logger,
	}
}

// RSVP creates or updates an RSVP for an event
func (s *RSVPService) RSVP(ctx context.Context, eventID, userID primitive.ObjectID, rsvpStatus string, guestCount int) (*models.EventAttendee, error) {
	// Validate event exists
	event, err := s.eventRepo.FindByID(ctx, eventID)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to find event")
	}

	// Validate user exists
	_, err = s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to find user")
	}

	// Validate RSVP status
	if !isValidRSVP(rsvpStatus) {
		return nil, errors.New(errors.CodeInvalidArgument, "Invalid RSVP status")
	}

	// Validate guest count
	if guestCount < 0 {
		return nil, errors.New(errors.CodeInvalidArgument, "Guest count cannot be negative")
	}

	// Check if event is cancelled
	if event.Status == "cancelled" {
		return nil, errors.New(errors.CodeInvalidOperation, "Cannot RSVP to a cancelled event")
	}

	// Check if event is already over
	if event.EndTime.Before(time.Now()) {
		return nil, errors.New(errors.CodeInvalidOperation, "Cannot RSVP to a past event")
	}

	// Check if attendee record already exists
	attendee, err := s.attendeeRepo.FindByEventAndUser(ctx, eventID, userID)
	var oldRSVP string

	if err != nil {
		if errors.Code(err) != errors.CodeNotFound {
			return nil, errors.Wrap(err, "Failed to check existing RSVP")
		}

		// Create new attendee record
		attendee = &models.EventAttendee{
			EventID:       eventID,
			UserID:        userID,
			RSVP:          rsvpStatus,
			RSVPTimestamp: time.Now(),
			CheckedIn:     false,
			GuestCount:    guestCount,
		}

		// Check if event has maximum attendees limit
		if rsvpStatus == "going" && event.MaxAttendees > 0 {
			// Count current "going" attendees
			goingCount, err := s.attendeeRepo.CountByEventIDAndRSVP(ctx, eventID, "going")
			if err != nil {
				s.logger.Warn("Failed to count attendees", "eventId", eventID.Hex(), "error", err)
			} else {
				// Check if adding this attendee would exceed the limit
				if goingCount+guestCount+1 > event.MaxAttendees {
					// Put on waitlist instead
					attendee.RSVP = "going"
					attendee.IsWaitlisted = true

					// Get current waitlist count for position
					waitlistCount, err := s.attendeeRepo.CountWaitlisted(ctx, eventID)
					if err != nil {
						s.logger.Warn("Failed to count waitlisted", "eventId", eventID.Hex(), "error", err)
					} else {
						attendee.WaitlistPosition = waitlistCount + 1
					}
				}
			}
		}

		// Create attendee record
		createdAttendee, err := s.attendeeRepo.Create(ctx, attendee)
		if err != nil {
			return nil, errors.Wrap(err, "Failed to create attendee record")
		}

		attendee = createdAttendee
		oldRSVP = ""
	} else {
		// Store old RSVP for updating counts
		oldRSVP = attendee.RSVP

		// Update existing attendee record
		attendee.RSVP = rsvpStatus
		attendee.RSVPTimestamp = time.Now()
		attendee.GuestCount = guestCount

		// Handle waitlist logic
		if rsvpStatus == "going" && event.MaxAttendees > 0 {
			// Count current "going" attendees (excluding this one if already going)
			filter := map[string]interface{}{
				"event_id":      eventID,
				"rsvp":          "going",
				"is_waitlisted": false,
			}

			if oldRSVP == "going" && !attendee.IsWaitlisted {
				filter["user_id"] = map[string]interface{}{
					"$ne": userID,
				}
			}

			goingCount, err := s.attendeeRepo.CountWithFilter(ctx, filter)
			if err != nil {
				s.logger.Warn("Failed to count attendees", "eventId", eventID.Hex(), "error", err)
			} else {
				// Check if adding this attendee would exceed the limit
				if goingCount+guestCount+1 > event.MaxAttendees {
					// Put on waitlist
					attendee.IsWaitlisted = true

					// Get current waitlist count for position
					waitlistCount, err := s.attendeeRepo.CountWaitlisted(ctx, eventID)
					if err != nil {
						s.logger.Warn("Failed to count waitlisted", "eventId", eventID.Hex(), "error", err)
					} else {
						attendee.WaitlistPosition = waitlistCount + 1
					}
				} else if attendee.IsWaitlisted {
					// Remove from waitlist
					attendee.IsWaitlisted = false
					attendee.WaitlistPosition = 0

					// Notify about waitlist change
					s.notificationSvc.NotifyEventWaitlistStatusChanged(ctx, event, attendee, true)
				}
			}
		} else if attendee.IsWaitlisted {
			// Remove from waitlist for non-going RSVPs
			attendee.IsWaitlisted = false
			attendee.WaitlistPosition = 0
		}

		// Update attendee record
		err = s.attendeeRepo.Update(ctx, attendee)
		if err != nil {
			return nil, errors.Wrap(err, "Failed to update attendee record")
		}
	}

	// Update RSVP counts on event
	s.updateEventRSVPCounts(ctx, event, oldRSVP, rsvpStatus)

	// Send notification
	if oldRSVP != rsvpStatus {
		s.notificationSvc.NotifyNewRSVP(ctx, event, attendee)
	}

	return attendee, nil
}

// GetAttendee retrieves an attendee record
func (s *RSVPService) GetAttendee(ctx context.Context, eventID, userID primitive.ObjectID) (*models.EventAttendee, error) {
	attendee, err := s.attendeeRepo.FindByEventAndUser(ctx, eventID, userID)
	if err != nil {
		if errors.Code(err) == errors.CodeNotFound {
			return nil, errors.New(errors.CodeNotFound, "No RSVP found for this event")
		}
		return nil, errors.Wrap(err, "Failed to find attendee")
	}

	return attendee, nil
}

// GetAttendees retrieves attendees for an event
func (s *RSVPService) GetAttendees(ctx context.Context, eventID primitive.ObjectID, rsvpStatus string, page, limit int) ([]models.EventAttendee, int, error) {
	// Validate pagination
	if page < 1 {
		page = 1
	}

	if limit < 1 {
		limit = 10
	} else if limit > 100 {
		limit = 100
	}

	// Validate RSVP status if provided
	if rsvpStatus != "" && !isValidRSVP(rsvpStatus) {
		return nil, 0, errors.New(errors.CodeInvalidArgument, "Invalid RSVP status")
	}

	// Get attendees
	if rsvpStatus != "" {
		return s.attendeeRepo.FindByEventIDAndRSVPWithPagination(ctx, eventID, rsvpStatus, page, limit)
	}

	return s.attendeeRepo.FindByEventIDWithPagination(ctx, eventID, page, limit)
}

// GetWaitlist retrieves waitlisted attendees for an event
func (s *RSVPService) GetWaitlist(ctx context.Context, eventID primitive.ObjectID, page, limit int) ([]models.EventAttendee, int, error) {
	// Validate pagination
	if page < 1 {
		page = 1
	}

	if limit < 1 {
		limit = 10
	} else if limit > 100 {
		limit = 100
	}

	// Get waitlisted attendees
	return s.attendeeRepo.FindWaitlistedWithPagination(ctx, eventID, page, limit)
}

// RemoveAttendee removes an attendee from an event
func (s *RSVPService) RemoveAttendee(ctx context.Context, eventID, userID, removerID primitive.ObjectID) error {
	// Get event
	event, err := s.eventRepo.FindByID(ctx, eventID)
	if err != nil {
		return errors.Wrap(err, "Failed to find event")
	}

	// Check permissions (only host, co-hosts, the user themselves, or admins can remove attendees)
	if removerID != userID && removerID != event.HostID {
		// Check if remover is a co-host
		isCoHost := false
		for _, id := range event.CoHosts {
			if id == removerID {
				isCoHost = true
				break
			}
		}

		if !isCoHost {
			// Check if remover is an admin
			remover, err := s.userRepo.FindByID(ctx, removerID)
			if err != nil || remover.Role != "admin" {
				return errors.New(errors.CodeForbidden, "You don't have permission to remove attendees")
			}
		}
	}

	// Get attendee
	attendee, err := s.attendeeRepo.FindByEventAndUser(ctx, eventID, userID)
	if err != nil {
		if errors.Code(err) == errors.CodeNotFound {
			return errors.New(errors.CodeNotFound, "No RSVP found for this event")
		}
		return errors.Wrap(err, "Failed to find attendee")
	}

	// Store RSVP for updating counts
	oldRSVP := attendee.RSVP

	// Delete attendee record
	err = s.attendeeRepo.Delete(ctx, attendee.ID)
	if err != nil {
		return errors.Wrap(err, "Failed to delete attendee record")
	}

	// Update RSVP counts on event
	s.updateEventRSVPCounts(ctx, event, oldRSVP, "")

	// Check if this creates an opening for someone on the waitlist
	if oldRSVP == "going" && !attendee.IsWaitlisted && event.MaxAttendees > 0 {
		s.promoteFromWaitlist(ctx, event)
	}

	return nil
}

// InviteToEvent invites a user to an event
func (s *RSVPService) InviteToEvent(ctx context.Context, eventID, userID, inviterID primitive.ObjectID) error {
	// Validate event exists
	event, err := s.eventRepo.FindByID(ctx, eventID)
	if err != nil {
		return errors.Wrap(err, "Failed to find event")
	}

	// Validate user exists
	_, err = s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return errors.Wrap(err, "Failed to find user")
	}

	// Check if already RSVP'd
	existing, err := s.attendeeRepo.FindByEventAndUser(ctx, eventID, userID)
	if err == nil && existing != nil {
		return errors.New(errors.CodeInvalidOperation, "User already has an RSVP for this event")
	}

	// Create attendee record with invitation
	attendee := &models.EventAttendee{
		EventID:       eventID,
		UserID:        userID,
		RSVP:          "no_reply",
		RSVPTimestamp: time.Now(),
		InvitedBy:     &inviterID,
		CheckedIn:     false,
		GuestCount:    0,
	}

	_, err = s.attendeeRepo.Create(ctx, attendee)
	if err != nil {
		return errors.Wrap(err, "Failed to create attendee record")
	}

	// Update RSVP counts
	event.RSVPCount.NoReply++
	err = s.eventRepo.Update(ctx, event)
	if err != nil {
		s.logger.Warn("Failed to update RSVP count", "eventId", eventID.Hex(), "error", err)
	}

	// Send notification
	s.notificationSvc.NotifyEventInvitation(ctx, event, userID, inviterID)

	return nil
}

// UpdateGuestCount updates the number of guests for an attendee
func (s *RSVPService) UpdateGuestCount(ctx context.Context, eventID, userID primitive.ObjectID, guestCount int) error {
	// Validate guest count
	if guestCount < 0 {
		return errors.New(errors.CodeInvalidArgument, "Guest count cannot be negative")
	}

	// Get event
	event, err := s.eventRepo.FindByID(ctx, eventID)
	if err != nil {
		return errors.Wrap(err, "Failed to find event")
	}

	// Get attendee
	attendee, err := s.attendeeRepo.FindByEventAndUser(ctx, eventID, userID)
	if err != nil {
		return errors.Wrap(err, "Failed to find attendee")
	}

	// Check if on waitlist
	if attendee.IsWaitlisted {
		return errors.New(errors.CodeInvalidOperation, "Cannot update guest count while on waitlist")
	}

	// Check if this would exceed max attendees
	if event.MaxAttendees > 0 && attendee.RSVP == "going" {
		// Calculate total without this attendee's current guests
		totalWithoutCurrentGuests := event.RSVPCount.Going - attendee.GuestCount

		// Check if new total would exceed max
		if totalWithoutCurrentGuests+guestCount > event.MaxAttendees {
			return errors.New(errors.CodeInvalidOperation, "Cannot add this many guests due to event capacity")
		}
	}

	// Update guest count
	attendee.GuestCount = guestCount

	// Update attendee record
	err = s.attendeeRepo.Update(ctx, attendee)
	if err != nil {
		return errors.Wrap(err, "Failed to update attendee record")
	}

	return nil
}

// Helper methods

// updateEventRSVPCounts updates RSVP counts on an event
func (s *RSVPService) updateEventRSVPCounts(ctx context.Context, event *models.Event, oldStatus, newStatus string) {
	// Skip if no change
	if oldStatus == newStatus {
		return
	}

	// Decrement old status count
	switch oldStatus {
	case "going":
		if event.RSVPCount.Going > 0 {
			event.RSVPCount.Going--
		}
	case "interested":
		if event.RSVPCount.Interested > 0 {
			event.RSVPCount.Interested--
		}
	case "not_going":
		if event.RSVPCount.NotGoing > 0 {
			event.RSVPCount.NotGoing--
		}
	case "no_reply":
		if event.RSVPCount.NoReply > 0 {
			event.RSVPCount.NoReply--
		}
	}

	// Increment new status count
	switch newStatus {
	case "going":
		event.RSVPCount.Going++
	case "interested":
		event.RSVPCount.Interested++
	case "not_going":
		event.RSVPCount.NotGoing++
	case "no_reply":
		event.RSVPCount.NoReply++
	}

	// Update event
	err := s.eventRepo.Update(ctx, event)
	if err != nil {
		s.logger.Warn("Failed to update RSVP counts", "eventId", event.ID.Hex(), "error", err)
	}
}

// promoteFromWaitlist promotes the next person from the waitlist
func (s *RSVPService) promoteFromWaitlist(ctx context.Context, event *models.Event) {
	// Get first person on waitlist
	waitlistAttendees, err := s.attendeeRepo.FindWaitlistedWithPagination(ctx, event.ID, 1, 1)
	if err != nil {
		s.logger.Warn("Failed to get waitlist", "eventId", event.ID.Hex(), "error", err)
		return
	}

	if len(waitlistAttendees) == 0 {
		// No one on waitlist
		return
	}

	// Get first attendee
	attendee := waitlistAttendees[0]

	// Remove from waitlist
	attendee.IsWaitlisted = false
	attendee.WaitlistPosition = 0

	// Update attendee
	err = s.attendeeRepo.Update(ctx, &attendee)
	if err != nil {
		s.logger.Warn("Failed to update waitlisted attendee", "attendeeId", attendee.ID.Hex(), "error", err)
		return
	}

	// Send notification
	s.notificationSvc.NotifyEventWaitlistStatusChanged(ctx, event, &attendee, true)

	// Reorder waitlist positions
	s.reorderWaitlist(ctx, event.ID)
}

// reorderWaitlist reorders waitlist positions after a promotion
func (s *RSVPService) reorderWaitlist(ctx context.Context, eventID primitive.ObjectID) {
	// Get all waitlisted attendees
	waitlistAttendees, err := s.attendeeRepo.FindWaitlistedWithPagination(ctx, eventID, 1, 1000)
	if err != nil {
		s.logger.Warn("Failed to get waitlist for reordering", "eventId", eventID.Hex(), "error", err)
		return
	}

	// Update positions
	for i, attendee := range waitlistAttendees {
		attendee.WaitlistPosition = i + 1

		err = s.attendeeRepo.Update(ctx, &attendee)
		if err != nil {
			s.logger.Warn("Failed to update waitlist position", "attendeeId", attendee.ID.Hex(), "error", err)
		}
	}
}

// isValidRSVP checks if an RSVP status is valid
func isValidRSVP(status string) bool {
	validStatuses := map[string]bool{
		"going":      true,
		"interested": true,
		"not_going":  true,
		"no_reply":   true,
	}

	return validStatuses[status]
}
