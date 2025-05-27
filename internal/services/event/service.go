package event

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/Caqil/vyrall/internal/internal/models"
	"github.com/Caqil/vyrall/internal/pkg/config"
	"github.com/Caqil/vyrall/internal/pkg/errors"
	"github.com/Caqil/vyrall/internal/pkg/logging"
	"github.com/Caqil/vyrall/internal/pkg/metrics"
)

// Service defines the interface for event-related operations
type Service interface {
	// Event management
	CreateEvent(ctx context.Context, event *models.Event) (*models.Event, error)
	UpdateEvent(ctx context.Context, eventID primitive.ObjectID, updates *EventUpdates, userID primitive.ObjectID) (*models.Event, error)
	DeleteEvent(ctx context.Context, eventID primitive.ObjectID, userID primitive.ObjectID) error
	GetEvent(ctx context.Context, eventID primitive.ObjectID) (*models.Event, error)
	GetEvents(ctx context.Context, filter *EventFilter, page, limit int) ([]models.Event, int, error)
	GetEventsByUser(ctx context.Context, userID primitive.ObjectID, status string, page, limit int) ([]models.Event, int, error)
	GetEventsAttendedByUser(ctx context.Context, userID primitive.ObjectID, rsvpStatus string, page, limit int) ([]models.Event, int, error)
	UpdateEventStatus(ctx context.Context, eventID primitive.ObjectID, status string, userID primitive.ObjectID) error
	AddCoHost(ctx context.Context, eventID, hostID, coHostID primitive.ObjectID) error
	RemoveCoHost(ctx context.Context, eventID, hostID, coHostID primitive.ObjectID) error
	CheckInAttendee(ctx context.Context, eventID, attendeeID, checkerID primitive.ObjectID) error
	CreateRecurringEvent(ctx context.Context, event *models.Event, recurrenceRule string, occurrences int) ([]models.Event, error)

	// RSVP management
	RSVP(ctx context.Context, eventID, userID primitive.ObjectID, rsvpStatus string, guestCount int) (*models.EventAttendee, error)
	GetAttendee(ctx context.Context, eventID, userID primitive.ObjectID) (*models.EventAttendee, error)
	GetAttendees(ctx context.Context, eventID primitive.ObjectID, rsvpStatus string, page, limit int) ([]models.EventAttendee, int, error)
	GetWaitlist(ctx context.Context, eventID primitive.ObjectID, page, limit int) ([]models.EventAttendee, int, error)
	RemoveAttendee(ctx context.Context, eventID, userID, removerID primitive.ObjectID) error
	InviteToEvent(ctx context.Context, eventID, userID, inviterID primitive.ObjectID) error
	UpdateGuestCount(ctx context.Context, eventID, userID primitive.ObjectID, guestCount int) error

	// Reminders
	CreateReminder(ctx context.Context, userID, eventID primitive.ObjectID, reminderTime time.Time, reminderType string) (*models.EventReminder, error)
	UpdateReminder(ctx context.Context, reminderID, userID primitive.ObjectID, reminderTime time.Time, reminderType string) (*models.EventReminder, error)
	DeleteReminder(ctx context.Context, reminderID, userID primitive.ObjectID) error
	GetReminder(ctx context.Context, reminderID, userID primitive.ObjectID) (*models.EventReminder, error)
	GetRemindersByUser(ctx context.Context, userID primitive.ObjectID) ([]models.EventReminder, error)
	GetRemindersByEvent(ctx context.Context, eventID primitive.ObjectID) ([]models.EventReminder, error)
	ProcessDueReminders(ctx context.Context) (int, error)
	CreateDefaultReminders(ctx context.Context, userID primitive.ObjectID) (int, error)
	CreateEventReminders(ctx context.Context, eventID primitive.ObjectID, hours int) (int, error)

	// Analytics
	RecordEventView(ctx context.Context, eventID primitive.ObjectID, userID *primitive.ObjectID, sessionID string) error
	RecordEventShare(ctx context.Context, eventID primitive.ObjectID, userID primitive.ObjectID, platform string) error
	RecordTicketSale(ctx context.Context, eventID primitive.ObjectID, ticketType string, price float64) error
	RecordCheckIn(ctx context.Context, eventID primitive.ObjectID, userID primitive.ObjectID) error
	GetEventAnalytics(ctx context.Context, eventID primitive.ObjectID) (*models.EventAnalytics, error)
	GetEventDemographics(ctx context.Context, eventID primitive.ObjectID) (map[string]interface{}, error)
	GetEventTrends(ctx context.Context, timeframe string, limit int) ([]EventTrend, error)
	GetEventAttendeeActivity(ctx context.Context, eventID, userID primitive.ObjectID) (*AttendeeActivity, error)

	// Notifications
	NotifyEventCreated(ctx context.Context, event *models.Event) error
	NotifyEventUpdated(ctx context.Context, event *models.Event, changes []string) error
	NotifyEventCancelled(ctx context.Context, event *models.Event) error
	NotifyEventStartingSoon(ctx context.Context, event *models.Event, minutesBefore int) error
	NotifyNewRSVP(ctx context.Context, event *models.Event, attendee *models.EventAttendee) error
	NotifyEventInvitation(ctx context.Context, event *models.Event, inviteeID, inviterID primitive.ObjectID) error
	NotifyEventWaitlistStatusChanged(ctx context.Context, event *models.Event, attendee *models.EventAttendee, promoted bool) error
	NotifyEventCheckIn(ctx context.Context, event *models.Event, attendee *models.EventAttendee) error
	SendEventReminder(ctx context.Context, reminder *models.EventReminder) error
}

