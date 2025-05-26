package events

import (
	"net/http"
	"strconv"
	"time"

	"github.com/Caqil/vyrall/internal/utils/response"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ListEvents retrieves a list of events based on various filters
func ListEvents(c *gin.Context) {
	// Get query parameters
	timeFilter := c.DefaultQuery("filter", "upcoming") // upcoming, past, all
	categoryFilter := c.DefaultQuery("category", "")
	typeFilter := c.DefaultQuery("type", "") // in-person, online, hybrid
	startDateStr := c.DefaultQuery("start_date", "")
	endDateStr := c.DefaultQuery("end_date", "")
	locationStr := c.DefaultQuery("location", "")
	radiusStr := c.DefaultQuery("radius", "10") // default 10km
	groupIDStr := c.DefaultQuery("group_id", "")
	userIDStr := c.DefaultQuery("user_id", "")
	limitStr := c.DefaultQuery("limit", "20")
	offsetStr := c.DefaultQuery("offset", "0")
	sortBy := c.DefaultQuery("sort_by", "start_time")
	sortOrder := c.DefaultQuery("sort_order", "asc")

	// Parse pagination params
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		limit = 20
	}
	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	// Build filter map
	filter := make(map[string]interface{})

	// Time filter
	now := time.Now()
	switch timeFilter {
	case "upcoming":
		filter["start_time"] = map[string]interface{}{
			"$gte": now,
		}
	case "past":
		filter["end_time"] = map[string]interface{}{
			"$lt": now,
		}
	case "today":
		startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		endOfDay := startOfDay.Add(24 * time.Hour)
		filter["$or"] = []map[string]interface{}{
			{
				"start_time": map[string]interface{}{
					"$gte": startOfDay,
					"$lt":  endOfDay,
				},
			},
			{
				"start_time": map[string]interface{}{
					"$lt": startOfDay,
				},
				"end_time": map[string]interface{}{
					"$gte": startOfDay,
				},
			},
		}
	case "week":
		startOfWeek := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		for startOfWeek.Weekday() != time.Sunday {
			startOfWeek = startOfWeek.AddDate(0, 0, -1)
		}
		endOfWeek := startOfWeek.AddDate(0, 0, 7)
		filter["start_time"] = map[string]interface{}{
			"$gte": startOfWeek,
			"$lt":  endOfWeek,
		}
	}

	// Category filter
	if categoryFilter != "" {
		filter["category"] = categoryFilter
	}

	// Type filter
	if typeFilter != "" {
		filter["type"] = typeFilter
	}

	// Date range filter
	if startDateStr != "" {
		startDate, err := time.Parse("2006-01-02", startDateStr)
		if err == nil {
			if filter["start_time"] == nil {
				filter["start_time"] = map[string]interface{}{}
			}
			filter["start_time"].(map[string]interface{})["$gte"] = startDate
		}
	}
	if endDateStr != "" {
		endDate, err := time.Parse("2006-01-02", endDateStr)
		if err == nil {
			endDate = endDate.Add(24 * time.Hour) // Include the entire end date
			if filter["end_time"] == nil {
				filter["end_time"] = map[string]interface{}{}
			}
			filter["end_time"].(map[string]interface{})["$lte"] = endDate
		}
	}

	// Location filter
	if locationStr != "" {
		// If location coordinates are provided, search by radius
		// This would typically use geospatial query features of MongoDB
		radius, err := strconv.ParseFloat(radiusStr, 64)
		if err != nil || radius <= 0 {
			radius = 10 // Default to 10km
		}

		// Parse lat,lng from locationStr
		// This is simplified and would need proper geo handling
		filter["location"] = locationStr
		filter["radius"] = radius
	}

	// Group filter
	if groupIDStr != "" {
		groupID, err := primitive.ObjectIDFromHex(groupIDStr)
		if err == nil {
			filter["group_id"] = groupID
		}
	}

	// User filter (events hosted by this user)
	if userIDStr != "" {
		userID, err := primitive.ObjectIDFromHex(userIDStr)
		if err == nil {
			filter["host_id"] = userID
		}
	}

	// Build sort options
	sort := make(map[string]int)
	if sortOrder == "desc" {
		sort[sortBy] = -1
	} else {
		sort[sortBy] = 1
	}

	// Get the event service
	eventService := c.MustGet("eventService").(EventService)

	// Check for current user (to handle private events)
	userID, _ := c.Get("userID")

	// Get events
	events, total, err := eventService.ListEvents(c.Request.Context(), filter, sort, userID, limit, offset)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve events", err)
		return
	}

	response.SuccessWithPagination(c, http.StatusOK, "Events retrieved successfully", events, limit, offset, total)
}

// GetEventDetail retrieves detailed information about a specific event
func GetEventDetail(c *gin.Context) {
	// Get event ID from URL parameter
	eventIDStr := c.Param("id")
	eventID, err := primitive.ObjectIDFromHex(eventIDStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid event ID format", err)
		return
	}

	// Get the event service
	eventService := c.MustGet("eventService").(EventService)

	// Get the current user ID if authenticated
	userID, _ := c.Get("userID")

	// Get the event
	event, err := eventService.GetEventByID(c.Request.Context(), eventID)
	if err != nil {
		response.Error(c, http.StatusNotFound, "Event not found", err)
		return
	}

	// Check if user can view this event
	if event.Privacy != "public" {
		if userID == nil {
			response.Error(c, http.StatusUnauthorized, "Authentication required to view this event", nil)
			return
		}

		// Check if user is host, co-host, or attendee
		hasAccess, err := eventService.CanUserAccessEvent(c.Request.Context(), eventID, userID.(primitive.ObjectID))
		if err != nil {
			response.Error(c, http.StatusInternalServerError, "Failed to check event access", err)
			return
		}

		if !hasAccess {
			response.Error(c, http.StatusForbidden, "You don't have permission to view this event", nil)
			return
		}
	}

	// Get additional event data
	eventDetail, err := eventService.GetEventDetailWithAttendees(c.Request.Context(), event, userID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve event details", err)
		return
	}

	response.Success(c, http.StatusOK, "Event details retrieved successfully", eventDetail)
}
