package events

import (
	"net/http"
	"strconv"
	"time"

	"github.com/Caqil/vyrall/internal/utils/response"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// AnalyticsService defines the interface for event analytics operations
type AnalyticsService interface {
	GetEventAnalytics(ctx context.Context, eventID primitive.ObjectID) (*models.EventAnalytics, error)
	GetEventViewsByTimeRange(ctx context.Context, eventID primitive.ObjectID, startDate, endDate time.Time) (map[string]int, error)
	GetEventAttendeeStats(ctx context.Context, eventID primitive.ObjectID) (map[string]interface{}, error)
	GetEventReferralSources(ctx context.Context, eventID primitive.ObjectID) (map[string]int, error)
	GetEventEngagementMetrics(ctx context.Context, eventID primitive.ObjectID) (map[string]interface{}, error)
	GetHostAnalytics(ctx context.Context, userID primitive.ObjectID, limit int) ([]map[string]interface{}, error)
	GetEventConversionRate(ctx context.Context, eventID primitive.ObjectID) (float64, error)
	GetEventDemographics(ctx context.Context, eventID primitive.ObjectID) (map[string]interface{}, error)
}

// GetEventAnalytics retrieves analytics data for a specific event
func GetEventAnalytics(c *gin.Context) {
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

	// Check if the user has permission to view analytics (must be host, co-host, or admin)
	if event.HostID != userID.(primitive.ObjectID) {
		// Check if user is a co-host
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
				response.Error(c, http.StatusForbidden, "Only the event host, co-host, or an admin can view analytics", nil)
				return
			}
		}
	}

	// Get the analytics service
	analyticsService := c.MustGet("analyticsService").(AnalyticsService)

	// Get the event analytics
	analytics, err := analyticsService.GetEventAnalytics(c.Request.Context(), eventID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve event analytics", err)
		return
	}

	response.Success(c, http.StatusOK, "Event analytics retrieved successfully", analytics)
}

// GetEventViewsByTimeRange retrieves view data for an event over a specific time period
func GetEventViewsByTimeRange(c *gin.Context) {
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

	// Get query parameters
	startDateStr := c.DefaultQuery("start_date", "")
	endDateStr := c.DefaultQuery("end_date", "")

	// Parse dates
	var startDate, endDate time.Time
	if startDateStr != "" {
		startDate, err = time.Parse("2006-01-02", startDateStr)
		if err != nil {
			response.Error(c, http.StatusBadRequest, "Invalid start date format", err)
			return
		}
	} else {
		// Default to 30 days ago
		startDate = time.Now().AddDate(0, 0, -30)
	}

	if endDateStr != "" {
		endDate, err = time.Parse("2006-01-02", endDateStr)
		if err != nil {
			response.Error(c, http.StatusBadRequest, "Invalid end date format", err)
			return
		}
		// Include the entire end date
		endDate = endDate.Add(24 * time.Hour)
	} else {
		// Default to now
		endDate = time.Now()
	}

	// Get the event service
	eventService := c.MustGet("eventService").(EventService)

	// Check if the event exists
	event, err := eventService.GetEventByID(c.Request.Context(), eventID)
	if err != nil {
		response.Error(c, http.StatusNotFound, "Event not found", err)
		return
	}

	// Check permissions
	if event.HostID != userID.(primitive.ObjectID) {
		isCoHost := false
		for _, coHostID := range event.CoHosts {
			if coHostID == userID.(primitive.ObjectID) {
				isCoHost = true
				break
			}
		}

		if !isCoHost {
			isAdmin, err := eventService.IsUserAdmin(c.Request.Context(), userID.(primitive.ObjectID))
			if err != nil {
				response.Error(c, http.StatusInternalServerError, "Failed to check user permissions", err)
				return
			}

			if !isAdmin {
				response.Error(c, http.StatusForbidden, "Only the event host, co-host, or an admin can view analytics", nil)
				return
			}
		}
	}

	// Get the analytics service
	analyticsService := c.MustGet("analyticsService").(AnalyticsService)

	// Get the view data by time range
	viewData, err := analyticsService.GetEventViewsByTimeRange(c.Request.Context(), eventID, startDate, endDate)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve event view data", err)
		return
	}

	response.Success(c, http.StatusOK, "Event view data retrieved successfully", viewData)
}

