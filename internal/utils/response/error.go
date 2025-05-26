package response

import (
	"net/http"
	"runtime"
	"strings"

	"github.com/Caqil/vyrall/pkg/errors"
	"github.com/gin-gonic/gin"
)

// ErrorResponse represents the structure of error responses
type ErrorResponse struct {
	Status  int         `json:"status"`
	Message string      `json:"message"`
	Error   string      `json:"error,omitempty"`
	Details interface{} `json:"details,omitempty"`
	Code    string      `json:"code,omitempty"`
	Source  string      `json:"source,omitempty"`
}

// Error sends a standardized error response
func Error(c *gin.Context, status int, message string, err error) {
	// Get caller information for better debugging
	_, file, line, _ := runtime.Caller(1)
	// Extract just the file name without the full path
	parts := strings.Split(file, "/")
	source := parts[len(parts)-1] + ":" + string(line)

	// Create the response
	resp := ErrorResponse{
		Status:  status,
		Message: message,
		Source:  source,
	}

	// If error is provided, add error details
	if err != nil {
		resp.Error = err.Error()

		// Check if it's a custom application error
		if appErr, ok := err.(*errors.AppError); ok {
			resp.Code = appErr.Code
			resp.Details = appErr.Details
		}

		// Log the error
		c.Error(err)
	}

	// Send the response
	c.JSON(status, resp)
}

// ValidationError sends a validation error response
func ValidationError(c *gin.Context, message string, details interface{}) {
	resp := ErrorResponse{
		Status:  http.StatusBadRequest,
		Message: message,
		Code:    "validation_error",
		Details: details,
	}

	// Get caller information for better debugging
	_, file, line, _ := runtime.Caller(1)
	// Extract just the file name without the full path
	parts := strings.Split(file, "/")
	resp.Source = parts[len(parts)-1] + ":" + string(line)

	c.JSON(http.StatusBadRequest, resp)
}

// NotFoundError sends a not found error response
func NotFoundError(c *gin.Context, message string) {
	resp := ErrorResponse{
		Status:  http.StatusNotFound,
		Message: message,
		Code:    "not_found",
	}

	// Get caller information for better debugging
	_, file, line, _ := runtime.Caller(1)
	// Extract just the file name without the full path
	parts := strings.Split(file, "/")
	resp.Source = parts[len(parts)-1] + ":" + string(line)

	c.JSON(http.StatusNotFound, resp)
}

// UnauthorizedError sends an unauthorized error response
func UnauthorizedError(c *gin.Context, message string) {
	if message == "" {
		message = "Unauthorized access"
	}

	resp := ErrorResponse{
		Status:  http.StatusUnauthorized,
		Message: message,
		Code:    "unauthorized",
	}

	// Get caller information for better debugging
	_, file, line, _ := runtime.Caller(1)
	// Extract just the file name without the full path
	parts := strings.Split(file, "/")
	resp.Source = parts[len(parts)-1] + ":" + string(line)

	c.JSON(http.StatusUnauthorized, resp)
}

// ForbiddenError sends a forbidden error response
func ForbiddenError(c *gin.Context, message string) {
	if message == "" {
		message = "Access forbidden"
	}

	resp := ErrorResponse{
		Status:  http.StatusForbidden,
		Message: message,
		Code:    "forbidden",
	}

	// Get caller information for better debugging
	_, file, line, _ := runtime.Caller(1)
	// Extract just the file name without the full path
	parts := strings.Split(file, "/")
	resp.Source = parts[len(parts)-1] + ":" + string(line)

	c.JSON(http.StatusForbidden, resp)
}

// InternalServerError sends a server error response
func InternalServerError(c *gin.Context, err error) {
	message := "Internal server error"

	resp := ErrorResponse{
		Status:  http.StatusInternalServerError,
		Message: message,
		Code:    "internal_server_error",
	}

	if err != nil {
		resp.Error = err.Error()
		c.Error(err)
	}

	// Get caller information for better debugging
	_, file, line, _ := runtime.Caller(1)
	// Extract just the file name without the full path
	parts := strings.Split(file, "/")
	resp.Source = parts[len(parts)-1] + ":" + string(line)

	c.JSON(http.StatusInternalServerError, resp)
}
