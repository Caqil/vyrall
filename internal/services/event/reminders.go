package event

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/Caqil/vyrall/internal/internal/models"
	"github.com/Caqil/vyrall/internal/pkg/errors"
	"github.com/Caqil/vyrall/internal/pkg/logging"
)

// RemindersService handles event reminders
type RemindersService struct {
	reminderRepo    ReminderRepository
	eventRepo       EventRepository
	attendeeRepo    AttendeeRepository
	userRepo        UserRepository
	notificationSvc *NotificationsService
	logger          logging.Logger
}

// NewRemindersService creates a new reminders service
func NewRemindersService(
	reminderRepo ReminderRepository,
	eventRepo EventRepository,
	attendeeRepo AttendeeRepository,
	userRepo UserRepository,
	notificationSvc *NotificationsService,
	logger logging.Logger,
) *RemindersService {
	return &RemindersService{
		reminderRepo:    reminderRepo,
		eventRepo:       eventRepo,
		attendeeRepo:    attendeeRepo,
		userRepo:        userRepo,
		notificationSvc: notificationSvc,
		logger:          logger,
	}
}

// CreateReminder creates a new event reminder
func (s *RemindersService) CreateReminder(ctx context.Context, userID, eventID primitive.ObjectID, reminderTime time.Time, reminderType string) (*models.EventReminder, error) {
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

	// Check if user is attending the event
	attendee, err := s.attendeeRepo.FindByEventAndUser(ctx, eventID, userID)
	if err != nil {
		if errors.Code(err) != errors.CodeNotFound {
			return nil, errors.Wrap(err, "Failed to check attendee status")
		}

		// User is not an attendee, automatically add as "interested"
		attendee = &models.EventAttendee{
			EventID:       eventID,
			UserID:        userID,
			RSVP:          "interested",
			RSVPTimestamp: time.Now(),
			CheckedIn:     false,
			GuestCount:    0,
		}

		_, err = s.attendeeRepo.Create(ctx, attendee)
		if err != nil {
			s.logger.Warn("Failed to create attendee record", "userId", userID.Hex(), "eventId", eventID.Hex(), "error", err)
		} else {
			// Update RSVP count
			event.RSVPCount.Interested++
			err = s.eventRepo.Update(ctx, event)
			if err != nil {
				s.logger.Warn("Failed to update RSVP count", "eventId", eventID.Hex(), "error", err)
			}
		}
	}

	// Validate reminder time
	if reminderTime.After(event.StartTime) {
		return nil, errors.New(errors.CodeInvalidArgument, "Reminder time must be before event start time")
	}

	// Validate reminder type
	if !isValidReminderType(reminderType) {
		return nil, errors.New(errors.CodeInvalidArgument, "Invalid reminder type")
	}

	// Check for existing reminder
	existing, err := s.reminderRepo.FindByEventAndUser(ctx, eventID, userID)
	if err == nil && existing != nil {
		return nil, errors.New(errors.CodeDuplicateEntity, "Reminder already exists for this event")
	}

	// Create reminder
	reminder := &models.EventReminder{
		EventID:      eventID,
		UserID:       userID,
		ReminderTime: reminderTime,
		Type:         reminderType,
		Status:       "pending",
		CreatedAt:    time.Now(),
	}

	// Save reminder
	createdReminder, err := s.reminderRepo.Create(ctx, reminder)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create reminder")
	}

	return createdReminder, nil
}

// UpdateReminder updates an existing reminder
func (s *RemindersService) UpdateReminder(ctx context.Context, reminderID, userID primitive.ObjectID, reminderTime time.Time, reminderType string) (*models.EventReminder, error) {
	// Get reminder
	reminder, err := s.reminderRepo.FindByID(ctx, reminderID)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to find reminder")
	}

	// Check ownership
	if reminder.UserID != userID {
		return nil, errors.New(errors.CodeForbidden, "You can only update your own reminders")
	}

	// Validate event exists
	event, err := s.eventRepo.FindByID(ctx, reminder.EventID)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to find event")
	}

	// Validate reminder time
	if reminderTime.After(event.StartTime) {
		return nil, errors.New(errors.CodeInvalidArgument, "Reminder time must be before event start time")
	}

	// Validate reminder type
	if !isValidReminderType(reminderType) {
		return nil, errors.New(errors.CodeInvalidArgument, "Invalid reminder type")
	}

	// Update reminder
	reminder.ReminderTime = reminderTime
	reminder.Type = reminderType

	// If reminder was already sent, mark it as pending again
	if reminder.Status == "sent" {
		reminder.Status = "pending"
		reminder.SentAt = nil
	}

	// Save reminder
	err = s.reminderRepo.Update(ctx, reminder)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to update reminder")
	}

	return reminder, nil
}