// EventService implements the Service interface
type EventService struct {
	managementSvc   *ManagementService
	rsvpSvc         *RSVPService
	reminderSvc     *RemindersService
	analyticsSvc    *AnalyticsService
	notificationSvc *NotificationsService
	eventRepo       EventRepository
	attendeeRepo    AttendeeRepository
	reminderRepo    ReminderRepository
	analyticsRepo   EventAnalyticsRepository
	config          *config.EventConfig
	metrics         metrics.Collector
	logger          logging.Logger
}

// NewEventService creates a new event service
func NewEventService(
	managementSvc *ManagementService,
	rsvpSvc *RSVPService,
	reminderSvc *RemindersService,
	analyticsSvc *AnalyticsService,
	notificationSvc *NotificationsService,
	eventRepo EventRepository,
	attendeeRepo AttendeeRepository,
	reminderRepo ReminderRepository,
	analyticsRepo EventAnalyticsRepository,
	config *config.EventConfig,
	metrics metrics.Collector,
	logger logging.Logger,
) Service {
	return &EventService{
		managementSvc:   managementSvc,
		rsvpSvc:         rsvpSvc,
		reminderSvc:     reminderSvc,
		analyticsSvc:    analyticsSvc,
		notificationSvc: notificationSvc,
		eventRepo:       eventRepo,
		attendeeRepo:    attendeeRepo,
		reminderRepo:    reminderRepo,
		analyticsRepo:   analyticsRepo,
		config:          config,
		metrics:         metrics,
		logger:          logger,
	}
}

// CreateEvent creates a new event
func (s *EventService) CreateEvent(ctx context.Context, event *models.Event) (*models.Event, error) {
	startTime := time.Now()
	defer func() {
		s.metrics.ObserveLatency("event.create", time.Since(startTime))
	}()

	// Create event
	createdEvent, err := s.managementSvc.CreateEvent(ctx, event)
	if err != nil {
		return nil, err
	}

	// Send notifications
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		err := s.notificationSvc.NotifyEventCreated(ctx, createdEvent)
		if err != nil {
			s.logger.Warn("Failed to send event creation notifications", "eventId", createdEvent.ID.Hex(), "error", err)
		}
	}()

	s.metrics.IncrementCounter("event.created")
	return createdEvent, nil
}

