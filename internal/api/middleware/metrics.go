package middleware

import (
	"strconv"
	"time"

	"github.com/Caqil/vyrall/internal/services/metrics"
	"github.com/gin-gonic/gin"
)

// Metrics middleware for collecting request metrics
func Metrics(metricsService *metrics.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()

		// Process the request
		c.Next()

		// Calculate request duration
		duration := time.Since(startTime)

		// Collect metrics
		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}

		// Record request metrics
		metricsService.RecordRequest(
			c.Request.Context(),
			path,
			c.Request.Method,
			c.Writer.Status(),
			duration.Milliseconds(),
		)
	}
}

// RouteMetrics middleware for collecting metrics for specific routes
func RouteMetrics(metricsService *metrics.Service, routeName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()

		// Process the request
		c.Next()

		// Calculate request duration
		duration := time.Since(startTime)

		// Record route metrics
		metricsService.RecordRouteRequest(
			c.Request.Context(),
			routeName,
			c.Request.Method,
			c.Writer.Status(),
			duration.Milliseconds(),
		)
	}
}

// UserMetrics middleware for collecting metrics per user
func UserMetrics(metricsService *metrics.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()

		// Get user ID from context (set by Auth middleware)
		var userID string
		if id, exists := c.Get("userID"); exists {
			userID = id.(string)
		} else {
			userID = "anonymous"
		}

		// Process the request
		c.Next()

		// Calculate request duration
		duration := time.Since(startTime)

		// Record user metrics
		metricsService.RecordUserRequest(
			c.Request.Context(),
			userID,
			c.FullPath(),
			c.Request.Method,
			c.Writer.Status(),
			duration.Milliseconds(),
		)
	}
}

// PrometheusMetrics middleware for exposing metrics in Prometheus format
func PrometheusMetrics() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()

		// Get the request path
		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}

		// Process the request
		c.Next()

		// Calculate request duration
		duration := time.Since(startTime)

		// Record metrics
		statusCode := strconv.Itoa(c.Writer.Status())
		method := c.Request.Method

		// Update Prometheus metrics
		metrics.HttpRequestsTotal.WithLabelValues(method, path, statusCode).Inc()
		metrics.HttpRequestDuration.WithLabelValues(method, path).Observe(duration.Seconds())
		metrics.HttpResponseSize.WithLabelValues(method, path).Observe(float64(c.Writer.Size()))
	}
}
