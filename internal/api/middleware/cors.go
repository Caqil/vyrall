package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
)

// CORS middleware for handling Cross-Origin Resource Sharing
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get origin from request
		origin := c.Request.Header.Get("Origin")

		// Check if the origin is allowed
		if origin != "" && isAllowedOrigin(origin) {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
			c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		}

		// Handle preflight requests
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// isAllowedOrigin checks if the origin is allowed
func isAllowedOrigin(origin string) bool {
	// Define allowed origins
	allowedOrigins := []string{
		"http://localhost:3000",
		"https://yourapp.com",
		"https://*.yourapp.com",
	}

	for _, allowed := range allowedOrigins {
		if allowed == origin {
			return true
		}

		// Handle wildcard domains
		if strings.HasPrefix(allowed, "https://*.") {
			domain := strings.TrimPrefix(allowed, "https://*.")
			if strings.HasSuffix(origin, domain) && strings.HasPrefix(origin, "https://") {
				return true
			}
		}
	}

	return false
}