// UpdateEvent updates an existing event
func (s *EventService) UpdateEvent(ctx context.Context, eventID primitive.ObjectID, updates *EventUpdates, userID primitive.ObjectID) (*models.Event, error) {
	startTime := time.Now()
	defer func() {
		s.metrics.ObserveLatency("event.update", time.Since(startTime))
	}()

	// Get original event for tracking changes
	originalEvent, err := s.eventRepo.FindByID(ctx, eventID)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to find event")
	}

	// Update event
	updatedEvent, err := s.managementSvc.UpdateEvent(ctx, eventID, updates, userID)
	if err != nil {
		return nil, err
	}

	// Track significant changes for notifications
	changes := detectChanges(originalEvent, updatedEvent)

	// Send notifications if there are significant changes
	if len(changes) > 0 {
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			err := s.notificationSvc.NotifyEventUpdated(ctx, updatedEvent, changes)
			if err != nil {
				s.logger.Warn("Failed to send event update notifications", "eventId", updatedEvent.ID.Hex(), "error", err)
			}
		}()
	}

	s.metrics.IncrementCounter("event.updated")
	return updatedEvent, nil
}

// DeleteEvent deletes an event
func (s *EventService) DeleteEvent(ctx context.Context, eventID primitive.ObjectID, userID primitive.ObjectID) error {
	startTime := time.Now()
	defer func() {
		s.metrics.ObserveLatency("event.delete", time.Since(startTime))
	}()

	// Get event before deletion for notifications
	event, err := s.eventRepo.FindByID(ctx, eventID)
	if err != nil {
		return errors.Wrap(err, "Failed to find event")
	}

	// Delete event (actually marks as cancelled)
	err = s.managementSvc.DeleteEvent(ctx, eventID, userID)
	if err != nil {
		return err
	}

	// Send cancellation notifications
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		err := s.notificationSvc.NotifyEventCancelled(ctx, event)
		if err != nil {
			s.logger.Warn("Failed to send event cancellation notifications", "eventId", eventID.Hex(), "error", err)
		}
	}()

	s.metrics.IncrementCounter("event.deleted")
	return nil
}

// GetEvent retrieves an event by ID
func (s *EventService) GetEvent(ctx context.Context, eventID primitive.ObjectID) (*models.Event, error) {
	return s.managementSvc.GetEvent(ctx, eventID)
}

// GetEvents retrieves events with filters and pagination
func (s *EventService) GetEvents(ctx context.Context, filter *EventFilter, page, limit int) ([]models.Event, int, error) {
	return s.managementSvc.GetEvents(ctx, filter, page, limit)
}

// GetEventsByUser retrieves events for a user
func (s *EventService) GetEventsByUser(ctx context.Context, userID primitive.ObjectID, status string, page, limit int) ([]models.Event, int, error) {
	return s.managementSvc.GetEventsByUser(ctx, userID, status, page, limit)
}

// GetEventsAttendedByUser retrieves events a user is attending
func (s *EventService) GetEventsAttendedByUser(ctx context.Context, userID primitive.ObjectID, rsvpStatus string, page, limit int) ([]models.Event, int, error) {
	return s.managementSvc.GetEventsAttendedByUser(ctx, userID, rsvpStatus, page, limit)
}

// UpdateEventStatus updates the status of an event
func (s *EventService) UpdateEventStatus(ctx context.Context, eventID primitive.ObjectID, status string, userID primitive.ObjectID) error {
	startTime := time.Now()
	defer func() {
		s.metrics.ObserveLatency("event.updateStatus", time.Since(startTime))
	}()

	// Get original event for notifications
	originalEvent, err := s.eventRepo.FindByID(ctx, eventID)
	if err != nil {
		return errors.Wrap(err, "Failed to find event")
	}

	// Update status
	err = s.managementSvc.UpdateEventStatus(ctx, eventID, status, userID)
	if err != nil {
		return err
	}

	// Send notifications if status changed to cancelled
	if status == "cancelled" && originalEvent.Status != "cancelled" {
		// Refresh event with updated status
		updatedEvent, err := s.eventRepo.FindByID(ctx, eventID)
		if err != nil {
			s.logger.Warn("Failed to fetch updated event", "eventId", eventID.Hex(), "error", err)
			return nil
		}

		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			err := s.notificationSvc.NotifyEventCancelled(ctx, updatedEvent)
			if err != nil {
				s.logger.Warn("Failed to send event cancellation notifications", "eventId", eventID.Hex(), "error", err)
			}
		}()
	}

	s.metrics.IncrementCounter("event.statusUpdated")
	return nil
}

