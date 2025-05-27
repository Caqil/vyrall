package event

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/Caqil/vyrall/internal/internal/models"
	"github.com/Caqil/vyrall/internal/pkg/errors"
	"github.com/Caqil/vyrall/internal/pkg/logging"
)

// ManagementService handles event management operations
type ManagementService struct {
	eventRepo    EventRepository
	attendeeRepo AttendeeRepository
	userRepo     UserRepository
	groupRepo    GroupRepository
	logger       logging.Logger
}

// NewManagementService creates a new event management service
func NewManagementService(
	eventRepo EventRepository,
	attendeeRepo AttendeeRepository,
	userRepo UserRepository,
	groupRepo GroupRepository,
	logger logging.Logger,
) *ManagementService {
	return &ManagementService{
		eventRepo:    eventRepo,
		attendeeRepo: attendeeRepo,
		userRepo:     userRepo,
		groupRepo:    groupRepo,
		logger:       logger,
	}
}

// CreateEvent creates a new event
func (s *ManagementService) CreateEvent(ctx context.Context, event *models.Event) (*models.Event, error) {
	// Validate host exists
	host, err := s.userRepo.FindByID(ctx, event.HostID)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to find host")
	}

	// Validate co-hosts exist
	for _, coHostID := range event.CoHosts {
		_, err := s.userRepo.FindByID(ctx, coHostID)
		if err != nil {
			s.logger.Warn("Co-host not found", "userId", coHostID.Hex(), "error", err)
			// Don't fail creation, just log warning
		}
	}

	// If group event, validate group exists
	if event.GroupID != nil {
		group, err := s.groupRepo.FindByID(ctx, *event.GroupID)
		if err != nil {
			return nil, errors.Wrap(err, "Failed to find group")
		}

		// Check if user is group admin or has permission to create events
		isAdmin := false
		for _, adminID := range group.Admins {
			if adminID == event.HostID {
				isAdmin = true
				break
			}
		}

		if !isAdmin && host.Role != "admin" {
			return nil, errors.New(errors.CodeForbidden, "Only group admins can create group events")
		}
	}

	// Set creation time
	now := time.Now()
	event.CreatedAt = now
	event.UpdatedAt = now

	// Initialize counts
	event.RSVPCount = models.EventRSVPCounts{
		Going:      0,
		Interested: 0,
		NotGoing:   0,
		NoReply:    0,
		Waitlist:   0,
	}

	// Set status based on start time
	if event.StartTime.Before(now) {
		event.Status = "live"
	} else {
		event.Status = "scheduled"
	}

	// Create event
	createdEvent, err := s.eventRepo.Create(ctx, event)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create event")
	}

	// Auto-add host as "going"
	attendee := &models.EventAttendee{
		EventID:       createdEvent.ID,
		UserID:        event.HostID,
		RSVP:          "going",
		RSVPTimestamp: now,
		CheckedIn:     false,
		GuestCount:    0,
	}

	_, err = s.attendeeRepo.Create(ctx, attendee)
	if err != nil {
		s.logger.Warn("Failed to add host as attendee", "eventId", createdEvent.ID.Hex(), "hostId", event.HostID.Hex(), "error", err)
	} else {
		// Update RSVP count
		createdEvent.RSVPCount.Going++
		err = s.eventRepo.Update(ctx, createdEvent)
		if err != nil {
			s.logger.Warn("Failed to update RSVP count", "eventId", createdEvent.ID.Hex(), "error", err)
		}
	}

	// Update group event count if applicable
	if event.GroupID != nil {
		err = s.groupRepo.IncrementEventCount(ctx, *event.GroupID)
		if err != nil {
			s.logger.Warn("Failed to increment group event count", "groupId", event.GroupID.Hex(), "error", err)
		}
	}

	return createdEvent, nil
}

