package middleware

import (
	"github.com/gin-gonic/gin"
)

// Security middleware for setting security headers
func Security() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Set security headers

		// Content Security Policy (CSP)
		c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self' https://cdnjs.cloudflare.com; style-src 'self' https://fonts.googleapis.com; font-src 'self' https://fonts.gstatic.com; img-src 'self' data: https:; connect-src 'self'")

		// X-Content-Type-Options
		c.Header("X-Content-Type-Options", "nosniff")

		// X-Frame-Options
		c.Header("X-Frame-Options", "SAMEORIGIN")

		// X-XSS-Protection
		c.Header("X-XSS-Protection", "1; mode=block")

		// Referrer-Policy
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		// Permissions-Policy
		c.Header("Permissions-Policy", "camera=(), microphone=(), geolocation=(), interest-cohort=()")

		// Strict-Transport-Security
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")

		c.Next()
	}
}

// CSRF middleware for CSRF protection
func CSRF() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip CSRF check for GET, HEAD, OPTIONS requests
		if c.Request.Method == "GET" || c.Request.Method == "HEAD" || c.Request.Method == "OPTIONS" {
			c.Next()
			return
		}

		// Check for CSRF token in header
		csrfToken := c.GetHeader("X-CSRF-Token")
		if csrfToken == "" {
			// Check for CSRF token in form
			csrfToken = c.PostForm("_csrf")
		}

		// Validate CSRF token (implementation depends on how tokens are generated and stored)
		if !validateCSRFToken(c, csrfToken) {
			c.JSON(403, gin.H{"error": "CSRF token validation failed"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// validateCSRFToken validates the CSRF token
func validateCSRFToken(c *gin.Context, token string) bool {
	// In a real implementation, this would check the token against a stored value
	// For example, comparing with a token stored in the user's session

	// This is a placeholder implementation
	return token != ""
}

// XSS middleware for XSS protection
func XSS() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Set XSS protection header
		c.Header("X-XSS-Protection", "1; mode=block")

		c.Next()
	}
}
