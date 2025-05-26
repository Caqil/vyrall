package media

import (
	"net/http"

	"github.com/Caqil/vyrall/internal/services/media"
	"github.com/Caqil/vyrall/internal/utils/response"
	"github.com/Caqil/vyrall/internal/utils/validation"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ProcessHandler handles media processing
type ProcessHandler struct {
	mediaService *media.Service
}

// NewProcessHandler creates a new process handler
func NewProcessHandler(mediaService *media.Service) *ProcessHandler {
	return &ProcessHandler{
		mediaService: mediaService,
	}
}

// Process handles the request to process a media file
func (h *ProcessHandler) Process(c *gin.Context) {
	// Get media ID from URL parameter
	mediaIDStr := c.Param("id")
	if !validation.IsValidObjectID(mediaIDStr) {
		response.ValidationError(c, "Invalid media ID", nil)
		return
	}

	mediaID, _ := primitive.ObjectIDFromHex(mediaIDStr)

	// Get processing options from request body
	var req struct {
		GenerateThumbnail bool `json:"generate_thumbnail"`
		Resize            bool `json:"resize"`
		Compress          bool `json:"compress"`
		OptimizeForWeb    bool `json:"optimize_for_web"`
		ExtractMetadata   bool `json:"extract_metadata"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		// If no body provided, use default options
		req.GenerateThumbnail = true
		req.Resize = true
		req.Compress = true
		req.OptimizeForWeb = true
		req.ExtractMetadata = true
	}

	// Check if the media exists
	media, err := h.mediaService.GetByID(c.Request.Context(), mediaID)
	if err != nil {
		response.NotFoundError(c, "Media not found")
		return
	}

	// Check if the media is already processed
	if media.IsProcessed {
		response.OK(c, "Media is already processed", media)
		return
	}

	// Process the media
	processedMedia, err := h.mediaService.Process(c.Request.Context(), mediaID, req.GenerateThumbnail, req.Resize, req.Compress, req.OptimizeForWeb, req.ExtractMetadata)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to process media", err)
		return
	}

	// Return success response
	response.OK(c, "Media processed successfully", processedMedia)
}

// UpdateMetadata handles updating media metadata
func (h *ProcessHandler) UpdateMetadata(c *gin.Context) {
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

	// Get media to check ownership
	media, err := h.mediaService.GetByID(c.Request.Context(), mediaID)
	if err != nil {
		response.NotFoundError(c, "Media not found")
		return
	}

	if media.UserID != userID.(primitive.ObjectID) {
		response.ForbiddenError(c, "You don't have permission to update this media")
		return
	}

	// Get metadata from request body
	var req struct {
		Caption  string                 `json:"caption"`
		AltText  string                 `json:"alt_text"`
		Metadata map[string]interface{} `json:"metadata"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "Invalid request body", err.Error())
		return
	}

	// Update the media
	updates := make(map[string]interface{})
	if req.Caption != "" {
		updates["caption"] = req.Caption
	}
	if req.AltText != "" {
		updates["alt_text"] = req.AltText
	}
	if req.Metadata != nil {
		updates["metadata"] = req.Metadata
	}

	if len(updates) == 0 {
		response.ValidationError(c, "No updates provided", nil)
		return
	}

	// Update the media
	updatedMedia, err := h.mediaService.UpdateMetadata(c.Request.Context(), mediaID, updates)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to update media metadata", err)
		return
	}

	// Return success response
	response.OK(c, "Media metadata updated successfully", updatedMedia)
}