// GetEventAttendeeStats retrieves statistics about event attendees
func GetEventAttendeeStats(c *gin.Context) {
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

	// Check permissions
	if event.HostID != userID.(primitive.ObjectID) {
		isCoHost := false
		for _, coHostID := range event.CoHosts {
			if coHostID == userID.(primitive.ObjectID) {
				isCoHost = true
				break
			}
		}

		if !isCoHost {
			isAdmin, err := eventService.IsUserAdmin(c.Request.Context(), userID.(primitive.ObjectID))
			if err != nil {
				response.Error(c, http.StatusInternalServerError, "Failed to check user permissions", err)
				return
			}

			if !isAdmin {
				response.Error(c, http.StatusForbidden, "Only the event host, co-host, or an admin can view analytics", nil)
				return
			}
		}
	}

	// Get the analytics service
	analyticsService := c.MustGet("analyticsService").(AnalyticsService)

	// Get attendee statistics
	stats, err := analyticsService.GetEventAttendeeStats(c.Request.Context(), eventID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve attendee statistics", err)
		return
	}

	response.Success(c, http.StatusOK, "Attendee statistics retrieved successfully", stats)
}

// GetEventReferralSources retrieves information about where attendees came from
func GetEventReferralSources(c *gin.Context) {
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

	// Check permissions
	if event.HostID != userID.(primitive.ObjectID) {
		isCoHost := false
		for _, coHostID := range event.CoHosts {
			if coHostID == userID.(primitive.ObjectID) {
				isCoHost = true
				break
			}
		}

		if !isCoHost {
			isAdmin, err := eventService.IsUserAdmin(c.Request.Context(), userID.(primitive.ObjectID))
			if err != nil {
				response.Error(c, http.StatusInternalServerError, "Failed to check user permissions", err)
				return
			}

			if !isAdmin {
				response.Error(c, http.StatusForbidden, "Only the event host, co-host, or an admin can view analytics", nil)
				return
			}
		}
	}

	// Get the analytics service
	analyticsService := c.MustGet("analyticsService").(AnalyticsService)

	// Get referral sources
	sources, err := analyticsService.GetEventReferralSources(c.Request.Context(), eventID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve referral sources", err)
		return
	}

	response.Success(c, http.StatusOK, "Referral sources retrieved successfully", sources)
}

// GetEventEngagementMetrics retrieves engagement metrics for an event
func GetEventEngagementMetrics(c *gin.Context) {
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

	// Check permissions
	if event.HostID != userID.(primitive.ObjectID) {
		isCoHost := false
		for _, coHostID := range event.CoHosts {
			if coHostID == userID.(primitive.ObjectID) {
				isCoHost = true
				break
			}
		}

		if !isCoHost {
			isAdmin, err := eventService.IsUserAdmin(c.Request.Context(), userID.(primitive.ObjectID))
			if err != nil {
				response.Error(c, http.StatusInternalServerError, "Failed to check user permissions", err)
				return
			}

			if !isAdmin {
				response.Error(c, http.StatusForbidden, "Only the event host, co-host, or an admin can view analytics", nil)
				return
			}
		}
	}

	// Get the analytics service
	analyticsService := c.MustGet("analyticsService").(AnalyticsService)

	// Get engagement metrics
	metrics, err := analyticsService.GetEventEngagementMetrics(c.Request.Context(), eventID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve engagement metrics", err)
		return
	}

	response.Success(c, http.StatusOK, "Engagement metrics retrieved successfully", metrics)
}