// AddCoHost adds a co-host to an event
func (s *EventService) AddCoHost(ctx context.Context, eventID, hostID, coHostID primitive.ObjectID) error {
	return s.managementSvc.AddCoHost(ctx, eventID, hostID, coHostID)
}

// RemoveCoHost removes a co-host from an event
func (s *EventService) RemoveCoHost(ctx context.Context, eventID, hostID, coHostID primitive.ObjectID) error {
	return s.managementSvc.RemoveCoHost(ctx, eventID, hostID, coHostID)
}

// CheckInAttendee checks in an attendee at an event
func (s *EventService) CheckInAttendee(ctx context.Context, eventID, attendeeID, checkerID primitive.ObjectID) error {
	startTime := time.Now()
	defer func() {
		s.metrics.ObserveLatency("event.checkIn", time.Since(startTime))
	}()

	// Get attendee before check-in for notifications
	attendee, err := s.attendeeRepo.FindByID(ctx, attendeeID)
	if err != nil {
		return errors.Wrap(err, "Failed to find attendee")
	}

	// Check in attendee
	err = s.managementSvc.CheckInAttendee(ctx, eventID, attendeeID, checkerID)
	if err != nil {
		return err
	}

	// Record check-in for analytics
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		err := s.analyticsSvc.RecordCheckIn(ctx, eventID, attendee.UserID)
		if err != nil {
			s.logger.Warn("Failed to record check-in analytics", "eventId", eventID.Hex(), "userId", attendee.UserID.Hex(), "error", err)
		}
	}()

	// Send check-in notification
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Get event for notification
		event, err := s.eventRepo.FindByID(ctx, eventID)
		if err != nil {
			s.logger.Warn("Failed to find event for notification", "eventId", eventID.Hex(), "error", err)
			return
		}

		// Refresh attendee with check-in info
		updatedAttendee, err := s.attendeeRepo.FindByID(ctx, attendeeID)
		if err != nil {
			s.logger.Warn("Failed to find updated attendee", "attendeeId", attendeeID.Hex(), "error", err)
			return
		}

		err = s.notificationSvc.NotifyEventCheckIn(ctx, event, updatedAttendee)
		if err != nil {
			s.logger.Warn("Failed to send check-in notification", "attendeeId", attendeeID.Hex(), "error", err)
		}
	}()

	s.metrics.IncrementCounter("event.checkIn")
	return nil
}

// CreateRecurringEvent creates a series of recurring events
func (s *EventService) CreateRecurringEvent(ctx context.Context, event *models.Event, recurrenceRule string, occurrences int) ([]models.Event, error) {
	startTime := time.Now()
	defer func() {
		s.metrics.ObserveLatency("event.createRecurring", time.Since(startTime))
	}()

	events, err := s.managementSvc.CreateRecurringEvent(ctx, event, recurrenceRule, occurrences)
	if err != nil {
		return nil, err
	}

	s.metrics.IncrementCounter("event.recurringCreated")
	return events, nil
}

// RSVP creates or updates an RSVP for an event
func (s *EventService) RSVP(ctx context.Context, eventID, userID primitive.ObjectID, rsvpStatus string, guestCount int) (*models.EventAttendee, error) {
	startTime := time.Now()
	defer func() {
		s.metrics.ObserveLatency("event.rsvp", time.Since(startTime))
	}()

	attendee, err := s.rsvpSvc.RSVP(ctx, eventID, userID, rsvpStatus, guestCount)
	if err != nil {
		return nil, err
	}

	s.metrics.IncrementCounter("event.rsvp")
	return attendee, nil
}

