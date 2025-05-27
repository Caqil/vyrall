package event

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/Caqil/vyrall/internal/internal/models"
	"github.com/Caqil/vyrall/internal/pkg/errors"
	"github.com/Caqil/vyrall/internal/pkg/logging"
)

// AnalyticsService handles analytics for events
type AnalyticsService struct {
	eventRepo     EventRepository
	attendeeRepo  AttendeeRepository
	analyticsRepo EventAnalyticsRepository
	userRepo      UserRepository
	logger        logging.Logger
}

// NewAnalyticsService creates a new analytics service for events
func NewAnalyticsService(
	eventRepo EventRepository,
	attendeeRepo AttendeeRepository,
	analyticsRepo EventAnalyticsRepository,
	userRepo UserRepository,
	logger logging.Logger,
) *AnalyticsService {
	return &AnalyticsService{
		eventRepo:     eventRepo,
		attendeeRepo:  attendeeRepo,
		analyticsRepo: analyticsRepo,
		userRepo:      userRepo,
		logger:        logger,
	}
}

// RecordEventView records a view of an event
func (s *AnalyticsService) RecordEventView(ctx context.Context, eventID primitive.ObjectID, userID *primitive.ObjectID, sessionID string) error {
	// Get event
	event, err := s.eventRepo.FindByID(ctx, eventID)
	if err != nil {
		return errors.Wrap(err, "Failed to find event")
	}

	// Update view count
	err = s.eventRepo.IncrementViewCount(ctx, eventID)
	if err != nil {
		s.logger.Warn("Failed to increment view count", "eventId", eventID.Hex(), "error", err)
	}

	// Get or create analytics
	analytics, err := s.analyticsRepo.FindByEventID(ctx, eventID)
	if err != nil {
		if errors.Code(err) == errors.CodeNotFound {
			// Create new analytics
			analytics = &models.EventAnalytics{
				EventID:         eventID,
				ViewCount:       1,
				UniqueViewCount: 1,
				UpdatedAt:       time.Now(),
			}

			_, err = s.analyticsRepo.Create(ctx, analytics)
			if err != nil {
				s.logger.Warn("Failed to create event analytics", "eventId", eventID.Hex(), "error", err)
			}

			return nil
		}

		return errors.Wrap(err, "Failed to find event analytics")
	}

	// Update analytics
	analytics.ViewCount++

	// Check if unique view
	if userID != nil {
		isNewViewer, err := s.analyticsRepo.IsNewViewer(ctx, eventID, *userID)
		if err != nil {
			s.logger.Warn("Failed to check if new viewer", "eventId", eventID.Hex(), "userId", userID.Hex(), "error", err)
		} else if isNewViewer {
			analytics.UniqueViewCount++

			// Record viewer
			err = s.analyticsRepo.AddViewer(ctx, eventID, *userID)
			if err != nil {
				s.logger.Warn("Failed to add viewer", "eventId", eventID.Hex(), "userId", userID.Hex(), "error", err)
			}
		}
	} else if sessionID != "" {
		isNewSession, err := s.analyticsRepo.IsNewSession(ctx, eventID, sessionID)
		if err != nil {
			s.logger.Warn("Failed to check if new session", "eventId", eventID.Hex(), "sessionId", sessionID, "error", err)
		} else if isNewSession {
			analytics.UniqueViewCount++

			// Record session
			err = s.analyticsRepo.AddSession(ctx, eventID, sessionID)
			if err != nil {
				s.logger.Warn("Failed to add session", "eventId", eventID.Hex(), "sessionId", sessionID, "error", err)
			}
		}
	}

	// Update analytics
	analytics.UpdatedAt = time.Now()
	err = s.analyticsRepo.Update(ctx, analytics)
	if err != nil {
		s.logger.Warn("Failed to update event analytics", "eventId", eventID.Hex(), "error", err)
	}

	return nil
}

