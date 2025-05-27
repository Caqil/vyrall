package middleware

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Caqil/vyrall/internal/services/cache"
	"github.com/gin-gonic/gin"
)

// Cache is the response caching middleware
func Cache(cacheService *cache.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip caching for non-GET requests
		if c.Request.Method != http.MethodGet {
			c.Next()
			return
		}

		// Skip caching for requests with query parameters that affect content (e.g., pagination)
		skipParams := []string{"page", "limit", "offset", "sort"}
		for _, param := range skipParams {
			if c.Query(param) != "" {
				c.Next()
				return
			}
		}

		// Generate cache key
		key := generateCacheKey(c)

		// Try to get from cache
		cachedResponse, err := cacheService.Get(c.Request.Context(), key)
		if err == nil && cachedResponse != nil {
			// Return cached response
			contentType := cachedResponse["content_type"].(string)
			body := cachedResponse["body"].([]byte)
			statusCode := int(cachedResponse["status_code"].(float64))

			c.Header("Content-Type", contentType)
			c.Header("X-Cache", "HIT")
			c.Data(statusCode, contentType, body)
			c.Abort()
			return
		}

		// Create a response writer proxy to capture the response
		responseWriter := &responseWriterProxy{
			ResponseWriter: c.Writer,
			body:           &bytes.Buffer{},
		}
		c.Writer = responseWriter

		// Process the request
		c.Next()

		// Cache the response if it's successful
		if c.Writer.Status() >= 200 && c.Writer.Status() < 400 {
			// Store response in cache
			cacheData := map[string]interface{}{
				"body":         responseWriter.body.Bytes(),
				"status_code":  c.Writer.Status(),
				"content_type": c.Writer.Header().Get("Content-Type"),
			}

			// Set cache TTL based on content type
			var ttl time.Duration
			contentType := c.Writer.Header().Get("Content-Type")
			if strings.Contains(contentType, "application/json") {
				ttl = 5 * time.Minute
			} else if strings.Contains(contentType, "text/html") {
				ttl = 1 * time.Hour
			} else if strings.Contains(contentType, "image/") {
				ttl = 24 * time.Hour
			} else {
				ttl = 10 * time.Minute
			}

			// Store in cache
			cacheService.Set(c.Request.Context(), key, cacheData, ttl)
		}
	}
}

// responseWriterProxy is a proxy for gin.ResponseWriter that captures the response body
type responseWriterProxy struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

// Write captures the response body
func (w *responseWriterProxy) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// generateCacheKey creates a unique key for caching based on the request
func generateCacheKey(c *gin.Context) string {
	// Base key on URL path
	key := c.Request.URL.Path

	// Add query parameters
	if c.Request.URL.RawQuery != "" {
		key += "?" + c.Request.URL.RawQuery
	}

	// Add user ID if available (for personalized content)
	if userID, exists := c.Get("userID"); exists {
		key += fmt.Sprintf("|user:%s", userID)
	}

	// Add Accept header for content negotiation
	key += "|accept:" + c.GetHeader("Accept")

	// Hash the key to ensure it's a valid cache key
	hasher := sha256.New()
	hasher.Write([]byte(key))
	return "cache:" + hex.EncodeToString(hasher.Sum(nil))
}

// CacheClear creates a middleware that clears the cache for specific patterns
func CacheClear(cacheService *cache.Service, patterns []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Process the request first
		c.Next()

		// Only clear cache for successful operations
		if c.Writer.Status() >= 200 && c.Writer.Status() < 400 {
			for _, pattern := range patterns {
				cacheService.DeleteByPattern(c.Request.Context(), pattern)
			}
		}
	}
}