// UpdateEvent updates an existing event
func (s *ManagementService) UpdateEvent(ctx context.Context, eventID primitive.ObjectID, updates *EventUpdates, userID primitive.ObjectID) (*models.Event, error) {
	// Get existing event
	event, err := s.eventRepo.FindByID(ctx, eventID)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to find event")
	}

	// Check permissions
	if event.HostID != userID {
		isCoHost := false
		for _, coHostID := range event.CoHosts {
			if coHostID == userID {
				isCoHost = true
				break
			}
		}

		if !isCoHost {
			// Check if user is admin
			user, err := s.userRepo.FindByID(ctx, userID)
			if err != nil || user.Role != "admin" {
				return nil, errors.New(errors.CodeForbidden, "Only host or co-hosts can update this event")
			}
		}
	}

	// Apply updates
	if updates.Title != "" {
		event.Title = updates.Title
	}

	if updates.Description != "" {
		event.Description = updates.Description
	}

	if updates.StartTime != nil {
		event.StartTime = *updates.StartTime
	}

	if updates.EndTime != nil {
		event.EndTime = *updates.EndTime
	}

	if updates.TimeZone != "" {
		event.TimeZone = updates.TimeZone
	}

	if updates.Location != nil {
		event.Location = *updates.Location
	}

	if updates.CoverImage != "" {
		event.CoverImage = updates.CoverImage
	}

	if updates.Type != "" {
		event.Type = updates.Type
	}

	if updates.Privacy != "" {
		event.Privacy = updates.Privacy
	}

	if updates.Category != "" {
		event.Category = updates.Category
	}

	if updates.Tags != nil {
		event.Tags = updates.Tags
	}

	if updates.MaxAttendees > 0 {
		event.MaxAttendees = updates.MaxAttendees
	}

	if updates.URL != "" {
		event.URL = updates.URL
	}

	if updates.TicketInfo != nil {
		event.TicketInfo = updates.TicketInfo
	}

	// Update timestamp
	event.UpdatedAt = time.Now()

	// Update event
	err = s.eventRepo.Update(ctx, event)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to update event")
	}

	return event, nil
}

// DeleteEvent deletes an event
func (s *ManagementService) DeleteEvent(ctx context.Context, eventID primitive.ObjectID, userID primitive.ObjectID) error {
	// Get event
	event, err := s.eventRepo.FindByID(ctx, eventID)
	if err != nil {
		return errors.Wrap(err, "Failed to find event")
	}

	// Check permissions
	if event.HostID != userID {
		// Check if user is admin
		user, err := s.userRepo.FindByID(ctx, userID)
		if err != nil || user.Role != "admin" {
			return errors.New(errors.CodeForbidden, "Only the host can delete this event")
		}
	}

	// Cancel event instead of deleting
	event.Status = "cancelled"
	event.UpdatedAt = time.Now()

	// Update event
	err = s.eventRepo.Update(ctx, event)
	if err != nil {
		return errors.Wrap(err, "Failed to cancel event")
	}

	// Decrement group event count if applicable
	if event.GroupID != nil {
		err = s.groupRepo.DecrementEventCount(ctx, *event.GroupID)
		if err != nil {
			s.logger.Warn("Failed to decrement group event count", "groupId", event.GroupID.Hex(), "error", err)
		}
	}

	return nil
}

// GetEvent retrieves an event by ID
func (s *ManagementService) GetEvent(ctx context.Context, eventID primitive.ObjectID) (*models.Event, error) {
	event, err := s.eventRepo.FindByID(ctx, eventID)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to find event")
	}

	return event, nil
}

// GetEvents retrieves events with filters and pagination
func (s *ManagementService) GetEvents(ctx context.Context, filter *EventFilter, page, limit int) ([]models.Event, int, error) {
	// Validate pagination
	if page < 1 {
		page = 1
	}

	if limit < 1 {
		limit = 10
	} else if limit > 100 {
		limit = 100
	}

	// Build filter
	filterMap := make(map[string]interface{})

	if filter.HostID != nil {
		filterMap["host_id"] = *filter.HostID
	}

	if filter.GroupID != nil {
		filterMap["group_id"] = *filter.GroupID
	}

	if filter.Category != "" {
		filterMap["category"] = filter.Category
	}

	if filter.Status != "" {
		filterMap["status"] = filter.Status
	} else {
		// By default, exclude cancelled events
		filterMap["status"] = map[string]interface{}{
			"$ne": "cancelled",
		}
	}

	if filter.Type != "" {
		filterMap["type"] = filter.Type
	}

	if filter.StartAfter != nil {
		filterMap["start_time"] = map[string]interface{}{
			"$gte": *filter.StartAfter,
		}
	}

	if filter.StartBefore != nil {
		if filterMap["start_time"] == nil {
			filterMap["start_time"] = map[string]interface{}{
				"$lte": *filter.StartBefore,
			}
		} else {
			filterMap["start_time"].(map[string]interface{})["$lte"] = *filter.StartBefore
		}
	}

	if filter.Location != nil {
		// For simplicity, just filter by city
		filterMap["location.city"] = filter.Location.City
	}

	if filter.Tags != nil && len(filter.Tags) > 0 {
		filterMap["tags"] = map[string]interface{}{
			"$in": filter.Tags,
		}
	}

	// Default sort by start time
	sortBy := "start_time"
	sortOrder := "asc"

	if filter.SortBy != "" {
		sortBy = filter.SortBy
	}

	if filter.SortOrder != "" {
		sortOrder = filter.SortOrder
	}

	// Get events
	events, total, err := s.eventRepo.FindWithFilter(ctx, filterMap, page, limit, sortBy, sortOrder)
	if err != nil {
		return nil, 0, errors.Wrap(err, "Failed to find events")
	}

	return events, total, nil
}

