package response

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// SuccessResponse represents a successful API response
type SuccessResponse struct {
	Status    int         `json:"status"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data,omitempty"`
	Metadata  interface{} `json:"metadata,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

// Success sends a standardized success response
func Success(c *gin.Context, status int, message string, data interface{}) {
	// Create success response
	resp := SuccessResponse{
		Status:    status,
		Message:   message,
		Data:      data,
		Timestamp: time.Now(),
	}

	c.JSON(status, resp)
}

// SuccessWithMetadata sends a success response with metadata
func SuccessWithMetadata(c *gin.Context, status int, message string, data interface{}, metadata interface{}) {
	// Create success response with metadata
	resp := SuccessResponse{
		Status:    status,
		Message:   message,
		Data:      data,
		Metadata:  metadata,
		Timestamp: time.Now(),
	}

	c.JSON(status, resp)
}

// SuccessWithPagination sends a success response with pagination
func SuccessWithPagination(c *gin.Context, status int, message string, data interface{}, limit, offset, total int) {
	// Create pagination info
	pagination := NewPaginationInfo(total, limit, offset)

	// Create links (optional)
	links := BuildPaginationLinks(c, pagination)

	// Create metadata
	metadata := map[string]interface{}{
		"pagination": pagination,
		"links":      links,
	}

	// Create success response with pagination
	resp := SuccessResponse{
		Status:    status,
		Message:   message,
		Data:      data,
		Metadata:  metadata,
		Timestamp: time.Now(),
	}

	c.JSON(status, resp)
}

// Created sends a success response for resource creation
func Created(c *gin.Context, message string, data interface{}) {
	Success(c, http.StatusCreated, message, data)
}

// OK sends a standard 200 OK response
func OK(c *gin.Context, message string, data interface{}) {
	Success(c, http.StatusOK, message, data)
}

// Accepted sends a 202 Accepted response
func Accepted(c *gin.Context, message string) {
	Success(c, http.StatusAccepted, message, nil)
}

// NoContent sends a 204 No Content response
func NoContentSuccess(c *gin.Context, message string) {
	// Create success response without data
	resp := SuccessResponse{
		Status:    http.StatusNoContent,
		Message:   message,
		Timestamp: time.Now(),
	}

	c.JSON(http.StatusNoContent, resp)
}
