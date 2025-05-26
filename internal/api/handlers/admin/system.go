package admin

import (
	"net/http"

	"github.com/Caqil/vyrall/internal/utils/response"
	"github.com/gin-gonic/gin"
)

// SystemService interface for system operations
type SystemService interface {
	GetSystemInfo() (map[string]interface{}, error)
	GetServiceStatus() (map[string]interface{}, error)
	RestartService(service string) error
	GetSystemLogs(service, level string, limit, offset int) ([]map[string]interface{}, int, error)
	GetSystemMetrics(metric string, duration string) ([]map[string]interface{}, error)
	GetDatabaseStats() (map[string]interface{}, error)
	RunHealthCheck() (map[string]interface{}, error)
	GetCacheStats() (map[string]interface{}, error)
	ClearCache(cache string) error
	GetJobQueueStats() (map[string]interface{}, error)
	ManageJobQueue(action string, jobID string) error
}

// GetSystemInfo returns information about the system
func GetSystemInfo(c *gin.Context) {
	systemService := c.MustGet("systemService").(SystemService)

	info, err := systemService.GetSystemInfo()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve system information", err)
		return
	}

	response.Success(c, http.StatusOK, "System information retrieved successfully", info)
}

// GetServiceStatus returns the status of system services
func GetServiceStatus(c *gin.Context) {
	systemService := c.MustGet("systemService").(SystemService)

	status, err := systemService.GetServiceStatus()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve service status", err)
		return
	}

	response.Success(c, http.StatusOK, "Service status retrieved successfully", status)
}

// RestartService restarts a specific system service
func RestartService(c *gin.Context) {
	service := c.Param("service")
	systemService := c.MustGet("systemService").(SystemService)

	if err := systemService.RestartService(service); err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to restart service", err)
		return
	}

	response.Success(c, http.StatusOK, "Service restarted successfully", nil)
}

// GetSystemLogs returns system logs
func GetSystemLogs(c *gin.Context) {
	service := c.DefaultQuery("service", "all")
	level := c.DefaultQuery("level", "all")
	limit, offset := getPaginationParams(c)

	systemService := c.MustGet("systemService").(SystemService)

	logs, total, err := systemService.GetSystemLogs(service, level, limit, offset)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve system logs", err)
		return
	}

	response.SuccessWithPagination(c, http.StatusOK, "System logs retrieved successfully", logs, limit, offset, total)
}

// GetSystemMetrics returns system performance metrics
func GetSystemMetrics(c *gin.Context) {
	metric := c.DefaultQuery("metric", "cpu")
	duration := c.DefaultQuery("duration", "1h")

	systemService := c.MustGet("systemService").(SystemService)

	// Validate metric
	validMetrics := map[string]bool{
		"cpu":        true,
		"memory":     true,
		"disk":       true,
		"network":    true,
		"requests":   true,
		"errors":     true,
		"latency":    true,
		"throughput": true,
	}

	if !validMetrics[metric] {
		response.Error(c, http.StatusBadRequest, "Invalid metric", nil)
		return
	}

	// Validate duration
	validDurations := map[string]bool{
		"1h":  true,
		"6h":  true,
		"12h": true,
		"24h": true,
		"7d":  true,
		"30d": true,
	}

	if !validDurations[duration] {
		response.Error(c, http.StatusBadRequest, "Invalid duration", nil)
		return
	}

	metrics, err := systemService.GetSystemMetrics(metric, duration)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve system metrics", err)
		return
	}

	response.Success(c, http.StatusOK, "System metrics retrieved successfully", metrics)
}

// GetDatabaseStats returns database statistics
func GetDatabaseStats(c *gin.Context) {
	systemService := c.MustGet("systemService").(SystemService)

	stats, err := systemService.GetDatabaseStats()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve database statistics", err)
		return
	}

	response.Success(c, http.StatusOK, "Database statistics retrieved successfully", stats)
}

// RunHealthCheck performs a system health check
func RunHealthCheck(c *gin.Context) {
	systemService := c.MustGet("systemService").(SystemService)

	health, err := systemService.RunHealthCheck()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to run health check", err)
		return
	}

	response.Success(c, http.StatusOK, "Health check completed successfully", health)
}

// GetCacheStats returns cache statistics
func GetCacheStats(c *gin.Context) {
	systemService := c.MustGet("systemService").(SystemService)

	stats, err := systemService.GetCacheStats()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve cache statistics", err)
		return
	}

	response.Success(c, http.StatusOK, "Cache statistics retrieved successfully", stats)
}

// ClearCache clears a specific cache
func ClearCache(c *gin.Context) {
	cache := c.Param("cache")
	systemService := c.MustGet("systemService").(SystemService)

	if err := systemService.ClearCache(cache); err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to clear cache", err)
		return
	}

	response.Success(c, http.StatusOK, "Cache cleared successfully", nil)
}

// GetJobQueueStats returns job queue statistics
func GetJobQueueStats(c *gin.Context) {
	systemService := c.MustGet("systemService").(SystemService)

	stats, err := systemService.GetJobQueueStats()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve job queue statistics", err)
		return
	}

	response.Success(c, http.StatusOK, "Job queue statistics retrieved successfully", stats)
}

// ManageJobQueue manages jobs in the queue
func ManageJobQueue(c *gin.Context) {
	var req struct {
		Action string `json:"action" binding:"required"`
		JobID  string `json:"job_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	systemService := c.MustGet("systemService").(SystemService)

	// Validate action
	validActions := map[string]bool{
		"pause":      true,
		"resume":     true,
		"cancel":     true,
		"retry":      true,
		"clear_all":  true,
		"pause_all":  true,
		"resume_all": true,
	}

	if !validActions[req.Action] {
		response.Error(c, http.StatusBadRequest, "Invalid action", nil)
		return
	}

	// Check if job ID is required for the action
	if (req.Action == "cancel" || req.Action == "retry") && req.JobID == "" {
		response.Error(c, http.StatusBadRequest, "Job ID is required for this action", nil)
		return
	}

	if err := systemService.ManageJobQueue(req.Action, req.JobID); err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to manage job queue", err)
		return
	}

	response.Success(c, http.StatusOK, "Job queue managed successfully", nil)
}
