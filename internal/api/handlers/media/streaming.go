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

// StreamingHandler handles media streaming
type StreamingHandler struct {
	mediaService *media.Service
}

// NewStreamingHandler creates a new streaming handler
func NewStreamingHandler(mediaService *media.Service) *StreamingHandler {
	return &StreamingHandler{
		mediaService: mediaService,
	}
}

// StreamMedia handles the request to stream a media file
func (h *StreamingHandler) StreamMedia(c *gin.Context) {
	// Get media ID from URL parameter
	mediaIDStr := c.Param("id")
	if !validation.IsValidObjectID(mediaIDStr) {
		response.ValidationError(c, "Invalid media ID", nil)
		return
	}

	mediaID, _ := primitive.ObjectIDFromHex(mediaIDStr)

	// Check if the media exists
	media, err := h.mediaService.GetByID(c.Request.Context(), mediaID)
	if err != nil {
		response.NotFoundError(c, "Media not found")
		return
	}

	// Parse range header for video/audio streaming
	rangeHeader := c.GetHeader("Range")

	// Get the media data with range support
	data, contentType, contentLength, err := h.mediaService.GetMediaData(c.Request.Context(), mediaID, rangeHeader)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get media data", err)
		return
	}

	// Set appropriate headers
	c.Header("Content-Type", contentType)
	c.Header("Accept-Ranges", "bytes")
	c.Header("Content-Length", strconv.FormatInt(contentLength, 10))

	// If range was requested, set partial content status and range headers
	if rangeHeader != "" {
		start, end, _ := h.mediaService.ParseRangeHeader(rangeHeader, contentLength)
		c.Header("Content-Range", "bytes "+strconv.FormatInt(start, 10)+"-"+strconv.FormatInt(end, 10)+"/"+strconv.FormatInt(contentLength, 10))
		c.Status(http.StatusPartialContent)
	} else {
		c.Status(http.StatusOK)
	}

	// Stream the data
	c.Writer.Write(data)
}

// StreamHLS handles the request to stream a video using HLS
func (h *StreamingHandler) StreamHLS(c *gin.Context) {
	// Get media ID from URL parameter
	mediaIDStr := c.Param("id")
	if !validation.IsValidObjectID(mediaIDStr) {
		response.ValidationError(c, "Invalid media ID", nil)
		return
	}

	mediaID, _ := primitive.ObjectIDFromHex(mediaIDStr)

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

	// Get segment parameter
	segment := c.Param("segment")
	if segment != "" {
		// Stream the specific segment
		segmentData, contentType, err := h.mediaService.GetHLSSegment(c.Request.Context(), mediaID, segment)
		if err != nil {
			response.Error(c, http.StatusInternalServerError, "Failed to get segment", err)
			return
		}

		// Set appropriate headers
		c.Header("Content-Type", contentType)
		c.Header("Cache-Control", "public, max-age=31536000") // Cache for 1 year

		// Stream the segment data
		c.Data(http.StatusOK, contentType, segmentData)
		return
	}

	// Stream the manifest file
	manifestData, err := h.mediaService.GetHLSManifest(c.Request.Context(), mediaID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get HLS manifest", err)
		return
	}

	// Set appropriate headers
	c.Header("Content-Type", "application/vnd.apple.mpegurl")
	c.Header("Cache-Control", "private, no-cache, no-store, must-revalidate")

	// Stream the manifest data
	c.Data(http.StatusOK, "application/vnd.apple.mpegurl", manifestData)
}