// GetAttendee retrieves an attendee record
func (s *EventService) GetAttendee(ctx context.Context, eventID, userID primitive.ObjectID) (*models.EventAttendee, error) {
	return s.rsvpSvc.GetAttendee(ctx, eventID, userID)
}

// GetAttendees retrieves attendees for an event
func (s *EventService) GetAttendees(ctx context.Context, eventID primitive.ObjectID, rsvpStatus string, page, limit int) ([]models.EventAttendee, int, error) {
	return s.rsvpSvc.GetAttendees(ctx, eventID, rsvpStatus, page, limit)
}

// GetWaitlist retrieves waitlisted attendees for an event
func (s *EventService) GetWaitlist(ctx context.Context, eventID primitive.ObjectID, page, limit int) ([]models.EventAttendee, int, error) {
	return s.rsvpSvc.GetWaitlist(ctx, eventID, page, limit)
}

// RemoveAttendee removes an attendee from an event
func (s *EventService) RemoveAttendee(ctx context.Context, eventID, userID, removerID primitive.ObjectID) error {
	return s.rsvpSvc.RemoveAttendee(ctx, eventID, userID, removerID)
}

// InviteToEvent invites a user to an event
func (s *EventService) InviteToEvent(ctx context.Context, eventID, userID, inviterID primitive.ObjectID) error {
	return s.rsvpSvc.InviteToEvent(ctx, eventID, userID, inviterID)
}

// UpdateGuestCount updates the number of guests for an attendee
func (s *EventService) UpdateGuestCount(ctx context.Context, eventID, userID primitive.ObjectID, guestCount int) error {
	return s.rsvpSvc.UpdateGuestCount(ctx, eventID, userID, guestCount)
}

// CreateReminder creates a new event reminder
func (s *EventService) CreateReminder(ctx context.Context, userID, eventID primitive.ObjectID, reminderTime time.Time, reminderType string) (*models.EventReminder, error) {
	return s.reminderSvc.CreateReminder(ctx, userID, eventID, reminderTime, reminderType)
}

// UpdateReminder updates an existing reminder
func (s *EventService) UpdateReminder(ctx context.Context, reminderID, userID primitive.ObjectID, reminderTime time.Time, reminderType string) (*models.EventReminder, error) {
	return s.reminderSvc.UpdateReminder(ctx, reminderID, userID, reminderTime, reminderType)
}

// DeleteReminder deletes a reminder
func (s *EventService) DeleteReminder(ctx context.Context, reminderID, userID primitive.ObjectID) error {
	return s.reminderSvc.DeleteReminder(ctx, reminderID, userID)
}

// GetReminder retrieves a reminder by ID
func (s *EventService) GetReminder(ctx context.Context, reminderID, userID primitive.ObjectID) (*models.EventReminder, error) {
	return s.reminderSvc.GetReminder(ctx, reminderID, userID)
}

// GetRemindersByUser retrieves reminders for a user
func (s *EventService) GetRemindersByUser(ctx context.Context, userID primitive.ObjectID) ([]models.EventReminder, error) {
	return s.reminderSvc.GetRemindersByUser(ctx, userID)
}

// GetRemindersByEvent retrieves reminders for an event
func (s *EventService) GetRemindersByEvent(ctx context.Context, eventID primitive.ObjectID) ([]models.EventReminder, error) {
	return s.reminderSvc.GetRemindersByEvent(ctx, eventID)
}

// ProcessDueReminders processes reminders that are due to be sent
func (s *EventService) ProcessDueReminders(ctx context.Context) (int, error) {
	startTime := time.Now()
	defer func() {
		s.metrics.ObserveLatency("event.processDueReminders", time.Since(startTime))
	}()

	count, err := s.reminderSvc.ProcessDueReminders(ctx)
	if err != nil {
		return 0, err
	}

	s.metrics.AddCounter("event.remindersSent", count)
	return count, nil
}

// CreateDefaultReminders creates default reminders for a user's upcoming events
func (s *EventService) CreateDefaultReminders(ctx context.Context, userID primitive.ObjectID) (int, error) {
	return s.reminderSvc.CreateDefaultReminders(ctx, userID)
}