// RecordEventShare records a share of an event
func (s *AnalyticsService) RecordEventShare(ctx context.Context, eventID primitive.ObjectID, userID primitive.ObjectID, platform string) error {
	// Update share count
	err := s.eventRepo.IncrementShareCount(ctx, eventID)
	if err != nil {
		s.logger.Warn("Failed to increment share count", "eventId", eventID.Hex(), "error", err)
	}

	// Update analytics
	analytics, err := s.analyticsRepo.FindByEventID(ctx, eventID)
	if err != nil {
		if errors.Code(err) == errors.CodeNotFound {
			// Create new analytics
			analytics = &models.EventAnalytics{
				EventID:    eventID,
				ShareCount: 1,
				UpdatedAt:  time.Now(),
			}

			_, err = s.analyticsRepo.Create(ctx, analytics)
			if err != nil {
				s.logger.Warn("Failed to create event analytics", "eventId", eventID.Hex(), "error", err)
			}

			return nil
		}

		return errors.Wrap(err, "Failed to find event analytics")
	}

	// Update analytics
	analytics.ShareCount++

	// Update referral sources if platform is provided
	if platform != "" {
		if analytics.ReferralSources == nil {
			analytics.ReferralSources = make(map[string]int)
		}

		analytics.ReferralSources[platform]++
	}

	// Update analytics
	analytics.UpdatedAt = time.Now()
	err = s.analyticsRepo.Update(ctx, analytics)
	if err != nil {
		s.logger.Warn("Failed to update event analytics", "eventId", eventID.Hex(), "error", err)
	}

	return nil
}

// RecordTicketSale records a ticket sale for an event
func (s *AnalyticsService) RecordTicketSale(ctx context.Context, eventID primitive.ObjectID, ticketType string, price float64) error {
	// Update analytics
	analytics, err := s.analyticsRepo.FindByEventID(ctx, eventID)
	if err != nil {
		if errors.Code(err) == errors.CodeNotFound {
			// Create new analytics
			analytics = &models.EventAnalytics{
				EventID:       eventID,
				TicketRevenue: price,
				UpdatedAt:     time.Now(),
			}

			_, err = s.analyticsRepo.Create(ctx, analytics)
			if err != nil {
				s.logger.Warn("Failed to create event analytics", "eventId", eventID.Hex(), "error", err)
			}

			return nil
		}

		return errors.Wrap(err, "Failed to find event analytics")
	}

	// Update analytics
	analytics.TicketRevenue += price

	// Update analytics
	analytics.UpdatedAt = time.Now()
	err = s.analyticsRepo.Update(ctx, analytics)
	if err != nil {
		s.logger.Warn("Failed to update event analytics", "eventId", eventID.Hex(), "error", err)
	}

	return nil
}

// RecordCheckIn records an attendee check-in at an event
func (s *AnalyticsService) RecordCheckIn(ctx context.Context, eventID primitive.ObjectID, userID primitive.ObjectID) error {
	// Update analytics
	analytics, err := s.analyticsRepo.FindByEventID(ctx, eventID)
	if err != nil {
		if errors.Code(err) == errors.CodeNotFound {
			// Create new analytics
			analytics = &models.EventAnalytics{
				EventID:     eventID,
				CheckInRate: 1.0, // First check-in, 100% rate
				UpdatedAt:   time.Now(),
			}

			_, err = s.analyticsRepo.Create(ctx, analytics)
			if err != nil {
				s.logger.Warn("Failed to create event analytics", "eventId", eventID.Hex(), "error", err)
			}

			return nil
		}

		return errors.Wrap(err, "Failed to find event analytics")
	}

	// Calculate new check-in rate
	attendeeCount, err := s.attendeeRepo.CountByEventIDAndRSVP(ctx, eventID, "going")
	if err != nil {
		s.logger.Warn("Failed to count attendees", "eventId", eventID.Hex(), "error", err)
	} else if attendeeCount > 0 {
		checkedInCount, err := s.attendeeRepo.CountCheckedIn(ctx, eventID)
		if err != nil {
			s.logger.Warn("Failed to count checked-in attendees", "eventId", eventID.Hex(), "error", err)
		} else {
			analytics.CheckInRate = float64(checkedInCount) / float64(attendeeCount)
		}
	}

	// Update analytics
	analytics.UpdatedAt = time.Now()
	err = s.analyticsRepo.Update(ctx, analytics)
	if err != nil {
		s.logger.Warn("Failed to update event analytics", "eventId", eventID.Hex(), "error", err)
	}

	return nil
}

