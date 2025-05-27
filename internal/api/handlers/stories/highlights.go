package stories

import (
	"net/http"

	"github.com/Caqil/vyrall/internal/services/story"
	"github.com/Caqil/vyrall/internal/utils/response"
	"github.com/Caqil/vyrall/internal/utils/validation"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// HighlightsHandler handles story highlights operations
type HighlightsHandler struct {
	storyService *story.Service
}

// NewHighlightsHandler creates a new highlights handler
func NewHighlightsHandler(storyService *story.Service) *HighlightsHandler {
	return &HighlightsHandler{
		storyService: storyService,
	}
}

// CreateHighlight handles the request to create a story highlight
func (h *HighlightsHandler) CreateHighlight(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Parse request body
	var req struct {
		Title      string   `json:"title" binding:"required"`
		CoverID    string   `json:"cover_id,omitempty"`
		StoryIDs   []string `json:"story_ids" binding:"required"`
		IsPrivate  bool     `json:"is_private,omitempty"`
		CategoryID string   `json:"category_id,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "Invalid request body", err.Error())
		return
	}

	if len(req.StoryIDs) == 0 {
		response.ValidationError(c, "At least one story is required", nil)
		return
	}

	// Convert story IDs to ObjectIDs
	storyIDs := make([]primitive.ObjectID, 0, len(req.StoryIDs))
	for _, idStr := range req.StoryIDs {
		if !validation.IsValidObjectID(idStr) {
			response.ValidationError(c, "Invalid story ID: "+idStr, nil)
			return
		}
		storyID, _ := primitive.ObjectIDFromHex(idStr)
		storyIDs = append(storyIDs, storyID)
	}

	// Convert cover ID to ObjectID if provided
	var coverID *primitive.ObjectID
	if req.CoverID != "" && validation.IsValidObjectID(req.CoverID) {
		id, _ := primitive.ObjectIDFromHex(req.CoverID)
		coverID = &id
	}

	// Convert category ID to ObjectID if provided
	var categoryID *primitive.ObjectID
	if req.CategoryID != "" && validation.IsValidObjectID(req.CategoryID) {
		id, _ := primitive.ObjectIDFromHex(req.CategoryID)
		categoryID = &id
	}

	// Create highlight
	highlight, err := h.storyService.CreateHighlight(
		c.Request.Context(),
		userID.(primitive.ObjectID),
		req.Title,
		coverID,
		storyIDs,
		req.IsPrivate,
		categoryID,
	)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to create highlight", err)
		return
	}

	// Return success response
	response.Created(c, "Highlight created successfully", highlight)
}

// GetUserHighlights handles the request to get user's highlights
func (h *HighlightsHandler) GetUserHighlights(c *gin.Context) {
	// Get target user ID from URL parameter
	userIDStr := c.Param("userId")
	if !validation.IsValidObjectID(userIDStr) {
		response.ValidationError(c, "Invalid user ID", nil)
		return
	}
	targetUserID, _ := primitive.ObjectIDFromHex(userIDStr)

	// Get authenticated user ID (may be nil for unauthenticated users)
	var authUserID primitive.ObjectID
	if id, exists := c.Get("userID"); exists {
		authUserID = id.(primitive.ObjectID)
	}

	// Get pagination parameters
	limit, offset := response.GetPaginationParams(c)

	// Get user highlights
	highlights, total, err := h.storyService.GetUserHighlights(c.Request.Context(), targetUserID, authUserID, limit, offset)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve highlights", err)
		return
	}

	// Return paginated response
	response.SuccessWithPagination(c, http.StatusOK, "Highlights retrieved successfully", highlights, limit, offset, total)
}

// GetHighlight handles the request to get a specific highlight
func (h *HighlightsHandler) GetHighlight(c *gin.Context) {
	// Get highlight ID from URL parameter
	highlightIDStr := c.Param("id")
	if !validation.IsValidObjectID(highlightIDStr) {
		response.ValidationError(c, "Invalid highlight ID", nil)
		return
	}
	highlightID, _ := primitive.ObjectIDFromHex(highlightIDStr)

	// Get authenticated user ID (may be nil for unauthenticated users)
	var userID primitive.ObjectID
	if id, exists := c.Get("userID"); exists {
		userID = id.(primitive.ObjectID)
	}

	// Get the highlight
	highlight, err := h.storyService.GetHighlight(c.Request.Context(), highlightID, userID)
	if err != nil {
		response.NotFoundError(c, "Highlight not found")
		return
	}

	// Return success response
	response.OK(c, "Highlight retrieved successfully", highlight)
}

// UpdateHighlight handles the request to update a highlight
func (h *HighlightsHandler) UpdateHighlight(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Get highlight ID from URL parameter
	highlightIDStr := c.Param("id")
	if !validation.IsValidObjectID(highlightIDStr) {
		response.ValidationError(c, "Invalid highlight ID", nil)
		return
	}
	highlightID, _ := primitive.ObjectIDFromHex(highlightIDStr)

	// Parse request body
	var req struct {
		Title      string   `json:"title,omitempty"`
		CoverID    string   `json:"cover_id,omitempty"`
		StoryIDs   []string `json:"story_ids,omitempty"`
		IsPrivate  *bool    `json:"is_private,omitempty"`
		CategoryID string   `json:"category_id,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "Invalid request body", err.Error())
		return
	}

	// Build updates map
	updates := make(map[string]interface{})

	if req.Title != "" {
		updates["title"] = req.Title
	}

	if req.CoverID != "" {
		if !validation.IsValidObjectID(req.CoverID) {
			response.ValidationError(c, "Invalid cover ID", nil)
			return
		}
		coverID, _ := primitive.ObjectIDFromHex(req.CoverID)
		updates["cover_id"] = coverID
	}

	if req.StoryIDs != nil && len(req.StoryIDs) > 0 {
		storyIDs := make([]primitive.ObjectID, 0, len(req.StoryIDs))
		for _, idStr := range req.StoryIDs {
			if !validation.IsValidObjectID(idStr) {
				response.ValidationError(c, "Invalid story ID: "+idStr, nil)
				return
			}
			storyID, _ := primitive.ObjectIDFromHex(idStr)
			storyIDs = append(storyIDs, storyID)
		}
		updates["story_ids"] = storyIDs
	}

	if req.IsPrivate != nil {
		updates["is_private"] = *req.IsPrivate
	}

	if req.CategoryID != "" {
		if !validation.IsValidObjectID(req.CategoryID) {
			response.ValidationError(c, "Invalid category ID", nil)
			return
		}
		categoryID, _ := primitive.ObjectIDFromHex(req.CategoryID)
		updates["category_id"] = categoryID
	}

	if len(updates) == 0 {
		response.ValidationError(c, "No updates provided", nil)
		return
	}

	// Update the highlight
	updatedHighlight, err := h.storyService.UpdateHighlight(c.Request.Context(), highlightID, userID.(primitive.ObjectID), updates)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to update highlight", err)
		return
	}

	// Return success response
	response.OK(c, "Highlight updated successfully", updatedHighlight)
}

// DeleteHighlight handles the request to delete a highlight
func (h *HighlightsHandler) DeleteHighlight(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Get highlight ID from URL parameter
	highlightIDStr := c.Param("id")
	if !validation.IsValidObjectID(highlightIDStr) {
		response.ValidationError(c, "Invalid highlight ID", nil)
		return
	}
	highlightID, _ := primitive.ObjectIDFromHex(highlightIDStr)

	// Delete the highlight
	err := h.storyService.DeleteHighlight(c.Request.Context(), highlightID, userID.(primitive.ObjectID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to delete highlight", err)
		return
	}

	// Return success response
	response.OK(c, "Highlight deleted successfully", nil)
}