// CreateEventReminders creates reminders for all attendees of an event
func (s *EventService) CreateEventReminders(ctx context.Context, eventID primitive.ObjectID, hours int) (int, error) {
	startTime := time.Now()
	defer func() {
		s.metrics.ObserveLatency("event.createEventReminders", time.Since(startTime))
	}()

	count, err := s.reminderSvc.CreateEventReminders(ctx, eventID, hours)
	if err != nil {
		return 0, err
	}

	s.metrics.AddCounter("event.remindersCreated", count)
	return count, nil
}

// RecordEventView records a view of an event
func (s *EventService) RecordEventView(ctx context.Context, eventID primitive.ObjectID, userID *primitive.ObjectID, sessionID string) error {
	return s.analyticsSvc.RecordEventView(ctx, eventID, userID, sessionID)
}

// RecordEventShare records a share of an event
func (s *EventService) RecordEventShare(ctx context.Context, eventID primitive.ObjectID, userID primitive.ObjectID, platform string) error {
	startTime := time.Now()
	defer func() {
		s.metrics.ObserveLatency("event.recordShare", time.Since(startTime))
	}()

	err := s.analyticsSvc.RecordEventShare(ctx, eventID, userID, platform)
	if err != nil {
		return err
	}

	s.metrics.IncrementCounter("event.shared")
	return nil
}

// RecordTicketSale records a ticket sale for an event
func (s *EventService) RecordTicketSale(ctx context.Context, eventID primitive.ObjectID, ticketType string, price float64) error {
	startTime := time.Now()
	defer func() {
		s.metrics.ObserveLatency("event.recordTicketSale", time.Since(startTime))
	}()

	err := s.analyticsSvc.RecordTicketSale(ctx, eventID, ticketType, price)
	if err != nil {
		return err
	}

	s.metrics.IncrementCounter("event.ticketSold")
	return nil
}

// RecordCheckIn records an attendee check-in at an event
func (s *EventService) RecordCheckIn(ctx context.Context, eventID primitive.ObjectID, userID primitive.ObjectID) error {
	return s.analyticsSvc.RecordCheckIn(ctx, eventID, userID)
}

// GetEventAnalytics retrieves analytics for an event
func (s *EventService) GetEventAnalytics(ctx context.Context, eventID primitive.ObjectID) (*models.EventAnalytics, error) {
	return s.analyticsSvc.GetEventAnalytics(ctx, eventID)
}

// GetEventDemographics retrieves demographic information about attendees
func (s *EventService) GetEventDemographics(ctx context.Context, eventID primitive.ObjectID) (map[string]interface{}, error) {
	return s.analyticsSvc.GetEventDemographics(ctx, eventID)
}

// GetEventTrends retrieves trending events
func (s *EventService) GetEventTrends(ctx context.Context, timeframe string, limit int) ([]EventTrend, error) {
	return s.analyticsSvc.GetEventTrends(ctx, timeframe, limit)
}

// GetEventAttendeeActivity retrieves activity of an attendee at an event
func (s *EventService) GetEventAttendeeActivity(ctx context.Context, eventID, userID primitive.ObjectID) (*AttendeeActivity, error) {
	return s.analyticsSvc.GetEventAttendeeActivity(ctx, eventID, userID)
}

// NotifyEventCreated sends notifications when an event is created
func (s *EventService) NotifyEventCreated(ctx context.Context, event *models.Event) error {
	return s.notificationSvc.NotifyEventCreated(ctx, event)
}

// NotifyEventUpdated sends notifications when an event is updated
func (s *EventService) NotifyEventUpdated(ctx context.Context, event *models.Event, changes []string) error {
	return s.notificationSvc.NotifyEventUpdated(ctx, event, changes)
}

// NotifyEventCancelled sends notifications when an event is cancelled
func (s *EventService) NotifyEventCancelled(ctx context.Context, event *models.Event) error {
	return s.notificationSvc.NotifyEventCancelled(ctx, event)
}