// GetHostAnalytics retrieves analytics for all events hosted by a user
func GetHostAnalytics(c *gin.Context) {
	// Get user ID from URL parameter or use authenticated user
	userIDStr := c.Param("user_id")
	var targetUserID primitive.ObjectID
	var err error

	if userIDStr != "" {
		targetUserID, err = primitive.ObjectIDFromHex(userIDStr)
		if err != nil {
			response.Error(c, http.StatusBadRequest, "Invalid user ID format", err)
			return
		}
	} else {
		// Use the authenticated user's ID
		currentUserID, exists := c.Get("userID")
		if !exists {
			response.Error(c, http.StatusUnauthorized, "User not authenticated", nil)
			return
		}
		targetUserID = currentUserID.(primitive.ObjectID)
	}

	// Get query parameters
	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		limit = 10
	}

	// Get the authenticated user's ID
	currentUserID, exists := c.Get("userID")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	// Check permissions if viewing another user's analytics
	if targetUserID != currentUserID.(primitive.ObjectID) {
		// Get the event service
		eventService := c.MustGet("eventService").(EventService)

		// Check if current user is an admin
		isAdmin, err := eventService.IsUserAdmin(c.Request.Context(), currentUserID.(primitive.ObjectID))
		if err != nil {
			response.Error(c, http.StatusInternalServerError, "Failed to check user permissions", err)
			return
		}

		if !isAdmin {
			response.Error(c, http.StatusForbidden, "You don't have permission to view this user's analytics", nil)
			return
		}
	}

	// Get the analytics service
	analyticsService := c.MustGet("analyticsService").(AnalyticsService)

	// Get host analytics
	analytics, err := analyticsService.GetHostAnalytics(c.Request.Context(), targetUserID, limit)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve host analytics", err)
		return
	}

	response.Success(c, http.StatusOK, "Host analytics retrieved successfully", analytics)
}

// GetEventConversionRate retrieves the conversion rate for an event (views to RSVPs)
func GetEventConversionRate(c *gin.Context) {
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

	// Check permissions
	if event.HostID != userID.(primitive.ObjectID) {
		isCoHost := false
		for _, coHostID := range event.CoHosts {
			if coHostID == userID.(primitive.ObjectID) {
				isCoHost = true
				break
			}
		}

		if !isCoHost {
			isAdmin, err := eventService.IsUserAdmin(c.Request.Context(), userID.(primitive.ObjectID))
			if err != nil {
				response.Error(c, http.StatusInternalServerError, "Failed to check user permissions", err)
				return
			}

			if !isAdmin {
				response.Error(c, http.StatusForbidden, "Only the event host, co-host, or an admin can view analytics", nil)
				return
			}
		}
	}

	// Get the analytics service
	analyticsService := c.MustGet("analyticsService").(AnalyticsService)

	// Get conversion rate
	rate, err := analyticsService.GetEventConversionRate(c.Request.Context(), eventID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve conversion rate", err)
		return
	}

	response.Success(c, http.StatusOK, "Conversion rate retrieved successfully", gin.H{
		"conversion_rate": rate,
	})
}

// GetEventDemographics retrieves demographic information about event attendees
func GetEventDemographics(c *gin.Context) {
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

	// Check permissions
	if event.HostID != userID.(primitive.ObjectID) {
		isCoHost := false
		for _, coHostID := range event.CoHosts {
			if coHostID == userID.(primitive.ObjectID) {
				isCoHost = true
				break
			}
		}

		if !isCoHost {
			isAdmin, err := eventService.IsUserAdmin(c.Request.Context(), userID.(primitive.ObjectID))
			if err != nil {
				response.Error(c, http.StatusInternalServerError, "Failed to check user permissions", err)
				return
			}

			if !isAdmin {
				response.Error(c, http.StatusForbidden, "Only the event host, co-host, or an admin can view analytics", nil)
				return
			}
		}
	}

	// Get the analytics service
	analyticsService := c.MustGet("analyticsService").(AnalyticsService)

	// Get demographics
	demographics, err := analyticsService.GetEventDemographics(c.Request.Context(), eventID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve demographic information", err)
		return
	}

	response.Success(c, http.StatusOK, "Demographic information retrieved successfully", demographics)
}