// GetEventAnalytics retrieves analytics for an event
func (s *AnalyticsService) GetEventAnalytics(ctx context.Context, eventID primitive.ObjectID) (*models.EventAnalytics, error) {
	// Get analytics
	analytics, err := s.analyticsRepo.FindByEventID(ctx, eventID)
	if err != nil {
		if errors.Code(err) == errors.CodeNotFound {
			// Create empty analytics
			return &models.EventAnalytics{
				EventID:          eventID,
				ViewCount:        0,
				UniqueViewCount:  0,
				ShareCount:       0,
				ClickThroughRate: 0,
				ConversionRate:   0,
				TicketRevenue:    0,
				CheckInRate:      0,
				UpdatedAt:        time.Now(),
			}, nil
		}

		return nil, errors.Wrap(err, "Failed to find event analytics")
	}

	return analytics, nil
}

// GetEventDemographics retrieves demographic information about attendees
func (s *AnalyticsService) GetEventDemographics(ctx context.Context, eventID primitive.ObjectID) (map[string]interface{}, error) {
	// Get attendees
	attendees, err := s.attendeeRepo.FindByEventID(ctx, eventID)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to find attendees")
	}

	// Initialize demographics
	demographics := map[string]interface{}{
		"gender":     map[string]int{},
		"age_groups": map[string]int{},
		"locations":  map[string]int{},
	}

	// Process attendees
	for _, attendee := range attendees {
		// Skip if not going
		if attendee.RSVP != "going" {
			continue
		}

		// Get user
		user, err := s.userRepo.FindByID(ctx, attendee.UserID)
		if err != nil {
			s.logger.Warn("Failed to find user", "userId", attendee.UserID.Hex(), "error", err)
			continue
		}

		// Update gender stats
		if user.Gender != "" {
			gender := user.Gender
			if _, ok := demographics["gender"].(map[string]int)[gender]; !ok {
				demographics["gender"].(map[string]int)[gender] = 0
			}
			demographics["gender"].(map[string]int)[gender]++
		}

		// Update age group stats
		if !user.DateOfBirth.IsZero() {
			age := calculateAge(user.DateOfBirth)
			ageGroup := getAgeGroup(age)
			if _, ok := demographics["age_groups"].(map[string]int)[ageGroup]; !ok {
				demographics["age_groups"].(map[string]int)[ageGroup] = 0
			}
			demographics["age_groups"].(map[string]int)[ageGroup]++
		}

		// Update location stats
		if user.Location != "" {
			location := user.Location
			if _, ok := demographics["locations"].(map[string]int)[location]; !ok {
				demographics["locations"].(map[string]int)[location] = 0
			}
			demographics["locations"].(map[string]int)[location]++
		}
	}

	return demographics, nil
}

// GetEventTrends retrieves trending events
func (s *AnalyticsService) GetEventTrends(ctx context.Context, timeframe string, limit int) ([]EventTrend, error) {
	// Set time range based on timeframe
	var startTime time.Time
	now := time.Now()

	switch timeframe {
	case "day":
		startTime = now.AddDate(0, 0, -1)
	case "week":
		startTime = now.AddDate(0, 0, -7)
	case "month":
		startTime = now.AddDate(0, -1, 0)
	default:
		startTime = now.AddDate(0, 0, -7) // Default to week
	}

	// Limit to reasonable value
	if limit <= 0 {
		limit = 10
	} else if limit > 100 {
		limit = 100
	}

	// Get events with high engagement
	events, err := s.eventRepo.FindTrending(ctx, startTime, limit)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to find trending events")
	}

	// Build trends
	trends := make([]EventTrend, 0, len(events))

	for _, event := range events {
		// Get view data
		analytics, err := s.analyticsRepo.FindByEventID(ctx, event.ID)
		if err != nil {
			s.logger.Warn("Failed to find event analytics", "eventId", event.ID.Hex(), "error", err)
			continue
		}

		// Get RSVP counts
		rsvpCounts := event.RSVPCount

		trend := EventTrend{
			EventID:    event.ID,
			Title:      event.Title,
			StartTime:  event.StartTime,
			ViewCount:  analytics.ViewCount,
			RSVPCount:  rsvpCounts.Going + rsvpCounts.Interested,
			ShareCount: analytics.ShareCount,
			TrendScore: calculateTrendScore(analytics.ViewCount, rsvpCounts.Going+rsvpCounts.Interested, analytics.ShareCount),
		}

		trends = append(trends, trend)
	}

	return trends, nil
}

