package middleware

import (
	"errors"
	"net/http"

	"github.com/Caqil/vyrall/internal/utils/response"
	"github.com/gin-gonic/gin"
)

// ErrorHandler middleware for handling errors globally
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Process the request
		c.Next()

		// Check if there are any errors
		if len(c.Errors) > 0 {
			// Get the last error
			err := c.Errors.Last()

			// Handle different types of errors
			switch e := err.Err.(type) {
			case *response.AppError:
				// Application error
				response.Error(c, e.StatusCode, e.Message, e.Err)
			default:
				// Default error handler
				handleDefaultError(c, err.Err)
			}
		}
	}
}

// handleDefaultError processes errors that don't have a specific handler
func handleDefaultError(c *gin.Context, err error) {
	// Check for common error types
	switch {
	case errors.Is(err, http.ErrBodyNotAllowed):
		response.Error(c, http.StatusBadRequest, "Request body not allowed", err)
	case errors.Is(err, http.ErrHandlerTimeout):
		response.Error(c, http.StatusGatewayTimeout, "Request timed out", err)
	case errors.Is(err, http.ErrMissingFile):
		response.Error(c, http.StatusBadRequest, "Missing file", err)
	case errors.Is(err, http.ErrNotSupported):
		response.Error(c, http.StatusUnsupportedMediaType, "Not supported", err)
	case errors.Is(err, http.ErrUnexpectedTrailer):
		response.Error(c, http.StatusBadRequest, "Unexpected trailer", err)
	case errors.Is(err, http.ErrBodyReadAfterClose):
		response.Error(c, http.StatusBadRequest, "Body read after close", err)
	case errors.Is(err, http.ErrHijacked):
		response.Error(c, http.StatusInternalServerError, "Connection hijacked", err)
	case errors.Is(err, http.ErrContentLength):
		response.Error(c, http.StatusBadRequest, "Content length error", err)
	case errors.Is(err, http.ErrWriteAfterFlush):
		response.Error(c, http.StatusInternalServerError, "Write after flush", err)
	default:
		// Default to internal server error
		response.Error(c, http.StatusInternalServerError, "Internal server error", err)
	}
}
