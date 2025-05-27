package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/Caqil/vyrall/internal/utils/logger"
	"github.com/Caqil/vyrall/internal/utils/response"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Recovery middleware for recovering from panics
func Recovery(log *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Get stack trace
				stack := debug.Stack()

				// Log the error
				log.Error("Panic recovered",
					zap.Any("error", err),
					zap.String("stack", string(stack)),
					zap.String("method", c.Request.Method),
					zap.String("path", c.Request.URL.Path),
					zap.String("client_ip", c.ClientIP()),
				)

				// Respond with error
				errorMessage := "Internal server error"
				if gin.Mode() == gin.DebugMode {
					errorMessage = fmt.Sprintf("Panic: %v", err)
				}

				response.Error(c, http.StatusInternalServerError, errorMessage, nil)

				// Abort the request
				c.Abort()
			}
		}()

		c.Next()
	}
}

// PanicHandler handles specific panic scenarios
func PanicHandler(log *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Get stack trace
				stack := debug.Stack()

				// Log the error
				log.Error("Panic in request handler",
					zap.Any("error", err),
					zap.String("stack", string(stack)),
					zap.String("method", c.Request.Method),
					zap.String("path", c.Request.URL.Path),
					zap.String("client_ip", c.ClientIP()),
				)

				// Check for specific types of panics
				switch e := err.(type) {
				case string:
					if e == "database connection lost" {
						response.Error(c, http.StatusServiceUnavailable, "Database connection lost", nil)
					} else {
						response.Error(c, http.StatusInternalServerError, "Internal server error", nil)
					}
				case error:
					if e.Error() == "database connection lost" {
						response.Error(c, http.StatusServiceUnavailable, "Database connection lost", nil)
					} else {
						response.Error(c, http.StatusInternalServerError, "Internal server error", nil)
					}
				default:
					response.Error(c, http.StatusInternalServerError, "Internal server error", nil)
				}

				// Abort the request
				c.Abort()
			}
		}()

		c.Next()
	}
}
