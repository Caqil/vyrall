package middleware

import (
	"net/http"
	"time"

	"github.com/Caqil/vyrall/internal/services/ratelimit"
	"github.com/Caqil/vyrall/internal/utils/response"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// RateLimit middleware for rate limiting requests
func RateLimit(rateLimitService *ratelimit.Service, limit int, duration time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get client IP
		clientIP := c.ClientIP()

		// Check if the client has exceeded the rate limit
		key := "ratelimit:" + clientIP
		allowed, remaining, resetTime, err := rateLimitService.Allow(c.Request.Context(), key, limit, duration)
		if err != nil {
			// If there's an error checking the rate limit, allow the request but log the error
			c.Next()
			return
		}

		// Set rate limit headers
		c.Header("X-RateLimit-Limit", itoa(limit))
		c.Header("X-RateLimit-Remaining", itoa(remaining))
		c.Header("X-RateLimit-Reset", itoa(int(resetTime.Unix())))

		if !allowed {
			// If the client has exceeded the rate limit, return a 429 Too Many Requests response
			c.Header("Retry-After", itoa(int(time.Until(resetTime).Seconds())))
			response.Error(c, http.StatusTooManyRequests, "Rate limit exceeded", nil)
			c.Abort()
			return
		}

		c.Next()
	}
}

// UserRateLimit middleware for rate limiting requests per user
func UserRateLimit(rateLimitService *ratelimit.Service, limit int, duration time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user ID from context (set by Auth middleware)
		userIDValue, exists := c.Get("userID")
		if !exists {
			// If the user is not authenticated, use the IP-based rate limit
			RateLimit(rateLimitService, limit, duration)(c)
			return
		}

		userID := userIDValue.(primitive.ObjectID)

		// Check if the user has exceeded the rate limit
		key := "ratelimit:user:" + userID.Hex()
		allowed, remaining, resetTime, err := rateLimitService.Allow(c.Request.Context(), key, limit, duration)
		if err != nil {
			// If there's an error checking the rate limit, allow the request but log the error
			c.Next()
			return
		}

		// Set rate limit headers
		c.Header("X-RateLimit-Limit", itoa(limit))
		c.Header("X-RateLimit-Remaining", itoa(remaining))
		c.Header("X-RateLimit-Reset", itoa(int(resetTime.Unix())))

		if !allowed {
			// If the user has exceeded the rate limit, return a 429 Too Many Requests response
			c.Header("Retry-After", itoa(int(time.Until(resetTime).Seconds())))
			response.Error(c, http.StatusTooManyRequests, "Rate limit exceeded", nil)
			c.Abort()
			return
		}

		c.Next()
	}
}

// EndpointRateLimit middleware for rate limiting specific endpoints
func EndpointRateLimit(rateLimitService *ratelimit.Service, limit int, duration time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get client IP
		clientIP := c.ClientIP()

		// Get request path
		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}

		// Check if the client has exceeded the rate limit for this endpoint
		key := "ratelimit:" + clientIP + ":" + path
		allowed, remaining, resetTime, err := rateLimitService.Allow(c.Request.Context(), key, limit, duration)
		if err != nil {
			// If there's an error checking the rate limit, allow the request but log the error
			c.Next()
			return
		}

		// Set rate limit headers
		c.Header("X-RateLimit-Limit", itoa(limit))
		c.Header("X-RateLimit-Remaining", itoa(remaining))
		c.Header("X-RateLimit-Reset", itoa(int(resetTime.Unix())))

		if !allowed {
			// If the client has exceeded the rate limit, return a 429 Too Many Requests response
			c.Header("Retry-After", itoa(int(time.Until(resetTime).Seconds())))
			response.Error(c, http.StatusTooManyRequests, "Rate limit exceeded for this endpoint", nil)
			c.Abort()
			return
		}

		c.Next()
	}
}

// itoa converts an integer to a string
func itoa(i int) string {
	return primitive.Itoa(i)
}
