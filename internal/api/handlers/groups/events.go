package groups

import (
	"net/http"
	"strconv"

	"github.com/Caqil/vyrall/internal/utils/response"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// EventService defines the interface for group event operations
type EventService interface {
	GetEventsByGroupID(ctx context.Context, groupID primitive.ObjectID, limit, offset int) ([]*models.Event, int, error)
	GetUpcomingEventsByGroupID(ctx context.Context, groupID primitive.ObjectID, limit, offset int) ([]*models.Event, int, error)
	GetPastEventsByGroupID(ctx context.Context, groupID primitive.ObjectID, limit, offset int) ([]*models.Event, int, error)
	CanCreateGroupEvent(ctx context.Context, groupID, userID primitive.ObjectID) (bool, error)
}

// GetGroupEvents retrieves events for a specific group
func GetGroupEvents(c *gin.Context) {
	// Get group ID from URL parameter
	groupIDStr := c.Param("id")
	groupID, err := primitive.ObjectIDFromHex(groupIDStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid group ID format", err)
		return
	}

	// Get query parameters
	filter := c.DefaultQuery("filter", "upcoming") // upcoming, past, all
	limitStr := c.DefaultQuery("limit", "20")
	offsetStr := c.DefaultQuery("offset", "0")

	// Parse pagination params
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		limit = 20
	}
	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	// Get the group service
	groupService := c.MustGet("groupService").(GroupService)

	// Check if the group exists
	group, err := groupService.GetGroupByID(c.Request.Context(), groupID)
	if err != nil {
		response.Error(c, http.StatusNotFound, "Group not found", err)
		return
	}

	// If the group is not public, check if the user is a member
	if !group.IsPublic {
		// Get the authenticated user's ID
		userID, exists := c.Get("userID")
		if !exists {
			response.Error(c, http.StatusUnauthorized, "Authentication required for private groups", nil)
			return
		}

		// Check if the user is a member of the group
		isMember, err := groupService.IsGroupMember(c.Request.Context(), groupID, userID.(primitive.ObjectID))
		if err != nil {
			response.Error(c, http.StatusInternalServerError, "Failed to check group membership", err)
			return
		}

		if !isMember {
			response.Error(c, http.StatusForbidden, "You must be a member to view group events", nil)
			return
		}
	}

	// Get the event service
	eventService := c.MustGet("eventService").(EventService)

	// Get events based on filter
	var events []*models.Event
	var total int
	switch filter {
	case "upcoming":
		events, total, err = eventService.GetUpcomingEventsByGroupID(c.Request.Context(), groupID, limit, offset)
	case "past":
		events, total, err = eventService.GetPastEventsByGroupID(c.Request.Context(), groupID, limit, offset)
	default:
		events, total, err = eventService.GetEventsByGroupID(c.Request.Context(), groupID, limit, offset)
	}

	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve group events", err)
		return
	}

	response.SuccessWithPagination(c, http.StatusOK, "Group events retrieved successfully", events, limit, offset, total)
}

// CanCreateEvent checks if the current user can create events in this group
func CanCreateEvent(c *gin.Context) {
	// Get group ID from URL parameter
	groupIDStr := c.Param("id")
	groupID, err := primitive.ObjectIDFromHex(groupIDStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid group ID format", err)
		return
	}

	// Get the authenticated user's ID
	userID, exists := c.Get("userID")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	// Get the group service
	groupService := c.MustGet("groupService").(GroupService)

	// Check if the group exists
	_, err = groupService.GetGroupByID(c.Request.Context(), groupID)
	if err != nil {
		response.Error(c, http.StatusNotFound, "Group not found", err)
		return
	}

	// Get the event service
	eventService := c.MustGet("eventService").(EventService)

	// Check if the user can create events
	canCreate, err := eventService.CanCreateGroupEvent(c.Request.Context(), groupID, userID.(primitive.ObjectID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to check permissions", err)
		return
	}

	response.Success(c, http.StatusOK, "Permission check completed", gin.H{
		"can_create_event": canCreate,
	})
}
