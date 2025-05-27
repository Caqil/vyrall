package search

import (
	"net/http"

	"github.com/Caqil/vyrall/internal/services/search"
	"github.com/Caqil/vyrall/internal/utils/response"
	"github.com/gin-gonic/gin"
)

// LocationSearchHandler handles location search operations
type LocationSearchHandler struct {
	searchService *search.Service
}

// NewLocationSearchHandler creates a new location search handler
func NewLocationSearchHandler(searchService *search.Service) *LocationSearchHandler {
	return &LocationSearchHandler{
		searchService: searchService,
	}
}

// SearchLocations handles the request to search for locations
func (h *LocationSearchHandler) SearchLocations(c *gin.Context) {
	// Get query parameters
	query := c.Query("q")
	if query == "" {
		response.ValidationError(c, "Search query is required", nil)
		return
	}

	// Get location parameters
	lat := c.DefaultQuery("lat", "")
	lng := c.DefaultQuery("lng", "")
	radius := c.DefaultQuery("radius", "50") // Default 50km radius

	// Get pagination parameters
	limit, offset := response.GetPaginationParams(c)

	// Search locations
	locations, total, err := h.searchService.SearchLocations(
		c.Request.Context(),
		query,
		lat,
		lng,
		radius,
		limit,
		offset,
	)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to search locations", err)
		return
	}

	// Return paginated response
	response.SuccessWithPagination(c, http.StatusOK, "Location search results retrieved successfully", locations, limit, offset, total)
}

// GetNearbyLocations handles the request to get nearby locations
func (h *LocationSearchHandler) GetNearbyLocations(c *gin.Context) {
	// Get location parameters
	lat := c.Query("lat")
	lng := c.Query("lng")

	if lat == "" || lng == "" {
		response.ValidationError(c, "Latitude and longitude are required", nil)
		return
	}

	radius := c.DefaultQuery("radius", "5") // Default 5km radius

	// Get pagination parameters
	limit, offset := response.GetPaginationParams(c)

	// Get nearby locations
	locations, total, err := h.searchService.GetNearbyLocations(
		c.Request.Context(),
		lat,
		lng,
		radius,
		limit,
		offset,
	)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get nearby locations", err)
		return
	}

	// Return paginated response
	response.SuccessWithPagination(c, http.StatusOK, "Nearby locations retrieved successfully", locations, limit, offset, total)
}
