package interfaces

import (
	"context"
	"time"

	"github.com/Caqil/vyrall/internal/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// EventRepository defines the interface for event data access
type EventRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, event *models.Event) (primitive.ObjectID, error)
	GetByID(ctx context.Context, id primitive.ObjectID) (*models.Event, error)
	Update(ctx context.Context, event *models.Event) error
	Delete(ctx context.Context, id primitive.ObjectID) error

	// Query operations
	GetByHostID(ctx context.Context, hostID primitive.ObjectID, limit, offset int) ([]*models.Event, int, error)
	GetByGroupID(ctx context.Context, groupID primitive.ObjectID, limit, offset int) ([]*models.Event, int, error)
	GetByIDs(ctx context.Context, ids []primitive.ObjectID) ([]*models.Event, error)
	GetByLocation(ctx context.Context, lat, lng float64, radiusKm float64, limit, offset int) ([]*models.Event, int, error)
	GetRecurringEvents(ctx context.Context, parentID primitive.ObjectID) ([]*models.Event, error)

	// Time-based queries
	GetUpcomingEvents(ctx context.Context, limit, offset int) ([]*models.Event, int, error)
	GetUpcomingEventsForUser(ctx context.Context, userID primitive.ObjectID, limit, offset int) ([]*models.Event, int, error)
	GetEventsByDateRange(ctx context.Context, startDate, endDate time.Time, limit, offset int) ([]*models.Event, int, error)
	GetPastEvents(ctx context.Context, limit, offset int) ([]*models.Event, int, error)
	GetOngoingEvents(ctx context.Context, limit, offset int) ([]*models.Event, int, error)

	// Attendee management
	AddAttendee(ctx context.Context, eventID, userID primitive.ObjectID, rsvp string) (primitive.ObjectID, error)
	UpdateAttendeeStatus(ctx context.Context, eventID, userID primitive.ObjectID, rsvp string) error
	RemoveAttendee(ctx context.Context, eventID, userID primitive.ObjectID) error
	GetAttendees(ctx context.Context, eventID primitive.ObjectID, rsvp string, limit, offset int) ([]*models.EventAttendee, int, error)
	GetAttendeeByID(ctx context.Context, eventID, userID primitive.ObjectID) (*models.EventAttendee, error)
	UpdateRSVPCounts(ctx context.Context, eventID primitive.ObjectID) error

	// Event status management
	UpdateStatus(ctx context.Context, eventID primitive.ObjectID, status string) error
	CancelEvent(ctx context.Context, eventID primitive.ObjectID) error

	// Check-in management
	CheckInAttendee(ctx context.Context, eventID, userID primitive.ObjectID) error
	GetCheckedInAttendees(ctx context.Context, eventID primitive.ObjectID, limit, offset int) ([]*models.EventAttendee, int, error)

	// Ticket management
	UpdateTicketInfo(ctx context.Context, eventID primitive.ObjectID, ticketInfo models.EventTicketInfo) error
	UpdateTicketSales(ctx context.Context, eventID primitive.ObjectID, ticketTypeID string, quantity int) error

	// Recurring event management
	CreateRecurringInstances(ctx context.Context, parentEventID primitive.ObjectID, rule string, endDate time.Time) ([]primitive.ObjectID, error)

	// Reminders
	CreateReminder(ctx context.Context, reminder *models.EventReminder) (primitive.ObjectID, error)
	GetRemindersToSend(ctx context.Context, before time.Time) ([]*models.EventReminder, error)
	MarkReminderSent(ctx context.Context, reminderID primitive.ObjectID) error

	// Analytics
	GetEventStats(ctx context.Context, eventID primitive.ObjectID) (models.EventAnalytics, error)

	// Search and discovery
	GetRecommendedEvents(ctx context.Context, userID primitive.ObjectID, limit int) ([]*models.Event, error)
	GetPopularEvents(ctx context.Context, limit int) ([]*models.Event, error)
	Search(ctx context.Context, query string, filter map[string]interface{}, limit, offset int) ([]*models.Event, int, error)

	// Advanced filtering
	List(ctx context.Context, filter map[string]interface{}, sort map[string]int, limit, offset int) ([]*models.Event, int, error)
}