// AddCoHost adds a co-host to an event
func (s *ManagementService) AddCoHost(ctx context.Context, eventID, hostID, coHostID primitive.ObjectID) error {
	// Get event
	event, err := s.eventRepo.FindByID(ctx, eventID)
	if err != nil {
		return errors.Wrap(err, "Failed to find event")
	}

	// Check permissions
	if event.HostID != hostID {
		return errors.New(errors.CodeForbidden, "Only the host can add co-hosts")
	}

	// Check if user exists
	_, err = s.userRepo.FindByID(ctx, coHostID)
	if err != nil {
		return errors.Wrap(err, "Failed to find co-host user")
	}

	// Check if already a co-host
	for _, id := range event.CoHosts {
		if id == coHostID {
			return errors.New(errors.CodeInvalidOperation, "User is already a co-host")
		}
	}

	// Add co-host
	event.CoHosts = append(event.CoHosts, coHostID)
	event.UpdatedAt = time.Now()

	// Update event
	err = s.eventRepo.Update(ctx, event)
	if err != nil {
		return errors.Wrap(err, "Failed to update event")
	}

	return nil
}

// RemoveCoHost removes a co-host from an event
func (s *ManagementService) RemoveCoHost(ctx context.Context, eventID, hostID, coHostID primitive.ObjectID) error {
	// Get event
	event, err := s.eventRepo.FindByID(ctx, eventID)
	if err != nil {
		return errors.Wrap(err, "Failed to find event")
	}

	// Check permissions
	if event.HostID != hostID {
		return errors.New(errors.CodeForbidden, "Only the host can remove co-hosts")
	}

	// Find and remove co-host
	found := false
	newCoHosts := make([]primitive.ObjectID, 0, len(event.CoHosts))

	for _, id := range event.CoHosts {
		if id != coHostID {
			newCoHosts = append(newCoHosts, id)
		} else {
			found = true
		}
	}

	if !found {
		return errors.New(errors.CodeNotFound, "User is not a co-host")
	}

	// Update co-hosts
	event.CoHosts = newCoHosts
	event.UpdatedAt = time.Now()

	// Update event
	err = s.eventRepo.Update(ctx, event)
	if err != nil {
		return errors.Wrap(err, "Failed to update event")
	}

	return nil
}

// CheckInAttendee checks in an attendee at an event
func (s *ManagementService) CheckInAttendee(ctx context.Context, eventID, attendeeID, checkerID primitive.ObjectID) error {
	// Get event
	event, err := s.eventRepo.FindByID(ctx, eventID)
	if err != nil {
		return errors.Wrap(err, "Failed to find event")
	}

	// Check permissions
	if event.HostID != checkerID {
		isCoHost := false
		for _, id := range event.CoHosts {
			if id == checkerID {
				isCoHost = true
				break
			}
		}

		if !isCoHost {
			return errors.New(errors.CodeForbidden, "Only host or co-hosts can check in attendees")
		}
	}

	// Get attendee
	attendee, err := s.attendeeRepo.FindByEventAndUser(ctx, eventID, attendeeID)
	if err != nil {
		return errors.Wrap(err, "Failed to find attendee")
	}

	// Check if already checked in
	if attendee.CheckedIn {
		return errors.New(errors.CodeInvalidOperation, "Attendee is already checked in")
	}

	// Check in attendee
	now := time.Now()
	attendee.CheckedIn = true
	attendee.CheckInTime = &now

	// Update attendee
	err = s.attendeeRepo.Update(ctx, attendee)
	if err != nil {
		return errors.Wrap(err, "Failed to update attendee")
	}

	return nil
}

