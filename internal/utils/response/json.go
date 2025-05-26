package response

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// JSONResponse is the base structure for all JSON responses
type JSONResponse struct {
	Status    int         `json:"status"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data,omitempty"`
	Metadata  interface{} `json:"metadata,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

// JSON sends a custom JSON response with the specified status code
func JSON(c *gin.Context, status int, message string, data interface{}, metadata interface{}) {
	resp := JSONResponse{
		Status:    status,
		Message:   message,
		Data:      data,
		Metadata:  metadata,
		Timestamp: time.Now(),
	}

	c.JSON(status, resp)
}

// JSONWithoutMetadata sends a JSON response without metadata
func JSONWithoutMetadata(c *gin.Context, status int, message string, data interface{}) {
	resp := JSONResponse{
		Status:    status,
		Message:   message,
		Data:      data,
		Timestamp: time.Now(),
	}

	c.JSON(status, resp)
}

// Stream sends data as a streaming response
func Stream(c *gin.Context, data []byte, contentType string) {
	// Set the content type
	c.Header("Content-Type", contentType)
	c.Header("Transfer-Encoding", "chunked")

	// Write the data directly to the response writer
	c.Writer.Write(data)
}

// File sends a file as a response
func File(c *gin.Context, filepath string, filename string) {
	// Set appropriate headers for file download
	if filename != "" {
		c.Header("Content-Disposition", "attachment; filename="+filename)
	}

	// Serve the file
	c.File(filepath)
}

// Raw sends raw data as a response with the specified content type
func Raw(c *gin.Context, status int, contentType string, data []byte) {
	c.Data(status, contentType, data)
}

// XML sends an XML response
func XML(c *gin.Context, status int, data interface{}) {
	c.XML(status, data)
}

// HTML sends an HTML response
func HTML(c *gin.Context, status int, name string, data interface{}) {
	c.HTML(status, name, data)
}

// Redirect sends a redirect response
func Redirect(c *gin.Context, status int, location string) {
	c.Redirect(status, location)
}

// NoContent sends a response with no content (204)
func NoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}
