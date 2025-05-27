package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/Caqil/vyrall/internal/services/content"
	"github.com/Caqil/vyrall/internal/utils/response"
	"github.com/gin-gonic/gin"
)

// ContentFilter middleware for filtering inappropriate content
func ContentFilter(contentService *content.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Only filter POST, PUT, and PATCH requests
		if c.Request.Method != http.MethodPost &&
			c.Request.Method != http.MethodPut &&
			c.Request.Method != http.MethodPatch {
			c.Next()
			return
		}

		// Check content type
		contentType := c.GetHeader("Content-Type")
		if !strings.Contains(contentType, "application/json") {
			c.Next()
			return
		}

		// Read the request body
		bodyBytes, err := io.ReadAll(c.Request.Body)
		if err != nil {
			response.Error(c, http.StatusBadRequest, "Failed to read request body", err)
			c.Abort()
			return
		}

		// Restore the request body for later handlers
		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		// Parse the JSON body
		var bodyMap map[string]interface{}
		if err := json.Unmarshal(bodyBytes, &bodyMap); err != nil {
			// If it's not a valid JSON, skip content filtering
			c.Next()
			return
		}

		// Fields to check for inappropriate content
		fieldsToCheck := []string{
			"content", "caption", "text", "message", "title",
			"description", "bio", "comment", "reply",
		}

		// Extract content to check
		var textsToCheck []string
		for _, field := range fieldsToCheck {
			if value, ok := bodyMap[field]; ok && value != nil {
				if text, ok := value.(string); ok && text != "" {
					textsToCheck = append(textsToCheck, text)
				}
			}
		}

		if len(textsToCheck) == 0 {
			c.Next()
			return
		}

		// Check content for inappropriate material
		for _, text := range textsToCheck {
			isInappropriate, reason, err := contentService.CheckInappropriateContent(c.Request.Context(), text)
			if err != nil {
				// If there's an error in checking, log it and continue
				c.Next()
				return
			}

			if isInappropriate {
				response.ValidationError(c, "Content contains inappropriate material: "+reason, nil)
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

// ProfanityFilter middleware for filtering profanity in content
func ProfanityFilter() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Only filter POST, PUT, and PATCH requests
		if c.Request.Method != http.MethodPost &&
			c.Request.Method != http.MethodPut &&
			c.Request.Method != http.MethodPatch {
			c.Next()
			return
		}

		// Check content type
		contentType := c.GetHeader("Content-Type")
		if !strings.Contains(contentType, "application/json") {
			c.Next()
			return
		}

		// Read the request body
		bodyBytes, err := io.ReadAll(c.Request.Body)
		if err != nil {
			response.Error(c, http.StatusBadRequest, "Failed to read request body", err)
			c.Abort()
			return
		}

		// Restore the request body for later handlers
		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		// Parse the JSON body
		var bodyMap map[string]interface{}
		if err := json.Unmarshal(bodyBytes, &bodyMap); err != nil {
			// If it's not a valid JSON, skip profanity filtering
			c.Next()
			return
		}

		// Filter profanity in fields
		bodyMap = filterProfanityInMap(bodyMap)

		// Replace the request body with the filtered body
		filteredBody, err := json.Marshal(bodyMap)
		if err != nil {
			c.Next()
			return
		}

		c.Request.Body = io.NopCloser(bytes.NewBuffer(filteredBody))
		c.Request.ContentLength = int64(len(filteredBody))

		c.Next()
	}
}

// filterProfanityInMap recursively filters profanity in map values
func filterProfanityInMap(m map[string]interface{}) map[string]interface{} {
	for key, value := range m {
		switch v := value.(type) {
		case string:
			m[key] = filterProfanity(v)
		case map[string]interface{}:
			m[key] = filterProfanityInMap(v)
		case []interface{}:
			m[key] = filterProfanityInArray(v)
		}
	}
	return m
}

// filterProfanityInArray recursively filters profanity in array values
func filterProfanityInArray(arr []interface{}) []interface{} {
	for i, value := range arr {
		switch v := value.(type) {
		case string:
			arr[i] = filterProfanity(v)
		case map[string]interface{}:
			arr[i] = filterProfanityInMap(v)
		case []interface{}:
			arr[i] = filterProfanityInArray(v)
		}
	}
	return arr
}

// filterProfanity replaces profanity with asterisks
func filterProfanity(text string) string {
	// This is a very basic implementation
	// In a real application, you would use a proper profanity filter library
	profanityList := []string{
		"badword1", "badword2", "badword3",
	}

	for _, word := range profanityList {
		replacement := strings.Repeat("*", len(word))
		text = strings.ReplaceAll(text, word, replacement)
	}

	return text
}
