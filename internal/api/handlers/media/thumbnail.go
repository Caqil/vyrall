package media

import (
	"net/http"
	"strconv"

	"github.com/Caqil/vyrall/internal/services/media"
	"github.com/Caqil/vyrall/internal/utils/response"
	"github.com/Caqil/vyrall/internal/utils/validation"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ThumbnailHandler handles thumbnail generation and retrieval
type ThumbnailHandler struct {
	mediaService *media.Service
}

// NewThumbnailHandler creates a new thumbnail handler
func NewThumbnailHandler(mediaService *media.Service) *ThumbnailHandler {
	return &ThumbnailHandler{
		mediaService: mediaService,
	}
}

// GenerateThumbnail handles the request to generate a thumbnail
func (h *ThumbnailHandler) GenerateThumbnail(c *gin.Context) {
	// Get media ID from URL parameter
	mediaIDStr := c.Param("id")
	if !validation.IsValidObjectID(mediaIDStr) {
		response.ValidationError(c, "Invalid media ID", nil)
		return
	}

	mediaID, _ := primitive.ObjectIDFromHex(mediaIDStr)

	// Get thumbnail options from request
	width, _ := strconv.Atoi(c.DefaultQuery("width", "300"))
	height, _ := strconv.Atoi(c.DefaultQuery("height", "300"))
	format := c.DefaultQuery("format", "jpeg")
	quality, _ := strconv.Atoi(c.DefaultQuery("quality", "80"))

	// Validate parameters
	if width <= 0 || width > 1200 {
		width = 300
	}
	if height <= 0 || height > 1200 {
		height = 300
	}
	if quality <= 0 || quality > 100 {
		quality = 80
	}
	if format != "jpeg" && format != "png" && format != "webp" {
		format = "jpeg"
	}

	// Generate the thumbnail
	thumbnailURL, err := h.mediaService.GenerateThumbnail(c.Request.Context(), mediaID, width, height, format, quality)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to generate thumbnail", err)
		return
	}

	// Return success response
	response.OK(c, "Thumbnail generated successfully", gin.H{
		"thumbnail_url": thumbnailURL,
	})
}

// StreamThumbnail streams the thumbnail directly to the client
func (h *ThumbnailHandler) StreamThumbnail(c *gin.Context) {
	// Get media ID from URL parameter
	mediaIDStr := c.Param("id")
	if !validation.IsValidObjectID(mediaIDStr) {
		response.ValidationError(c, "Invalid media ID", nil)
		return
	}

	mediaID, _ := primitive.ObjectIDFromHex(mediaIDStr)

	// Get thumbnail options from request
	width, _ := strconv.Atoi(c.DefaultQuery("width", "300"))
	height, _ := strconv.Atoi(c.DefaultQuery("height", "300"))
	format := c.DefaultQuery("format", "jpeg")
	quality, _ := strconv.Atoi(c.DefaultQuery("quality", "80"))

	// Validate parameters
	if width <= 0 || width > 1200 {
		width = 300
	}
	if height <= 0 || height > 1200 {
		height = 300
	}
	if quality <= 0 || quality > 100 {
		quality = 80
	}
	if format != "jpeg" && format != "png" && format != "webp" {
		format = "jpeg"
	}

	// Get the thumbnail data
	thumbnailData, contentType, err := h.mediaService.GetThumbnailData(c.Request.Context(), mediaID, width, height, format, quality)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get thumbnail", err)
		return
	}

	// Set content type and other headers
	c.Header("Content-Type", contentType)
	c.Header("Cache-Control", "public, max-age=31536000") // Cache for 1 year

	// Stream the thumbnail data
	c.Data(http.StatusOK, contentType, thumbnailData)
}
