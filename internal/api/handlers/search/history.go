package search

import (
	"net/http"

	"github.com/Caqil/vyrall/internal/services/search"
	"github.com/Caqil/vyrall/internal/utils/response"
	"github.com/Caqil/vyrall/internal/utils/validation"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// HistoryHandler handles search history operations
type HistoryHandler struct {
	searchService *search.Service
}

// NewHistoryHandler creates a new history handler
func NewHistoryHandler(searchService *search.Service) *HistoryHandler {
	return &HistoryHandler{
		searchService: searchService,
	}
}

// GetSearchHistory handles the request to get user's search history
func (h *HistoryHandler) GetSearchHistory(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Get query parameters
	searchType := c.DefaultQuery("type", "all") // all, users, posts, hashtags, etc.

	// Get pagination parameters
	limit, offset := response.GetPaginationParams(c)

	// Get search history
	history, total, err := h.searchService.GetSearchHistory(
		c.Request.Context(),
		userID.(primitive.ObjectID),
		searchType,
		limit,
		offset,
	)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get search history", err)
		return
	}

	// Return paginated response
	response.SuccessWithPagination(c, http.StatusOK, "Search history retrieved successfully", history, limit, offset, total)
}

// ClearSearchHistory handles the request to clear user's search history
func (h *HistoryHandler) ClearSearchHistory(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Get query parameter for specific search ID (optional)
	searchID := c.Query("id")
	var searchIDObj *primitive.ObjectID

	if searchID != "" && validation.IsValidObjectID(searchID) {
		id, _ := primitive.ObjectIDFromHex(searchID)
		searchIDObj = &id
	}

	// Clear search history
	err := h.searchService.ClearSearchHistory(
		c.Request.Context(),
		userID.(primitive.ObjectID),
		searchIDObj,
	)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to clear search history", err)
		return
	}

	// Return success response
	response.OK(c, "Search history cleared successfully", nil)
}

// DeleteSearchHistoryItem handles the request to delete a specific search history item
func (h *HistoryHandler) DeleteSearchHistoryItem(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Get search history ID from URL parameter
	historyIDStr := c.Param("id")
	if !validation.IsValidObjectID(historyIDStr) {
		response.ValidationError(c, "Invalid history ID", nil)
		return
	}
	historyID, _ := primitive.ObjectIDFromHex(historyIDStr)

	// Delete search history item
	err := h.searchService.DeleteSearchHistoryItem(
		c.Request.Context(),
		userID.(primitive.ObjectID),
		historyID,
	)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to delete search history item", err)
		return
	}

	// Return success response
	response.OK(c, "Search history item deleted successfully", nil)
}