// NotifyEventStartingSoon sends notifications for upcoming events
func (s *EventService) NotifyEventStartingSoon(ctx context.Context, event *models.Event, minutesBefore int) error {
	return s.notificationSvc.NotifyEventStartingSoon(ctx, event, minutesBefore)
}

// NotifyNewRSVP sends a notification to the host when someone RSVPs
func (s *EventService) NotifyNewRSVP(ctx context.Context, event *models.Event, attendee *models.EventAttendee) error {
	return s.notificationSvc.NotifyNewRSVP(ctx, event, attendee)
}

// NotifyEventInvitation sends a notification when a user is invited to an event
func (s *EventService) NotifyEventInvitation(ctx context.Context, event *models.Event, inviteeID, inviterID primitive.ObjectID) error {
	return s.notificationSvc.NotifyEventInvitation(ctx, event, inviteeID, inviterID)
}

// NotifyEventWaitlistStatusChanged sends a notification when a user's waitlist status changes
func (s *EventService) NotifyEventWaitlistStatusChanged(ctx context.Context, event *models.Event, attendee *models.EventAttendee, promoted bool) error {
	return s.notificationSvc.NotifyEventWaitlistStatusChanged(ctx, event, attendee, promoted)
}

// NotifyEventCheckIn sends a notification to the attendee when they're checked in
func (s *EventService) NotifyEventCheckIn(ctx context.Context, event *models.Event, attendee *models.EventAttendee) error {
	return s.notificationSvc.NotifyEventCheckIn(ctx, event, attendee)
}

// SendEventReminder sends a reminder notification for an event
func (s *EventService) SendEventReminder(ctx context.Context, reminder *models.EventReminder) error {
	return s.notificationSvc.SendEventReminder(ctx, reminder)
}

// Repository interfaces

// EventRepository defines operations for event data access
type EventRepository interface {
	FindByID(ctx context.Context, id primitive.ObjectID) (*models.Event, error)
	FindByIDs(ctx context.Context, ids []primitive.ObjectID) ([]models.Event, error)
	FindWithFilter(ctx context.Context, filter map[string]interface{}, page, limit int, sortBy, sortOrder string) ([]models.Event, int, error)
	FindTrending(ctx context.Context, since time.Time, limit int) ([]models.Event, error)
	Create(ctx context.Context, event *models.Event) (*models.Event, error)
	Update(ctx context.Context, event *models.Event) error
	Delete(ctx context.Context, id primitive.ObjectID) error
	IncrementViewCount(ctx context.Context, id primitive.ObjectID) error
	IncrementShareCount(ctx context.Context, id primitive.ObjectID) error
	Count(ctx context.Context, filter map[string]interface{}) (int, error)
}

// AttendeeRepository defines operations for event attendee data access
type AttendeeRepository interface {
	FindByID(ctx context.Context, id primitive.ObjectID) (*models.EventAttendee, error)
	FindByEventID(ctx context.Context, eventID primitive.ObjectID) ([]models.EventAttendee, error)
	FindByEventIDWithPagination(ctx context.Context, eventID primitive.ObjectID, page, limit int) ([]models.EventAttendee, int, error)
	FindByEventIDAndRSVP(ctx context.Context, eventID primitive.ObjectID, rsvp string) ([]models.EventAttendee, error)
	FindByEventIDAndRSVPWithPagination(ctx context.Context, eventID primitive.ObjectID, rsvp string, page, limit int) ([]models.EventAttendee, int, error)
	FindByEventAndUser(ctx context.Context, eventID, userID primitive.ObjectID) (*models.EventAttendee, error)
	FindByUserID(ctx context.Context, userID primitive.ObjectID) ([]models.EventAttendee, error)
	FindByUserIDAndRSVP(ctx context.Context, userID primitive.ObjectID, rsvp string, page, limit int) ([]models.EventAttendee, int, error)
	FindWaitlistedWithPagination(ctx context.Context, eventID primitive.ObjectID, page, limit int) ([]models.EventAttendee, int, error)
	Create(ctx context.Context, attendee *models.EventAttendee) (*models.EventAttendee, error)
	Update(ctx context.Context, attendee *models.EventAttendee) error
	Delete(ctx context.Context, id primitive.ObjectID) error
	CountByEventIDAndRSVP(ctx context.Context, eventID primitive.ObjectID, rsvp string) (int, error)
	CountWaitlisted(ctx context.Context, eventID primitive.ObjectID) (int, error)
	CountCheckedIn(ctx context.Context, eventID primitive.ObjectID) (int, error)
	CountWithFilter(ctx context.Context, filter map[string]interface{}) (int, error)
}

