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

// CompressHandler handles media compression
type CompressHandler struct {
	mediaService *media.Service
}

// NewCompressHandler creates a new compress handler
func NewCompressHandler(mediaService *media.Service) *CompressHandler {
	return &CompressHandler{
		mediaService: mediaService,
	}
}

// CompressImage handles the request to compress an image
func (h *CompressHandler) CompressImage(c *gin.Context) {
	// Get media ID from URL parameter
	mediaIDStr := c.Param("id")
	if !validation.IsValidObjectID(mediaIDStr) {
		response.ValidationError(c, "Invalid media ID", nil)
		return
	}

	mediaID, _ := primitive.ObjectIDFromHex(mediaIDStr)

	// Get compression options from request
	quality, _ := strconv.Atoi(c.DefaultQuery("quality", "80"))
	format := c.DefaultQuery("format", "")
	progressive := c.DefaultQuery("progressive", "false") == "true"

	// Validate parameters
	if quality < 1 || quality > 100 {
		quality = 80
	}
	if format != "" && format != "jpeg" && format != "png" && format != "webp" {
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

	// Compress the image
	result, err := h.mediaService.CompressImage(c.Request.Context(), mediaID, quality, format, progressive)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to compress image", err)
		return
	}

	// Return success response
	response.OK(c, "Image compressed successfully", gin.H{
		"compressed_url":    result.URL,
		"original_size":     result.OriginalSize,
		"compressed_size":   result.CompressedSize,
		"compression_ratio": result.CompressionRatio,
		"size_reduction":    result.SizeReduction,
		"quality":           quality,
		"format":            result.Format,
	})
}

// CompressVideo handles the request to compress a video
func (h *CompressHandler) CompressVideo(c *gin.Context) {
	// Get media ID from URL parameter
	mediaIDStr := c.Param("id")
	if !validation.IsValidObjectID(mediaIDStr) {
		response.ValidationError(c, "Invalid media ID", nil)
		return
	}

	mediaID, _ := primitive.ObjectIDFromHex(mediaIDStr)

	// Get compression options from request
	quality := c.DefaultQuery("quality", "medium")
	bitrate, _ := strconv.Atoi(c.DefaultQuery("bitrate", "0"))
	format := c.DefaultQuery("format", "")
	audio := c.DefaultQuery("audio", "true") == "true"

	// Validate parameters
	if quality != "low" && quality != "medium" && quality != "high" {
		quality = "medium"
	}
	if format != "" && format != "mp4" && format != "webm" {
		format = "" // Use original format if specified format is invalid
	}

	// Check if the media exists and is a video
	media, err := h.mediaService.GetByID(c.Request.Context(), mediaID)
	if err != nil {
		response.NotFoundError(c, "Media not found")
		return
	}

	if media.Type != "video" {
		response.ValidationError(c, "Media is not a video", nil)
		return
	}

	// Compress the video
	result, err := h.mediaService.CompressVideo(c.Request.Context(), mediaID, quality, bitrate, format, audio)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to compress video", err)
		return
	}

	// Return success response
	response.OK(c, "Video compression started", gin.H{
		"task_id":            result.TaskID,
		"status":             result.Status,
		"estimated_duration": result.EstimatedDuration,
	})
}

// GetCompressionStatus handles the request to get the status of a compression task
func (h *CompressHandler) GetCompressionStatus(c *gin.Context) {
	// Get task ID from URL parameter
	taskID := c.Param("taskId")
	if taskID == "" {
		response.ValidationError(c, "Task ID is required", nil)
		return
	}

	// Get the compression status
	status, err := h.mediaService.GetCompressionStatus(c.Request.Context(), taskID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get compression status", err)
		return
	}

	// Return the status
	response.OK(c, "Compression status retrieved", status)
}
