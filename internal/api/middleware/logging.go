package middleware

import (
	"bytes"
	"io"
	"time"

	"github.com/Caqil/vyrall/internal/utils/logger"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Logging middleware for request logging
func Logging(log *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start time
		startTime := time.Now()

		// Get request path and method
		path := c.Request.URL.Path
		method := c.Request.Method

		// Get request ID (or generate one if not present)
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
			c.Request.Header.Set("X-Request-ID", requestID)
		}

		// Set the request ID in the response header
		c.Writer.Header().Set("X-Request-ID", requestID)

		// Log request details
		log.Info("Request started",
			zap.String("request_id", requestID),
			zap.String("method", method),
			zap.String("path", path),
			zap.String("ip", c.ClientIP()),
			zap.String("user_agent", c.Request.UserAgent()),
		)

		// Process the request
		c.Next()

		// Log response details
		latency := time.Since(startTime)
		statusCode := c.Writer.Status()
		bodySize := c.Writer.Size()

		// Determine log level based on status code
		var logFunc func(msg string, fields ...zap.Field)
		if statusCode >= 500 {
			logFunc = log.Error
		} else if statusCode >= 400 {
			logFunc = log.Warn
		} else {
			logFunc = log.Info
		}

		logFunc("Request completed",
			zap.String("request_id", requestID),
			zap.String("method", method),
			zap.String("path", path),
			zap.Int("status", statusCode),
			zap.Int("size", bodySize),
			zap.Duration("latency", latency),
			zap.Int("errors", len(c.Errors)),
		)
	}
}

// RequestBodyLogging middleware for logging request bodies
func RequestBodyLogging(log *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip if not a POST, PUT, PATCH, or DELETE request
		if c.Request.Method != "POST" && c.Request.Method != "PUT" && c.Request.Method != "PATCH" && c.Request.Method != "DELETE" {
			c.Next()
			return
		}

		// Get request ID
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
			c.Request.Header.Set("X-Request-ID", requestID)
		}

		// Read the request body
		var bodyBytes []byte
		if c.Request.Body != nil {
			bodyBytes, _ = io.ReadAll(c.Request.Body)
			// Restore the body for subsequent middleware and handlers
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		// Log the request body
		if len(bodyBytes) > 0 {
			// Truncate if too large
			const maxBodyLogSize = 10000
			bodyLog := string(bodyBytes)
			if len(bodyLog) > maxBodyLogSize {
				bodyLog = bodyLog[:maxBodyLogSize] + "... (truncated)"
			}

			log.Debug("Request body",
				zap.String("request_id", requestID),
				zap.String("method", c.Request.Method),
				zap.String("path", c.Request.URL.Path),
				zap.String("body", bodyLog),
			)
		}

		c.Next()
	}
}

// ResponseBodyLogging middleware for logging response bodies
func ResponseBodyLogging(log *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Create a response writer proxy to capture the response
		blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = blw

		// Process the request
		c.Next()

		// Get request ID
		requestID := c.GetHeader("X-Request-ID")

		// Log the response body
		if blw.body.Len() > 0 {
			// Truncate if too large
			const maxBodyLogSize = 10000
			bodyLog := blw.body.String()
			if len(bodyLog) > maxBodyLogSize {
				bodyLog = bodyLog[:maxBodyLogSize] + "... (truncated)"
			}

			log.Debug("Response body",
				zap.String("request_id", requestID),
				zap.String("method", c.Request.Method),
				zap.String("path", c.Request.URL.Path),
				zap.Int("status", c.Writer.Status()),
				zap.String("body", bodyLog),
			)
		}
	}
}

// bodyLogWriter is a proxy for gin.ResponseWriter that captures the response body
type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

// Write captures the response body
func (w *bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// generateRequestID generates a unique request ID
func generateRequestID() string {
	// In a real application, this would use a more sophisticated method
	// like UUID generation
	return "req-" + time.Now().Format("20060102-150405.000000")
}
