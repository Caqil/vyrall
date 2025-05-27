package middleware

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/Caqil/vyrall/internal/utils/response"
	"github.com/gin-gonic/gin"
)

// Timeout middleware for setting a timeout for request processing
func Timeout(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Create a context with timeout
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()

		// Update the request with the new context
		c.Request = c.Request.WithContext(ctx)

		// Create a channel to signal when the request is completed
		done := make(chan struct{}, 1)

		// Execute the request in a goroutine
		go func() {
			c.Next()
			done <- struct{}{}
		}()

		// Wait for either the request to complete or the timeout to expire
		select {
		case <-done:
			// Request completed before timeout
			return
		case <-ctx.Done():
			// Timeout occurred
			if errors.Is(ctx.Err(), context.DeadlineExceeded) {
				// Check if the response has already been written
				if !c.Writer.Written() {
					response.Error(c, http.StatusGatewayTimeout, "Request timed out", nil)
				}
				c.Abort()
			}
		}
	}
}

// TimeoutForRoute middleware for setting different timeouts for different routes
func TimeoutForRoute(timeouts map[string]time.Duration, defaultTimeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the route path
		path := c.FullPath()

		// Get the timeout for this route, or use the default timeout
		timeout, exists := timeouts[path]
		if !exists {
			timeout = defaultTimeout
		}

		// Use the Timeout middleware with the selected timeout
		Timeout(timeout)(c)
	}
}

// SlowRequestNotifier middleware for notifying about slow requests
func SlowRequestNotifier(threshold time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()

		// Process the request
		c.Next()

		// Calculate request duration
		duration := time.Since(startTime)

		// Check if the request exceeded the threshold
		if duration > threshold {
			// Log the slow request
			// In a real implementation, this might send a notification to a monitoring system
			// or log to a special slow request log
			// For now, we'll just add a header to the response
			c.Header("X-Slow-Request", "true")
			c.Header("X-Request-Duration", duration.String())
		}
	}
}