// GetEventsByUser retrieves events for a user
func (s *ManagementService) GetEventsByUser(ctx context.Context, userID primitive.ObjectID, status string, page, limit int) ([]models.Event, int, error) {
	// Validate pagination
	if page < 1 {
		page = 1
	}

	if limit < 1 {
		limit = 10
	} else if limit > 100 {
		limit = 100
	}

	// Build filter
	filterMap := make(map[string]interface{})

	// Filter by user (host or co-host)
	filterMap["$or"] = []map[string]interface{}{
		{"host_id": userID},
		{"co_hosts": userID},
	}

	// Filter by status if provided
	if status != "" {
		filterMap["status"] = status
	} else {
		// By default, exclude cancelled events
		filterMap["status"] = map[string]interface{}{
			"$ne": "cancelled",
		}
	}

	// Get events
	events, total, err := s.eventRepo.FindWithFilter(ctx, filterMap, page, limit, "start_time", "asc")
	if err != nil {
		return nil, 0, errors.Wrap(err, "Failed to find events")
	}

	return events, total, nil
}

// GetEventsAttendedByUser retrieves events a user is attending
func (s *ManagementService) GetEventsAttendedByUser(ctx context.Context, userID primitive.ObjectID, rsvpStatus string, page, limit int) ([]models.Event, int, error) {
	// Validate pagination
	if page < 1 {
		page = 1
	}

	if limit < 1 {
		limit = 10
	} else if limit > 100 {
		limit = 100
	}

	// Set default RSVP status if not provided
	if rsvpStatus == "" {
		rsvpStatus = "going"
	}

	// Get attendee records
	attendees, total, err := s.attendeeRepo.FindByUserIDAndRSVP(ctx, userID, rsvpStatus, page, limit)
	if err != nil {
		return nil, 0, errors.Wrap(err, "Failed to find attendee records")
	}

	if len(attendees) == 0 {
		return []models.Event{}, 0, nil
	}

	// Get event IDs
	eventIDs := make([]primitive.ObjectID, 0, len(attendees))
	for _, attendee := range attendees {
		eventIDs = append(eventIDs, attendee.EventID)
	}

	// Get events
	events, err := s.eventRepo.FindByIDs(ctx, eventIDs)
	if err != nil {
		return nil, 0, errors.Wrap(err, "Failed to find events")
	}

	return events, total, nil
}

// UpdateEventStatus updates the status of an event
func (s *ManagementService) UpdateEventStatus(ctx context.Context, eventID primitive.ObjectID, status string, userID primitive.ObjectID) error {
	// Get event
	event, err := s.eventRepo.FindByID(ctx, eventID)
	if err != nil {
		return errors.Wrap(err, "Failed to find event")
	}

	// Check permissions
	if event.HostID != userID {
		isCoHost := false
		for _, id := range event.CoHosts {
			if id == userID {
				isCoHost = true
				break
			}
		}

		if !isCoHost {
			// Check if user is admin
			user, err := s.userRepo.FindByID(ctx, userID)
			if err != nil || user.Role != "admin" {
				return errors.New(errors.CodeForbidden, "Only host, co-hosts, or admins can update event status")
			}
		}
	}

	// Validate status
	validStatuses := map[string]bool{
		"scheduled": true,
		"live":      true,
		"ended":     true,
		"cancelled": true,
	}

	if !validStatuses[status] {
		return errors.New(errors.CodeInvalidArgument, "Invalid status. Must be one of: scheduled, live, ended, cancelled")
	}

	// Update status
	event.Status = status
	event.UpdatedAt = time.Now()

	// Update event
	err = s.eventRepo.Update(ctx, event)
	if err != nil {
		return errors.Wrap(err, "Failed to update event")
	}

	return nil
}

