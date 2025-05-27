package users

import (
	"net/http"

	"github.com/Caqil/vyrall/internal/services/user"
	"github.com/Caqil/vyrall/internal/utils/response"
	"github.com/Caqil/vyrall/internal/utils/validation"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// SuggestionsHandler handles user suggestions operations
type SuggestionsHandler struct {
	userService *user.Service
}

// NewSuggestionsHandler creates a new suggestions handler
func NewSuggestionsHandler(userService *user.Service) *SuggestionsHandler {
	return &SuggestionsHandler{
		userService: userService,
	}
}

// GetSuggestedUsers handles the request to get suggested users to follow
func (h *SuggestionsHandler) GetSuggestedUsers(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Get pagination parameters
	limit, offset := response.GetPaginationParams(c)

	// Get suggested users
	suggestions, total, err := h.userService.GetSuggestedUsers(c.Request.Context(), userID.(primitive.ObjectID), limit, offset)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get suggested users", err)
		return
	}

	// Return paginated response
	response.SuccessWithPagination(c, http.StatusOK, "Suggested users retrieved successfully", suggestions, limit, offset, total)
}

// GetSuggestedUsersByInterests handles the request to get suggested users based on interests
func (h *SuggestionsHandler) GetSuggestedUsersByInterests(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Get pagination parameters
	limit, offset := response.GetPaginationParams(c)

	// Get suggested users by interests
	suggestions, total, err := h.userService.GetSuggestedUsersByInterests(c.Request.Context(), userID.(primitive.ObjectID), limit, offset)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get suggested users by interests", err)
		return
	}

	// Return paginated response
	response.SuccessWithPagination(c, http.StatusOK, "Suggested users retrieved successfully", suggestions, limit, offset, total)
}

// GetSuggestedUsersByMutualFriends handles the request to get suggested users based on mutual friends
func (h *SuggestionsHandler) GetSuggestedUsersByMutualFriends(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Get pagination parameters
	limit, offset := response.GetPaginationParams(c)

	// Get suggested users by mutual friends
	suggestions, total, err := h.userService.GetSuggestedUsersByMutualFriends(c.Request.Context(), userID.(primitive.ObjectID), limit, offset)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get suggested users by mutual friends", err)
		return
	}

	// Return paginated response
	response.SuccessWithPagination(c, http.StatusOK, "Suggested users retrieved successfully", suggestions, limit, offset, total)
}

// GetSuggestedUsersByLocation handles the request to get suggested users based on location
func (h *SuggestionsHandler) GetSuggestedUsersByLocation(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Get pagination parameters
	limit, offset := response.GetPaginationParams(c)

	// Get suggested users by location
	suggestions, total, err := h.userService.GetSuggestedUsersByLocation(c.Request.Context(), userID.(primitive.ObjectID), limit, offset)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get suggested users by location", err)
		return
	}

	// Return paginated response
	response.SuccessWithPagination(c, http.StatusOK, "Suggested users retrieved successfully", suggestions, limit, offset, total)
}

// DismissSuggestion handles the request to dismiss a user suggestion
func (h *SuggestionsHandler) DismissSuggestion(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Get suggested user ID from URL parameter
	suggestedIDStr := c.Param("id")
	if !validation.IsValidObjectID(suggestedIDStr) {
		response.ValidationError(c, "Invalid user ID", nil)
		return
	}
	suggestedID, _ := primitive.ObjectIDFromHex(suggestedIDStr)

	// Dismiss the suggestion
	err := h.userService.DismissSuggestion(c.Request.Context(), userID.(primitive.ObjectID), suggestedID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to dismiss suggestion", err)
		return
	}

	// Return success response
	response.OK(c, "Suggestion dismissed successfully", nil)
}
