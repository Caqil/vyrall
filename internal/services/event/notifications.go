package event

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/Caqil/vyrall/internal/internal/models"
	"github.com/Caqil/vyrall/internal/internal/notification"
	"github.com/Caqil/vyrall/internal/pkg/errors"
	"github.com/Caqil/vyrall/internal/pkg/logging"
)

// NotificationsService handles notifications for events
type NotificationsService struct {
	notifier     notification.Service
	eventRepo    EventRepository
	attendeeRepo AttendeeRepository
	userRepo     UserRepository
	groupRepo    GroupRepository
	logger       logging.Logger
}

// NewNotificationsService creates a new notifications service for events
func NewNotificationsService(
	notifier notification.Service,
	eventRepo EventRepository,
	attendeeRepo AttendeeRepository,
	userRepo UserRepository,
	groupRepo GroupRepository,
	logger logging.Logger,
) *NotificationsService {
	return &NotificationsService{
		notifier:     notifier,
		eventRepo:    eventRepo,
		attendeeRepo: attendeeRepo,
		userRepo:     userRepo,
		groupRepo:    groupRepo,
		logger:       logger,
	}
}

// NotifyEventCreated sends notifications when an event is created
func (s *NotificationsService) NotifyEventCreated(ctx context.Context, event *models.Event) error {
	// Get host
	host, err := s.userRepo.FindByID(ctx, event.HostID)
	if err != nil {
		s.logger.Warn("Failed to find host", "hostId", event.HostID.Hex(), "error", err)
		return errors.Wrap(err, "Failed to find host")
	}

	// Notify followers of the host
	followers, err := s.userRepo.FindFollowers(ctx, event.HostID)
	if err != nil {
		s.logger.Warn("Failed to find followers", "hostId", event.HostID.Hex(), "error", err)
	} else {
		for _, follower := range followers {
			// Create notification
			notif := &models.Notification{
				UserID:    follower.ID,
				Type:      "new_event",
				Actor:     event.HostID,
				Subject:   "event",
				SubjectID: event.ID,
				Message:   fmt.Sprintf("%s created a new event: %s", host.DisplayName, event.Title),
				ActionURL: fmt.Sprintf("/events/%s", event.ID.Hex()),
				IsRead:    false,
			}

			// Send notification
			err = s.notifier.Send(ctx, notif)
			if err != nil {
				s.logger.Warn("Failed to send new event notification", "followerId", follower.ID.Hex(), "error", err)
			}
		}
	}

	// If group event, notify group members
	if event.GroupID != nil {
		// Get group
		group, err := s.groupRepo.FindByID(ctx, *event.GroupID)
		if err != nil {
			s.logger.Warn("Failed to find group", "groupId", event.GroupID.Hex(), "error", err)
		} else {
			// Get group members
			members, err := s.groupRepo.FindMembers(ctx, *event.GroupID)
			if err != nil {
				s.logger.Warn("Failed to find group members", "groupId", event.GroupID.Hex(), "error", err)
			} else {
				for _, member := range members {
					// Skip host
					if member.UserID == event.HostID {
						continue
					}

					// Create notification
					notif := &models.Notification{
						UserID:    member.UserID,
						Type:      "group_event",
						Actor:     event.HostID,
						Subject:   "event",
						SubjectID: event.ID,
						Message:   fmt.Sprintf("New event in %s: %s", group.Name, event.Title),
						ActionURL: fmt.Sprintf("/events/%s", event.ID.Hex()),
						IsRead:    false,
					}

					// Send notification
					err = s.notifier.Send(ctx, notif)
					if err != nil {
						s.logger.Warn("Failed to send group event notification", "memberId", member.UserID.Hex(), "error", err)
					}
				}
			}
		}
	}

	return nil
}

