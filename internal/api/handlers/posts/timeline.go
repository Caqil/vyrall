package posts

import (
	"net/http"

	"github.com/Caqil/vyrall/internal/services/post"
	"github.com/Caqil/vyrall/internal/utils/response"
	"github.com/Caqil/vyrall/internal/utils/validation"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// TimelineHandler handles post timeline operations
type TimelineHandler struct {
	postService *post.Service
}

// NewTimelineHandler creates a new timeline handler
func NewTimelineHandler(postService *post.Service) *TimelineHandler {
	return &TimelineHandler{
		postService: postService,
	}
}

// GetUserTimeline handles the request to get a user's timeline
func (h *TimelineHandler) GetUserTimeline(c *gin.Context) {
	// Get authenticated user ID (may be nil for unauthenticated users)
	var authUserID primitive.ObjectID
	if id, exists := c.Get("userID"); exists {
		authUserID = id.(primitive.ObjectID)
	}

	// Get target user ID from URL parameter
	userIDStr := c.Param("userId")
	if !validation.IsValidObjectID(userIDStr) {
		response.ValidationError(c, "Invalid user ID", nil)
		return
	}
	userID, _ := primitive.ObjectIDFromHex(userIDStr)

	// Get pagination parameters
	limit, offset := response.GetPaginationParams(c)

	// Get user timeline
	posts, total, err := h.postService.GetUserTimeline(c.Request.Context(), userID, authUserID, limit, offset)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve user timeline", err)
		return
	}

	// Return paginated response
	response.SuccessWithPagination(c, http.StatusOK, "User timeline retrieved successfully", posts, limit, offset, total)
}

// GetHomeTimeline handles the request to get the authenticated user's home timeline
func (h *TimelineHandler) GetHomeTimeline(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Get pagination parameters
	limit, offset := response.GetPaginationParams(c)

	// Get home timeline
	posts, total, err := h.postService.GetHomeTimeline(c.Request.Context(), userID.(primitive.ObjectID), limit, offset)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve home timeline", err)
		return
	}

	// Return paginated response
	response.SuccessWithPagination(c, http.StatusOK, "Home timeline retrieved successfully", posts, limit, offset, total)
}

// GetGroupTimeline handles the request to get a group's timeline
func (h *TimelineHandler) GetGroupTimeline(c *gin.Context) {
	// Get user ID from context (may be nil for unauthenticated users)
	var userID primitive.ObjectID
	if id, exists := c.Get("userID"); exists {
		userID = id.(primitive.ObjectID)
	}

	// Get group ID from URL parameter
	groupIDStr := c.Param("groupId")
	if !validation.IsValidObjectID(groupIDStr) {
		response.ValidationError(c, "Invalid group ID", nil)
		return
	}
	groupID, _ := primitive.ObjectIDFromHex(groupIDStr)

	// Get pagination parameters
	limit, offset := response.GetPaginationParams(c)

	// Get group timeline
	posts, total, err := h.postService.GetGroupTimeline(c.Request.Context(), groupID, userID, limit, offset)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve group timeline", err)
		return
	}

	// Return paginated response
	response.SuccessWithPagination(c, http.StatusOK, "Group timeline retrieved successfully", posts, limit, offset, total)
}

// GetLocationTimeline handles the request to get a location's timeline
func (h *TimelineHandler) GetLocationTimeline(c *gin.Context) {
	// Get user ID from context (may be nil for unauthenticated users)
	var userID primitive.ObjectID
	if id, exists := c.Get("userID"); exists {
		userID = id.(primitive.ObjectID)
	}

	// Parse query parameters
	locationName := c.Query("name")
	lat, _ := primitive.ParseFloat(c.Query("lat"))
	lng, _ := primitive.ParseFloat(c.Query("lng"))
	radius, _ := primitive.ParseFloat(c.DefaultQuery("radius", "5")) // Default 5km radius

	// Validate parameters
	if locationName == "" && (lat == 0 || lng == 0) {
		response.ValidationError(c, "Location name or coordinates (lat/lng) are required", nil)
		return
	}

	// Get pagination parameters
	limit, offset := response.GetPaginationParams(c)

	// Get location timeline
	posts, total, err := h.postService.GetLocationTimeline(
		c.Request.Context(),
		locationName,
		lat,
		lng,
		radius,
		userID,
		limit,
		offset,
	)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve location timeline", err)
		return
	}

	// Return paginated response
	response.SuccessWithPagination(c, http.StatusOK, "Location timeline retrieved successfully", posts, limit, offset, total)
}