// CreateRecurringEvent creates a series of recurring events
func (s *ManagementService) CreateRecurringEvent(ctx context.Context, event *models.Event, recurrenceRule string, occurrences int) ([]models.Event, error) {
	// Create parent event
	parentEvent, err := s.CreateEvent(ctx, event)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create parent event")
	}

	// Set recurring flag
	parentEvent.IsRecurring = true
	parentEvent.RecurrenceRule = recurrenceRule

	// Update parent event
	err = s.eventRepo.Update(ctx, parentEvent)
	if err != nil {
		s.logger.Warn("Failed to update parent event", "eventId", parentEvent.ID.Hex(), "error", err)
	}

	// Calculate recurring dates based on rule
	dates, err := calculateRecurringDates(event.StartTime, event.EndTime, recurrenceRule, occurrences)
	if err != nil {
		return []models.Event{*parentEvent}, errors.Wrap(err, "Failed to calculate recurring dates")
	}

	// Skip the first date (it's the parent event)
	if len(dates) > 1 {
		dates = dates[1:]
	}

	// Create child events
	events := []models.Event{*parentEvent}

	for _, date := range dates {
		// Clone parent event
		childEvent := *event
		childEvent.ID = primitive.NewObjectID()
		childEvent.ParentEventID = &parentEvent.ID
		childEvent.StartTime = date.Start
		childEvent.EndTime = date.End
		childEvent.IsRecurring = true
		childEvent.RecurrenceRule = recurrenceRule

		// Create child event
		child, err := s.CreateEvent(ctx, &childEvent)
		if err != nil {
			s.logger.Warn("Failed to create child event", "error", err)
			continue
		}

		events = append(events, *child)
	}

	return events, nil
}

// EventUpdates represents fields that can be updated on an event
type EventUpdates struct {
	Title        string                  `json:"title"`
	Description  string                  `json:"description"`
	StartTime    *time.Time              `json:"start_time"`
	EndTime      *time.Time              `json:"end_time"`
	TimeZone     string                  `json:"time_zone"`
	Location     *models.EventLocation   `json:"location"`
	CoverImage   string                  `json:"cover_image"`
	Type         string                  `json:"type"`
	Privacy      string                  `json:"privacy"`
	Category     string                  `json:"category"`
	Tags         []string                `json:"tags"`
	MaxAttendees int                     `json:"max_attendees"`
	URL          string                  `json:"url"`
	TicketInfo   *models.EventTicketInfo `json:"ticket_info"`
}

// EventFilter represents filters for event retrieval
type EventFilter struct {
	HostID      *primitive.ObjectID   `json:"host_id"`
	GroupID     *primitive.ObjectID   `json:"group_id"`
	Category    string                `json:"category"`
	Status      string                `json:"status"`
	Type        string                `json:"type"`
	StartAfter  *time.Time            `json:"start_after"`
	StartBefore *time.Time            `json:"start_before"`
	Location    *models.EventLocation `json:"location"`
	Tags        []string              `json:"tags"`
	SortBy      string                `json:"sort_by"`
	SortOrder   string                `json:"sort_order"`
}

// RecurringDate represents a start and end time for a recurring event
type RecurringDate struct {
	Start time.Time
	End   time.Time
}

// Helper functions

// calculateRecurringDates calculates dates for recurring events
func calculateRecurringDates(start, end time.Time, rule string, occurrences int) ([]RecurringDate, error) {
	// This is a simplified implementation
	// A production-ready version would parse and implement the full iCalendar RRULE

	duration := end.Sub(start)
	dates := make([]RecurringDate, 0, occurrences)

	// Add first occurrence
	dates = append(dates, RecurringDate{
		Start: start,
		End:   end,
	})

	// Parse rule (simplified)
	var interval int
	var unit string

	switch rule {
	case "DAILY":
		interval = 1
		unit = "day"
	case "WEEKLY":
		interval = 1
		unit = "week"
	case "BIWEEKLY":
		interval = 2
		unit = "week"
	case "MONTHLY":
		interval = 1
		unit = "month"
	default:
		// Default to weekly
		interval = 1
		unit = "week"
	}

	// Calculate recurring dates
	currentStart := start
	currentEnd := end

	for i := 1; i < occurrences; i++ {
		switch unit {
		case "day":
			currentStart = currentStart.AddDate(0, 0, interval)
			currentEnd = currentStart.Add(duration)
		case "week":
			currentStart = currentStart.AddDate(0, 0, 7*interval)
			currentEnd = currentStart.Add(duration)
		case "month":
			currentStart = currentStart.AddDate(0, interval, 0)
			currentEnd = currentStart.Add(duration)
		}

		dates = append(dates, RecurringDate{
			Start: currentStart,
			End:   currentEnd,
		})
	}

	return dates, nil
}