// DeleteReminder deletes a reminder
func (s *RemindersService) DeleteReminder(ctx context.Context, reminderID, userID primitive.ObjectID) error {
	// Get reminder
	reminder, err := s.reminderRepo.FindByID(ctx, reminderID)
	if err != nil {
		return errors.Wrap(err, "Failed to find reminder")
	}

	// Check ownership
	if reminder.UserID != userID {
		return errors.New(errors.CodeForbidden, "You can only delete your own reminders")
	}

	// Delete reminder
	err = s.reminderRepo.Delete(ctx, reminderID)
	if err != nil {
		return errors.Wrap(err, "Failed to delete reminder")
	}

	return nil
}

// GetReminder retrieves a reminder by ID
func (s *RemindersService) GetReminder(ctx context.Context, reminderID, userID primitive.ObjectID) (*models.EventReminder, error) {
	// Get reminder
	reminder, err := s.reminderRepo.FindByID(ctx, reminderID)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to find reminder")
	}

	// Check ownership
	if reminder.UserID != userID {
		return nil, errors.New(errors.CodeForbidden, "You can only view your own reminders")
	}

	return reminder, nil
}

// GetRemindersByUser retrieves reminders for a user
func (s *RemindersService) GetRemindersByUser(ctx context.Context, userID primitive.ObjectID) ([]models.EventReminder, error) {
	// Get reminders
	reminders, err := s.reminderRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to find reminders")
	}

	return reminders, nil
}

// GetRemindersByEvent retrieves reminders for an event
func (s *RemindersService) GetRemindersByEvent(ctx context.Context, eventID primitive.ObjectID) ([]models.EventReminder, error) {
	// Get reminders
	reminders, err := s.reminderRepo.FindByEventID(ctx, eventID)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to find reminders")
	}

	return reminders, nil
}

// ProcessDueReminders processes reminders that are due to be sent
func (s *RemindersService) ProcessDueReminders(ctx context.Context) (int, error) {
	// Get reminders that are due
	now := time.Now()
	reminders, err := s.reminderRepo.FindDueReminders(ctx, now)
	if err != nil {
		return 0, errors.Wrap(err, "Failed to find due reminders")
	}

	// Process each reminder
	count := 0
	for _, reminder := range reminders {
		// Check if event still exists and is not cancelled
		event, err := s.eventRepo.FindByID(ctx, reminder.EventID)
		if err != nil {
			s.logger.Warn("Failed to find event for reminder", "eventId", reminder.EventID.Hex(), "error", err)
			continue
		}

		if event.Status == "cancelled" {
			// Delete reminder for cancelled event
			err = s.reminderRepo.Delete(ctx, reminder.ID)
			if err != nil {
				s.logger.Warn("Failed to delete reminder for cancelled event", "reminderId", reminder.ID.Hex(), "error", err)
			}
			continue
		}

		// Send reminder notification
		err = s.notificationSvc.SendEventReminder(ctx, &reminder)
		if err != nil {
			s.logger.Warn("Failed to send reminder notification", "reminderId", reminder.ID.Hex(), "error", err)
			continue
		}

		// Mark reminder as sent
		reminder.Status = "sent"
		reminder.SentAt = timePtr(now)

		err = s.reminderRepo.Update(ctx, &reminder)
		if err != nil {
			s.logger.Warn("Failed to update reminder status", "reminderId", reminder.ID.Hex(), "error", err)
			continue
		}

		count++
	}

	return count, nil
}

