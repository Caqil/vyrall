package media

import (
	"net/http"

	"github.com/Caqil/vyrall/internal/services/media"
	"github.com/Caqil/vyrall/internal/utils/response"
	"github.com/Caqil/vyrall/internal/utils/validation"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// DeleteHandler handles media deletion
type DeleteHandler struct {
	mediaService *media.Service
}

// NewDeleteHandler creates a new delete handler
func NewDeleteHandler(mediaService *media.Service) *DeleteHandler {
	return &DeleteHandler{
		mediaService: mediaService,
	}
}

// Delete handles the media deletion request
func (h *DeleteHandler) Delete(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Get media ID from URL parameter
	mediaIDStr := c.Param("id")
	if !validation.IsValidObjectID(mediaIDStr) {
		response.ValidationError(c, "Invalid media ID", nil)
		return
	}

	mediaID, _ := primitive.ObjectIDFromHex(mediaIDStr)

	// Check if the media belongs to the user
	media, err := h.mediaService.GetByID(c.Request.Context(), mediaID)
	if err != nil {
		response.NotFoundError(c, "Media not found")
		return
	}

	if media.UserID != userID.(primitive.ObjectID) {
		response.ForbiddenError(c, "You don't have permission to delete this media")
		return
	}

	// Delete the media
	err = h.mediaService.Delete(c.Request.Context(), mediaID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to delete media", err)
		return
	}

	// Return success response
	response.OK(c, "Media deleted successfully", nil)
}

// BulkDelete handles multiple media deletion
func (h *DeleteHandler) BulkDelete(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Get media IDs from request body
	var req struct {
		MediaIDs []string `json:"media_ids" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "Invalid request body", err.Error())
		return
	}

	if len(req.MediaIDs) == 0 {
		response.ValidationError(c, "No media IDs provided", nil)
		return
	}

	// Validate and convert media IDs
	mediaIDs := make([]primitive.ObjectID, 0, len(req.MediaIDs))
	for _, idStr := range req.MediaIDs {
		if !validation.IsValidObjectID(idStr) {
			continue // Skip invalid IDs
		}
		mediaID, _ := primitive.ObjectIDFromHex(idStr)
		mediaIDs = append(mediaIDs, mediaID)
	}

	if len(mediaIDs) == 0 {
		response.ValidationError(c, "No valid media IDs provided", nil)
		return
	}

	// Verify ownership of all media files
	mediaList, err := h.mediaService.GetByIDs(c.Request.Context(), mediaIDs)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve media files", err)
		return
	}

	// Filter out media that doesn't belong to the user
	userMediaIDs := make([]primitive.ObjectID, 0, len(mediaList))
	for _, media := range mediaList {
		if media.UserID == userID.(primitive.ObjectID) {
			userMediaIDs = append(userMediaIDs, media.ID)
		}
	}

	if len(userMediaIDs) == 0 {
		response.ForbiddenError(c, "You don't have permission to delete any of these media files")
		return
	}

	// Delete the media files
	deletedCount, err := h.mediaService.BulkDelete(c.Request.Context(), userMediaIDs)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to delete media files", err)
		return
	}

	// Return success response
	response.OK(c, "Media files deleted successfully", gin.H{
		"deleted_count": deletedCount,
	})
}
