package media

import (
	"net/http"
	"time"

	"github.com/Caqil/vyrall/internal/services/media"
	"github.com/Caqil/vyrall/internal/utils/response"
	"github.com/Caqil/vyrall/internal/utils/validation"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// CDNHandler handles CDN-related operations for media
type CDNHandler struct {
	mediaService *media.Service
}

// NewCDNHandler creates a new CDN handler
func NewCDNHandler(mediaService *media.Service) *CDNHandler {
	return &CDNHandler{
		mediaService: mediaService,
	}
}

// GetCDNURL returns a CDN URL for a media file
func (h *CDNHandler) GetCDNURL(c *gin.Context) {
	// Get media ID from URL parameter
	mediaIDStr := c.Param("id")
	if !validation.IsValidObjectID(mediaIDStr) {
		response.ValidationError(c, "Invalid media ID", nil)
		return
	}

	mediaID, _ := primitive.ObjectIDFromHex(mediaIDStr)

	// Get expiration parameter (in seconds)
	expirationStr := c.DefaultQuery("expiration", "3600") // Default: 1 hour
	expiration := 3600
	if validation.MatchesPattern(expirationStr, `^\d+$`) {
		// If it's a valid number, convert it
		if val, err := primitive.ParseInt32(expirationStr); err == nil {
			expiration = int(val)
		}
	}

	// Limit expiration to reasonable values
	if expiration < 60 {
		expiration = 60 // Minimum 1 minute
	} else if expiration > 86400*7 {
		expiration = 86400 * 7 // Maximum 7 days
	}

	// Get CDN URL for the media
	cdnURL, expiresAt, err := h.mediaService.GetCDNURL(c.Request.Context(), mediaID, time.Duration(expiration)*time.Second)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to generate CDN URL", err)
		return
	}

	// Return success response
	response.OK(c, "CDN URL generated successfully", gin.H{
		"cdn_url":     cdnURL,
		"expires_at":  expiresAt,
		"valid_until": expiresAt.Format(time.RFC3339),
	})
}

// GetSignedURL returns a signed URL for a media file
func (h *CDNHandler) GetSignedURL(c *gin.Context) {
	// Get media ID from URL parameter
	mediaIDStr := c.Param("id")
	if !validation.IsValidObjectID(mediaIDStr) {
		response.ValidationError(c, "Invalid media ID", nil)
		return
	}

	mediaID, _ := primitive.ObjectIDFromHex(mediaIDStr)

	// Get expiration parameter (in seconds)
	expirationStr := c.DefaultQuery("expiration", "3600") // Default: 1 hour
	expiration := 3600
	if validation.MatchesPattern(expirationStr, `^\d+$`) {
		// If it's a valid number, convert it
		if val, err := primitive.ParseInt32(expirationStr); err == nil {
			expiration = int(val)
		}
	}

	// Limit expiration to reasonable values
	if expiration < 60 {
		expiration = 60 // Minimum 1 minute
	} else if expiration > 86400*7 {
		expiration = 86400 * 7 // Maximum 7 days
	}

	// Get signed URL for the media
	signedURL, expiresAt, err := h.mediaService.GetSignedURL(c.Request.Context(), mediaID, time.Duration(expiration)*time.Second)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to generate signed URL", err)
		return
	}

	// Return success response
	response.OK(c, "Signed URL generated successfully", gin.H{
		"signed_url":  signedURL,
		"expires_at":  expiresAt,
		"valid_until": expiresAt.Format(time.RFC3339),
	})
}

// PurgeCache purges a media file from the CDN cache
func (h *CDNHandler) PurgeCache(c *gin.Context) {
	// Check admin permissions (this might be an admin-only endpoint)
	isAdmin, exists := c.Get("isAdmin")
	if !exists || !isAdmin.(bool) {
		response.ForbiddenError(c, "Admin access required")
		return
	}

	// Get media ID from URL parameter
	mediaIDStr := c.Param("id")
	if !validation.IsValidObjectID(mediaIDStr) {
		response.ValidationError(c, "Invalid media ID", nil)
		return
	}

	mediaID, _ := primitive.ObjectIDFromHex(mediaIDStr)

	// Purge the media from CDN cache
	err := h.mediaService.PurgeCDNCache(c.Request.Context(), mediaID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to purge CDN cache", err)
		return
	}

	// Return success response
	response.OK(c, "CDN cache purged successfully", nil)
}