// CreateDefaultReminders creates default reminders for a user's upcoming events
func (s *RemindersService) CreateDefaultReminders(ctx context.Context, userID primitive.ObjectID) (int, error) {
	// Get user's upcoming events with RSVP "going"
	attendees, err := s.attendeeRepo.FindByUserIDAndRSVP(ctx, userID, "going", 1, 100)
	if err != nil {
		return 0, errors.Wrap(err, "Failed to find user's events")
	}

	// Process each event
	count := 0
	now := time.Now()

	for _, attendee := range attendees {
		// Get event
		event, err := s.eventRepo.FindByID(ctx, attendee.EventID)
		if err != nil {
			s.logger.Warn("Failed to find event", "eventId", attendee.EventID.Hex(), "error", err)
			continue
		}

		// Skip past events or cancelled events
		if event.StartTime.Before(now) || event.Status == "cancelled" {
			continue
		}

		// Check if reminder already exists
		existing, err := s.reminderRepo.FindByEventAndUser(ctx, event.ID, userID)
		if err == nil && existing != nil {
			continue
		}

		// Calculate reminder time (1 day before)
		reminderTime := event.StartTime.AddDate(0, 0, -1)

		// If event is less than a day away, set reminder to 1 hour before
		if reminderTime.Before(now) {
			reminderTime = event.StartTime.Add(-1 * time.Hour)
		}

		// If event is less than an hour away, skip creating reminder
		if reminderTime.Before(now) {
			continue
		}

		// Create reminder
		reminder := &models.EventReminder{
			EventID:      event.ID,
			UserID:       userID,
			ReminderTime: reminderTime,
			Type:         "push",
			Status:       "pending",
			CreatedAt:    now,
		}

		_, err = s.reminderRepo.Create(ctx, reminder)
		if err != nil {
			s.logger.Warn("Failed to create default reminder", "eventId", event.ID.Hex(), "userId", userID.Hex(), "error", err)
			continue
		}

		count++
	}

	return count, nil
}

// CreateEventReminders creates reminders for all attendees of an event
func (s *RemindersService) CreateEventReminders(ctx context.Context, eventID primitive.ObjectID, hours int) (int, error) {
	// Get event
	event, err := s.eventRepo.FindByID(ctx, eventID)
	if err != nil {
		return 0, errors.Wrap(err, "Failed to find event")
	}

	// Calculate reminder time
	reminderTime := event.StartTime.Add(-1 * time.Duration(hours) * time.Hour)

	// If reminder time is in the past, don't create reminders
	if reminderTime.Before(time.Now()) {
		return 0, errors.New(errors.CodeInvalidArgument, "Reminder time is in the past")
	}

	// Get attendees with RSVP "going"
	attendees, err := s.attendeeRepo.FindByEventIDAndRSVP(ctx, eventID, "going")
	if err != nil {
		return 0, errors.Wrap(err, "Failed to find attendees")
	}

	// Create reminders for each attendee
	count := 0
	for _, attendee := range attendees {
		// Check if reminder already exists
		existing, err := s.reminderRepo.FindByEventAndUser(ctx, eventID, attendee.UserID)
		if err == nil && existing != nil {
			// Update existing reminder
			existing.ReminderTime = reminderTime
			existing.Status = "pending"
			existing.SentAt = nil

			err = s.reminderRepo.Update(ctx, existing)
			if err != nil {
				s.logger.Warn("Failed to update reminder", "reminderId", existing.ID.Hex(), "error", err)
				continue
			}
		} else {
			// Create new reminder
			reminder := &models.EventReminder{
				EventID:      eventID,
				UserID:       attendee.UserID,
				ReminderTime: reminderTime,
				Type:         "push",
				Status:       "pending",
				CreatedAt:    time.Now(),
			}

			_, err = s.reminderRepo.Create(ctx, reminder)
			if err != nil {
				s.logger.Warn("Failed to create reminder", "eventId", eventID.Hex(), "userId", attendee.UserID.Hex(), "error", err)
				continue
			}
		}

		count++
	}

	return count, nil
}

// Helper functions

// isValidReminderType checks if a reminder type is valid
func isValidReminderType(reminderType string) bool {
	validTypes := map[string]bool{
		"email": true,
		"push":  true,
		"sms":   true,
	}

	return validTypes[reminderType]
}

// timePtr returns a pointer to a time.Time
func timePtr(t time.Time) *time.Time {
	return &t
}
