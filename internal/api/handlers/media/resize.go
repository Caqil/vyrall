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

// ResizeHandler handles image resizing
type ResizeHandler struct {
	mediaService *media.Service
}

// NewResizeHandler creates a new resize handler
func NewResizeHandler(mediaService *media.Service) *ResizeHandler {
	return &ResizeHandler{
		mediaService: mediaService,
	}
}

// ResizeImage handles the request to resize an image
func (h *ResizeHandler) ResizeImage(c *gin.Context) {
	// Get media ID from URL parameter
	mediaIDStr := c.Param("id")
	if !validation.IsValidObjectID(mediaIDStr) {
		response.ValidationError(c, "Invalid media ID", nil)
		return
	}

	mediaID, _ := primitive.ObjectIDFromHex(mediaIDStr)

	// Get resize options from request
	width, _ := strconv.Atoi(c.DefaultQuery("width", "0"))
	height, _ := strconv.Atoi(c.DefaultQuery("height", "0"))
	preserveAspectRatio := c.DefaultQuery("preserve_aspect_ratio", "true") == "true"
	format := c.DefaultQuery("format", "")
	quality, _ := strconv.Atoi(c.DefaultQuery("quality", "90"))

	// Validate parameters
	if width < 0 || width > 5000 {
		width = 0 // Auto
	}
	if height < 0 || height > 5000 {
		height = 0 // Auto
	}
	if width == 0 && height == 0 {
		response.ValidationError(c, "At least one dimension (width or height) must be specified", nil)
		return
	}
	if quality <= 0 || quality > 100 {
		quality = 90
	}
	if format != "" && format != "jpeg" && format != "png" && format != "webp" && format != "gif" {
		format = "" // Use original format if specified format is invalid
	}

	// Check if the media exists and is an image
	media, err := h.mediaService.GetByID(c.Request.Context(), mediaID)
	if err != nil {
		response.NotFoundError(c, "Media not found")
		return
	}

	if media.Type != "image" {
		response.ValidationError(c, "Media is not an image", nil)
		return
	}

	// Resize the image
	resizedURL, err := h.mediaService.ResizeImage(c.Request.Context(), mediaID, width, height, preserveAspectRatio, format, quality)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to resize image", err)
		return
	}

	// Return success response
	response.OK(c, "Image resized successfully", gin.H{
		"resized_url": resizedURL,
		"width":       width,
		"height":      height,
		"format":      format,
	})
}

// StreamResizedImage streams a resized image directly to the client
func (h *ResizeHandler) StreamResizedImage(c *gin.Context) {
	// Get media ID from URL parameter
	mediaIDStr := c.Param("id")
	if !validation.IsValidObjectID(mediaIDStr) {
		response.ValidationError(c, "Invalid media ID", nil)
		return
	}

	mediaID, _ := primitive.ObjectIDFromHex(mediaIDStr)

	// Get resize options from request
	width, _ := strconv.Atoi(c.DefaultQuery("width", "0"))
	height, _ := strconv.Atoi(c.DefaultQuery("height", "0"))
	preserveAspectRatio := c.DefaultQuery("preserve_aspect_ratio", "true") == "true"
	format := c.DefaultQuery("format", "")
	quality, _ := strconv.Atoi(c.DefaultQuery("quality", "90"))

	// Validate parameters
	if width < 0 || width > 5000 {
		width = 0 // Auto
	}
	if height < 0 || height > 5000 {
		height = 0 // Auto
	}
	if width == 0 && height == 0 {
		response.ValidationError(c, "At least one dimension (width or height) must be specified", nil)
		return
	}
	if quality <= 0 || quality > 100 {
		quality = 90
	}
	if format != "" && format != "jpeg" && format != "png" && format != "webp" && format != "gif" {
		format = "" // Use original format if specified format is invalid
	}

	// Get the resized image data
	imageData, contentType, err := h.mediaService.GetResizedImageData(c.Request.Context(), mediaID, width, height, preserveAspectRatio, format, quality)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get resized image", err)
		return
	}

	// Set content type and other headers
	c.Header("Content-Type", contentType)
	c.Header("Cache-Control", "public, max-age=31536000") // Cache for 1 year

	// Stream the image data
	c.Data(http.StatusOK, contentType, imageData)
}
