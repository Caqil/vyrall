package search

import (
	"net/http"

	"github.com/Caqil/vyrall/internal/services/search"
	"github.com/Caqil/vyrall/internal/utils/response"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// SuggestionHandler handles search suggestion operations
type SuggestionHandler struct {
	searchService *search.Service
}

// NewSuggestionHandler creates a new suggestion handler
func NewSuggestionHandler(searchService *search.Service) *SuggestionHandler {
	return &SuggestionHandler{
		searchService: searchService,
	}
}

// GetSearchSuggestions handles the request to get search suggestions
func (h *SuggestionHandler) GetSearchSuggestions(c *gin.Context) {
	// Get user ID from context (may be nil for unauthenticated users)
	var userID primitive.ObjectID
	if id, exists := c.Get("userID"); exists {
		userID = id.(primitive.ObjectID)
	}

	// Get query parameters
	query := c.Query("q")
	if query == "" {
		response.ValidationError(c, "Search query is required", nil)
		return
	}

	// Get suggestion type
	suggestionType := c.DefaultQuery("type", "all") // all, users, hashtags, locations

	// Get limit for suggestions
	limit := 10
	if limitParam := c.Query("limit"); limitParam != "" {
		if parsedLimit, err := primitive.ParseInt(limitParam); err == nil && parsedLimit > 0 {
			limit = int(parsedLimit)
		}
	}

	// Get search suggestions
	suggestions, err := h.searchService.GetSearchSuggestions(
		c.Request.Context(),
		query,
		suggestionType,
		userID,
		limit,
	)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get search suggestions", err)
		return
	}

	// Return success response
	response.OK(c, "Search suggestions retrieved successfully", suggestions)
}

// GetAutocompleteSuggestions handles the request to get autocomplete suggestions
func (h *SuggestionHandler) GetAutocompleteSuggestions(c *gin.Context) {
	// Get user ID from context (may be nil for unauthenticated users)
	var userID primitive.ObjectID
	if id, exists := c.Get("userID"); exists {
		userID = id.(primitive.ObjectID)
	}

	// Get query parameters
	query := c.Query("q")
	if query == "" {
		response.ValidationError(c, "Search query is required", nil)
		return
	}

	// Get suggestion type
	suggestionType := c.DefaultQuery("type", "all") // all, users, hashtags, locations

	// Get limit for suggestions
	limit := 5
	if limitParam := c.Query("limit"); limitParam != "" {
		if parsedLimit, err := primitive.ParseInt(limitParam); err == nil && parsedLimit > 0 {
			limit = int(parsedLimit)
		}
	}

	// Get autocomplete suggestions
	suggestions, err := h.searchService.GetAutocompleteSuggestions(
		c.Request.Context(),
		query,
		suggestionType,
		userID,
		limit,
	)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get autocomplete suggestions", err)
		return
	}

	// Return success response
	response.OK(c, "Autocomplete suggestions retrieved successfully", suggestions)
}