// NotifyEventUpdated sends notifications when an event is updated
func (s *NotificationsService) NotifyEventUpdated(ctx context.Context, event *models.Event, changes []string) error {
	// Only notify if significant changes
	if len(changes) == 0 {
		return nil
	}

	// Get host
	host, err := s.userRepo.FindByID(ctx, event.HostID)
	if err != nil {
		s.logger.Warn("Failed to find host", "hostId", event.HostID.Hex(), "error", err)
		return errors.Wrap(err, "Failed to find host")
	}

	// Get attendees
	attendees, err := s.attendeeRepo.FindByEventID(ctx, event.ID)
	if err != nil {
		s.logger.Warn("Failed to find attendees", "eventId", event.ID.Hex(), "error", err)
		return errors.Wrap(err, "Failed to find attendees")
	}

	// Create message
	message := fmt.Sprintf("%s updated the event '%s'", host.DisplayName, event.Title)

	// Add change details if there are few changes
	if len(changes) <= 3 {
		message += ": " + joinStringSlice(changes, ", ")
	}

	// Notify all attendees
	for _, attendee := range attendees {
		// Skip host
		if attendee.UserID == event.HostID {
			continue
		}

		// Create notification
		notif := &models.Notification{
			UserID:    attendee.UserID,
			Type:      "event_update",
			Actor:     event.HostID,
			Subject:   "event",
			SubjectID: event.ID,
			Message:   message,
			ActionURL: fmt.Sprintf("/events/%s", event.ID.Hex()),
			IsRead:    false,
		}

		// Send notification
		err = s.notifier.Send(ctx, notif)
		if err != nil {
			s.logger.Warn("Failed to send event update notification", "attendeeId", attendee.UserID.Hex(), "error", err)
		}
	}

	return nil
}

// NotifyEventCancelled sends notifications when an event is cancelled
func (s *NotificationsService) NotifyEventCancelled(ctx context.Context, event *models.Event) error {
	// Get attendees
	attendees, err := s.attendeeRepo.FindByEventID(ctx, event.ID)
	if err != nil {
		s.logger.Warn("Failed to find attendees", "eventId", event.ID.Hex(), "error", err)
		return errors.Wrap(err, "Failed to find attendees")
	}

	// Notify all attendees
	for _, attendee := range attendees {
		// Skip host
		if attendee.UserID == event.HostID {
			continue
		}

		// Create notification
		notif := &models.Notification{
			UserID:    attendee.UserID,
			Type:      "event_cancelled",
			Actor:     event.HostID,
			Subject:   "event",
			SubjectID: event.ID,
			Message:   fmt.Sprintf("Event '%s' has been cancelled", event.Title),
			ActionURL: fmt.Sprintf("/events/%s", event.ID.Hex()),
			IsRead:    false,
			Priority:  "high",
		}

		// Send notification
		err = s.notifier.Send(ctx, notif)
		if err != nil {
			s.logger.Warn("Failed to send event cancellation notification", "attendeeId", attendee.UserID.Hex(), "error", err)
		}
	}

	return nil
}

// NotifyEventStartingSoon sends notifications for upcoming events
func (s *NotificationsService) NotifyEventStartingSoon(ctx context.Context, event *models.Event, minutesBefore int) error {
	// Get attendees with RSVP "going"
	attendees, err := s.attendeeRepo.FindByEventIDAndRSVP(ctx, event.ID, "going")
	if err != nil {
		s.logger.Warn("Failed to find attendees", "eventId", event.ID.Hex(), "error", err)
		return errors.Wrap(err, "Failed to find attendees")
	}

	// Create message
	var message string
	if minutesBefore >= 60 {
		hours := minutesBefore / 60
		message = fmt.Sprintf("Event '%s' starts in %d hour(s)", event.Title, hours)
	} else {
		message = fmt.Sprintf("Event '%s' starts in %d minutes", event.Title, minutesBefore)
	}

	// Notify all attendees
	for _, attendee := range attendees {
		// Create notification
		notif := &models.Notification{
			UserID:    attendee.UserID,
			Type:      "event_starting_soon",
			Subject:   "event",
			SubjectID: event.ID,
			Message:   message,
			ActionURL: fmt.Sprintf("/events/%s", event.ID.Hex()),
			IsRead:    false,
			Priority:  "high",
		}

		// Send notification
		err = s.notifier.Send(ctx, notif)
		if err != nil {
			s.logger.Warn("Failed to send event starting soon notification", "attendeeId", attendee.UserID.Hex(), "error", err)
		}
	}

	return nil
}