// ReminderRepository defines operations for event reminder data access
type ReminderRepository interface {
	FindByID(ctx context.Context, id primitive.ObjectID) (*models.EventReminder, error)
	FindByEventID(ctx context.Context, eventID primitive.ObjectID) ([]models.EventReminder, error)
	FindByUserID(ctx context.Context, userID primitive.ObjectID) ([]models.EventReminder, error)
	FindByEventAndUser(ctx context.Context, eventID, userID primitive.ObjectID) (*models.EventReminder, error)
	FindDueReminders(ctx context.Context, before time.Time) ([]models.EventReminder, error)
	Create(ctx context.Context, reminder *models.EventReminder) (*models.EventReminder, error)
	Update(ctx context.Context, reminder *models.EventReminder) error
	Delete(ctx context.Context, id primitive.ObjectID) error
	Count(ctx context.Context, filter map[string]interface{}) (int, error)
}

// EventAnalyticsRepository defines operations for event analytics data access
type EventAnalyticsRepository interface {
	FindByEventID(ctx context.Context, eventID primitive.ObjectID) (*models.EventAnalytics, error)
	Create(ctx context.Context, analytics *models.EventAnalytics) (*models.EventAnalytics, error)
	Update(ctx context.Context, analytics *models.EventAnalytics) error
	IsNewViewer(ctx context.Context, eventID, userID primitive.ObjectID) (bool, error)
	IsNewSession(ctx context.Context, eventID primitive.ObjectID, sessionID string) (bool, error)
	AddViewer(ctx context.Context, eventID, userID primitive.ObjectID) error
	AddSession(ctx context.Context, eventID primitive.ObjectID, sessionID string) error
}

// Helper functions

// detectChanges identifies significant changes between event versions
func detectChanges(original, updated *models.Event) []string {
	changes := []string{}

	// Check for title change
	if original.Title != updated.Title {
		changes = append(changes, "title")
	}

	// Check for date/time changes
	if !original.StartTime.Equal(updated.StartTime) {
		changes = append(changes, "start time")
	}

	if !original.EndTime.Equal(updated.EndTime) {
		changes = append(changes, "end time")
	}

	// Check for location change
	if original.Location.Name != updated.Location.Name ||
		original.Location.Address != updated.Location.Address ||
		original.Location.City != updated.Location.City {
		changes = append(changes, "location")
	}

	// Check for type change
	if original.Type != updated.Type {
		changes = append(changes, "event type")
	}

	// Check for privacy change
	if original.Privacy != updated.Privacy {
		changes = append(changes, "privacy setting")
	}

	// Check for capacity change
	if original.MaxAttendees != updated.MaxAttendees {
		changes = append(changes, "capacity")
	}

	// Check for URL change
	if original.URL != updated.URL {
		changes = append(changes, "event URL")
	}

	// Check for ticket changes
	if (original.TicketInfo == nil && updated.TicketInfo != nil) ||
		(original.TicketInfo != nil && updated.TicketInfo == nil) {
		changes = append(changes, "ticket information")
	} else if original.TicketInfo != nil && updated.TicketInfo != nil {
		if original.TicketInfo.IsPaid != updated.TicketInfo.IsPaid ||
			original.TicketInfo.Currency != updated.TicketInfo.Currency ||
			!original.TicketInfo.SalesStartAt.Equal(updated.TicketInfo.SalesStartAt) ||
			!original.TicketInfo.SalesEndAt.Equal(updated.TicketInfo.SalesEndAt) {
			changes = append(changes, "ticket information")
		}
	}

	return changes
}