// GetEventAttendeeActivity retrieves activity of an attendee at an event
func (s *AnalyticsService) GetEventAttendeeActivity(ctx context.Context, eventID, userID primitive.ObjectID) (*AttendeeActivity, error) {
	// Get attendee
	attendee, err := s.attendeeRepo.FindByEventAndUser(ctx, eventID, userID)
	if err != nil {
		if errors.Code(err) == errors.CodeNotFound {
			return nil, errors.New(errors.CodeNotFound, "User is not an attendee of this event")
		}
		return nil, errors.Wrap(err, "Failed to find attendee")
	}

	// Build activity
	activity := &AttendeeActivity{
		RSVP:          attendee.RSVP,
		RSVPTimestamp: attendee.RSVPTimestamp,
		CheckedIn:     attendee.CheckedIn,
		CheckInTime:   attendee.CheckInTime,
		InvitedBy:     attendee.InvitedBy,
		GuestCount:    attendee.GuestCount,
	}

	// Get user
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		s.logger.Warn("Failed to find user", "userId", userID.Hex(), "error", err)
	} else {
		activity.Username = user.Username
		activity.DisplayName = user.DisplayName
		activity.ProfilePicture = user.ProfilePicture
	}

	// Get event
	event, err := s.eventRepo.FindByID(ctx, eventID)
	if err != nil {
		s.logger.Warn("Failed to find event", "eventId", eventID.Hex(), "error", err)
	} else {
		activity.EventTitle = event.Title
		activity.EventStartTime = event.StartTime
		activity.EventEndTime = event.EndTime
	}

	return activity, nil
}

// EventTrend represents a trending event
type EventTrend struct {
	EventID    primitive.ObjectID `json:"event_id"`
	Title      string             `json:"title"`
	StartTime  time.Time          `json:"start_time"`
	ViewCount  int                `json:"view_count"`
	RSVPCount  int                `json:"rsvp_count"`
	ShareCount int                `json:"share_count"`
	TrendScore float64            `json:"trend_score"`
}

// AttendeeActivity represents activity of an attendee at an event
type AttendeeActivity struct {
	Username       string              `json:"username"`
	DisplayName    string              `json:"display_name"`
	ProfilePicture string              `json:"profile_picture"`
	EventTitle     string              `json:"event_title"`
	EventStartTime time.Time           `json:"event_start_time"`
	EventEndTime   time.Time           `json:"event_end_time"`
	RSVP           string              `json:"rsvp"`
	RSVPTimestamp  time.Time           `json:"rsvp_timestamp"`
	CheckedIn      bool                `json:"checked_in"`
	CheckInTime    *time.Time          `json:"check_in_time,omitempty"`
	InvitedBy      *primitive.ObjectID `json:"invited_by,omitempty"`
	GuestCount     int                 `json:"guest_count"`
}

// Helper functions

// calculateAge calculates age from date of birth
func calculateAge(dob time.Time) int {
	now := time.Now()
	years := now.Year() - dob.Year()

	// Adjust for months/days
	if now.Month() < dob.Month() || (now.Month() == dob.Month() && now.Day() < dob.Day()) {
		years--
	}

	return years
}

// getAgeGroup determines the age group for a given age
func getAgeGroup(age int) string {
	switch {
	case age < 18:
		return "under_18"
	case age < 25:
		return "18_24"
	case age < 35:
		return "25_34"
	case age < 45:
		return "35_44"
	case age < 55:
		return "45_54"
	case age < 65:
		return "55_64"
	default:
		return "65_plus"
	}
}

// calculateTrendScore calculates a score for trending events
func calculateTrendScore(views, rsvps, shares int) float64 {
	return float64(views)*0.2 + float64(rsvps)*1.0 + float64(shares)*0.5
}