// NotifyNewRSVP sends a notification to the host when someone RSVPs
func (s *NotificationsService) NotifyNewRSVP(ctx context.Context, event *models.Event, attendee *models.EventAttendee) error {
	// Skip if attendee is the host
	if attendee.UserID == event.HostID {
		return nil
	}

	// Get user who RSVPed
	user, err := s.userRepo.FindByID(ctx, attendee.UserID)
	if err != nil {
		s.logger.Warn("Failed to find user", "userId", attendee.UserID.Hex(), "error", err)
		return errors.Wrap(err, "Failed to find user")
	}

	// Create message based on RSVP status
	var message string
	switch attendee.RSVP {
	case "going":
		message = fmt.Sprintf("%s is going to your event '%s'", user.DisplayName, event.Title)
	case "interested":
		message = fmt.Sprintf("%s is interested in your event '%s'", user.DisplayName, event.Title)
	case "not_going":
		message = fmt.Sprintf("%s is not going to your event '%s'", user.DisplayName, event.Title)
	default:
		return nil // Don't notify for other statuses
	}

	// Create notification for host
	notif := &models.Notification{
		UserID:    event.HostID,
		Type:      "event_rsvp",
		Actor:     attendee.UserID,
		Subject:   "event",
		SubjectID: event.ID,
		Message:   message,
		ActionURL: fmt.Sprintf("/events/%s/attendees", event.ID.Hex()),
		IsRead:    false,
	}

	// Send notification
	err = s.notifier.Send(ctx, notif)
	if err != nil {
		s.logger.Warn("Failed to send RSVP notification", "hostId", event.HostID.Hex(), "error", err)
	}

	// Also notify co-hosts if this is a "going" RSVP
	if attendee.RSVP == "going" {
		for _, coHostID := range event.CoHosts {
			// Skip if co-host is the attendee
			if coHostID == attendee.UserID {
				continue
			}

			// Create notification for co-host
			notif := &models.Notification{
				UserID:    coHostID,
				Type:      "event_rsvp",
				Actor:     attendee.UserID,
				Subject:   "event",
				SubjectID: event.ID,
				Message:   message,
				ActionURL: fmt.Sprintf("/events/%s/attendees", event.ID.Hex()),
				IsRead:    false,
			}

			// Send notification
			err = s.notifier.Send(ctx, notif)
			if err != nil {
				s.logger.Warn("Failed to send RSVP notification to co-host", "coHostId", coHostID.Hex(), "error", err)
			}
		}
	}

	return nil
}

// NotifyEventInvitation sends a notification when a user is invited to an event
func (s *NotificationsService) NotifyEventInvitation(ctx context.Context, event *models.Event, inviteeID, inviterID primitive.ObjectID) error {
	// Get inviter
	inviter, err := s.userRepo.FindByID(ctx, inviterID)
	if err != nil {
		s.logger.Warn("Failed to find inviter", "inviterId", inviterID.Hex(), "error", err)
		return errors.Wrap(err, "Failed to find inviter")
	}

	// Create notification
	notif := &models.Notification{
		UserID:    inviteeID,
		Type:      "event_invitation",
		Actor:     inviterID,
		Subject:   "event",
		SubjectID: event.ID,
		Message:   fmt.Sprintf("%s invited you to the event '%s'", inviter.DisplayName, event.Title),
		ActionURL: fmt.Sprintf("/events/%s", event.ID.Hex()),
		IsRead:    false,
	}

	// Send notification
	err = s.notifier.Send(ctx, notif)
	if err != nil {
		s.logger.Warn("Failed to send event invitation notification", "inviteeId", inviteeID.Hex(), "error", err)
	}

	return nil
}

