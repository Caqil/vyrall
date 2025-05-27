package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/Caqil/vyrall/internal/utils/response"
	"github.com/Caqil/vyrall/internal/utils/validation"
	"github.com/gin-gonic/gin"
)

// Validation middleware for request validation
func Validation() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip validation for GET, HEAD, OPTIONS requests
		if c.Request.Method == "GET" || c.Request.Method == "HEAD" || c.Request.Method == "OPTIONS" {
			c.Next()
			return
		}

		// Check content type
		contentType := c.GetHeader("Content-Type")
		if !strings.Contains(contentType, "application/json") {
			c.Next()
			return
		}

		// Read request body
		bodyBytes, err := io.ReadAll(c.Request.Body)
		if err != nil {
			response.Error(c, http.StatusBadRequest, "Failed to read request body", err)
			c.Abort()
			return
		}

		// Restore request body
		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		// Parse JSON
		var bodyMap map[string]interface{}
		if err := json.Unmarshal(bodyBytes, &bodyMap); err != nil {
			response.ValidationError(c, "Invalid JSON in request body", err.Error())
			c.Abort()
			return
		}

		// Get the route path
		path := c.FullPath()

		// Validate based on the route
		validationErrors := validateRequest(path, bodyMap)
		if len(validationErrors) > 0 {
			response.ValidationError(c, "Validation failed", validationErrors)
			c.Abort()
			return
		}

		c.Next()
	}
}

// validateRequest validates the request body based on the route
func validateRequest(path string, body map[string]interface{}) map[string]string {
	// Define validation rules for different routes
	switch {
	case strings.Contains(path, "/auth/register"):
		return validateRegistration(body)
	case strings.Contains(path, "/auth/login"):
		return validateLogin(body)
	case strings.Contains(path, "/users/profile"):
		return validateProfileUpdate(body)
	case strings.Contains(path, "/posts"):
		return validatePost(body)
	case strings.Contains(path, "/comments"):
		return validateComment(body)
	default:
		// For other routes, no specific validation
		return nil
	}
}

// validateRegistration validates a registration request
func validateRegistration(body map[string]interface{}) map[string]string {
	errors := make(map[string]string)

	// Validate email
	if email, ok := body["email"].(string); ok {
		if !validation.IsValidEmail(email) {
			errors["email"] = "Invalid email format"
		}
	} else {
		errors["email"] = "Email is required"
	}

	// Validate username
	if username, ok := body["username"].(string); ok {
		if !validation.IsValidUsername(username) {
			errors["username"] = "Username must be 3-30 characters and contain only letters, numbers, dots, and underscores"
		}
	} else {
		errors["username"] = "Username is required"
	}

	// Validate password
	if password, ok := body["password"].(string); ok {
		if !validation.IsValidPassword(password) {
			errors["password"] = "Password must be at least 8 characters long and include a number, a lowercase letter, an uppercase letter, and a special character"
		}
	} else {
		errors["password"] = "Password is required"
	}

	return errors
}

// validateLogin validates a login request
func validateLogin(body map[string]interface{}) map[string]string {
	errors := make(map[string]string)

	// Check if either email or username is provided
	hasEmail := false
	hasUsername := false

	if email, ok := body["email"].(string); ok && email != "" {
		hasEmail = true
	}

	if username, ok := body["username"].(string); ok && username != "" {
		hasUsername = true
	}

	if !hasEmail && !hasUsername {
		errors["auth"] = "Either email or username is required"
	}

	// Validate password
	if password, ok := body["password"].(string); !ok || password == "" {
		errors["password"] = "Password is required"
	}

	return errors
}

// validateProfileUpdate validates a profile update request
func validateProfileUpdate(body map[string]interface{}) map[string]string {
	errors := make(map[string]string)

	// Validate display name
	if displayName, ok := body["display_name"].(string); ok {
		if len(displayName) < 2 || len(displayName) > 50 {
			errors["display_name"] = "Display name must be between 2 and 50 characters"
		}
	}

	// Validate bio
	if bio, ok := body["bio"].(string); ok {
		if len(bio) > 500 {
			errors["bio"] = "Bio cannot exceed 500 characters"
		}
	}

	// Validate website
	if website, ok := body["website"].(string); ok && website != "" {
		if !validation.IsValidURL(website) {
			errors["website"] = "Invalid website URL"
		}
	}

	return errors
}

// validatePost validates a post creation/update request
func validatePost(body map[string]interface{}) map[string]string {
	errors := make(map[string]string)

	// Check if either content or media is provided
	hasContent := false
	hasMedia := false

	if content, ok := body["content"].(string); ok && content != "" {
		hasContent = true

		// Validate content length
		if len(content) > 5000 {
			errors["content"] = "Content cannot exceed 5000 characters"
		}
	}

	if mediaIDs, ok := body["media_ids"].([]interface{}); ok && len(mediaIDs) > 0 {
		hasMedia = true

		// Validate media count
		if len(mediaIDs) > 10 {
			errors["media_ids"] = "Cannot include more than 10 media files"
		}
	}

	if !hasContent && !hasMedia {
		errors["post"] = "Post must contain either text content or media"
	}

	return errors
}

// validateComment validates a comment creation/update request
func validateComment(body map[string]interface{}) map[string]string {
	errors := make(map[string]string)

	// Validate content
	if content, ok := body["content"].(string); ok {
		if content == "" {
			errors["content"] = "Comment content is required"
		} else if len(content) > 1000 {
			errors["content"] = "Comment cannot exceed 1000 characters"
		}
	} else {
		errors["content"] = "Comment content is required"
	}

	return errors
}
