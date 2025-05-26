package response

import (
	"math"
	"strconv"

	"github.com/gin-gonic/gin"
)

// PaginationInfo represents pagination metadata
type PaginationInfo struct {
	TotalItems   int  `json:"total_items"`
	TotalPages   int  `json:"total_pages"`
	CurrentPage  int  `json:"current_page"`
	ItemsPerPage int  `json:"items_per_page"`
	HasNext      bool `json:"has_next"`
	HasPrevious  bool `json:"has_previous"`
	NextPage     int  `json:"next_page,omitempty"`
	PreviousPage int  `json:"previous_page,omitempty"`
}

// NewPaginationInfo creates a new pagination info object
func NewPaginationInfo(total, limit, offset int) *PaginationInfo {
	// Calculate current page (1-based)
	currentPage := offset/limit + 1

	// Calculate total pages
	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	// Create pagination info
	pagination := &PaginationInfo{
		TotalItems:   total,
		TotalPages:   totalPages,
		CurrentPage:  currentPage,
		ItemsPerPage: limit,
		HasNext:      currentPage < totalPages,
		HasPrevious:  currentPage > 1,
	}

	// Set next page if exists
	if pagination.HasNext {
		pagination.NextPage = currentPage + 1
	}

	// Set previous page if exists
	if pagination.HasPrevious {
		pagination.PreviousPage = currentPage - 1
	}

	return pagination
}

// GetPaginationParams extracts pagination parameters from the request
func GetPaginationParams(c *gin.Context) (limit, offset int) {
	// Get limit and offset from query parameters
	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")
	pageStr := c.DefaultQuery("page", "")

	// Parse limit
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		limit = 10
	}

	// Parse offset
	offset, err = strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	// If page is provided, calculate offset
	if pageStr != "" {
		page, err := strconv.Atoi(pageStr)
		if err == nil && page > 0 {
			offset = (page - 1) * limit
		}
	}

	// Enforce reasonable limits
	if limit > 100 {
		limit = 100
	}

	return limit, offset
}

// BuildPaginationLinks builds pagination links for HATEOAS
func BuildPaginationLinks(c *gin.Context, pagination *PaginationInfo) map[string]string {
	baseURL := c.Request.URL.Path
	query := c.Request.URL.Query()

	// Remove existing pagination parameters
	query.Del("page")
	query.Del("offset")

	// Build links
	links := make(map[string]string)

	// Set limit parameter
	query.Set("limit", strconv.Itoa(pagination.ItemsPerPage))

	// First page
	query.Set("page", "1")
	links["first"] = baseURL + "?" + query.Encode()

	// Last page
	query.Set("page", strconv.Itoa(pagination.TotalPages))
	links["last"] = baseURL + "?" + query.Encode()

	// Next page
	if pagination.HasNext {
		query.Set("page", strconv.Itoa(pagination.NextPage))
		links["next"] = baseURL + "?" + query.Encode()
	}

	// Previous page
	if pagination.HasPrevious {
		query.Set("page", strconv.Itoa(pagination.PreviousPage))
		links["prev"] = baseURL + "?" + query.Encode()
	}

	return links
}