// NotifyEventWaitlistStatusChanged sends a notification when a user's waitlist status changes
func (s *NotificationsService) NotifyEventWaitlistStatusChanged(ctx context.Context, event *models.Event, attendee *models.EventAttendee, promoted bool) error {
	// Create message
	var message string
	if promoted {
		message = fmt.Sprintf("You've been moved from the waitlist to the attendee list for '%s'", event.Title)
	} else {
		message = fmt.Sprintf("You've been added to the waitlist for '%s'", event.Title)
	}

	// Create notification
	notif := &models.Notification{
		UserID:    attendee.UserID,
		Type:      "event_waitlist",
		Subject:   "event",
		SubjectID: event.ID,
		Message:   message,
		ActionURL: fmt.Sprintf("/events/%s", event.ID.Hex()),
		IsRead:    false,
		Priority:  "high",
	}

	// Send notification
	err := s.notifier.Send(ctx, notif)
	if err != nil {
		s.logger.Warn("Failed to send waitlist notification", "attendeeId", attendee.UserID.Hex(), "error", err)
	}

	return nil
}

// NotifyEventCheckIn sends a notification to the attendee when they're checked in
func (s *NotificationsService) NotifyEventCheckIn(ctx context.Context, event *models.Event, attendee *models.EventAttendee) error {
	// Create notification
	notif := &models.Notification{
		UserID:    attendee.UserID,
		Type:      "event_check_in",
		Subject:   "event",
		SubjectID: event.ID,
		Message:   fmt.Sprintf("You have been checked in to '%s'", event.Title),
		ActionURL: fmt.Sprintf("/events/%s", event.ID.Hex()),
		IsRead:    false,
	}

	// Send notification
	err := s.notifier.Send(ctx, notif)
	if err != nil {
		s.logger.Warn("Failed to send check-in notification", "attendeeId", attendee.UserID.Hex(), "error", err)
	}

	return nil
}

// SendEventReminder sends a reminder notification for an event
func (s *NotificationsService) SendEventReminder(ctx context.Context, reminder *models.EventReminder) error {
	// Get event
	event, err := s.eventRepo.FindByID(ctx, reminder.EventID)
	if err != nil {
		s.logger.Warn("Failed to find event", "eventId", reminder.EventID.Hex(), "error", err)
		return errors.Wrap(err, "Failed to find event")
	}

	// Calculate time until event
	timeUntil := event.StartTime.Sub(time.Now())
	var timeDesc string

	if timeUntil.Hours() > 24 {
		days := int(timeUntil.Hours() / 24)
		timeDesc = fmt.Sprintf("%d day(s)", days)
	} else if timeUntil.Hours() >= 1 {
		hours := int(timeUntil.Hours())
		timeDesc = fmt.Sprintf("%d hour(s)", hours)
	} else {
		minutes := int(timeUntil.Minutes())
		timeDesc = fmt.Sprintf("%d minute(s)", minutes)
	}

	// Create notification
	notif := &models.Notification{
		UserID:    reminder.UserID,
		Type:      "event_reminder",
		Subject:   "event",
		SubjectID: event.ID,
		Message:   fmt.Sprintf("Reminder: '%s' starts in %s", event.Title, timeDesc),
		ActionURL: fmt.Sprintf("/events/%s", event.ID.Hex()),
		IsRead:    false,
		Priority:  "high",
	}

	// Send notification
	err = s.notifier.Send(ctx, notif)
	if err != nil {
		s.logger.Warn("Failed to send reminder notification", "reminderId", reminder.ID.Hex(), "error", err)
	}

	// Update reminder status
	reminder.Status = "sent"
	reminder.SentAt = timePtr(time.Now())

	// This would be handled by a reminders service in a real implementation

	return nil
}

// Helper functions

// joinStringSlice joins a slice of strings with a separator
func joinStringSlice(slice []string, sep string) string {
	if len(slice) == 0 {
		return ""
	}

	result := slice[0]
	for i := 1; i < len(slice); i++ {
		result += sep + slice[i]
	}

	return result
}

// timePtr returns a pointer to a time.Time
func timePtr(t time.Time) *time.Time {
	return &t
}
