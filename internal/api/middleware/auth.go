package middleware

import (
	"strings"

	"github.com/Caqil/vyrall/internal/services/auth"
	"github.com/Caqil/vyrall/internal/utils/response"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Auth is the authentication middleware
func Auth(authService *auth.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.UnauthorizedError(c, "Authorization header is required")
			c.Abort()
			return
		}

		// Check if the header has the Bearer prefix
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			response.UnauthorizedError(c, "Authorization header must be in the format: Bearer {token}")
			c.Abort()
			return
		}

		token := parts[1]
		if token == "" {
			response.UnauthorizedError(c, "Token is required")
			c.Abort()
			return
		}

		// Validate token and get user ID
		userID, err := authService.ValidateToken(c.Request.Context(), token)
		if err != nil {
			response.UnauthorizedError(c, "Invalid or expired token")
			c.Abort()
			return
		}

		// Set user ID in context
		c.Set("userID", userID)
		c.Next()
	}
}

// OptionalAuth is a middleware that tries to authenticate the user but continues regardless
func OptionalAuth(authService *auth.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		// Check if the header has the Bearer prefix
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.Next()
			return
		}

		token := parts[1]
		if token == "" {
			c.Next()
			return
		}

		// Validate token and get user ID
		userID, err := authService.ValidateToken(c.Request.Context(), token)
		if err != nil {
			c.Next()
			return
		}

		// Set user ID in context
		c.Set("userID", userID)
		c.Next()
	}
}

// AdminOnly ensures that the authenticated user has admin privileges
func AdminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user ID from context (set by Auth middleware)
		userIDValue, exists := c.Get("userID")
		if !exists {
			response.UnauthorizedError(c, "User not authenticated")
			c.Abort()
			return
		}

		userID := userIDValue.(primitive.ObjectID)

		// Check if user is an admin (usually this would be done by checking a role in the database)
		isAdmin, err := checkUserIsAdmin(c, userID)
		if err != nil || !isAdmin {
			response.ForbiddenError(c, "Admin privileges required")
			c.Abort()
			return
		}

		// Set admin flag in context
		c.Set("isAdmin", true)
		c.Next()
	}
}

// checkUserIsAdmin checks if the user has admin privileges
func checkUserIsAdmin(c *gin.Context, userID primitive.ObjectID) (bool, error) {
	// In a real application, this would check the user's role in the database
	// For example, you might have a user service that can check if a user has a specific role
	// This is just a placeholder implementation

	// userService := c.MustGet("userService").(*user.Service)
	// return userService.IsAdmin(c.Request.Context(), userID)

	// For demonstration purposes, I'm returning a placeholder
	return true, nil
}
